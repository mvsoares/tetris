package engine

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"tetris/internal/logger"
)

func readMeasurements(t *testing.T, path string) ([]logger.MoveLog, []logger.SessionEndLog) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var moves []logger.MoveLog
	var ends []logger.SessionEndLog
	s := bufio.NewScanner(f)
	for s.Scan() {
		var h struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(s.Bytes(), &h); err != nil {
			t.Fatal(err)
		}
		if h.Type == "SESSION_END" {
			var end logger.SessionEndLog
			if err := json.Unmarshal(s.Bytes(), &end); err != nil {
				t.Fatal(err)
			}
			ends = append(ends, end)
		} else {
			var move logger.MoveLog
			if err := json.Unmarshal(s.Bytes(), &move); err != nil {
				t.Fatal(err)
			}
			moves = append(moves, move)
		}
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	return moves, ends
}

func TestLoggedRewardsIncludeAllDropPoints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "rewards.jsonl")
	l, err := logger.NewLoggerWithOptions(path, logger.Options{BufferSize: 1024})
	if err != nil {
		t.Fatal(err)
	}
	g := NewGameWithSeed(42)
	g.SetLogger(l)
	g.SoftDrop()
	g.HardDrop()
	firstScore := g.Score
	g.HardDrop()
	if err := g.CloseLogger(); err != nil {
		t.Fatal(err)
	}
	moves, ends := readMeasurements(t, path)
	if len(moves) != 2 || len(ends) != 1 {
		t.Fatalf("moves=%d ends=%d", len(moves), len(ends))
	}
	if moves[0].ScoreGained != firstScore || moves[1].ScoreGained != g.Score-firstScore {
		t.Fatalf("rewards do not match score deltas: %+v", moves)
	}
	if len(moves[0].BoardStateBefore) != BoardHeight || moves[0].MaxHeightBefore != 0 {
		t.Fatal("missing pre-move board")
	}
	if ends[0].EndReason != EndClosed {
		t.Fatalf("reason=%s", ends[0].EndReason)
	}
}

func TestHeadlessDropDoesNotRewardHiddenRows(t *testing.T) {
	g := NewGameWithSeed(42)
	g.CanHold = false
	plan := FindBestMove(g)
	manual := NewGameWithSeed(42)
	manual.CurrentPiece.Rotation = plan.TargetRotation
	manual.CurrentPiece.X = plan.TargetX
	manual.HardDrop()
	if !g.StepAIImmediate() {
		t.Fatal("no headless move")
	}
	if g.Score != manual.Score || g.Board.Cells != manual.Board.Cells {
		t.Fatalf("headless score=%d manual=%d", g.Score, manual.Score)
	}
}

func TestHoldTopOutLogsFailedPieceOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hold.jsonl")
	l, err := logger.NewLoggerWithOptions(path, logger.Options{BufferSize: 1024})
	if err != nil {
		t.Fatal(err)
	}
	g := NewGameWithSeed(7)
	g.SetLogger(l)
	g.HoldPiece = NewPiece(PieceO)
	g.Board.Cells[0][4].Filled = true
	if g.Hold() {
		t.Fatal("hold should top out")
	}
	_ = g.CloseWithReason(EndMoveLimit)
	if err := g.CloseLogger(); err != nil {
		t.Fatal(err)
	}
	_, ends := readMeasurements(t, path)
	if len(ends) != 1 || ends[0].EndReason != EndTopOut || ends[0].FailedPiece != "O" || ends[0].Seed != 7 {
		t.Fatalf("incorrect terminal event: %+v", ends)
	}
}

func TestCleanupMeasurementUsesDecisionState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cleanup.jsonl")
	l, err := logger.NewLoggerWithOptions(path, logger.Options{BufferSize: 1024})
	if err != nil {
		t.Fatal(err)
	}
	g := NewGameWithSeed(7)
	g.SetLogger(l)
	g.SetupFourLines()
	g.CurrentPiece.Rotation, g.CurrentPiece.X = 1, 7
	g.CurrentAIMove = &AIMove{CleanupMode: true}
	g.HardDrop()
	if err := g.CloseLogger(); err != nil {
		t.Fatal(err)
	}
	moves, _ := readMeasurements(t, path)
	if !moves[0].IsCleanupMode || moves[0].LinesCleared != 4 || moves[0].MaxHeightAfter != 0 {
		t.Fatalf("bad cleanup event: %+v", moves[0])
	}
}

func TestAIReportsNoPlacement(t *testing.T) {
	g := NewGameWithSeed(7)
	g.CanHold = false
	for y := range g.Board.Cells {
		for x := range g.Board.Cells[y] {
			g.Board.Cells[y][x].Filled = true
		}
	}
	if FindBestMove(g) != nil {
		t.Fatal("full board must have no move")
	}
	if g.StepAIImmediate() || g.EndReason != EndNoLegalMove || g.FailedPiece == "" {
		t.Fatal("missing no-move termination")
	}
}

func TestSeededRandomizerIsReproducible(t *testing.T) {
	a, b := NewRandomizerWithSeed(42), NewRandomizerWithSeed(42)
	for i := 0; i < 100; i++ {
		if a.Next() != b.Next() {
			t.Fatalf("different sequences at draw %d", i)
		}
	}
}

func TestBlockedPlanReplansWithoutBlindDrop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missed.jsonl")
	l, err := logger.NewLoggerWithOptions(path, logger.Options{BufferSize: 1024})
	if err != nil {
		t.Fatal(err)
	}
	g := NewGameWithSeed(7)
	g.SetLogger(l)
	g.AutoPlay = true
	g.CurrentPiece = NewPiece(PieceO)
	g.CurrentPiece.X = 8
	g.CurrentAIMove = &AIMove{TargetX: 9, TargetRotation: 0, Actions: []AIAction{AIRight}, Expected: []Piece{*g.CurrentPiece}}
	for i := 0; i < 30 && g.MoveCount == 0; i++ {
		g.StepAI()
	}
	if err := g.CloseLogger(); err != nil {
		t.Fatal(err)
	}
	moves, ends := readMeasurements(t, path)
	if len(moves) != 1 || !moves[0].HadAIPlan || !moves[0].AIPlanMatched || ends[0].PlanMisses != 0 || ends[0].WatchdogDrops != 0 || ends[0].AIReplans != 1 {
		t.Fatalf("blocked plan was not safely replaced: moves=%+v ends=%+v", moves, ends)
	}
}
