package engine

import (
	"fmt"
	"time"

	"tetris/internal/logger"
)

type GameState int

const (
	StatePlaying GameState = iota
	StatePaused
	StateGameOver
)

type Game struct {
	Board            *Board
	CurrentPiece     *Piece
	HoldPiece        *Piece
	CanHold          bool
	NextQueue        []TetrominoType
	Randomizer       *Randomizer
	Score            int
	HighScore        int
	Level            int
	Lines            int
	State            GameState
	LastAction       string
	LastScoreGained  int
	AutoPlay         bool
	CurrentAIMove    *AIMove
	stuckTicks       int
	Logger           *logger.Logger
	SessionID        string
	MoveCount        int
	UsedHoldThisTurn bool
	Singles          int
	Doubles          int
	Triples          int
	Tetrises         int
	sessionEnded     bool
	lockDelayTicks   int
	lockDelayActive  bool
	lockResets       int
}

const (
	NextQueueSize = 3
)

func NewGame() *Game {
	g := &Game{
		Board:      NewBoard(),
		Randomizer: NewRandomizer(),
		Level:      1,
		CanHold:    true,
		State:      StatePlaying,
		SessionID:  fmt.Sprintf("session-%s", time.Now().Format("20060102-150405.000")),
	}

	// Initialize the next queue
	for i := 0; i < NextQueueSize; i++ {
		g.NextQueue = append(g.NextQueue, g.Randomizer.Next())
	}

	g.spawnPiece()
	return g
}

func (g *Game) spawnPiece() {
	nextType := g.NextQueue[0]
	g.NextQueue = append(g.NextQueue[1:], g.Randomizer.Next())

	g.CurrentPiece = NewPiece(nextType)
	g.CanHold = true
	g.CurrentAIMove = nil
	g.stuckTicks = 0
	g.lockDelayActive = false
	g.lockDelayTicks = 0
	g.lockResets = 0

	// If spawned piece collides immediately, game over
	if !g.Board.IsValidPosition(g.CurrentPiece) {
		g.State = StateGameOver
		g.LastAction = "GAME OVER"
		g.logSessionEnd()
	}
}

// MoveLeft moves current piece left by 1 if valid.
func (g *Game) MoveLeft() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}
	test := *g.CurrentPiece
	test.X--
	if g.Board.IsValidPosition(&test) {
		g.CurrentPiece.X--
		if g.lockDelayActive && g.lockResets < 15 {
			g.lockDelayTicks = 2
			g.lockResets++
		}
		return true
	}
	return false
}

// MoveRight moves current piece right by 1 if valid.
func (g *Game) MoveRight() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}
	test := *g.CurrentPiece
	test.X++
	if g.Board.IsValidPosition(&test) {
		g.CurrentPiece.X++
		if g.lockDelayActive && g.lockResets < 15 {
			g.lockDelayTicks = 2
			g.lockResets++
		}
		return true
	}
	return false
}

// SoftDrop moves piece down by 1; awards 1 point.
func (g *Game) SoftDrop() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}
	test := *g.CurrentPiece
	test.Y++
	if g.Board.IsValidPosition(&test) {
		g.CurrentPiece.Y++
		g.addScore(1)
		return true
	}
	// Reached bottom, lock it
	g.lockDelayActive = false
	g.lockCurrentPiece()
	return false
}

// HardDrop instantly drops piece to ghost position, locks it, and awards 2 points per row dropped.
func (g *Game) HardDrop() int {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return 0
	}
	ghostY := g.Board.GetGhostY(g.CurrentPiece)
	dropDistance := ghostY - g.CurrentPiece.Y
	if dropDistance < 0 {
		dropDistance = 0
	}
	g.CurrentPiece.Y = ghostY
	g.addScore(dropDistance * 2)
	g.lockDelayActive = false
	g.lockCurrentPiece()
	return dropDistance
}

// RotateCW rotates the piece 90 degrees clockwise with wall kicks.
func (g *Game) RotateCW() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}
	rotated := g.Board.TryRotate(g.CurrentPiece, 1)
	if rotated && g.lockDelayActive && g.lockResets < 15 {
		g.lockDelayTicks = 2
		g.lockResets++
	}
	return rotated
}

// RotateCCW rotates the piece 90 degrees counter-clockwise with wall kicks.
func (g *Game) RotateCCW() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}
	rotated := g.Board.TryRotate(g.CurrentPiece, -1)
	if rotated && g.lockDelayActive && g.lockResets < 15 {
		g.lockDelayTicks = 2
		g.lockResets++
	}
	return rotated
}

// Hold swaps current piece with hold piece or stores it and spawns next.
func (g *Game) Hold() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil || !g.CanHold {
		return false
	}

	g.UsedHoldThisTurn = true
	g.lockDelayActive = false
	g.lockDelayTicks = 0
	g.lockResets = 0
	currentType := g.CurrentPiece.Type
	if g.HoldPiece == nil {
		g.HoldPiece = NewPiece(currentType)
		g.spawnPiece()
	} else {
		heldType := g.HoldPiece.Type
		g.HoldPiece = NewPiece(currentType)
		g.CurrentPiece = NewPiece(heldType)
		if !g.Board.IsValidPosition(g.CurrentPiece) {
			g.State = StateGameOver
			g.LastAction = "GAME OVER"
			return false
		}
	}

	g.CanHold = false
	return true
}

// Tick represents the standard gravity cycle step with lock delay support for manual play.
func (g *Game) Tick() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}

	test := *g.CurrentPiece
	test.Y++
	if g.Board.IsValidPosition(&test) {
		g.CurrentPiece.Y++
		testGround := *g.CurrentPiece
		testGround.Y++
		if g.Board.IsValidPosition(&testGround) {
			g.lockDelayActive = false
			g.lockDelayTicks = 0
		}
		return true
	}

	// Piece is grounded
	if g.AutoPlay {
		g.lockCurrentPiece()
		return false
	}

	// Manual play: Lock delay gives player time to slide/rotate
	if !g.lockDelayActive {
		g.lockDelayActive = true
		g.lockDelayTicks = 2 // ~500ms at typical gravity
		return true
	}

	g.lockDelayTicks--
	if g.lockDelayTicks <= 0 {
		g.lockDelayActive = false
		g.lockCurrentPiece()
		return false
	}
	return true
}

func (g *Game) lockCurrentPiece() {
	maxHBefore := GetMaxHeight(g.Board)
	pType := string(g.CurrentPiece.Type)
	rot := g.CurrentPiece.Rotation
	pX := g.CurrentPiece.X
	pY := g.CurrentPiece.Y
	usedHold := g.UsedHoldThisTurn
	playerType := "HUMAN"
	if g.AutoPlay {
		playerType = "AI"
	}
	scoreBefore := g.Score

	g.Board.LockPiece(g.CurrentPiece)
	cleared := g.Board.ClearLines()
	if cleared > 0 {
		g.Lines += cleared
		scoreMultiplier := 0
		actionText := ""

		switch cleared {
		case 1:
			scoreMultiplier = 100
			actionText = "SINGLE"
			g.Singles++
		case 2:
			scoreMultiplier = 300
			actionText = "DOUBLE!"
			g.Doubles++
		case 3:
			scoreMultiplier = 500
			actionText = "TRIPLE!!"
			g.Triples++
		case 4:
			scoreMultiplier = 800
			actionText = "💥 TETRIS! 💥"
			g.Tetrises++
		}

		gained := scoreMultiplier * g.Level
		g.addScore(gained)
		g.LastAction = actionText
		g.LastScoreGained = gained

		// Level up every 10 lines
		newLevel := (g.Lines / 10) + 1
		if newLevel > g.Level {
			g.Level = newLevel
			g.LastAction = "LEVEL UP!"
		}
	}

	g.MoveCount++

	if g.Logger != nil {
		maxHAfter := GetMaxHeight(g.Board)
		holesAfter := g.Board.CountHoles()
		bumpAfter := g.Board.Bumpiness()
		gridAfter := g.Board.ToStringGrid()

		var hScore float64
		if g.CurrentAIMove != nil {
			hScore = g.CurrentAIMove.Score
		}

		g.Logger.LogMove(logger.MoveLog{
			SessionID:       g.SessionID,
			Timestamp:       time.Now(),
			MoveNumber:      g.MoveCount,
			PlayerType:      playerType,
			PieceType:       pType,
			UsedHold:        usedHold,
			Rotation:        rot,
			X:               pX,
			Y:               pY,
			LinesCleared:    cleared,
			ScoreGained:     g.Score - scoreBefore,
			TotalScore:      g.Score,
			TotalLines:      g.Lines,
			Level:           g.Level,
			MaxHeightBefore: maxHBefore,
			MaxHeightAfter:  maxHAfter,
			HolesAfter:      holesAfter,
			BumpinessAfter:  bumpAfter,
			IsCleanupMode:   IsCleanupMode(g),
			HeuristicScore:  hScore,
			BoardStateAfter: gridAfter,
		})
	}

	g.UsedHoldThisTurn = false
	g.spawnPiece()
}

func (g *Game) addScore(pts int) {
	g.Score += pts
	if g.Score > g.HighScore {
		g.HighScore = g.Score
	}
}

// SetLogger sets the play logger for recording moves.
func (g *Game) SetLogger(l *logger.Logger) {
	g.Logger = l
}

// Close finalizes logging for the current game session without closing shared logger.
func (g *Game) Close() error {
	g.logSessionEnd()
	return nil
}

// CloseLogger finalizes the session and closes the underlying logger file.
func (g *Game) CloseLogger() error {
	g.logSessionEnd()
	if g.Logger != nil {
		return g.Logger.Close()
	}
	return nil
}

func (g *Game) logSessionEnd() {
	if g.Logger != nil && !g.sessionEnded {
		g.sessionEnded = true
		var tetrisRate float64
		if g.Lines > 0 {
			tetrisRate = (float64(g.Tetrises*4) / float64(g.Lines)) * 100.0
		}
		g.Logger.LogSessionEnd(logger.SessionEndLog{
			SessionID:         g.SessionID,
			Timestamp:         time.Now(),
			TotalMoves:        g.MoveCount,
			FinalScore:        g.Score,
			FinalLevel:        g.Level,
			TotalLines:        g.Lines,
			Singles:           g.Singles,
			Doubles:           g.Doubles,
			Triples:           g.Triples,
			Tetrises:          g.Tetrises,
			TetrisRatePercent: tetrisRate,
		})
	}
}

// TogglePause toggles pause state.
func (g *Game) TogglePause() {
	if g.State == StatePlaying {
		g.State = StatePaused
	} else if g.State == StatePaused {
		g.State = StatePlaying
	}
}

// Restart resets the game to initial state, keeping the HighScore, Logger, and AutoPlay.
func (g *Game) Restart() {
	g.logSessionEnd()
	high := g.HighScore
	savedLogger := g.Logger
	savedAutoPlay := g.AutoPlay

	*g = *NewGame()
	g.HighScore = high
	g.Logger = savedLogger
	g.AutoPlay = savedAutoPlay
}

// TickInterval calculates the gravity speed based on current level.
func (g *Game) TickInterval() time.Duration {
	// Base: 800ms at level 1, decreases by 60ms each level, minimum 80ms
	ms := 800 - (g.Level-1)*60
	if ms < 80 {
		ms = 80
	}
	return time.Duration(ms) * time.Millisecond
}

// QueueLinePieces queues consecutive I-tetrominoes ("linhas").
// The current piece becomes an I-piece, and the next pieces in the queue become I-pieces.
func (g *Game) QueueLinePieces(count int) {
	if g.State != StatePlaying {
		return
	}
	g.CurrentPiece = NewPiece(PieceI)
	g.NextQueue = make([]TetrominoType, count)
	for i := 0; i < count; i++ {
		g.NextQueue[i] = PieceI
	}
	g.LastAction = "4 LINHAS SEGUIDAS (I)!"
}

// SetupFourLines fills the bottom 4 rows with 9 blocks (leaving col 9 open for a Tetris)
// and equips the player with an I-piece ("linha").
func (g *Game) SetupFourLines() {
	if g.State != StatePlaying {
		return
	}
	colorPalette := []string{"#0055ff", "#ff7700", "#f0f000", "#00f000"}
	for i, y := range []int{16, 17, 18, 19} {
		color := colorPalette[i%len(colorPalette)]
		for x := 0; x < BoardWidth; x++ {
			if x < BoardWidth-1 {
				g.Board.Cells[y][x] = Cell{Filled: true, Color: color}
			} else {
				g.Board.Cells[y][x] = Cell{Filled: false}
			}
		}
	}
	g.CurrentPiece = NewPiece(PieceI)
	g.LastAction = "SETUP 4 LINHAS PRONTO!"
}

// ToggleAutoPlay switches the AI auto-player on or off.
func (g *Game) ToggleAutoPlay() bool {
	g.AutoPlay = !g.AutoPlay
	g.CurrentAIMove = nil
	g.stuckTicks = 0
	if g.AutoPlay {
		g.LastAction = "🤖 AUTO-PLAY ON"
	} else {
		g.LastAction = "🎮 MANUAL ON"
	}
	return g.AutoPlay
}

// StepAI executes one action step towards the computed optimal placement.
// Returns true if an action was taken.
func (g *Game) StepAI() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil || !g.AutoPlay {
		return false
	}

	// Compute plan if absent
	if g.CurrentAIMove == nil {
		g.CurrentAIMove = FindBestMove(g)
		g.stuckTicks = 0
		if g.CurrentAIMove == nil {
			return false
		}

		// If optimal plan requires hold, execute hold and re-plan
		if g.CurrentAIMove.UseHold && g.CanHold {
			g.Hold()
			g.CurrentAIMove = FindBestMove(g)
			if g.CurrentAIMove == nil {
				return false
			}
		}
	}

	plan := g.CurrentAIMove
	if plan.CleanupMode && (g.LastAction == "" || g.LastAction == "🤖 AUTO-PLAY ON" || g.LastAction == "🎮 MANUAL ON") {
		g.LastAction = "🚨 LIMPEZA (65%+)"
	}

	// 1. Rotate towards target rotation
	if g.CurrentPiece.Rotation != plan.TargetRotation {
		if !g.RotateCW() {
			g.stuckTicks++
		} else {
			g.stuckTicks = 0
			return true
		}
	}

	// 2. Move towards target X
	if g.CurrentPiece.X < plan.TargetX {
		if !g.MoveRight() {
			g.stuckTicks++
		} else {
			g.stuckTicks = 0
			return true
		}
	} else if g.CurrentPiece.X > plan.TargetX {
		if !g.MoveLeft() {
			g.stuckTicks++
		} else {
			g.stuckTicks = 0
			return true
		}
	}

	// 3. Target reached: hard drop!
	if g.CurrentPiece.Rotation == plan.TargetRotation && g.CurrentPiece.X == plan.TargetX {
		g.HardDrop()
		g.CurrentAIMove = nil
		g.stuckTicks = 0
		return true
	}

	// Watchdog: if piece gets blocked by obstacles, hard drop to unblock
	if g.stuckTicks >= 2 {
		g.HardDrop()
		g.CurrentAIMove = nil
		g.stuckTicks = 0
		return true
	}

	return false
}

// StepAIImmediate calculates the optimal move and drops the piece immediately.
// Designed for headless simulations, high-volume log generation, and background training.
// Returns true if a piece was placed, false if game over or no valid move.
func (g *Game) StepAIImmediate() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}

	plan := FindBestMove(g)
	if plan == nil {
		g.State = StateGameOver
		g.logSessionEnd()
		return false
	}

	if plan.UseHold && g.CanHold {
		if !g.Hold() {
			return false
		}
		plan = FindBestMove(g)
		if plan == nil {
			g.State = StateGameOver
			g.logSessionEnd()
			return false
		}
	}

	g.CurrentPiece.Rotation = plan.TargetRotation
	g.CurrentPiece.X = plan.TargetX
	g.CurrentPiece.Y = -2
	for g.CurrentPiece.Y < 0 && !g.Board.IsValidPosition(g.CurrentPiece) {
		g.CurrentPiece.Y++
	}

	if !g.Board.IsValidPosition(g.CurrentPiece) {
		g.CurrentPiece.Y = 0
		if !g.Board.IsValidPosition(g.CurrentPiece) {
			g.State = StateGameOver
			g.logSessionEnd()
			return false
		}
	}

	g.CurrentAIMove = plan
	g.HardDrop()
	g.CurrentAIMove = nil
	return true
}


