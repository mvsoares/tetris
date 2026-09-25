package engine

import (
	"testing"

	"tetris/internal/logger"
)

func TestRandomizer7Bag(t *testing.T) {
	rand := NewRandomizer()
	counts := make(map[TetrominoType]int)
	for i := 0; i < 7; i++ {
		p := rand.Next()
		counts[p]++
	}

	for _, pieceType := range AllPieces {
		if counts[pieceType] != 1 {
			t.Errorf("expected 1 of piece %s in first 7 draws, got %d", pieceType, counts[pieceType])
		}
	}
}

func TestMovementAndBounds(t *testing.T) {
	g := NewGame()
	g.CurrentPiece = NewPiece(PieceO) // 2x2 piece
	g.CurrentPiece.X = 0
	g.CurrentPiece.Y = 5

	// Moving left when at X=0 should fail
	if g.MoveLeft() {
		t.Errorf("expected MoveLeft to fail at left wall")
	}

	// Move right across the board
	moved := 0
	for g.MoveRight() {
		moved++
	}
	// O piece is 2 wide, board is 10 wide, max X is 8
	if g.CurrentPiece.X != 8 {
		t.Errorf("expected piece to stop at X=8, stopped at X=%d", g.CurrentPiece.X)
	}
}

func TestLineClear(t *testing.T) {
	b := NewBoard()
	// Fill row 19 completely except for 1 block
	for x := 0; x < BoardWidth-1; x++ {
		b.Cells[19][x] = Cell{Filled: true, Color: "#ffffff"}
	}
	if b.ClearLines() != 0 {
		t.Errorf("incomplete line should not clear")
	}

	// Fill the remaining block
	b.Cells[19][BoardWidth-1] = Cell{Filled: true, Color: "#ffffff"}
	cleared := b.ClearLines()
	if cleared != 1 {
		t.Errorf("expected 1 line cleared, got %d", cleared)
	}

	// Row 19 should now be empty
	for x := 0; x < BoardWidth; x++ {
		if b.Cells[19][x].Filled {
			t.Errorf("expected row 19 to be empty after clear, got filled cell at x=%d", x)
		}
	}
}

func TestHardDropAndScore(t *testing.T) {
	g := NewGame()
	g.CurrentPiece = NewPiece(PieceO)
	g.CurrentPiece.X = 4
	g.CurrentPiece.Y = 0

	initialScore := g.Score
	distance := g.HardDrop()

	// O piece height is 2, board height is 20 -> drops to Y = 18 -> distance = 18
	if distance != 18 {
		t.Errorf("expected drop distance 18, got %d", distance)
	}
	expectedScore := initialScore + distance*2
	if g.Score != expectedScore {
		t.Errorf("expected score %d, got %d", expectedScore, g.Score)
	}
}

func TestHoldMechanic(t *testing.T) {
	g := NewGame()
	firstType := g.CurrentPiece.Type

	// First hold should succeed
	ok := g.Hold()
	if !ok {
		t.Fatalf("first hold should succeed")
	}
	if g.HoldPiece == nil || g.HoldPiece.Type != firstType {
		t.Fatalf("held piece should match first type %s", firstType)
	}

	// Second hold immediately should fail before piece is locked
	ok2 := g.Hold()
	if ok2 {
		t.Fatalf("second consecutive hold without lock should fail")
	}

	// Lock the piece
	g.HardDrop()

	// Now hold should be available again
	if !g.CanHold {
		t.Fatalf("expected CanHold to be true after piece lock")
	}
}

func TestRotationWallKick(t *testing.T) {
	b := NewBoard()
	p := NewPiece(PieceI)
	p.X = 8 // Near right edge
	p.Y = 5
	p.Rotation = 0 // horizontal 4x4: row 1 has blocks from c=0 to c=3 -> X+3 = 11 (out of bounds)

	// Trying to rotate I piece into right wall: wall kick should kick it left
	success := b.TryRotate(p, 1)
	if !success {
		t.Errorf("expected wall kick to successfully rotate piece")
	}
	if !b.IsValidPosition(p) {
		t.Errorf("rotated piece ended up in invalid position")
	}
}

func TestQueueLinePieces(t *testing.T) {
	g := NewGame()
	g.QueueLinePieces(4)

	if g.CurrentPiece.Type != PieceI {
		t.Errorf("expected current piece to be I, got %s", g.CurrentPiece.Type)
	}
	if len(g.NextQueue) != 4 {
		t.Fatalf("expected next queue of length 4, got %d", len(g.NextQueue))
	}
	for i, p := range g.NextQueue {
		if p != PieceI {
			t.Errorf("expected next piece %d to be I, got %s", i, p)
		}
	}
}

func TestSetupFourLines(t *testing.T) {
	g := NewGame()
	g.SetupFourLines()

	if g.CurrentPiece.Type != PieceI {
		t.Errorf("expected current piece to be I, got %s", g.CurrentPiece.Type)
	}

	// Verify rows 16..19 have cols 0..8 filled and col 9 empty
	for y := 16; y <= 19; y++ {
		for x := 0; x < BoardWidth; x++ {
			filled := g.Board.Cells[y][x].Filled
			if x < 9 && !filled {
				t.Errorf("expected cell at (%d, %d) to be filled", x, y)
			} else if x == 9 && filled {
				t.Errorf("expected cell at (9, %d) to be empty for Tetris", y)
			}
		}
	}

	// Rotate I piece vertically and position at x=9
	g.CurrentPiece.Rotation = 1 // vertical
	g.CurrentPiece.X = 7        // with rotation 1, col 2 is filled -> 7+2 = 9
	g.HardDrop()

	// Should have cleared 4 lines (TETRIS!)
	if g.Lines != 4 {
		t.Errorf("expected 4 lines cleared, got %d", g.Lines)
	}
	if g.LastAction != "💥 TETRIS! 💥" {
		t.Errorf("expected action to be TETRIS, got %s", g.LastAction)
	}
}

func TestGameLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := tmpDir + "/game_plays.jsonl"

	playLogger, err := logger.NewLogger(logPath)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	g := NewGame()
	g.SetLogger(playLogger)

	// Play a hard drop move
	g.HardDrop()

	if g.MoveCount != 1 {
		t.Errorf("expected MoveCount to be 1, got %d", g.MoveCount)
	}

	// Close game logger
	if err := g.CloseLogger(); err != nil {
		t.Fatalf("failed to close game logger: %v", err)
	}
}

func TestStandardSRSKicks(t *testing.T) {
	b := NewBoard()

	// 1. Test T piece rotation against right wall: 0 -> 1 (CW)
	// Rotation 0: T points up. Placed at right wall X=8 (3x3 piece spanning X 8, 9, 10).
	pT := NewPiece(PieceT)
	pT.X = 8
	pT.Y = 5
	pT.Rotation = 0

	// Rotating CW (0 -> 1) should kick left from X=8 to X=7 because rotation 1 has blocks at col 1 & 2 (8+2=10 out of bounds)
	if !b.TryRotate(pT, 1) {
		t.Fatalf("expected T piece to rotate with SRS kick against right wall")
	}
	if pT.Rotation != 1 {
		t.Errorf("expected rotation 1, got %d", pT.Rotation)
	}
	if !b.IsValidPosition(pT) {
		t.Errorf("expected valid position after SRS kick, got invalid")
	}

	// 2. Test counter-clockwise rotation with kick: 0 -> 3 (CCW)
	pT2 := NewPiece(PieceT)
	pT2.X = 8
	pT2.Y = 5
	pT2.Rotation = 0
	if !b.TryRotate(pT2, -1) {
		t.Fatalf("expected T piece to rotate CCW with SRS kick")
	}
	if pT2.Rotation != 3 {
		t.Errorf("expected rotation 3, got %d", pT2.Rotation)
	}
	if !b.IsValidPosition(pT2) {
		t.Errorf("expected valid position after CCW SRS kick")
	}

	// 3. Test I piece rotation against left wall: 1 -> 0 (CCW)
	pI := NewPiece(PieceI)
	pI.Rotation = 1 // vertical, blocks at col 2
	pI.X = -2       // col 2 lands at X=0
	pI.Y = 5
	if !b.IsValidPosition(pI) {
		t.Fatalf("expected initial vertical I at X=-2 to be valid")
	}
	// Rotating to 0 (horizontal) puts blocks at cols 0..3: without kick, col 0 is at X=-2 (out of bounds).
	// SRS kick should kick it right!
	if !b.TryRotate(pI, -1) {
		t.Fatalf("expected I piece to rotate CCW with SRS kick against left wall")
	}
	if pI.Rotation != 0 {
		t.Errorf("expected rotation 0, got %d", pI.Rotation)
	}
	if !b.IsValidPosition(pI) {
		t.Errorf("expected valid position after I piece SRS kick")
	}
}

func TestBoardHeightAndHolesMethods(t *testing.T) {
	b := NewBoard()

	// Fill col 4 up to height 6 (rows 14..19)
	for y := 14; y < BoardHeight; y++ {
		b.Cells[y][4] = Cell{Filled: true, Color: "#ffffff"}
	}
	// Fill col 1 up to height 3 (rows 17..19)
	for y := 17; y < BoardHeight; y++ {
		b.Cells[y][1] = Cell{Filled: true, Color: "#ffffff"}
	}

	if b.MaxHeight() != 6 {
		t.Errorf("expected MaxHeight 6, got %d", b.MaxHeight())
	}
	if b.CenterHeight() != 6 {
		t.Errorf("expected CenterHeight 6 (col 4), got %d", b.CenterHeight())
	}

	colH := b.ColHeights()
	if colH[4] != 6 || colH[1] != 3 || colH[0] != 0 {
		t.Errorf("unexpected ColHeights: col 4=%d, col 1=%d, col 0=%d", colH[4], colH[1], colH[0])
	}

	// Create a hole: empty cell at (4, 16) with filled above at (4, 14), (4, 15)
	b.Cells[16][4] = Cell{Filled: false}
	if b.CountHoles() != 1 {
		t.Errorf("expected 1 hole, got %d", b.CountHoles())
	}

	// Test ToStringGrid and RowMask
	grid := b.ToStringGrid()
	if len(grid) != BoardHeight {
		t.Fatalf("expected 20 rows in ToStringGrid, got %d", len(grid))
	}
	// Row 14 should have '1' at col 4 (index 4)
	if grid[14] != "0000100000" {
		t.Errorf("expected row 14 to be '0000100000', got '%s'", grid[14])
	}
	expectedMask := uint16(1 << (BoardWidth - 1 - 4))
	if b.RowMask(14) != expectedMask {
		t.Errorf("expected RowMask(14) %d, got %d", expectedMask, b.RowMask(14))
	}
}

func TestLockDelayManualPlay(t *testing.T) {
	g := NewGame()
	g.AutoPlay = false
	g.CurrentPiece = NewPiece(PieceO)
	g.CurrentPiece.X = 4
	g.CurrentPiece.Y = 18 // On the floor (rows 18 and 19 filled for 2x2 piece O)

	// In manual play, piece is grounded. First tick should NOT lock piece yet (lock delay activated)
	lockedMoves := g.MoveCount
	g.Tick()
	if g.MoveCount != lockedMoves {
		t.Errorf("expected piece not to lock on first grounded tick due to lock delay")
	}
	if !g.lockDelayActive {
		t.Errorf("expected lockDelayActive to be true")
	}

	// Move right resets the lock delay
	g.MoveRight()
	if g.lockResets != 1 {
		t.Errorf("expected lockResets to be 1, got %d", g.lockResets)
	}

	// Ticking down until lock
	g.Tick() // ticks remaining: 1
	if g.MoveCount != lockedMoves {
		t.Errorf("expected piece not to lock while delay remains")
	}
	g.Tick() // ticks remaining: 0 -> locks!
	if g.MoveCount != lockedMoves+1 {
		t.Errorf("expected piece to lock after lock delay expires")
	}

	// In AutoPlay mode, grounded piece locks immediately without delay
	gAuto := NewGame()
	gAuto.AutoPlay = true
	gAuto.CurrentPiece = NewPiece(PieceO)
	gAuto.CurrentPiece.X = 4
	gAuto.CurrentPiece.Y = 18
	gAuto.Tick()
	if gAuto.MoveCount != 1 {
		t.Errorf("expected piece to lock immediately in AutoPlay mode")
	}
}



