package logger

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoggerWritesJSONL(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_plays.jsonl")

	l, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	move := MoveLog{
		SessionID:       "test-session-1",
		Timestamp:       time.Now(),
		MoveNumber:      1,
		PlayerType:      "AI",
		PieceType:       "I",
		UsedHold:        false,
		Rotation:        1,
		X:               9,
		Y:               16,
		LinesCleared:    4,
		ScoreGained:     800,
		TotalScore:      800,
		TotalLines:      4,
		Level:           1,
		MaxHeightBefore: 4,
		MaxHeightAfter:  0,
		BoardStateAfter: []string{"0000000000"},
	}

	l.LogMove(move)

	sessionEnd := SessionEndLog{
		SessionID:         "test-session-1",
		Timestamp:         time.Now(),
		TotalMoves:        1,
		FinalScore:        800,
		FinalLevel:        1,
		TotalLines:        4,
		Tetrises:          1,
		TetrisRatePercent: 100.0,
	}

	l.LogSessionEnd(sessionEnd)

	if err := l.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	// Verify file contents
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("failed to open log file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
		var raw map[string]interface{}
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			t.Errorf("line %d is not valid JSON: %v", lines, err)
		}
	}

	if lines != 2 {
		t.Errorf("expected 2 log lines, got %d", lines)
	}
}

func TestReadLogStats(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "stats_plays.jsonl")

	l, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Log 3 moves: 1 Tetris (4 lines), 1 Double (2 lines), 1 Single (1 line)
	l.LogMove(MoveLog{
		SessionID:    "session-A",
		MoveNumber:   1,
		PlayerType:   "AI",
		LinesCleared: 4,
		TotalScore:   800,
		HolesAfter:   1,
	})
	l.LogMove(MoveLog{
		SessionID:    "session-A",
		MoveNumber:   2,
		PlayerType:   "AI",
		LinesCleared: 2,
		TotalScore:   1100,
		HolesAfter:   0,
	})
	l.LogMove(MoveLog{
		SessionID:    "session-B",
		MoveNumber:   1,
		PlayerType:   "HUMAN",
		LinesCleared: 1,
		TotalScore:   100,
		HolesAfter:   2,
	})

	if err := l.Close(); err != nil {
		t.Fatalf("failed to close logger: %v", err)
	}

	stats, err := ReadLogStats(logPath)
	if err != nil {
		t.Fatalf("ReadLogStats failed: %v", err)
	}

	if stats.TotalSessions != 2 {
		t.Errorf("expected 2 sessions, got %d", stats.TotalSessions)
	}
	if stats.TotalMoves != 3 {
		t.Errorf("expected 3 moves, got %d", stats.TotalMoves)
	}
	if stats.AIMoves != 2 {
		t.Errorf("expected 2 AI moves, got %d", stats.AIMoves)
	}
	if stats.HumanMoves != 1 {
		t.Errorf("expected 1 Human move, got %d", stats.HumanMoves)
	}
	if stats.Tetrises != 1 {
		t.Errorf("expected 1 Tetris, got %d", stats.Tetrises)
	}
	if stats.TotalLines != 7 {
		t.Errorf("expected 7 total lines, got %d", stats.TotalLines)
	}
	expectedTetrisRate := (4.0 / 7.0) * 100.0
	if stats.TetrisRate < expectedTetrisRate-0.01 || stats.TetrisRate > expectedTetrisRate+0.01 {
		t.Errorf("expected TetrisRate ~%.2f%%, got %.2f%%", expectedTetrisRate, stats.TetrisRate)
	}
	if stats.MaxScore != 1100 {
		t.Errorf("expected MaxScore 1100, got %d", stats.MaxScore)
	}
}

func TestCleanLog(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "clean_test.jsonl")

	l, err := NewLogger(logPath)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	l.LogMove(MoveLog{SessionID: "s1", MoveNumber: 1})
	_ = l.Close()

	info, err := os.Stat(logPath)
	if err != nil || info.Size() == 0 {
		t.Fatalf("file should have content before clean")
	}

	if err := CleanLog(logPath); err != nil {
		t.Fatalf("CleanLog failed: %v", err)
	}

	infoAfter, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("failed to stat file after clean: %v", err)
	}
	if infoAfter.Size() != 0 {
		t.Errorf("expected file size 0 after clean, got %d", infoAfter.Size())
	}
}

func TestBufferedLoggerHighThroughput(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "throughput.jsonl")

	// 1 MB buffer
	l, err := NewLoggerWithOptions(logPath, Options{
		BufferSize: 1024 * 1024,
		Truncate:   true,
	})
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	start := time.Now()
	totalMoves := 5000
	for i := 1; i <= totalMoves; i++ {
		l.LogMove(MoveLog{
			SessionID:    fmt.Sprintf("sess-%d", i%10),
			MoveNumber:   i,
			PlayerType:   "AI",
			LinesCleared: i % 5,
			TotalScore:   i * 100,
		})
	}
	if err := l.Close(); err != nil {
		t.Fatalf("failed to close: %v", err)
	}
	duration := time.Since(start)

	stats, err := ReadLogStats(logPath)
	if err != nil {
		t.Fatalf("failed to read log stats: %v", err)
	}
	if stats.TotalMoves != totalMoves {
		t.Errorf("expected %d moves, got %d", totalMoves, stats.TotalMoves)
	}

	t.Logf("Wrote and flushed %d moves in %v (%.0f moves/sec)", totalMoves, duration, float64(totalMoves)/duration.Seconds())
}
