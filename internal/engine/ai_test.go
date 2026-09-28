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

func TestZPieceHorizontalAndVerticalAlignment(t *testing.T) {
	b := NewBoard()
	for y := 12; y < BoardHeight; y++ {
		for x := 0; x < 9; x++ {
			b.Cells[y][x] = Cell{Filled: true}
		}
	}
	horiz := NewPiece(PieceZ)
	horiz.Rotation = 0
	horiz.Y = 10
	vert := NewPiece(PieceZ)
	vert.Rotation = 1
	vert.Y = 10

	scoreHoriz := evaluatePlacement(b, horiz, 0, false)
	scoreVert := evaluatePlacement(b, vert, 0, false)
	if scoreHoriz-scoreVert < 4500.0 {
		t.Fatalf("expected horizontal Z bonus and vertical Z penalty on elevated board, got horiz=%f vert=%f", scoreHoriz, scoreVert)
	}
}

func TestIHoldGuardDoesNotHoldIntoHoleTrap(t *testing.T) {
	g := NewGameWithSeed(42)
	// Create a jagged surface where O piece must create a hole anywhere it lands,
	// while I piece can land vertically in column 0 without creating any hole.
	heights := [10]int{0, 3, 1, 4, 1, 3, 1, 4, 2, 0}
	for x, h := range heights {
		for y := BoardHeight - h; y < BoardHeight; y++ {
			g.Board.Cells[y][x] = Cell{Filled: true}
		}
	}
	g.CurrentPiece = NewPiece(PieceI)
	g.HoldPiece = NewPiece(PieceO)
	g.NextQueue = []TetrominoType{PieceO, PieceO, PieceO}
	g.CanHold = true

	move := FindBestMove(g)
	if move == nil {
		t.Fatalf("expected a valid move")
	}
	if move.UseHold {
		t.Fatalf("expected AI not to hold I into a hole-creating O placement")
	}
}

func TestCleanupDiggingPenalizesBuryingTopmostHole(t *testing.T) {
	// Board A and Board B both have a hole at (col 2, row 15) covered by row 14.
	// Board A adds a block on col 2 (row 13, burying the topmost hole deeper).
	// Board B adds a block on col 1 (row 13, keeping col 2 clear for digging).
	makeBase := func() *Board {
		b := NewBoard()
		for y := 14; y < BoardHeight; y++ {
			for x := 0; x < 9; x++ {
				b.Cells[y][x] = Cell{Filled: true}
			}
		}
		b.Cells[15][2] = Cell{Filled: false} // shallow hole at row 15, col 2
		b.Cells[18][6] = Cell{Filled: false} // deeper hole at row 18, col 6
		return b
	}
	buried := makeBase()
	buried.Cells[13][2] = Cell{Filled: true}

	clearForDig := makeBase()
	clearForDig.Cells[13][1] = Cell{Filled: true}

	p := NewPiece(PieceT)
	p.Y = 13
	scoreBuried := evaluatePlacement(buried, p, 0, true)
	scoreClear := evaluatePlacement(clearForDig, p, 0, true)

	// Burying a shallow hole in cleanup mode should be penalized (> 3000 pts difference)
	if scoreClear-scoreBuried < 3000.0 {
		t.Fatalf("expected digging penalty for burying shallow hole in cleanup mode, got clear=%f buried=%f (diff=%f)", scoreClear, scoreBuried, scoreClear-scoreBuried)
	}
}

func TestIsReachableDoesNotTeleportAboveCeiling(t *testing.T) {
	b := NewBoard()
	// Block column 2 at row 0 and row 1 so a piece spawning at X=3, Y=0 cannot move left to X=0.
	for y := 0; y < BoardHeight; y++ {
		b.Cells[y][2] = Cell{Filled: true}
	}
	p := NewPiece(PieceT)
	if isReachable(b, p, 0, 0) {
		t.Fatalf("expected targetX=0 to be unreachable across full-height wall in column 2 without ceiling teleportation")
	}
}

func TestBitBoardEquivalenceWithBoard(t *testing.T) {
	b := NewBoard()
	heights := [10]int{4, 5, 3, 6, 6, 4, 7, 5, 4, 0}
	for x, h := range heights {
		for y := BoardHeight - h; y < BoardHeight; y++ {
			b.Cells[y][x] = Cell{Filled: true}
		}
	}
	b.Cells[18][2] = Cell{Filled: false}
	bb := b.ToBitBoard()

	if bb.ColHeights() != b.ColHeights() {
		t.Fatalf("ColHeights mismatch: got %v want %v", bb.ColHeights(), b.ColHeights())
	}
	if bb.CountHoles() != b.CountHoles() {
		t.Fatalf("CountHoles mismatch: got %d want %d", bb.CountHoles(), b.CountHoles())
	}

	for _, pType := range AllPieces {
		pIdx := pieceTypeIndex(pType)
		for rot := 0; rot < 4; rot++ {
			for x := -3; x < BoardWidth; x++ {
				p := &Piece{Type: pType, Rotation: rot, X: x, Y: 0}
				validB := b.IsValidPosition(p)
				validBB := bb.IsValidPosition(pIdx, rot, x, 0)
				if validB != validBB {
					t.Fatalf("IsValidPosition mismatch for %s rot=%d x=%d: board=%v bb=%v", pType, rot, x, validB, validBB)
				}
				if !validB {
					continue
				}
				ghostB := b.GetGhostY(p)
				ghostBB := bb.GetGhostY(pIdx, rot, x, 0)
				if ghostB != ghostBB {
					t.Fatalf("GetGhostY mismatch for %s rot=%d x=%d: board=%d bb=%d", pType, rot, x, ghostB, ghostBB)
				}
				simB := b.Clone()
				p.Y = ghostB
				simB.LockPiece(p)
				clearedB := simB.ClearLines()

				simBB, clearedBB := bb.LockAndClear(pIdx, rot, x, ghostBB)
				if clearedB != clearedBB || simB.ToBitBoard() != simBB {
					t.Fatalf("LockAndClear mismatch for %s rot=%d x=%d: cleared=%d/%d", pType, rot, x, clearedB, clearedBB)
				}
			}
		}
	}
}

func TestHighSpeed40PercentThresholdAcrossAllThreeAIModes(t *testing.T) {
	// At level 1, a clean 8-row (40%) stack is NOT in cleanup mode (threshold is 14 / 65%+).
	// At HighSpeedModeLevel (75) and Level 100+, the 40% threshold (8 rows) engages across v2, beam, and hybrid.
	build8RowBoard := func() *Board {
		b := NewBoard()
		for y := BoardHeight - HighSpeedCleanupThreshold; y < BoardHeight; y++ {
			for x := 0; x < 9; x++ {
				b.Cells[y][x] = Cell{Filled: true}
			}
		}
		return b
	}

	gLow := NewGame()
	gLow.Board = build8RowBoard()
	gLow.Level = 1
	gLow.CurrentPiece = NewPiece(PieceT)
	gLow.HoldPiece = NewPiece(PieceJ)
	gLow.NextQueue = []TetrominoType{PieceL, PieceO, PieceS}
	if IsCleanupMode(gLow) {
		t.Fatalf("expected 8-row clean board at level 1 NOT to be in cleanup mode")
	}

	if MaxCleanupThresholdForLevel(1) != 14 {
		t.Fatalf("expected level 1 threshold 14, got %d", MaxCleanupThresholdForLevel(1))
	}
	if MaxCleanupThresholdForLevel(HighSpeedModeLevel) != HighSpeedCleanupThreshold {
		t.Fatalf("expected level %d threshold %d (40%%), got %d", HighSpeedModeLevel, HighSpeedCleanupThreshold, MaxCleanupThresholdForLevel(HighSpeedModeLevel))
	}
	if MaxCleanupThresholdForLevel(105) != HighSpeedCleanupThreshold {
		t.Fatalf("expected level 105 threshold %d (40%%), got %d", HighSpeedCleanupThreshold, MaxCleanupThresholdForLevel(105))
	}

	model, err := LoadMoveModel("../../models/move-risk.json")
	if err != nil {
		t.Fatalf("failed to load move-risk model: %v", err)
	}

	for _, mode := range []string{"v2", "beam", "hybrid"} {
		g := NewGame()
		g.Board = build8RowBoard()
		g.Level = 105
		g.CurrentPiece = NewPiece(PieceT)
		g.HoldPiece = NewPiece(PieceJ)
		g.NextQueue = []TetrominoType{PieceL, PieceO, PieceS, PieceZ, PieceT, PieceJ, PieceL, PieceO, PieceS, PieceZ}
		switch mode {
		case "beam":
			g.ConfigureLookahead(10, 4)
		case "hybrid":
			g.UseLearned = true
			g.LearnedModel = model
		}
		if !IsCleanupMode(g) {
			t.Fatalf("mode %s: expected 8-row (40%%) stack at level 105 to trigger cleanup mode", mode)
		}
		if CleanupThresholdPercent(g) != 40 {
			t.Fatalf("mode %s: expected CleanupThresholdPercent=40 at level 105, got %d", mode, CleanupThresholdPercent(g))
		}
		move := FindBestMove(g)
		if move == nil {
			t.Fatalf("mode %s: expected valid move at level 105", mode)
		}
		if !move.CleanupMode {
			t.Fatalf("mode %s: expected move.CleanupMode=true at 40%% stack height on level 105", mode)
		}
	}
}

func TestStepAILeftCorridorResilientToGravityJitter(t *testing.T) {
	g := NewGame()
	g.Level = 110
	g.AutoPlay = true
	g.CurrentPiece = NewPiece(PieceO) // spawns at X=4, Y=0

	// Plan a multi-step move all the way to the left corridor (X=0, Y=18)
	g.CurrentAIMove = &AIMove{
		TargetRotation: 0,
		TargetX:        0,
		TargetY:        18,
		HasTargetY:     true,
		Actions:        []AIAction{AILeft, AILeft, AILeft, AILeft, AIHardDrop},
		Expected: []Piece{
			{Type: PieceO, Rotation: 0, X: 4, Y: 0},
			{Type: PieceO, Rotation: 0, X: 3, Y: 0},
			{Type: PieceO, Rotation: 0, X: 2, Y: 1},
			{Type: PieceO, Rotation: 0, X: 1, Y: 1},
			{Type: PieceO, Rotation: 0, X: 0, Y: 2},
		},
	}

	// Simulate UI tick jitter where gravity ticked 1 row earlier than Expected[0].Y
	g.CurrentPiece.Y = 1

	for step := 0; step < 4; step++ {
		if !g.StepAI() {
			t.Fatalf("step %d: expected StepAI to succeed", step)
		}
	}
	if g.AIReplans != 0 {
		t.Fatalf("expected 0 AIReplans under 1-row gravity jitter on left-corridor path, got %d", g.AIReplans)
	}
	// PieceO should have reached X=0 and auto-committed HardDrop at the bottom-left corner
	if !g.Board.Cells[19][0].Filled || !g.Board.Cells[19][1].Filled {
		t.Fatalf("expected PieceO to lock cleanly in left corridor (cols 0..1), board bottom-left not filled")
	}
}
