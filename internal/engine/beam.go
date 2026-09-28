package engine

import (
	"math"
	"sort"
)

// ConfigureLookahead enables an experimental, bounded search. Depth counts
// placements including the current piece. Zero preserves the v2 policy.
// Enlarging the preview draws future pieces in order, never guesses them or
// changes the generated piece stream. One extra preview supports empty hold.
func (g *Game) ConfigureLookahead(depth, width int) {
	if depth < 0 {
		depth = 0
	}
	if depth > 10 {
		depth = 10
	}
	if width <= 0 {
		width = 4
	}
	if width > 64 {
		width = 64
	}
	g.LookaheadDepth, g.BeamWidth = depth, width
	if depth > 0 {
		for len(g.NextQueue) < depth {
			g.NextQueue = append(g.NextQueue, g.Randomizer.Next())
		}
	}
	g.CurrentAIMove = nil
}

type beamNode struct {
	board   *Board
	bits    BitBoard
	hasBits bool
	current TetrominoType
	hold    TetrominoType
	next    int
	root    *AIMove
	score   float64
}

type beamKey struct {
	rows          [BoardHeight]uint16
	current, hold TetrominoType
	next          int
}

func (n *beamNode) bitBoard() BitBoard {
	if n.hasBits {
		return n.bits
	}
	if n.board != nil {
		return n.board.ToBitBoard()
	}
	return BitBoard{}
}

func sameRootMove(a, b *AIMove) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.UseHold == b.UseHold && a.TargetX == b.TargetX && a.TargetY == b.TargetY && a.TargetRotation == b.TargetRotation
}

func selectBeam(children []beamNode, width int) []beamNode {
	// Equivalent rotations/routes must not use up the narrow beam. Color has
	// no gameplay effect, so the state identity uses occupancy, hold and queue.
	seen := make(map[beamKey]bool, width)
	result := make([]beamNode, 0, width)
	hasDistinctRoot := false
	for _, child := range children {
		key := beamKey{
			rows:    child.bitBoard(),
			current: child.current,
			hold:    child.hold,
			next:    child.next,
		}
		if seen[key] {
			continue
		}
		if width >= 4 && len(result) == width-1 && !hasDistinctRoot && result[0].root != nil {
			// Scan ahead for the highest-scoring distinct root move to prevent beam collapse
			var alt *beamNode
			var altKey beamKey
			for i := range children {
				c := &children[i]
				if c.root == nil || sameRootMove(c.root, result[0].root) {
					continue
				}
				k := beamKey{rows: c.bitBoard(), current: c.current, hold: c.hold, next: c.next}
				if !seen[k] {
					alt, altKey = c, k
					break
				}
			}
			if alt != nil {
				seen[altKey] = true
				result = append(result, *alt)
				break
			}
		}
		seen[key] = true
		if len(result) > 0 && !sameRootMove(child.root, result[0].root) {
			hasDistinctRoot = true
		}
		result = append(result, child)
		if len(result) == width {
			break
		}
	}
	return result
}

// findBeamMove performs receding-horizon search: execute only the first move,
// then search again on the actual board. Future placements are approximations,
// not full gravity/SRS routes. Beam pruning cannot guarantee a global optimum.
func findBeamMove(g *Game) *AIMove {
	width := g.BeamWidth
	if width <= 0 {
		width = 4
	}
	initial := beamNode{
		board:   g.Board,
		bits:    g.Board.ToBitBoard(),
		hasBits: true,
		current: g.CurrentPiece.Type,
	}
	if g.HoldPiece != nil {
		initial.hold = g.HoldPiece.Type
	}
	beam := []beamNode{initial}
	nodes, reached := 0, 0
	var fallback *AIMove
	for depth := 0; depth < g.LookaheadDepth; depth++ {
		children := make([]beamNode, 0, len(beam)*64)
		discount := math.Pow(0.85, float64(depth))
		for _, parent := range beam {
			if parent.current == "" {
				continue
			}
			parentBits := parent.bitBoard()
			for option := 0; option < 2; option++ {
				useHold := option == 1
				pieceType, holdType, next := parent.current, parent.hold, parent.next
				if useHold {
					if depth == 0 && !g.CanHold {
						continue
					}
					holdType = parent.current
					if parent.hold != "" {
						pieceType = parent.hold
					} else {
						if next >= len(g.NextQueue) {
							continue
						}
						pieceType = g.NextQueue[next]
						next++
					}
				}
				plyLevel := g.Level + depth/3
				cleanup := isCleanupModeBitBoardWithLevel(parentBits, plyLevel, pieceType, holdType, g.NextQueue[next:])
				var placements []candidatePlacement
				if depth == 0 {
					piece := NewPiece(pieceType)
					if !useHold {
						piece = g.CurrentPiece
					}
					placements = reachablePlacements(g, piece, cleanup)
				} else {
					placements = getCandidateBitPlacements(parentBits, pieceType, cleanup, !g.ReserveWell, plyLevel >= HighSpeedAwarenessStartLevel)
				}
				for _, placement := range placements {
					nodes++
					childBits := placement.bits
					if !placement.hasBits && placement.board != nil {
						childBits = placement.board.ToBitBoard()
					}
					child := beamNode{
						bits:    childBits,
						hasBits: true,
						hold:    holdType,
						next:    next,
						root:    parent.root,
						score:   parent.score + discount*placement.score,
					}
					if next < len(g.NextQueue) {
						child.current = g.NextQueue[next]
						child.next++
						// Locking also spawns the next piece: reject immediate top-out,
						// including at the search horizon.
						if !childBits.CanSpawn(child.current) {
							continue
						}
					}
					if depth == 0 {
						child.root = &AIMove{UseHold: useHold, TargetRotation: placement.rotation,
							TargetX: placement.x, TargetY: placement.y, HasTargetY: true,
							Actions: placement.actions, Expected: placement.expected, CleanupMode: cleanup}
					}
					children = append(children, child)
				}
			}
		}
		if len(children) == 0 {
			break
		}
		// Stable ties make runs independent of worker scheduling.
		sort.SliceStable(children, func(i, j int) bool { return children[i].score > children[j].score })
		// Keep distinct states, without retaining discarded boards between plies.
		beam, reached = selectBeam(children, width), depth+1
		chosen := *beam[0].root
		chosen.Score = beam[0].score
		fallback = &chosen
	}
	if fallback == nil {
		// If all searched continuations die, preserve an executable best-effort
		// move instead of ending the actual game merely due to search pruning.
		copyGame := *g
		copyGame.LookaheadDepth = 0
		fallback = FindBestMove(&copyGame)
	}
	if fallback != nil {
		fallback.SearchDepth, fallback.SearchNodes = reached, nodes
	}
	return fallback
}

