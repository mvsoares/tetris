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

var rowStrings [1024]string

func init() {
	for mask := 0; mask < 1024; mask++ {
		row := make([]byte, BoardWidth)
		for x := 0; x < BoardWidth; x++ {
			if (mask & (1 << (BoardWidth - 1 - x))) != 0 {
				row[x] = '1'
			} else {
				row[x] = '0'
			}
		}
		rowStrings[mask] = string(row)
	}
}

// RowMask returns the 10-bit integer bitmask for row y.
func (b *Board) RowMask(y int) uint16 {
	var mask uint16
	for x := 0; x < BoardWidth; x++ {
		if b.Cells[y][x].Filled {
			mask |= 1 << (BoardWidth - 1 - x)
		}
	}
	return mask
}

// ToStringGrid returns 20 strings of length 10 ("0" for empty, "1" for filled).
// Uses precomputed row strings to eliminate string heap allocations.
func (b *Board) ToStringGrid() []string {
	res := make([]string, BoardHeight)
	for y := 0; y < BoardHeight; y++ {
		res[y] = rowStrings[b.RowMask(y)]
	}
	return res
}

// IsValidPosition checks if a piece can exist at its current position without collision.
// Directly checks the shape matrix without heap allocations.
func (b *Board) IsValidPosition(p *Piece) bool {
	shape := p.Shape()
	for r := 0; r < len(shape); r++ {
		for c := 0; c < len(shape[r]); c++ {
			if shape[r][c] != 0 {
				x := p.X + c
				y := p.Y + r
				// Horizontal boundary check
				if x < 0 || x >= BoardWidth {
					return false
				}
				// Bottom boundary check
				if y >= BoardHeight {
					return false
				}
				// Cells above the board (y < 0) are valid during spawn
				if y >= 0 && b.Cells[y][x].Filled {
					return false
				}
			}
		}
	}
	return true
}

// LockPiece permanently places a piece onto the board without slice allocations.
func (b *Board) LockPiece(p *Piece) {
	shape := p.Shape()
	for r := 0; r < len(shape); r++ {
		for c := 0; c < len(shape[r]); c++ {
			if shape[r][c] != 0 {
				x := p.X + c
				y := p.Y + r
				if y >= 0 && y < BoardHeight && x >= 0 && x < BoardWidth {
					b.Cells[y][x] = Cell{
						Filled: true,
						Color:  p.Color,
					}
				}
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

// GetGhostY calculates the lowest valid Y position for the given piece without allocations.
func (b *Board) GetGhostY(p *Piece) int {
	origY := p.Y
	for b.IsValidPosition(p) {
		p.Y++
	}
	ghostY := p.Y - 1
	p.Y = origY
	return ghostY
}

// Standard SRS (Super Rotation System) Directional Kick Tables.
// Rotations: 0 = spawn, 1 = 90 deg CW, 2 = 180 deg, 3 = 270 deg CW (90 CCW).
// Offsets are [dx, dy] where +x is right and +y is DOWN (board coordinates).

// srsKicksJLSTZ maps [fromRotation][toRotation] to kick offsets for J, L, S, T, Z pieces.
var srsKicksJLSTZ = [4][4][][2]int{
	0: {
		1: {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}}, // 0 -> 1 (CW)
		3: {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},   // 0 -> 3 (CCW)
	},
	1: {
		0: {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},   // 1 -> 0 (CCW)
		2: {{0, 0}, {1, 0}, {1, 1}, {0, -2}, {1, -2}},   // 1 -> 2 (CW)
	},
	2: {
		1: {{0, 0}, {-1, 0}, {-1, -1}, {0, 2}, {-1, 2}}, // 2 -> 1 (CCW)
		3: {{0, 0}, {1, 0}, {1, -1}, {0, 2}, {1, 2}},   // 2 -> 3 (CW)
	},
	3: {
		2: {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 3 -> 2 (CCW)
		0: {{0, 0}, {-1, 0}, {-1, 1}, {0, -2}, {-1, -2}}, // 3 -> 0 (CW)
	},
}

// srsKicksI maps [fromRotation][toRotation] to kick offsets for the I piece.
var srsKicksI = [4][4][][2]int{
	0: {
		1: {{0, 0}, {-2, 0}, {1, 0}, {-2, 1}, {1, -2}}, // 0 -> 1 (CW)
		3: {{0, 0}, {-1, 0}, {2, 0}, {-1, -2}, {2, 1}}, // 0 -> 3 (CCW)
	},
	1: {
		0: {{0, 0}, {2, 0}, {-1, 0}, {2, -1}, {-1, 2}}, // 1 -> 0 (CCW)
		2: {{0, 0}, {-1, 0}, {2, 0}, {-1, -2}, {2, 1}}, // 1 -> 2 (CW)
	},
	2: {
		1: {{0, 0}, {1, 0}, {-2, 0}, {1, 2}, {-2, -1}}, // 2 -> 1 (CCW)
		3: {{0, 0}, {2, 0}, {-1, 0}, {2, -1}, {-1, 2}}, // 2 -> 3 (CW)
	},
	3: {
		2: {{0, 0}, {-2, 0}, {1, 0}, {-2, 1}, {1, -2}}, // 3 -> 2 (CCW)
		0: {{0, 0}, {1, 0}, {-2, 0}, {1, 2}, {-2, -1}}, // 3 -> 0 (CW)
	},
}

// TryRotate attempts to rotate the piece in the specified direction (+1 CW, -1 CCW),
// applying official SRS directional wall kicks. Returns true if rotation succeeded.
func (b *Board) TryRotate(p *Piece, direction int) bool {
	targetRotation := (p.Rotation + direction + 4) % 4

	var kicks [][2]int
	if p.Type == PieceO {
		kicks = [][2]int{{0, 0}}
	} else if p.Type == PieceI {
		kicks = srsKicksI[p.Rotation][targetRotation]
	} else {
		kicks = srsKicksJLSTZ[p.Rotation][targetRotation]
	}

	if len(kicks) == 0 {
		kicks = [][2]int{{0, 0}}
	}

	origX, origY, origRot := p.X, p.Y, p.Rotation
	p.Rotation = targetRotation

	for _, kick := range kicks {
		p.X = origX + kick[0]
		p.Y = origY + kick[1]
		if b.IsValidPosition(p) {
			return true
		}
	}

	// Restore original state if all kicks failed
	p.X = origX
	p.Y = origY
	p.Rotation = origRot
	return false
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

// MaxHeight calculates the highest stack height among all columns.
func (b *Board) MaxHeight() int {
	maxH := 0
	for x := 0; x < BoardWidth; x++ {
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				h := BoardHeight - y
				if h > maxH {
					maxH = h
				}
				break
			}
		}
	}
	return maxH
}

// CenterHeight returns the maximum height in the critical piece-spawn corridor (columns 3, 4, 5, 6).
func (b *Board) CenterHeight() int {
	maxH := 0
	for x := 3; x <= 6; x++ {
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				h := BoardHeight - y
				if h > maxH {
					maxH = h
				}
				break
			}
		}
	}
	return maxH
}

// ColHeights returns the height of each column as a stack array with zero heap allocations.
func (b *Board) ColHeights() [BoardWidth]int {
	var heights [BoardWidth]int
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

// ColumnHeights returns the height of each column as a slice.
func (b *Board) ColumnHeights() []int {
	h := b.ColHeights()
	res := make([]int, BoardWidth)
	copy(res, h[:])
	return res
}

// Bumpiness returns the sum of absolute height differences between adjacent columns.
func (b *Board) Bumpiness() int {
	heights := b.ColHeights()
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

