package engine

import (
	"math"
	"reflect"
	"testing"
)

func TestBeamOnePlacementMatchesEnumerationIncludingHold(t *testing.T) {
	for _, seed := range []int64{1, 42, 100} {
		g := NewGameWithSeed(seed)
		g.ConfigureLookahead(1, 4)
		best := -math.MaxFloat64
		for _, useHold := range []bool{false, true} {
			r := NewGameWithSeed(seed)
			if useHold && !r.Hold() {
				continue
			}
			for _, c := range reachablePlacements(r, r.CurrentPiece, IsCleanupMode(r)) {
				if c.board.IsValidPosition(NewPiece(r.NextQueue[0])) && c.score > best {
					best = c.score
				}
			}
		}
		m := FindBestMove(g)
		if m == nil || m.Score != best || m.SearchDepth != 1 {
			t.Fatalf("incorrect root ranking: %+v best=%f", m, best)
		}
	}
}

func TestBeamSearchTenPlacementsAndNoMutation(t *testing.T) {
	g := NewGameWithSeed(42)
	g.ConfigureLookahead(10, 8)
	beforeBoard, beforePiece := *g.Board, *g.CurrentPiece
	beforeQueue := append([]TetrominoType(nil), g.NextQueue...)
	m := FindBestMove(g)
	if m == nil || m.SearchDepth != 10 || m.SearchNodes == 0 || !m.HasTargetY || len(m.Actions) == 0 {
		t.Fatalf("not a ten-placement legal search: %+v", m)
	}
	if *g.Board != beforeBoard || *g.CurrentPiece != beforePiece || !reflect.DeepEqual(g.NextQueue, beforeQueue) {
		t.Fatal("search mutated game")
	}
	if !reflect.DeepEqual(m, FindBestMove(g)) {
		t.Fatal("non-deterministic search")
	}
}

func TestLookaheadPreservesPieceStreamAndRestart(t *testing.T) {
	a, b := NewGameWithSeed(123), NewGameWithSeed(123)
	b.ConfigureLookahead(10, 4)
	for i := 0; i < 100; i++ {
		if a.CurrentPiece.Type != b.CurrentPiece.Type {
			t.Fatalf("different piece at %d", i)
		}
		a.Board, b.Board = NewBoard(), NewBoard()
		a.spawnPiece()
		b.spawnPiece()
	}
	b.Restart()
	if b.LookaheadDepth != 10 || b.BeamWidth != 4 || len(b.NextQueue) != 10 || b.policyVersion() != "beam-v2-depth10-width4" {
		t.Fatal("restart lost search configuration")
	}
}

func TestBeamDeduplicatesOccupancyButPreservesHoldAndQueue(t *testing.T) {
	a, b := NewBoard(), NewBoard()
	a.Cells[19][0] = Cell{Filled: true, Color: "a"}
	b.Cells[19][0] = Cell{Filled: true, Color: "b"}
	children := []beamNode{
		{board: a, current: PieceT, hold: PieceI, next: 1, score: 10},
		{board: b, current: PieceT, hold: PieceI, next: 1, score: 9},
		{board: b, current: PieceT, hold: PieceO, next: 1, score: 8},
		{board: b, current: PieceT, hold: PieceI, next: 2, score: 7},
	}
	beam := selectBeam(children, 3)
	if len(beam) != 3 || beam[0].score != 10 || beam[1].hold != PieceO || beam[2].next != 2 {
		t.Fatalf("incorrect deduplication: %+v", beam)
	}
}

func TestBeamWithoutHoldAndShortQueue(t *testing.T) {
	g := NewGameWithSeed(42)
	g.ConfigureLookahead(10, 4)
	g.CanHold = false
	g.NextQueue = g.NextQueue[:2]
	m := FindBestMove(g)
	if m == nil || m.UseHold || m.SearchDepth != 3 {
		t.Fatalf("bad short-queue search: %+v", m)
	}
}

func TestBeamSelectedMoveExecutes(t *testing.T) {
	for _, seed := range []int64{1, 42, 100} {
		g := NewGameWithSeed(seed)
		g.ConfigureLookahead(10, 4)
		g.AutoPlay = true
		for i := 0; i < 10; i++ {
			if !g.StepAIImmediate() {
				t.Fatal("failed beam placement")
			}
		}
		if g.MoveCount != 10 || g.PlanMisses != 0 {
			t.Fatalf("bad execution: %+v", g)
		}
	}
}

func BenchmarkLookaheadDecision(b *testing.B) {
	for _, config := range []struct {
		name         string
		depth, width int
	}{{"v2", 0, 8}, {"depth10-width4", 10, 4}, {"depth10-width8", 10, 8}, {"depth10-width16", 10, 16}} {
		b.Run(config.name, func(b *testing.B) {
			g := NewGameWithSeed(42)
			g.ConfigureLookahead(config.depth, config.width)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				FindBestMove(g)
			}
			b.StopTimer()
			if m := FindBestMove(g); m != nil && config.depth > 0 {
				b.ReportMetric(float64(m.SearchDepth), "plies")
			}
		})
	}
}

func BenchmarkLookaheadStressedDecision(b *testing.B) {
	for _, depth := range []int{0, 10} {
		name := "v2"
		if depth == 10 {
			name = "depth10-width4"
		}
		b.Run(name, func(b *testing.B) {
			g := NewGameWithSeed(42)
			g.ConfigureLookahead(depth, 4)
			g.Level = 13
			for x, h := range [10]int{8, 9, 8, 6, 6, 7, 10, 11, 9, 0} {
				for y := BoardHeight - h; y < BoardHeight; y++ {
					g.Board.Cells[y][x].Filled = true
				}
			}
			g.Board.Cells[18][1] = Cell{}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				FindBestMove(g)
			}
			b.StopTimer()
			if m := FindBestMove(g); m != nil && depth > 0 {
				b.ReportMetric(float64(m.SearchDepth), "plies")
			}
		})
	}
}
