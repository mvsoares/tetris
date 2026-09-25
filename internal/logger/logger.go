package logger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// MoveLog represents a single move played in a game session.
// Designed for offline reinforcement learning, imitation learning, and performance analysis.
type MoveLog struct {
	SessionID       string    `json:"session_id"`
	Timestamp       time.Time `json:"timestamp"`
	MoveNumber      int       `json:"move_number"`
	PlayerType      string    `json:"player_type"` // "AI" or "HUMAN"
	PieceType       string    `json:"piece_type"`
	UsedHold        bool      `json:"used_hold"`
	Rotation        int       `json:"rotation"`
	X               int       `json:"x"`
	Y               int       `json:"y"`
	LinesCleared    int       `json:"lines_cleared"`
	ScoreGained     int       `json:"score_gained"`
	TotalScore      int       `json:"total_score"`
	TotalLines      int       `json:"total_lines"`
	Level           int       `json:"level"`
	MaxHeightBefore int       `json:"max_height_before"`
	MaxHeightAfter  int       `json:"max_height_after"`
	HolesAfter      int       `json:"holes_after"`
	BumpinessAfter  int       `json:"bumpiness_after"`
	IsCleanupMode   bool      `json:"is_cleanup_mode"`
	HeuristicScore  float64   `json:"heuristic_score,omitempty"`
	// BoardStateAfter stores the 20x10 grid as 20 strings of "0" and "1"
	BoardStateAfter []string  `json:"board_state_after"`
}

// SessionEndLog represents aggregate stats at the end of a game session.
type SessionEndLog struct {
	Type              string    `json:"type"` // "SESSION_END"
	SessionID         string    `json:"session_id"`
	Timestamp         time.Time `json:"timestamp"`
	TotalMoves        int       `json:"total_moves"`
	FinalScore        int       `json:"final_score"`
	FinalLevel        int       `json:"final_level"`
	TotalLines        int       `json:"total_lines"`
	Singles           int       `json:"singles"`
	Doubles           int       `json:"doubles"`
	Triples           int       `json:"triples"`
	Tetrises          int       `json:"tetrises"`
	TetrisRatePercent float64   `json:"tetris_rate_percent"`
}

// Options provides configuration for log creation, buffering and truncation.
type Options struct {
	Truncate   bool // If true, truncates/cleans the log file upon opening
	BufferSize int  // Buffer size in bytes for the writer (default: 1MB)
	DropOnFull bool // If true, drops logs when channel is full (for UI mode)
}

// LogStats stores aggregate metrics read from a JSONL log file.
type LogStats struct {
	TotalSessions  int
	TotalMoves     int
	AIMoves        int
	HumanMoves     int
	TotalLines     int
	Singles        int
	Doubles        int
	Triples        int
	Tetrises       int
	CleanupMoves   int
	TotalHoles     int
	TotalBumpiness int
	MaxScore       int
	TetrisRate     float64
	AvgHoles       float64
	AvgBumpiness   float64
}

// Logger handles buffered, asynchronous JSONL logging.
type Logger struct {
	file       *os.File
	ch         chan any
	wg         sync.WaitGroup
	closed     bool
	mu         sync.Mutex
	dropOnFull bool
	bufferSize int
}

// CleanLog cleans (truncates to 0 bytes) the log file at the given path.
func CleanLog(filePath string) error {
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to clean log file: %w", err)
	}
	return file.Close()
}

// NewLoggerWithOptions creates a Logger with custom options (such as buffering and truncation).
func NewLoggerWithOptions(filePath string, opts Options) (*Logger, error) {
	dir := filepath.Dir(filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	flags := os.O_CREATE | os.O_WRONLY
	if opts.Truncate {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_APPEND
	}

	file, err := os.OpenFile(filePath, flags, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	bufSize := opts.BufferSize
	if bufSize <= 0 {
		bufSize = 100 * 1024 * 1024 // 100 MB default buffer for massive write throughput
	}

	l := &Logger{
		file:       file,
		ch:         make(chan any, 131072), // 128K buffer for high concurrency
		dropOnFull: opts.DropOnFull,
		bufferSize: bufSize,
	}

	l.wg.Add(1)
	go l.worker()

	return l, nil
}

// NewLogger creates or opens a JSONL log file with a 100 MB output write buffer.
func NewLogger(filePath string) (*Logger, error) {
	return NewLoggerWithOptions(filePath, Options{
		BufferSize: 100 * 1024 * 1024,
		Truncate:   false,
	})
}

// SetDropOnFull configures whether moves should be dropped if the buffer is full.
func (l *Logger) SetDropOnFull(drop bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.dropOnFull = drop
}

func (l *Logger) worker() {
	defer l.wg.Done()
	writer := bufio.NewWriterSize(l.file, l.bufferSize)
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case item, ok := <-l.ch:
			if !ok {
				_ = writer.Flush()
				return
			}
			_ = encoder.Encode(item)

			// Batch drain up to 512 pending items to maximize buffer utilization
			batchCount := 0
			for batchCount < 512 {
				select {
				case nextItem, nextOk := <-l.ch:
					if !nextOk {
						_ = writer.Flush()
						return
					}
					_ = encoder.Encode(nextItem)
					batchCount++
				default:
					break
				}
				if len(l.ch) == 0 {
					break
				}
			}

		case <-ticker.C:
			_ = writer.Flush()
		}
	}
}

// LogMove logs a single move asynchronously.
func (l *Logger) LogMove(m MoveLog) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return
	}

	if l.dropOnFull {
		select {
		case l.ch <- m:
		default:
		}
	} else {
		l.ch <- m
	}
}

// LogSessionEnd logs the session summary.
func (l *Logger) LogSessionEnd(s SessionEndLog) {
	s.Type = "SESSION_END"
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return
	}

	if l.dropOnFull {
		select {
		case l.ch <- s:
		default:
		}
	} else {
		l.ch <- s
	}
}

// Close flushes all queued logs and closes the underlying file.
func (l *Logger) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	close(l.ch)
	l.mu.Unlock()

	l.wg.Wait()
	return l.file.Close()
}

// ReadLogStats reads and aggregates metrics from a JSONL log file.
func ReadLogStats(filePath string) (LogStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return LogStats{}, fmt.Errorf("failed to open log file %s: %w", filePath, err)
	}
	defer file.Close()

	return ReadLogStatsFromReader(file)
}

// ReadLogStatsFromReader reads and aggregates metrics from an io.Reader.
func ReadLogStatsFromReader(r io.Reader) (LogStats, error) {
	scanner := bufio.NewScanner(r)
	buf := make([]byte, 256*1024)
	scanner.Buffer(buf, 2*1024*1024)

	var stats LogStats
	sessionsMap := make(map[string]bool)

	for scanner.Scan() {
		rawBytes := bytes.TrimSpace(scanner.Bytes())
		if len(rawBytes) == 0 {
			continue
		}

		var move struct {
			SessionID      string `json:"session_id"`
			MoveNumber     int    `json:"move_number"`
			PlayerType     string `json:"player_type"`
			LinesCleared   int    `json:"lines_cleared"`
			HolesAfter     int    `json:"holes_after"`
			BumpinessAfter int    `json:"bumpiness_after"`
			TotalScore     int    `json:"total_score"`
			IsCleanupMode  bool   `json:"is_cleanup_mode"`
		}
		if err := json.Unmarshal(rawBytes, &move); err == nil && move.SessionID != "" && move.MoveNumber > 0 {
			stats.TotalMoves++
			sessionsMap[move.SessionID] = true
			if move.PlayerType == "AI" {
				stats.AIMoves++
			} else {
				stats.HumanMoves++
			}

			if move.IsCleanupMode {
				stats.CleanupMoves++
			}

			stats.TotalLines += move.LinesCleared
			switch move.LinesCleared {
			case 1:
				stats.Singles++
			case 2:
				stats.Doubles++
			case 3:
				stats.Triples++
			case 4:
				stats.Tetrises++
			}

			stats.TotalHoles += move.HolesAfter
			stats.TotalBumpiness += move.BumpinessAfter

			if move.TotalScore > stats.MaxScore {
				stats.MaxScore = move.TotalScore
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return stats, fmt.Errorf("error reading log lines: %w", err)
	}

	stats.TotalSessions = len(sessionsMap)
	if stats.TotalLines > 0 {
		stats.TetrisRate = (float64(stats.Tetrises*4) / float64(stats.TotalLines)) * 100.0
	}
	if stats.TotalMoves > 0 {
		stats.AvgHoles = float64(stats.TotalHoles) / float64(stats.TotalMoves)
		stats.AvgBumpiness = float64(stats.TotalBumpiness) / float64(stats.TotalMoves)
	}

	return stats, nil
}
