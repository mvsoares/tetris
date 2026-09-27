package simulator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"tetris/internal/engine"
	"tetris/internal/logger"
)

// Config configures the background simulation runner.
type Config struct {
	Workers         int    // Number of concurrent worker games (up to 50)
	TotalGames      int    // Target games to play (0 for infinite until cancelled)
	MaxMovesPerGame int    // Safety cap per game (default: 10000)
	LogPath         string // Destination JSONL path
	Clean           bool   // Clean/truncate log file before starting
	BufferSize      int    // Buffer size in bytes for writer (default: 100MB)
	Seed            int64  // Base seed; game n uses Seed+n, independent of scheduling.
	Mode            string // "placement" (default) or "gameplay" (55ms AI and level gravity).
	ReserveWell     bool   // Keep strict emergency well penalties for comparison.
	LookaheadDepth  int
	BeamWidth       int
	LearnedModel    *engine.MoveModel
	UseLearned      bool
}

// Progress holds real-time simulation metrics.
type Progress struct {
	ActiveWorkers  int
	CompletedGames int64
	TotalMoves     int64
	TotalLines     int64
	TotalTetrises  int64
	TotalScore     int64
	HighScore      int64
	MovesPerSec    float64
	Elapsed        time.Duration
}

// Simulator manages concurrent headless Tetris matches.
type Simulator struct {
	config Config
	logger *logger.Logger

	completedGames atomic.Int64
	startedGames   atomic.Int64
	activeWorkers  atomic.Int64
	totalMoves     atomic.Int64
	totalLines     atomic.Int64
	totalTetrises  atomic.Int64
	totalScore     atomic.Int64
	highScore      atomic.Int64

	startTime time.Time
	endTime   atomic.Int64
}

// New creates a new Simulator with up to 50 workers.
func New(cfg Config) (*Simulator, error) {
	if cfg.LookaheadDepth < 0 || cfg.LookaheadDepth > 10 || cfg.BeamWidth < 0 || cfg.BeamWidth > 64 {
		return nil, fmt.Errorf("lookahead must be 0..10 and beam width 0..64")
	}
	if cfg.Mode == "" {
		cfg.Mode = "placement"
	}
	if cfg.Mode != "placement" && cfg.Mode != "gameplay" {
		return nil, fmt.Errorf("unknown simulation mode %q", cfg.Mode)
	}
	if cfg.Workers < 1 {
		cfg.Workers = 1
	}
	if cfg.Workers > 50 {
		cfg.Workers = 50
	}
	if cfg.MaxMovesPerGame <= 0 {
		cfg.MaxMovesPerGame = 10000
	}
	if cfg.LogPath == "" {
		cfg.LogPath = "logs/plays.jsonl"
	}

	bufSize := cfg.BufferSize
	if bufSize <= 0 {
		bufSize = 100 * 1024 * 1024 // 100 MB buffer
	}

	l, err := logger.NewLoggerWithOptions(cfg.LogPath, logger.Options{
		Truncate:   cfg.Clean,
		BufferSize: bufSize,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create simulation logger: %w", err)
	}

	return &Simulator{
		config: cfg,
		logger: l,
	}, nil
}

// Run starts the background workers and blocks until target games are reached or ctx is cancelled.
// onProgress is an optional callback called every interval (e.g. 500ms).
func (s *Simulator) Run(ctx context.Context, onProgress func(Progress)) error {
	s.startTime = time.Now()
	var wg sync.WaitGroup
	var progressWG sync.WaitGroup

	// Progress ticker
	doneCh := make(chan struct{})
	if onProgress != nil {
		progressWG.Add(1)
		go func() {
			defer progressWG.Done()
			ticker := time.NewTicker(500 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-doneCh:
					onProgress(s.GetProgress(s.config.Workers))
					return
				case <-ticker.C:
					onProgress(s.GetProgress(s.config.Workers))
				}
			}
		}()
	}

	for workerID := 0; workerID < s.config.Workers; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s.activeWorkers.Add(1)
			defer s.activeWorkers.Add(-1)
			s.worker(ctx, id)
		}(workerID)
	}

	wg.Wait()
	close(doneCh)
	progressWG.Wait()
	err := s.logger.Close()
	s.endTime.Store(time.Now().UnixNano())
	return errors.Join(ctx.Err(), err)
}

func (s *Simulator) worker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		gameIndex := s.startedGames.Add(1) - 1
		if s.config.TotalGames > 0 && gameIndex >= int64(s.config.TotalGames) {
			return
		}

		// Play one complete game
		g := engine.NewGameWithSeed(s.config.Seed + gameIndex)
		g.AutoPlay = true
		g.ReserveWell = s.config.ReserveWell
		g.ConfigureLookahead(s.config.LookaheadDepth, s.config.BeamWidth)
		g.LearnedModel, g.UseLearned = s.config.LearnedModel, s.config.UseLearned
		g.ExecutionMode = "headless_placement"
		if s.config.Mode == "gameplay" {
			g.ExecutionMode = "timed_gameplay"
		}
		g.SessionID = fmt.Sprintf("sim-%d-game%d", s.startTime.UnixNano(), gameIndex)
		g.SetLogger(s.logger)

		moves := 0
		nextAI, nextGravity := engine.AIActionInterval, g.TickInterval()
		maxLimit := s.config.MaxMovesPerGame
		if maxLimit <= 0 {
			maxLimit = 10000
		}
		for moves < maxLimit && g.State == engine.StatePlaying {
			select {
			case <-ctx.Done():
				_ = g.CloseWithReason(engine.EndCancelled)
				return
			default:
			}

			movesBefore, linesBefore, tetrisesBefore, scoreBefore := g.MoveCount, g.Lines, g.Tetrises, g.Score
			if s.config.Mode == "gameplay" {
				// Advance virtual time without sleeping, using the same movement
				// and gravity methods as the terminal UI.
				if nextAI <= nextGravity {
					g.SetAIGravityRemaining(nextGravity - nextAI)
					g.StepAI()
					nextAI += engine.AIActionInterval
				} else {
					g.Tick()
					nextGravity += g.TickInterval()
				}
			} else {
				g.StepAIImmediate()
			}
			placed := g.MoveCount - movesBefore
			moves += placed
			s.totalMoves.Add(int64(placed))
			s.totalLines.Add(int64(g.Lines - linesBefore))
			s.totalTetrises.Add(int64(g.Tetrises - tetrisesBefore))
			s.totalScore.Add(int64(g.Score - scoreBefore))
			for {
				curHigh := s.highScore.Load()
				if int64(g.Score) <= curHigh || s.highScore.CompareAndSwap(curHigh, int64(g.Score)) {
					break
				}
			}
		}

		_ = g.CloseWithReason(engine.EndMoveLimit)

		s.completedGames.Add(1)

	}
}

// GetProgress returns the current progress snapshot.
func (s *Simulator) GetProgress(_ int) Progress {
	elapsed := time.Since(s.startTime)
	if end := s.endTime.Load(); end != 0 {
		elapsed = time.Unix(0, end).Sub(s.startTime)
	}
	moves := s.totalMoves.Load()

	movesPerSec := 0.0
	if elapsed.Seconds() > 0 {
		movesPerSec = float64(moves) / elapsed.Seconds()
	}

	return Progress{
		ActiveWorkers:  int(s.activeWorkers.Load()),
		CompletedGames: s.completedGames.Load(),
		TotalMoves:     moves,
		TotalLines:     s.totalLines.Load(),
		TotalTetrises:  s.totalTetrises.Load(),
		TotalScore:     s.totalScore.Load(),
		HighScore:      s.highScore.Load(),
		MovesPerSec:    movesPerSec,
		Elapsed:        elapsed,
	}
}
