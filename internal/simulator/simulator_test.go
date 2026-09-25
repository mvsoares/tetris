package simulator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tetris/internal/logger"
)

func TestSimulatorRunsConcurrently(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "concurrent_plays.jsonl")

	cfg := Config{
		Workers:         4,
		TotalGames:      4,
		MaxMovesPerGame: 50,
		LogPath:         logPath,
	}

	sim, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create simulator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = sim.Run(ctx, nil)
	if err != nil {
		t.Fatalf("simulator Run failed: %v", err)
	}

	prog := sim.GetProgress(0)
	if prog.CompletedGames < 4 {
		t.Errorf("expected at least 4 completed games, got %d", prog.CompletedGames)
	}
	if prog.TotalMoves == 0 {
		t.Errorf("expected moves to be recorded")
	}

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("log file should exist: %v", err)
	}
	if info.Size() == 0 {
		t.Errorf("log file should not be empty")
	}
}

func TestSimulatorWorkersClamping(t *testing.T) {
	sim, err := New(Config{Workers: 100})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if sim.config.Workers != 50 {
		t.Errorf("expected workers to be clamped to 50, got %d", sim.config.Workers)
	}

	simZero, err := New(Config{Workers: 0})
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	if simZero.config.Workers != 1 {
		t.Errorf("expected workers to be clamped to 1, got %d", simZero.config.Workers)
	}
}

func TestSimulatorCleanAndReadLogs50Games(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "sim_50_games.jsonl")

	// Pre-create file with dirty data to ensure Clean: true truncates it
	if err := os.WriteFile(logPath, []byte("dirty old log data\n"), 0644); err != nil {
		t.Fatalf("failed to write dirty data: %v", err)
	}

	cfg := Config{
		Workers:         10,
		TotalGames:      50,
		MaxMovesPerGame: 50, // Capped per game for fast unit test execution
		LogPath:         logPath,
		Clean:           true,
		BufferSize:      1024 * 1024,
	}

	sim, err := New(cfg)
	if err != nil {
		t.Fatalf("failed to create simulator: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	err = sim.Run(ctx, nil)
	if err != nil {
		t.Fatalf("sim.Run failed: %v", err)
	}

	prog := sim.GetProgress(0)
	if prog.CompletedGames < 50 {
		t.Errorf("expected 50 completed games, got %d", prog.CompletedGames)
	}

	// Read and verify logs using ReadLogStats
	stats, err := logger.ReadLogStats(logPath)
	if err != nil {
		t.Fatalf("ReadLogStats failed: %v", err)
	}

	if stats.TotalSessions < 50 {
		t.Errorf("expected at least 50 logged sessions, got %d", stats.TotalSessions)
	}
	if stats.TotalMoves == 0 {
		t.Errorf("expected moves to be recorded in log")
	}
	t.Logf("Read stats from 50 games: %d moves, %d lines cleared, %d tetrises, rate: %.1f%%",
		stats.TotalMoves, stats.TotalLines, stats.Tetrises, stats.TetrisRate)
}
