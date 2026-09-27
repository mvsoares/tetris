package engine

import (
	"testing"
	"time"
)

// Replay generated routes with the real game methods, independently of the
// search's transition code, including gravity at AI/gravity timer ties.
func TestReachableRoutesReplayWithGravity(t *testing.T) {
	for _, level := range []int{1, 13} {
		for _, phase := range []time.Duration{0, 40 * time.Millisecond, 80 * time.Millisecond} {
			for _, pieceType := range AllPieces {
				g := NewGameWithSeed(42)
				g.Level, g.CurrentPiece = level, NewPiece(pieceType)
				for x, h := range [10]int{8, 9, 8, 6, 6, 7, 10, 11, 9, 0} {
					for y := BoardHeight - h; y < BoardHeight; y++ {
						g.Board.Cells[y][x] = Cell{Filled: true, Color: "test"}
					}
				}
				g.Board.Cells[18][1] = Cell{}
				g.SetAIGravityRemaining(phase)
				candidates := reachablePlacements(g, g.CurrentPiece, IsCleanupMode(g))
				if len(candidates) == 0 {
					t.Fatalf("no paths: level=%d phase=%v type=%s", level, phase, pieceType)
				}
				for _, c := range candidates {
					r := NewGameWithSeed(42)
					r.Board, r.CurrentPiece, r.Level, r.AutoPlay = g.Board.Clone(), g.CurrentPiece.Clone(), level, true
					r.CurrentAIMove = &AIMove{TargetX: c.x, TargetRotation: c.rotation, TargetY: c.y, HasTargetY: true}
					nextGravity := phase
					for i, action := range c.actions {
						if *r.CurrentPiece != c.expected[i] {
							t.Fatalf("wrong expected state: type=%s level=%d action=%d got=%+v want=%+v", pieceType, level, i, r.CurrentPiece, c.expected[i])
						}
						ok := true
						switch action {
						case AILeft:
							ok = r.MoveLeft()
						case AIRight:
							ok = r.MoveRight()
						case AIRotateCW:
							ok = r.RotateCW()
						case AIRotateCCW:
							ok = r.RotateCCW()
						case AIDown:
							ok = r.SoftDrop()
						case AIHardDrop:
							r.HardDrop()
						}
						if !ok {
							t.Fatal("path contains an illegal action")
						}
						until := time.Duration(i+1) * AIActionInterval
						for r.MoveCount == 0 && nextGravity < until {
							r.Tick()
							nextGravity += r.TickInterval()
						}
						if r.MoveCount > 0 {
							if i != len(c.actions)-1 {
								t.Fatal("gravity locked before route completed")
							}
							break
						}
					}
					if r.MoveCount != 1 || r.PlanMisses != 0 || r.Board.Cells != c.board.Cells {
						t.Fatalf("route did not reach evaluated board: type=%s level=%d phase=%v", pieceType, level, phase)
					}
				}
			}
		}
	}
}

func TestPathUsesSingleCounterclockwiseRotation(t *testing.T) {
	g := NewGameWithSeed(42)
	g.CurrentPiece = NewPiece(PieceI)
	for _, c := range reachablePlacements(g, g.CurrentPiece, false) {
		if c.rotation == 3 && c.x == g.CurrentPiece.X {
			if len(c.actions) != 2 || c.actions[0] != AIRotateCCW || c.actions[1] != AIHardDrop {
				t.Fatalf("not the shortest rotation: %v", c.actions)
			}
			return
		}
	}
	t.Fatal("missing counterclockwise candidate")
}

func TestEmergencyWellAllowsPartialIWithoutWeakeningNormalPolicy(t *testing.T) {
	b := NewBoard()
	for y := 18; y < 20; y++ {
		for x := 0; x < 9; x++ {
			b.Cells[y][x].Filled = true
		}
	}
	p := NewPiece(PieceI)
	p.Rotation, p.X = 1, 7
	p.Y = b.GetGhostY(p)
	after := b.Clone()
	after.LockPiece(p)
	lines := after.ClearLines()
	strict := evaluatePlacementWithSupport(b, after, p, lines, true, false)
	emergency := evaluatePlacementWithSupport(b, after, p, lines, true, true)
	if lines != 2 || emergency <= strict+100000 {
		t.Fatalf("emergency partial clear still blocked: strict=%f emergency=%f lines=%d", strict, emergency, lines)
	}
	if evaluatePlacementWithSupport(b, after, p, lines, false, false) != evaluatePlacementWithSupport(b, after, p, lines, false, true) {
		t.Fatal("normal well reservation changed")
	}
}

func TestORewardUsesSupportBeforePlacement(t *testing.T) {
	flat, step := NewBoard(), NewBoard()
	for _, b := range []*Board{flat, step} {
		b.Cells[19][4].Filled = true
		b.Cells[19][5].Filled = true
	}
	step.Cells[18][4].Filled = true
	p := NewPiece(PieceO)
	after := NewBoard()
	a := evaluatePlacementWithSupport(flat, after, p, 0, false)
	b := evaluatePlacementWithSupport(step, after, p, 0, false)
	if a-b != 3500 {
		t.Fatalf("unequal support got flat-platform reward: %f vs %f", a, b)
	}
}

func TestUnexpectedGravityReplansFromActualPosition(t *testing.T) {
	g := NewGameWithSeed(42)
	g.AutoPlay = true
	g.CurrentPiece = NewPiece(PieceO)
	g.CanHold = false
	old := *g.CurrentPiece
	g.CurrentAIMove = &AIMove{TargetX: 8, TargetRotation: 0, TargetY: 18, HasTargetY: true, Actions: []AIAction{AIRight, AIHardDrop}, Expected: []Piece{old, old}}
	g.Tick() // An unpredicted tick arrives before the planned action.
	if !g.StepAI() || g.AIReplans != 1 || g.WatchdogDrops != 0 {
		t.Fatal("unexpected gravity was not safely replanned")
	}
	if !g.Board.IsValidPosition(g.CurrentPiece) {
		t.Fatal("replanning produced an invalid piece")
	}
}

func TestRestartPreservesWellPolicy(t *testing.T) {
	g := NewGameWithSeed(42)
	g.ReserveWell = true
	g.Restart()
	if !g.ReserveWell || g.policyVersion() != PolicyVersion+"-strict-well" {
		t.Fatal("restart changed well policy")
	}
}

func TestChangedBoardReplansBeforeDropping(t *testing.T) {
	g := NewGameWithSeed(42)
	g.AutoPlay, g.CanHold = true, false
	g.CurrentPiece = NewPiece(PieceO)
	g.CurrentAIMove = &AIMove{TargetX: 4, TargetRotation: 0, TargetY: 18, HasTargetY: true, Actions: []AIAction{AIHardDrop}, Expected: []Piece{*g.CurrentPiece}}
	g.Board.Cells[19][4].Filled = true // The previously evaluated landing changed.
	for i := 0; i < 30 && g.MoveCount == 0; i++ {
		g.StepAI()
	}
	if g.MoveCount != 1 || g.AIReplans != 1 || g.PlanMisses != 0 || g.WatchdogDrops != 0 {
		t.Fatalf("stale board plan was dropped: replans=%d misses=%d", g.AIReplans, g.PlanMisses)
	}
}
