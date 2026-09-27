package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

	"tetris/internal/engine"
	"tetris/internal/logger"
)

type distribution struct {
	Mean   float64 `json:"mean"`
	Min    int     `json:"min"`
	P10    int     `json:"p10"`
	Median int     `json:"median"`
	P90    int     `json:"p90"`
	Max    int     `json:"max"`
}

type lossReport struct {
	Sessions                 int            `json:"sessions"`
	FinishedSessions         int            `json:"finished_sessions"`
	UnknownSessions          int            `json:"unknown_sessions"`
	Losses                   int            `json:"losses"`
	EndReasons               map[string]int `json:"end_reasons"`
	TotalMoves               int            `json:"total_moves"`
	TotalLines               int            `json:"total_lines"`
	TetrisLinePercent        float64        `json:"tetris_line_percent"`
	MovesPerSession          distribution   `json:"moves_per_session"`
	LinesPerSession          distribution   `json:"lines_per_session"`
	FailedPieces             map[string]int `json:"failed_pieces"`
	PlacedPieces             map[string]int `json:"placed_pieces"`
	LossesInCleanup          int            `json:"losses_in_cleanup"`
	MeanDeathHeight          float64        `json:"mean_death_height"`
	MeanDeathHoles           float64        `json:"mean_death_holes"`
	MeanDeathBumpiness       float64        `json:"mean_death_bumpiness"`
	MeanDeathColumnHeights   [10]float64    `json:"mean_death_column_heights"`
	DeathHeightDistribution  map[int]int    `json:"death_height_distribution"`
	MeanHoles                float64        `json:"mean_holes_after_move"`
	MeanBumpiness            float64        `json:"mean_bumpiness_after_move"`
	CleanupMoves             int            `json:"cleanup_moves"`
	HoleCreatingMoves        int            `json:"hole_creating_moves"`
	HoleCreatingPieces       map[string]int `json:"hole_creating_pieces"`
	PlannedMoves             int            `json:"planned_moves"`
	PlanTelemetryMoves       int            `json:"plan_telemetry_moves"`
	PlanMisses               int            `json:"plan_misses"`
	WatchdogDrops            int            `json:"watchdog_drops"`
	AIReplans                int            `json:"ai_replans"`
	SessionsWithPlanMisses   int            `json:"sessions_with_plan_misses"`
	LossesWithPlanMisses     int            `json:"losses_with_plan_misses"`
	LossesWithRecentPlanMiss int            `json:"losses_with_plan_miss_in_last_10_moves"`
}

func describe(values []int) distribution {
	if len(values) == 0 {
		return distribution{}
	}
	sort.Ints(values)
	sum := 0
	for _, v := range values {
		sum += v
	}
	n := len(values)
	return distribution{Mean: float64(sum) / float64(n), Min: values[0], P10: values[(n-1)/10], Median: values[(n-1)/2], P90: values[(n-1)*9/10], Max: values[n-1]}
}

// Legacy and incomplete sessions are unknown, not assumed defeats.
func analyzeLosses(r io.Reader) (lossReport, error) {
	report := lossReport{EndReasons: map[string]int{}, FailedPieces: map[string]int{}, PlacedPieces: map[string]int{}, DeathHeightDistribution: map[int]int{}, HoleCreatingPieces: map[string]int{}}
	sessions := map[string]bool{}
	ends := map[string]logger.SessionEndLog{}
	type observedTotals struct {
		moves, lines, score int
		recentPlanMisses    [10]bool
	}
	observed := map[string]observedTotals{}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	totalHoles, totalBumpiness, tetrises, line := 0, 0, 0, 0
	for scanner.Scan() {
		line++
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}
		var header struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
			HadAIPlan *bool  `json:"had_ai_plan"`
		}
		if err := json.Unmarshal(raw, &header); err != nil {
			return report, fmt.Errorf("line %d: %w", line, err)
		}
		if header.SessionID == "" {
			return report, fmt.Errorf("line %d: missing session_id", line)
		}
		sessions[header.SessionID] = true
		if header.Type == "SESSION_END" {
			var end logger.SessionEndLog
			if err := json.Unmarshal(raw, &end); err != nil {
				return report, fmt.Errorf("line %d: %w", line, err)
			}
			if _, exists := ends[end.SessionID]; exists {
				return report, fmt.Errorf("line %d: duplicate session end", line)
			}
			ends[end.SessionID] = end
			continue
		}
		if header.Type != "" {
			return report, fmt.Errorf("line %d: unknown record type %q", line, header.Type)
		}
		var move logger.MoveLog
		if err := json.Unmarshal(raw, &move); err != nil {
			return report, fmt.Errorf("line %d: %w", line, err)
		}
		if move.MoveNumber <= 0 {
			return report, fmt.Errorf("line %d: invalid move number", line)
		}
		prev := observed[move.SessionID]
		if move.MoveNumber != prev.moves+1 {
			return report, fmt.Errorf("line %d: missing, duplicate or unordered move", line)
		}
		if move.BoardStateBefore != nil && move.ScoreGained != move.TotalScore-prev.score {
			return report, fmt.Errorf("line %d: reward does not match score delta", line)
		}
		prev.recentPlanMisses[prev.moves%10] = move.HadAIPlan && !move.AIPlanMatched
		prev.moves++
		prev.lines += move.LinesCleared
		prev.score = move.TotalScore
		observed[move.SessionID] = prev
		report.TotalMoves++
		if header.HadAIPlan != nil {
			report.PlanTelemetryMoves++
		}
		if move.HadAIPlan {
			report.PlannedMoves++
		}
		report.TotalLines += move.LinesCleared
		report.PlacedPieces[move.PieceType]++
		totalHoles += move.HolesAfter
		totalBumpiness += move.BumpinessAfter
		if move.LinesCleared == 4 {
			tetrises++
		}
		if move.IsCleanupMode {
			report.CleanupMoves++
		}
		if move.BoardStateBefore != nil && move.HolesAfter > move.HolesBefore {
			report.HoleCreatingMoves++
			report.HoleCreatingPieces[move.PieceType]++
		}
	}
	if err := scanner.Err(); err != nil {
		return report, err
	}
	report.Sessions, report.FinishedSessions = len(sessions), len(ends)
	moves, lines := []int{}, []int{}
	for id := range sessions {
		end, ok := ends[id]
		if !ok || end.EndReason == "" {
			report.UnknownSessions++
			report.EndReasons["unknown"]++
			continue
		}
		report.EndReasons[end.EndReason]++
		report.PlanMisses += end.PlanMisses
		report.WatchdogDrops += end.WatchdogDrops
		report.AIReplans += end.AIReplans
		if end.PlanMisses > 0 {
			report.SessionsWithPlanMisses++
		}
		totals := observed[id]
		if totals.moves != end.TotalMoves || totals.lines != end.TotalLines {
			return report, fmt.Errorf("session %s: summary disagrees with recorded moves or lines", id)
		}
		moves = append(moves, end.TotalMoves)
		lines = append(lines, end.TotalLines)
		if end.EndReason != engine.EndTopOut && end.EndReason != engine.EndNoLegalMove {
			continue
		}
		if end.FailedPiece == "" {
			return report, fmt.Errorf("session %s: loss has no failed piece", id)
		}
		if len(end.BoardState) != engine.BoardHeight {
			return report, fmt.Errorf("session %s: missing terminal board", id)
		}
		for _, row := range end.BoardState {
			if len(row) != engine.BoardWidth {
				return report, fmt.Errorf("session %s: invalid terminal row width", id)
			}
			for _, c := range row {
				if c != '0' && c != '1' {
					return report, fmt.Errorf("session %s: invalid board cell", id)
				}
			}
		}
		report.Losses++
		if end.PlanMisses > 0 {
			report.LossesWithPlanMisses++
		}
		for _, missed := range totals.recentPlanMisses {
			if missed {
				report.LossesWithRecentPlanMiss++
				break
			}
		}
		report.FailedPieces[end.FailedPiece]++
		if end.IsCleanupMode {
			report.LossesInCleanup++
		}
		report.MeanDeathHeight += float64(end.MaxHeight)
		report.MeanDeathHoles += float64(end.Holes)
		report.MeanDeathBumpiness += float64(end.Bumpiness)
		report.DeathHeightDistribution[end.MaxHeight]++
		for x := 0; x < engine.BoardWidth; x++ {
			for y := 0; y < engine.BoardHeight; y++ {
				if end.BoardState[y][x] == '1' {
					report.MeanDeathColumnHeights[x] += float64(engine.BoardHeight - y)
					break
				}
			}
		}
	}
	report.MovesPerSession, report.LinesPerSession = describe(moves), describe(lines)
	if report.TotalLines > 0 {
		report.TetrisLinePercent = float64(tetrises*4) * 100 / float64(report.TotalLines)
	}
	if report.TotalMoves > 0 {
		report.MeanHoles = float64(totalHoles) / float64(report.TotalMoves)
		report.MeanBumpiness = float64(totalBumpiness) / float64(report.TotalMoves)
	}
	if report.Losses > 0 {
		n := float64(report.Losses)
		report.MeanDeathHeight /= n
		report.MeanDeathHoles /= n
		report.MeanDeathBumpiness /= n
		for x := range report.MeanDeathColumnHeights {
			report.MeanDeathColumnHeights[x] /= n
		}
	}
	return report, nil
}

func main() {
	path := flag.String("file", "logs/plays.jsonl", "Log file path")
	asJSON := flag.Bool("json", false, "Output measured statistics as JSON")
	flag.Parse()
	f, err := os.Open(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	r, err := analyzeLosses(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(r); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	fmt.Printf("Sessions: %d | Finished: %d | Unknown: %d | Confirmed losses: %d\n", r.Sessions, r.FinishedSessions, r.UnknownSessions, r.Losses)
	fmt.Printf("End reasons: %v\n", r.EndReasons)
	fmt.Printf("Moves: %d | Lines: %d | Tetris share of lines: %.2f%%\n", r.TotalMoves, r.TotalLines, r.TetrisLinePercent)
	fmt.Printf("Moves/session: mean %.1f, median %d, p10 %d, p90 %d, min %d, max %d\n", r.MovesPerSession.Mean, r.MovesPerSession.Median, r.MovesPerSession.P10, r.MovesPerSession.P90, r.MovesPerSession.Min, r.MovesPerSession.Max)
	fmt.Printf("Lines/session: mean %.1f, median %d, p10 %d, p90 %d\n", r.LinesPerSession.Mean, r.LinesPerSession.Median, r.LinesPerSession.P10, r.LinesPerSession.P90)
	fmt.Printf("Holes after move: %.2f | Bumpiness after move: %.2f | Cleanup moves: %d\n", r.MeanHoles, r.MeanBumpiness, r.CleanupMoves)
	fmt.Printf("Hole-creating moves: %d | By placed piece: %v\n", r.HoleCreatingMoves, r.HoleCreatingPieces)
	fmt.Printf("Planned moves: %d | Missed plans: %d | Watchdog drops: %d\n", r.PlannedMoves, r.PlanMisses, r.WatchdogDrops)
	if r.Losses == 0 {
		fmt.Println("No confirmed losses; capped and incomplete games are excluded.")
		return
	}
	fmt.Printf("At loss: height %.2f, holes %.2f, bumpiness %.2f | Cleanup: %d/%d\n", r.MeanDeathHeight, r.MeanDeathHoles, r.MeanDeathBumpiness, r.LossesInCleanup, r.Losses)
	fmt.Printf("Failed pieces (not last placed pieces): %v\n", r.FailedPieces)
	fmt.Printf("Mean terminal column heights: %.2f\n", r.MeanDeathColumnHeights)
	fmt.Println("Failure counts describe association with terminal states, not causal piece risk.")
}
