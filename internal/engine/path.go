package engine

import "time"

const AIActionInterval = 55 * time.Millisecond

type AIAction uint8

const (
	AILeft AIAction = iota
	AIRight
	AIRotateCW
	AIRotateCCW
	AIDown
	AIHardDrop
)

type pathNode struct {
	piece  Piece
	parent int
	action AIAction
	depth  int
}

// SetAIGravityRemaining supplies the time from this AI event to the next
// scheduled gravity event. Callers without a clock use one full gravity period.
func (g *Game) SetAIGravityRemaining(d time.Duration) {
	if d < 0 {
		d = 0
	}
	g.aiGravityRemaining = d
	g.aiGravityKnown = true
}

func (g *Game) remainingGravity() time.Duration {
	if g.aiGravityKnown {
		return g.aiGravityRemaining
	}
	return g.TickInterval()
}

func pathTo(nodes []pathNode, index int, last AIAction) ([]AIAction, []Piece) {
	n := nodes[index].depth
	actions, expected := make([]AIAction, n+1), make([]Piece, n+1)
	actions[n], expected[n] = last, nodes[index].piece
	for n > 0 {
		node := nodes[index]
		n--
		actions[n], expected[n] = node.action, nodes[node.parent].piece
		index = node.parent
	}
	return actions, expected
}

// reachablePlacements searches from the actual position, not hidden spawn rows.
// Every edge is one UI action; gravity is simulated before the next AI event.
// First visits retain shortest paths. The bounded search may conservatively
// omit a longer path to the same position with a different gravity phase.
func reachablePlacements(g *Game, start *Piece, cleanup bool) []candidatePlacement {
	b := g.Board
	if !b.IsValidPosition(start) {
		return nil
	}
	beforeBits := b.ToBitBoard()
	pIdx := pieceTypeIndex(start.Type)
	nodes := make([]pathNode, 1, 1024)
	nodes[0] = pathNode{piece: *start, parent: -1}
	visited := make(map[[3]int]bool, 1024)
	visited[[3]int{start.X, start.Y, start.Rotation}] = true
	landings := make(map[[3]int]bool, 64)
	candidates := make([]candidatePlacement, 0, 40)
	add := func(p Piece, index int, last AIAction) {
		if p.Y < 0 {
			return
		}
		key := [3]int{p.X, p.Y, p.Rotation}
		if landings[key] {
			return
		}
		landings[key] = true
		actions, expected := pathTo(nodes, index, last)
		after := b.Clone()
		after.LockPiece(&p)
		cleared := after.ClearLines()
		afterBits, _ := beforeBits.LockAndClear(pIdx, p.Rotation, p.X, p.Y)
		candidates = append(candidates, candidatePlacement{
			rotation: p.Rotation, x: p.X, y: p.Y, board: after,
			bits:    afterBits,
			hasBits: true,
			score:   evaluateBitBoardWithSupport(beforeBits, afterBits, p.Type, p.Rotation, p.X, p.Y, cleared, cleanup, !g.ReserveWell),
			actions: actions, expected: expected,
		})
	}
	period, firstGravity := g.TickInterval(), g.remainingGravity()
	allowDown := beforeBits.CountHoles() > 0
	for index := 0; index < len(nodes); index++ {
		node := nodes[index]
		p := node.piece
		p.Y = b.GetGhostY(&p)
		add(p, index, AIHardDrop)
		if node.depth >= 32 {
			continue
		}
		for _, action := range [...]AIAction{AILeft, AIRight, AIRotateCW, AIRotateCCW, AIDown} {
			// Descending before a hard drop is useful for tucking beneath
			// overhangs. A board without holes has no such overhang to enter.
			if action == AIDown && !allowDown {
				continue
			}
			if start.Type == PieceO && (action == AIRotateCW || action == AIRotateCCW) {
				continue
			}
			p = node.piece
			valid := true
			switch action {
			case AILeft:
				p.X--
				valid = b.IsValidPosition(&p)
			case AIRight:
				p.X++
				valid = b.IsValidPosition(&p)
			case AIRotateCW:
				valid = b.TryRotate(&p, 1)
			case AIRotateCCW:
				valid = b.TryRotate(&p, -1)
			case AIDown:
				p.Y++
				valid = b.IsValidPosition(&p)
			}
			if !valid || p.Y < -4 {
				continue
			}
			from, until := time.Duration(node.depth)*AIActionInterval, time.Duration(node.depth+1)*AIActionInterval
			gravity := firstGravity
			if gravity < from {
				gravity += ((from - gravity + period - 1) / period) * period
			}
			locked := false
			for ; gravity < until; gravity += period {
				down := p
				down.Y++
				if !b.IsValidPosition(&down) {
					locked = true
					break
				}
				p = down
			}
			key := [3]int{p.X, p.Y, p.Rotation}
			if locked {
				add(p, index, action)
				continue
			}
			if !visited[key] && len(nodes) < 2048 {
				visited[key] = true
				nodes = append(nodes, pathNode{piece: p, parent: index, action: action, depth: node.depth + 1})
			}
		}
	}
	return candidates
}
