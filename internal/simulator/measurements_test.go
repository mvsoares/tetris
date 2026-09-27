package simulator

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"tetris/internal/engine"
	"tetris/internal/logger"
)

func simulateSeeds(t *testing.T, workers, games int, mode string, lookahead ...int) map[int64]logger.SessionEndLog {
	t.Helper()
	path := filepath.Join(t.TempDir(), "seeds.jsonl")
	depth := 0
	if len(lookahead) > 0 {
		depth = lookahead[0]
	}
	s, err := New(Config{Workers: workers, TotalGames: games, MaxMovesPerGame: 20, Seed: 42, LogPath: path, BufferSize: 1024, Mode: mode, LookaheadDepth: depth, BeamWidth: 2})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Run(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	p := s.GetProgress(99)
	if p.CompletedGames != int64(games) || p.ActiveWorkers != 0 || p.TotalMoves != int64(games*20) {
		t.Fatalf("wrong progress: %+v", p)
	}
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	ends := map[int64]logger.SessionEndLog{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var end logger.SessionEndLog
		if err := json.Unmarshal(scanner.Bytes(), &end); err != nil {
			t.Fatal(err)
		}
		if end.Type != "SESSION_END" {
			continue
		}
		executionMode := "headless_placement"
		if mode == "gameplay" {
			executionMode = "timed_gameplay"
		}
		if end.EndReason != engine.EndMoveLimit || end.ExecutionMode != executionMode || end.PolicyVersion == "" {
			t.Fatalf("bad metadata: %+v", end)
		}
		end.SessionID, end.Timestamp = "", time.Time{}
		ends[end.Seed] = end
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	if len(ends) != games {
		t.Fatalf("expected %d ends, got %d", games, len(ends))
	}
	return ends
}

func TestSimulationSeedsDoNotDependOnWorkers(t *testing.T) {
	a, b := simulateSeeds(t, 1, 4, "placement"), simulateSeeds(t, 8, 4, "placement")
	if !reflect.DeepEqual(a, b) {
		t.Fatal("worker scheduling changed seeded game outcomes")
	}
	_ = simulateSeeds(t, 8, 1, "placement")
}

func TestGameplaySeedsDoNotDependOnWorkers(t *testing.T) {
	a, b := simulateSeeds(t, 1, 4, "gameplay"), simulateSeeds(t, 8, 4, "gameplay")
	if !reflect.DeepEqual(a, b) {
		t.Fatal("worker scheduling changed timed gameplay outcomes")
	}
}

func TestBeamGameplaySeedsDoNotDependOnWorkers(t *testing.T) {
	a, b := simulateSeeds(t, 1, 2, "gameplay", 10), simulateSeeds(t, 4, 2, "gameplay", 10)
	if !reflect.DeepEqual(a, b) {
		t.Fatal("worker scheduling changed beam search outcomes")
	}
}

func TestInvalidSearchConfig(t *testing.T) {
	for _, cfg := range []Config{{LookaheadDepth: -1}, {LookaheadDepth: 11}, {BeamWidth: -1}, {BeamWidth: 65}} {
		if _, err := New(cfg); err == nil {
			t.Fatalf("invalid config accepted: %+v", cfg)
		}
	}
}

func TestInvalidSimulationMode(t *testing.T) {
	if _, err := New(Config{Mode: "invalid"}); err == nil {
		t.Fatal("invalid mode accepted")
	}
}

func TestRunWaitsForFinalProgressCallback(t *testing.T) {
	s, err := New(Config{Workers: 1, TotalGames: 1, MaxMovesPerGame: 1, LogPath: filepath.Join(t.TempDir(), "progress.jsonl"), BufferSize: 1024})
	if err != nil {
		t.Fatal(err)
	}
	entered, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- s.Run(context.Background(), func(p Progress) {
			if p.ActiveWorkers == 0 {
				close(entered)
				<-release
			}
		})
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("no final progress callback")
	}
	select {
	case <-done:
		t.Fatal("Run returned while callback was still active")
	default:
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestRunReturnsCancellation(t *testing.T) {
	s, err := New(Config{Workers: 2, TotalGames: 1, LogPath: filepath.Join(t.TempDir(), "cancelled.jsonl"), BufferSize: 1024})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := s.Run(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if p := s.GetProgress(0); p.CompletedGames != 0 || p.TotalMoves != 0 {
		t.Fatalf("cancelled run played games: %+v", p)
	}
}
