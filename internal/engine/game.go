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

const PolicyVersion = "heuristic-9-0-v2"

const (
	EndTopOut      = "top_out"
	EndNoLegalMove = "no_legal_move"
	EndMoveLimit   = "move_limit"
	EndCancelled   = "cancelled"
	EndClosed      = "closed"
	EndRestarted   = "restarted"
)

type Game struct {
	Board               *Board
	CurrentPiece        *Piece
	HoldPiece           *Piece
	CanHold             bool
	NextQueue           []TetrominoType
	Randomizer          *Randomizer
	Score               int
	HighScore           int
	Level               int
	Lines               int
	State               GameState
	LastAction          string
	LastScoreGained     int
	AutoPlay            bool
	CurrentAIMove       *AIMove
	Logger              *logger.Logger
	SessionID           string
	MoveCount           int
	UsedHoldThisTurn    bool
	Singles             int
	Doubles             int
	Triples             int
	Tetrises            int
	Seed                int64
	EndReason           string
	FailedPiece         TetrominoType
	ExecutionMode       string
	PlanMisses          int
	WatchdogDrops       int
	AIReplans           int
	ReserveWell         bool
	LookaheadDepth      int
	BeamWidth           int
	LearnedModel        *MoveModel
	UseLearned          bool
	learnedMoveThisTurn bool
	aiGravityRemaining  time.Duration
	aiGravityKnown      bool
	scoreAtLastLock     int
	sessionEnded        bool
	lockDelayTicks      int
	lockDelayActive     bool
	lockResets          int
}

const (
	NextQueueSize = 3
)

func NewGame() *Game {
	return NewGameWithSeed(time.Now().UnixNano())
}

func NewGameWithSeed(seed int64) *Game {
	g := &Game{
		Board:         NewBoard(),
		Randomizer:    NewRandomizerWithSeed(seed),
		Seed:          seed,
		ExecutionMode: "interactive",
		Level:         1,
		CanHold:       true,
		State:         StatePlaying,
		SessionID:     fmt.Sprintf("session-%d", time.Now().UnixNano()),
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
	g.lockDelayActive = false
	g.lockDelayTicks = 0
	g.lockResets = 0

	// If spawned piece collides immediately, game over
	if !g.Board.IsValidPosition(g.CurrentPiece) {
		g.gameOver(EndTopOut)
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
	return g.hardDrop(g.CurrentPiece.Y)
}

func (g *Game) hardDrop(startY int) int {
	ghostY := g.Board.GetGhostY(g.CurrentPiece)
	dropDistance := ghostY - startY
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
		if g.State != StatePlaying {
			return false
		}
	} else {
		heldType := g.HoldPiece.Type
		g.HoldPiece = NewPiece(currentType)
		g.CurrentPiece = NewPiece(heldType)
		if !g.Board.IsValidPosition(g.CurrentPiece) {
			g.gameOver(EndTopOut)
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
	scoreBefore := g.scoreAtLastLock
	cleanupMode := IsCleanupMode(g)
	if g.CurrentAIMove != nil {
		cleanupMode = g.CurrentAIMove.CleanupMode
	}
	var gridBefore []string
	hadAIPlan, planMatched := g.CurrentAIMove != nil, false
	if hadAIPlan {
		planMatched = rot == g.CurrentAIMove.TargetRotation && pX == g.CurrentAIMove.TargetX
		if g.CurrentAIMove.HasTargetY {
			planMatched = planMatched && pY == g.CurrentAIMove.TargetY
		}
		if !planMatched {
			g.PlanMisses++
		}
	}
	holesBefore, bumpBefore := 0, 0
	if g.Logger != nil {
		gridBefore = g.Board.ToStringGrid()
		holesBefore = g.Board.CountHoles()
		bumpBefore = g.Board.Bumpiness()
	}

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
		var searchDepth, searchNodes int
		var decision *logger.DecisionLog
		var probability *float64
		var learnedUsed bool
		var learnedFallback string
		if g.CurrentAIMove != nil {
			hScore = g.CurrentAIMove.Score
			searchDepth, searchNodes = g.CurrentAIMove.SearchDepth, g.CurrentAIMove.SearchNodes
			decision, probability = g.CurrentAIMove.Decision, g.CurrentAIMove.MoveProbability
			learnedUsed, learnedFallback = g.CurrentAIMove.LearnedUsed || g.learnedMoveThisTurn, g.CurrentAIMove.LearnedFallback
		}

		g.Logger.LogMove(logger.MoveLog{
			SessionID:        g.SessionID,
			Timestamp:        time.Now(),
			MoveNumber:       g.MoveCount,
			PlayerType:       playerType,
			PieceType:        pType,
			UsedHold:         usedHold,
			Rotation:         rot,
			X:                pX,
			Y:                pY,
			LinesCleared:     cleared,
			ScoreGained:      g.Score - scoreBefore,
			TotalScore:       g.Score,
			TotalLines:       g.Lines,
			Level:            g.Level,
			MaxHeightBefore:  maxHBefore,
			HolesBefore:      holesBefore,
			BumpinessBefore:  bumpBefore,
			MaxHeightAfter:   maxHAfter,
			HolesAfter:       holesAfter,
			BumpinessAfter:   bumpAfter,
			IsCleanupMode:    cleanupMode,
			HeuristicScore:   hScore,
			HadAIPlan:        hadAIPlan,
			AIPlanMatched:    planMatched,
			SearchDepth:      searchDepth,
			SearchNodes:      searchNodes,
			Decision:         decision,
			MoveProbability:  probability,
			LearnedUsed:      learnedUsed,
			LearnedFallback:  learnedFallback,
			BoardStateAfter:  gridAfter,
			BoardStateBefore: gridBefore,
		})
	}

	g.UsedHoldThisTurn = false
	g.learnedMoveThisTurn = false
	g.scoreAtLastLock = g.Score
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
	return g.CloseWithReason(EndClosed)
}

func (g *Game) CloseWithReason(reason string) error {
	if g.EndReason == "" {
		g.EndReason = reason
	}
	g.logSessionEnd()
	return nil
}

// CloseLogger finalizes the session and closes the underlying logger file.
func (g *Game) CloseLogger() error {
	_ = g.Close()
	if g.Logger != nil {
		return g.Logger.Close()
	}
	return nil
}

func (g *Game) gameOver(reason string) {
	g.State = StateGameOver
	g.LastAction = "GAME OVER"
	g.EndReason = reason
	if g.CurrentPiece != nil {
		g.FailedPiece = g.CurrentPiece.Type
	}
	g.logSessionEnd()
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
			EndReason:         g.EndReason,
			FailedPiece:       string(g.FailedPiece),
			Seed:              g.Seed,
			PolicyVersion:     g.policyVersion(),
			ExecutionMode:     g.ExecutionMode,
			MaxHeight:         g.Board.MaxHeight(),
			Holes:             g.Board.CountHoles(),
			Bumpiness:         g.Board.Bumpiness(),
			IsCleanupMode:     IsCleanupMode(g),
			BoardState:        g.Board.ToStringGrid(),
			PlanMisses:        g.PlanMisses,
			WatchdogDrops:     g.WatchdogDrops,
			AIReplans:         g.AIReplans,
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
	_ = g.CloseWithReason(EndRestarted)
	high := g.HighScore
	savedLogger := g.Logger
	savedAutoPlay := g.AutoPlay
	savedReserveWell := g.ReserveWell
	savedDepth, savedWidth := g.LookaheadDepth, g.BeamWidth
	savedModel, savedLearned := g.LearnedModel, g.UseLearned

	*g = *NewGame()
	g.HighScore = high
	g.Logger = savedLogger
	g.AutoPlay = savedAutoPlay
	g.ReserveWell = savedReserveWell
	g.ConfigureLookahead(savedDepth, savedWidth)
	g.LearnedModel, g.UseLearned = savedModel, savedLearned
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
	if g.AutoPlay {
		g.LastAction = "🤖 AUTO-PLAY ON"
	} else {
		g.LastAction = "🎮 MANUAL ON"
	}
	return g.AutoPlay
}

// canFollowRemainingPlan checks whether the remaining planned actions can still
// legally reach the planned landing when UI timer jitter causes the active piece's
// Y coordinate to differ from Expected[0].Y while X, Rotation, and Type match.
func canFollowRemainingPlan(b *Board, cur Piece, plan *AIMove) bool {
	if b == nil || plan == nil || len(plan.Expected) == 0 || len(plan.Actions) == 0 {
		return false
	}
	if cur.X != plan.Expected[0].X || cur.Rotation != plan.Expected[0].Rotation || cur.Type != plan.Expected[0].Type {
		return false
	}
	dy := cur.Y - plan.Expected[0].Y
	if dy == 0 {
		return true
	}
	p := cur
	for i, act := range plan.Actions {
		if i > 0 {
			if i >= len(plan.Expected) {
				return false
			}
			// Apply the vertical drop that occurred in the original plan between steps i-1 and i
			origDrop := plan.Expected[i].Y - plan.Expected[i-1].Y
			if plan.Actions[i-1] == AIDown {
				origDrop--
			}
			for d := 0; d < origDrop; d++ {
				down := p
				down.Y++
				if !b.IsValidPosition(&down) {
					return false
				}
				p = down
			}
		}
		switch act {
		case AILeft:
			p.X--
			if !b.IsValidPosition(&p) {
				return false
			}
		case AIRight:
			p.X++
			if !b.IsValidPosition(&p) {
				return false
			}
		case AIRotateCW:
			if !b.TryRotate(&p, 1) {
				return false
			}
		case AIRotateCCW:
			if !b.TryRotate(&p, -1) {
				return false
			}
		case AIDown:
			p.Y++
			if !b.IsValidPosition(&p) {
				return false
			}
		case AIHardDrop:
			ghostY := b.GetGhostY(&p)
			if p.X != plan.TargetX || p.Rotation != plan.TargetRotation {
				return false
			}
			if plan.HasTargetY && ghostY != plan.TargetY {
				return false
			}
			for j := range plan.Expected {
				plan.Expected[j].Y += dy
			}
			return true
		}
	}
	ghostY := b.GetGhostY(&p)
	if p.X != plan.TargetX || p.Rotation != plan.TargetRotation {
		return false
	}
	if plan.HasTargetY && ghostY != plan.TargetY {
		return false
	}
	for j := range plan.Expected {
		plan.Expected[j].Y += dy
	}
	return true
}

// StepAI executes one action step towards the computed optimal placement.
// Returns true if an action was taken.
func (g *Game) StepAI() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil || !g.AutoPlay {
		return false
	}
	for attempt := 0; attempt < 2; attempt++ {
		if g.CurrentAIMove != nil {
			p := g.CurrentAIMove
			if len(p.Expected) == 0 || len(p.Actions) == 0 || p.Expected[0].X != g.CurrentPiece.X || p.Expected[0].Rotation != g.CurrentPiece.Rotation || p.Expected[0].Type != g.CurrentPiece.Type {
				g.CurrentAIMove = nil
				g.AIReplans++
			} else if p.Expected[0].Y != g.CurrentPiece.Y && !canFollowRemainingPlan(g.Board, *g.CurrentPiece, p) {
				g.CurrentAIMove = nil
				g.AIReplans++
			}
		}
		if g.CurrentAIMove == nil {
			g.CurrentAIMove = findPolicyMove(g)
			if g.CurrentAIMove == nil {
				g.gameOver(EndNoLegalMove)
				return false
			}
			if g.CurrentAIMove.UseHold && g.CanHold {
				learnedHold := g.CurrentAIMove.LearnedUsed
				if !g.Hold() {
					return false
				}
				g.learnedMoveThisTurn = g.learnedMoveThisTurn || learnedHold
				g.CurrentAIMove = findPolicyMove(g)
				if g.CurrentAIMove == nil {
					g.gameOver(EndNoLegalMove)
					return false
				}
			}
			if g.Logger != nil {
				g.CurrentAIMove.Decision = captureDecision(g, g.CurrentAIMove)
			}
		}
		plan := g.CurrentAIMove
		if plan.CleanupMode {
			g.LastAction = "🚨 LIMPEZA"
		}
		action := plan.Actions[0]
		moved := true
		switch action {
		case AILeft:
			moved = g.MoveLeft()
		case AIRight:
			moved = g.MoveRight()
		case AIRotateCW:
			moved = g.RotateCW()
		case AIRotateCCW:
			moved = g.RotateCCW()
		case AIDown:
			moved = g.SoftDrop()
		case AIHardDrop:
			if plan.HasTargetY && g.Board.GetGhostY(g.CurrentPiece) != plan.TargetY {
				g.CurrentAIMove = nil
				g.AIReplans++
				continue
			}
			g.learnedMoveThisTurn = g.learnedMoveThisTurn || plan.LearnedUsed
			g.HardDrop()
			return true
		}
		if !moved {
			g.CurrentAIMove = nil
			g.AIReplans++
			continue
		}
		g.learnedMoveThisTurn = g.learnedMoveThisTurn || plan.LearnedUsed
		plan.Actions, plan.Expected = plan.Actions[1:], plan.Expected[1:]
		// Commit as soon as the selected landing is reachable by hard drop.
		if g.CurrentPiece.X == plan.TargetX && g.CurrentPiece.Rotation == plan.TargetRotation && g.Board.GetGhostY(g.CurrentPiece) == plan.TargetY {
			g.HardDrop()
		}
		return true
	}
	return false
}

func (g *Game) policyVersion() string {
	if g.UseLearned {
		copyGame := *g
		copyGame.UseLearned = false
		if g.LearnedModel == nil {
			return "hybrid-risk-v1-fallback-no-model/" + copyGame.policyVersion()
		}
		return "hybrid-risk-v1-" + g.LearnedModel.ID + "/" + copyGame.policyVersion()
	}
	if g.LookaheadDepth > 0 {
		version := fmt.Sprintf("beam-v2-depth%d-width%d", g.LookaheadDepth, g.BeamWidth)
		if g.ReserveWell {
			version += "-strict-well"
		}
		return version
	}
	if g.ReserveWell {
		return PolicyVersion + "-strict-well"
	}
	return PolicyVersion
}

// StepAIImmediate calculates the optimal move and drops the piece immediately.
// Designed for headless simulations, high-volume log generation, and background training.
// Returns true if a piece was placed, false if game over or no valid move.
func (g *Game) StepAIImmediate() bool {
	if g.State != StatePlaying || g.CurrentPiece == nil {
		return false
	}
	g.ExecutionMode = "headless_placement"

	plan := FindBestMove(g)
	if plan == nil {
		g.gameOver(EndNoLegalMove)
		return false
	}

	if plan.UseHold && g.CanHold {
		learnedHold := plan.LearnedUsed
		if !g.Hold() {
			return false
		}
		g.learnedMoveThisTurn = g.learnedMoveThisTurn || learnedHold
		plan = FindBestMove(g)
		if plan == nil {
			g.gameOver(EndNoLegalMove)
			return false
		}
	}

	g.CurrentPiece.Rotation = plan.TargetRotation
	g.CurrentPiece.X = plan.TargetX
	g.CurrentPiece.Y = plan.TargetY
	if !g.Board.IsValidPosition(g.CurrentPiece) {
		g.gameOver(EndNoLegalMove)
		return false
	}

	g.CurrentAIMove = plan
	// Direct placement awards drop distance from spawn, not hidden search rows.
	g.hardDrop(NewPiece(g.CurrentPiece.Type).Y)
	g.CurrentAIMove = nil
	return true
}
