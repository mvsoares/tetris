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


