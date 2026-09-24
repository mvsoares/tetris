package engine

type TetrominoType string

const (
	PieceI TetrominoType = "I"
	PieceJ TetrominoType = "J"
	PieceL TetrominoType = "L"
	PieceO TetrominoType = "O"
	PieceS TetrominoType = "S"
	PieceT TetrominoType = "T"
	PieceZ TetrominoType = "Z"
)

var AllPieces = []TetrominoType{
	PieceI, PieceJ, PieceL, PieceO, PieceS, PieceT, PieceZ,
}

// Colors according to standard Tetris guidelines
var PieceColors = map[TetrominoType]string{
	PieceI: "#00f0f0", // Cyan
	PieceJ: "#0055ff", // Blue
	PieceL: "#ff7700", // Orange
	PieceO: "#f0f000", // Yellow
	PieceS: "#00f000", // Green
	PieceT: "#aa00ff", // Purple
	PieceZ: "#f00000", // Red
}

// Standard SRS 4x4 or 3x3 matrices for all 4 rotations (0: spawn, 1: 90 CW, 2: 180, 3: 270 CW)
var pieceShapes = map[TetrominoType][4][][]int{
	PieceI: {
		{
			{0, 0, 0, 0},
			{1, 1, 1, 1},
			{0, 0, 0, 0},
			{0, 0, 0, 0},
		},
		{
			{0, 0, 1, 0},
			{0, 0, 1, 0},
			{0, 0, 1, 0},
			{0, 0, 1, 0},
		},
		{
			{0, 0, 0, 0},
			{0, 0, 0, 0},
			{1, 1, 1, 1},
			{0, 0, 0, 0},
		},
		{
			{0, 1, 0, 0},
			{0, 1, 0, 0},
			{0, 1, 0, 0},
			{0, 1, 0, 0},
		},
	},
	PieceO: {
		{
			{1, 1},
			{1, 1},
		},
		{
			{1, 1},
			{1, 1},
		},
		{
			{1, 1},
			{1, 1},
		},
		{
			{1, 1},
			{1, 1},
		},
	},
	PieceT: {
		{
			{0, 1, 0},
			{1, 1, 1},
			{0, 0, 0},
		},
		{
			{0, 1, 0},
			{0, 1, 1},
			{0, 1, 0},
		},
		{
			{0, 0, 0},
			{1, 1, 1},
			{0, 1, 0},
		},
		{
			{0, 1, 0},
			{1, 1, 0},
			{0, 1, 0},
		},
	},
	PieceS: {
		{
			{0, 1, 1},
			{1, 1, 0},
			{0, 0, 0},
		},
		{
			{0, 1, 0},
			{0, 1, 1},
			{0, 0, 1},
		},
		{
			{0, 0, 0},
			{0, 1, 1},
			{1, 1, 0},
		},
		{
			{1, 0, 0},
			{1, 1, 0},
			{0, 1, 0},
		},
	},
	PieceZ: {
		{
			{1, 1, 0},
			{0, 1, 1},
			{0, 0, 0},
		},
		{
			{0, 0, 1},
			{0, 1, 1},
			{0, 1, 0},
		},
		{
			{0, 0, 0},
			{1, 1, 0},
			{0, 1, 1},
		},
		{
			{0, 1, 0},
			{1, 1, 0},
			{1, 0, 0},
		},
	},
	PieceJ: {
		{
			{1, 0, 0},
			{1, 1, 1},
			{0, 0, 0},
		},
		{
			{0, 1, 1},
			{0, 1, 0},
			{0, 1, 0},
		},
		{
			{0, 0, 0},
			{1, 1, 1},
			{0, 0, 1},
		},
		{
			{0, 1, 0},
			{0, 1, 0},
			{1, 1, 0},
		},
	},
	PieceL: {
		{
			{0, 0, 1},
			{1, 1, 1},
			{0, 0, 0},
		},
		{
			{0, 1, 0},
			{0, 1, 0},
			{0, 1, 1},
		},
		{
			{0, 0, 0},
			{1, 1, 1},
			{1, 0, 0},
		},
		{
			{1, 1, 0},
			{0, 1, 0},
			{0, 1, 0},
		},
	},
}

type Piece struct {
	Type     TetrominoType
	Rotation int // 0 to 3
	X        int // Board X coordinate (col)
	Y        int // Board Y coordinate (row)
	Color    string
}

// NewPiece creates a new tetromino at its initial spawn position.
func NewPiece(t TetrominoType) *Piece {
	spawnX := 3
	spawnY := 0
	if t == PieceO {
		spawnX = 4
	}
	return &Piece{
		Type:     t,
		Rotation: 0,
		X:        spawnX,
		Y:        spawnY,
		Color:    PieceColors[t],
	}
}

// Clone creates a deep copy of the piece.
func (p *Piece) Clone() *Piece {
	return &Piece{
		Type:     p.Type,
		Rotation: p.Rotation,
		X:        p.X,
		Y:        p.Y,
		Color:    p.Color,
	}
}

// Shape returns the 2D matrix of the piece at its current rotation.
func (p *Piece) Shape() [][]int {
	return pieceShapes[p.Type][p.Rotation%4]
}

// ShapeFor returns the 2D matrix for a specific type and rotation.
func ShapeFor(t TetrominoType, rotation int) [][]int {
	return pieceShapes[t][rotation%4]
}

// BlockCoords returns absolute board coordinates (col, row) of the piece's filled blocks.
func (p *Piece) BlockCoords() [][2]int {
	shape := p.Shape()
	coords := make([][2]int, 0, 4)
	for r := 0; r < len(shape); r++ {
		for c := 0; c < len(shape[r]); c++ {
			if shape[r][c] != 0 {
				coords = append(coords, [2]int{p.X + c, p.Y + r})
			}
		}
	}
	return coords
}
