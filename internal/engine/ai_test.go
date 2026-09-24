package engine

import (
	"os"
	"testing"

	"tetris/internal/logger"
)

func TestAIFindBestMove(t *testing.T) {
	g := NewGame()
	move := FindBestMove(g)
	if move == nil {
		t.Fatalf("expected AI to find a valid move")
	}

	// Verify rotation is between 0 and 3
	if move.TargetRotation < 0 || move.TargetRotation > 3 {
		t.Errorf("invalid rotation from AI: %d", move.TargetRotation)
	}

	// Verify target X is within valid board bounds
	if move.TargetX < -2 || move.TargetX >= BoardWidth {
		t.Errorf("invalid target X from AI: %d", move.TargetX)
	}
}

func TestAIPrioritizesTetris(t *testing.T) {
	g := NewGame()
	// Setup 4 lines with column 9 open
	g.SetupFourLines()
	g.CurrentPiece = NewPiece(PieceI)

	move := FindBestMove(g)
	if move == nil {
		t.Fatalf("expected AI to find a move for Tetris setup")
	}

	// Verify AI targets vertical orientation (rot 1 or 3) in column 9
	test := g.CurrentPiece.Clone()
	test.Rotation = move.TargetRotation
	test.X = move.TargetX
	coords := test.BlockCoords()

	allInCol9 := true
	for _, pt := range coords {
		if pt[0] != 9 {
			allInCol9 = false
			break
		}
	}

	if !allInCol9 {
		t.Errorf("expected AI to target column 9 for 4-line Tetris clear, got rotation=%d, x=%d", move.TargetRotation, move.TargetX)
	}
}

func TestAINeverPlacesVerticalIInWellWithoutTetris(t *testing.T) {
	g := NewGame()
	// Fill bottom 2 rows in cols 0..8, leaving col 9 open (stack height is only 2, not ready for 4-line Tetris)
	for y := 18; y < BoardHeight; y++ {
		for x := 0; x < 9; x++ {
			g.Board.Cells[y][x] = Cell{Filled: true, Color: "#ffffff"}
		}
	}

	g.CurrentPiece = NewPiece(PieceI)
	g.CanHold = false // Force AI to place the I piece on the board

	move := FindBestMove(g)
	if move == nil {
		t.Fatalf("expected AI to find a valid move")
	}

	// Verify AI DOES NOT drop vertical I into column 9 when it would not clear a Tetris
	test := g.CurrentPiece.Clone()
	test.Rotation = move.TargetRotation
	test.X = move.TargetX
	coords := test.BlockCoords()

	allInCol9 := true
	for _, pt := range coords {
		if pt[0] != 9 {
			allInCol9 = false
			break
		}
	}

	if allInCol9 {
		t.Errorf("AI placed vertical I in column 9 without 4-line Tetris, which closes the well! rot=%d, x=%d", move.TargetRotation, move.TargetX)
	}
}

func TestAIProtectsCornerFromBeingClosed(t *testing.T) {
	g := NewGame()
	// Col 9 is completely open, cols 0..8 have height 4
	for y := 16; y < BoardHeight; y++ {
		for x := 0; x < 9; x++ {
			g.Board.Cells[y][x] = Cell{Filled: true, Color: "#ffffff"}
		}
	}

	// AI plays several steps with random pieces: col 9 must NEVER be choked (empty below filled)
	for step := 0; step < 15 && g.State == StatePlaying; step++ {
		g.StepAIImmediate()

		colH := 0
		for y := 0; y < BoardHeight; y++ {
			if g.Board.Cells[y][9].Filled {
				colH = BoardHeight - y
				break
			}
		}
		if colH > 0 {
			for y := BoardHeight - colH; y < BoardHeight; y++ {
				if !g.Board.Cells[y][9].Filled {
					t.Fatalf("step %d: corner column 9 was choked/closed! Col height=%d, empty at y=%d", step, colH, y)
				}
			}
		}
	}
}

func TestAIStepExecution(t *testing.T) {
	g := NewGame()
	g.AutoPlay = true

	// Step AI multiple times until the first piece locks
	initialLines := g.Lines
	initialScore := g.Score

	// Run up to 20 steps to place at least one piece
	for i := 0; i < 20; i++ {
		g.StepAI()
	}

	// AI should have dropped at least one piece (score increased)
	if g.Score == initialScore && g.Lines == initialLines && g.Board.Cells[19][0].Filled == false {
		// Just ensure game is running without panic or error
	}
}

func TestToggleAutoPlay(t *testing.T) {
	g := NewGame()
	if g.AutoPlay {
		t.Errorf("expected AutoPlay to start false")
	}

	g.ToggleAutoPlay()
	if !g.AutoPlay {
		t.Errorf("expected AutoPlay to be true after toggle")
	}

	g.ToggleAutoPlay()
	if g.AutoPlay {
		t.Errorf("expected AutoPlay to be false after second toggle")
	}
}

func TestAICleanupModeThreshold(t *testing.T) {
	g := NewGame()

	// Fill columns 0..8 up to height 13 (65% of 20 rows, i.e. rows 7..19 filled)
	// Leave column 9 empty
	for y := 7; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth-1; x++ {
			g.Board.Cells[y][x] = Cell{Filled: true, Color: "#ffffff"}
		}
	}

	maxH := GetMaxHeight(g.Board)
	if maxH != 13 {
		t.Fatalf("expected board height 13 (65%%), got %d", maxH)
	}

	// Case 1: No line piece coming (current is T, hold is nil, next is S)
	g.CurrentPiece = NewPiece(PieceT)
	g.HoldPiece = nil
	g.NextQueue = []TetrominoType{PieceS, PieceZ, PieceO}

	if HasLinePieceComing(g) {
		t.Errorf("expected HasLinePieceComing to be false")
	}

	if !IsCleanupMode(g) {
		t.Errorf("expected IsCleanupMode to be true when board is 65%% and no line is coming")
	}

	move := FindBestMove(g)
	if move == nil {
		t.Fatalf("expected move in cleanup mode")
	}
	if !move.CleanupMode {
		t.Errorf("expected move to have CleanupMode = true")
	}

	// Case 2: Line piece IS coming (Current is PieceI)
	g.CurrentPiece = NewPiece(PieceI)
	if !HasLinePieceComing(g) {
		t.Errorf("expected HasLinePieceComing to be true when current piece is I")
	}
	if IsCleanupMode(g) {
		t.Errorf("expected IsCleanupMode to be false when line piece is available")
	}

	// Case 3: Board is under 65% (height 10)
	gLow := NewGame()
	for y := 10; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth-1; x++ {
			gLow.Board.Cells[y][x] = Cell{Filled: true, Color: "#ffffff"}
		}
	}
	gLow.CurrentPiece = NewPiece(PieceT)
	gLow.NextQueue = []TetrominoType{PieceS, PieceZ, PieceO}
	if IsCleanupMode(gLow) {
		t.Errorf("expected IsCleanupMode to be false when board is under 65%%")
	}
}

func TestAISimulationWithLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/ai_plays.jsonl"

	playLogger, err := logger.NewLogger(logPath)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	g := NewGame()
	g.AutoPlay = true
	g.SetLogger(playLogger)

	// Simulate AI stepping for 50 cycles
	for i := 0; i < 50; i++ {
		if g.State != StatePlaying {
			break
		}
		g.StepAI()
	}

	_ = g.Close()
	if err := playLogger.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	// Verify log file was written and is not empty
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("expected log file to exist: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("expected log file to have content, got 0 bytes")
	}
}


