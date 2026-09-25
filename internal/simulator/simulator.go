package simulator

import (
	"context"
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
	MaxMovesPerGame int    // Safety cap per game (default: 2000)
	LogPath         string // Destination JSONL path
	Clean           bool   // Clean/truncate log file before starting
	BufferSize      int    // Buffer size in bytes for writer (default: 1MB)
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
	totalMoves     atomic.Int64
	totalLines     atomic.Int64
	totalTetrises  atomic.Int64
	totalScore     atomic.Int64
	highScore      atomic.Int64

	startTime time.Time
}

// New creates a new Simulator with up to 50 workers.
func New(cfg Config) (*Simulator, error) {
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

	// Progress ticker
	doneCh := make(chan struct{})
	if onProgress != nil {
		go func() {
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
			s.worker(ctx, id)
		}(workerID)
	}

	wg.Wait()
	close(doneCh)

	return s.logger.Close()
}

func (s *Simulator) worker(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if s.config.TotalGames > 0 {
			if s.completedGames.Load() >= int64(s.config.TotalGames) {
				return
			}
		}

		// Play one complete game
		g := engine.NewGame()
		g.AutoPlay = true
		g.SessionID = fmt.Sprintf("sim-w%d-%s", workerID, time.Now().Format("20060102-150405.000000"))
		g.SetLogger(s.logger)

		moves := 0
		maxLimit := s.config.MaxMovesPerGame
		if maxLimit <= 0 {
			maxLimit = 10000
		}
		for moves < maxLimit && g.Lines < maxLimit && g.State == engine.StatePlaying {
			select {
			case <-ctx.Done():
				_ = g.Close()
				return
			default:
			}

			if !g.StepAIImmediate() {
				break
			}
			moves++
			s.totalMoves.Add(1)
		}

		_ = g.Close()

		s.completedGames.Add(1)
		s.totalLines.Add(int64(g.Lines))
		s.totalTetrises.Add(int64(g.Tetrises))
		s.totalScore.Add(int64(g.Score))

		// Update high score atomically
		for {
			curHigh := s.highScore.Load()
			if int64(g.Score) <= curHigh || s.highScore.CompareAndSwap(curHigh, int64(g.Score)) {
				break
			}
		}
	}
}

// GetProgress returns the current progress snapshot.
func (s *Simulator) GetProgress(activeWorkers int) Progress {
	elapsed := time.Since(s.startTime)
	moves := s.totalMoves.Load()

	movesPerSec := 0.0
	if elapsed.Seconds() > 0 {
		movesPerSec = float64(moves) / elapsed.Seconds()
	}

	return Progress{
		ActiveWorkers:  activeWorkers,
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
