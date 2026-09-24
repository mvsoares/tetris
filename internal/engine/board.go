package engine

const (
	BoardWidth  = 10
	BoardHeight = 20
)

type Cell struct {
	Filled bool
	Color  string
}

type Board struct {
	Cells [BoardHeight][BoardWidth]Cell
}

func NewBoard() *Board {
	return &Board{}
}

// Clone creates a deep copy of the board matrix.
func (b *Board) Clone() *Board {
	copy := *b
	return &copy
}

// IsValidPosition checks if a piece can exist at its current position without collision.
func (b *Board) IsValidPosition(p *Piece) bool {
	coords := p.BlockCoords()
	for _, pt := range coords {
		x, y := pt[0], pt[1]
		// Horizontal boundary check
		if x < 0 || x >= BoardWidth {
			return false
		}
		// Bottom boundary check
		if y >= BoardHeight {
			return false
		}
		// Cells above the board (y < 0) are valid during spawn
		if y >= 0 {
			if b.Cells[y][x].Filled {
				return false
			}
		}
	}
	return true
}

// LockPiece permanently places a piece onto the board.
func (b *Board) LockPiece(p *Piece) {
	coords := p.BlockCoords()
	for _, pt := range coords {
		x, y := pt[0], pt[1]
		if y >= 0 && y < BoardHeight && x >= 0 && x < BoardWidth {
			b.Cells[y][x] = Cell{
				Filled: true,
				Color:  p.Color,
			}
		}
	}
}

// ClearLines checks for and clears completed rows, shifting lines above downwards.
// Returns the count of cleared lines.
func (b *Board) ClearLines() int {
	cleared := 0
	for y := BoardHeight - 1; y >= 0; y-- {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if !b.Cells[y][x].Filled {
				full = false
				break
			}
		}

		if full {
			cleared++
			// Shift all rows above down by 1
			for row := y; row > 0; row-- {
				b.Cells[row] = b.Cells[row-1]
			}
			// Reset top row
			b.Cells[0] = [BoardWidth]Cell{}
			// Re-check this row index since a new row moved into it
			y++
		}
	}
	return cleared
}

// GetGhostY calculates the lowest valid Y position for the given piece.
func (b *Board) GetGhostY(p *Piece) int {
	test := p.Clone()
	for b.IsValidPosition(test) {
		test.Y++
	}
	return test.Y - 1
}

// TryRotate attempts to rotate the piece in the specified direction (+1 CW, -1 CCW),
// applying wall kicks if necessary. Returns true if rotation succeeded.
func (b *Board) TryRotate(p *Piece, direction int) bool {
	targetRotation := (p.Rotation + direction + 4) % 4
	test := p.Clone()
	test.Rotation = targetRotation

	// Kick offsets to test: (dx, dy)
	var kicks [][2]int
	if p.Type == PieceI {
		kicks = [][2]int{
			{0, 0},
			{-2, 0},
			{1, 0},
			{-1, 0},
			{2, 0},
			{0, -1},
			{-2, -1},
			{1, -1},
		}
	} else if p.Type == PieceO {
		kicks = [][2]int{{0, 0}}
	} else {
		kicks = [][2]int{
			{0, 0},
			{-1, 0},
			{1, 0},
			{0, -1},
			{-1, -1},
			{1, -1},
			{-2, 0},
			{2, 0},
		}
	}

	for _, kick := range kicks {
		kicked := test.Clone()
		kicked.X += kick[0]
		kicked.Y += kick[1]
		if b.IsValidPosition(kicked) {
			p.Rotation = targetRotation
			p.X = kicked.X
			p.Y = kicked.Y
			return true
		}
	}

	return false
}

// ToStringGrid returns 20 strings of length 10 ("0" for empty, "1" for filled).
func (b *Board) ToStringGrid() []string {
	res := make([]string, BoardHeight)
	for y := 0; y < BoardHeight; y++ {
		row := make([]byte, BoardWidth)
		for x := 0; x < BoardWidth; x++ {
			if b.Cells[y][x].Filled {
				row[x] = '1'
			} else {
				row[x] = '0'
			}
		}
		res[y] = string(row)
	}
	return res
}

// CountHoles returns the number of empty cells beneath filled cells.
func (b *Board) CountHoles() int {
	holes := 0
	for x := 0; x < BoardWidth; x++ {
		foundFilled := false
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				foundFilled = true
			} else if foundFilled {
				holes++
			}
		}
	}
	return holes
}

// ColumnHeights returns the height of each column.
func (b *Board) ColumnHeights() []int {
	heights := make([]int, BoardWidth)
	for x := 0; x < BoardWidth; x++ {
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				heights[x] = BoardHeight - y
				break
			}
		}
	}
	return heights
}

// Bumpiness returns the sum of absolute height differences between adjacent columns.
func (b *Board) Bumpiness() int {
	heights := b.ColumnHeights()
	bumpy := 0
	for x := 0; x < BoardWidth-1; x++ {
		diff := heights[x] - heights[x+1]
		if diff < 0 {
			diff = -diff
		}
		bumpy += diff
	}
	return bumpy
}

