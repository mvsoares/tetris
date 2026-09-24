package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
)

type MoveRecord struct {
	SessionID       string   `json:"session_id"`
	MoveNumber      int      `json:"move_number"`
	PlayerType      string   `json:"player_type"`
	PieceType       string   `json:"piece_type"`
	LinesCleared    int      `json:"lines_cleared"`
	ScoreGained     int      `json:"score_gained"`
	TotalScore      int      `json:"total_score"`
	TotalLines      int      `json:"total_lines"`
	Level           int      `json:"level"`
	MaxHeightBefore int      `json:"max_height_before"`
	MaxHeightAfter  int      `json:"max_height_after"`
	HolesAfter      int      `json:"holes_after"`
	BumpinessAfter  int      `json:"bumpiness_after"`
	IsCleanupMode   bool     `json:"is_cleanup_mode"`
	BoardStateAfter []string `json:"board_state_after"`
}

type SessionEndRecord struct {
	Type              string  `json:"type"`
	SessionID         string  `json:"session_id"`
	TotalMoves        int     `json:"total_moves"`
	FinalScore        int     `json:"final_score"`
	FinalLevel        int     `json:"final_level"`
	TotalLines        int     `json:"total_lines"`
	Tetrises          int     `json:"tetrises"`
	TetrisRatePercent float64 `json:"tetris_rate_percent"`
}

func main() {
	logFile := flag.String("file", "logs/plays.jsonl", "Log file path")
	flag.Parse()

	f, err := os.Open(*logFile)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	lastMoveBySession := make(map[string]MoveRecord)
	sessionEnds := make(map[string]SessionEndRecord)
	totalMovesCount := 0

	for scanner.Scan() {
		raw := bytes.TrimSpace(scanner.Bytes())
		if len(raw) == 0 {
			continue
		}

		if bytes.Contains(raw, []byte(`"type":"SESSION_END"`)) {
			var se SessionEndRecord
			if err := json.Unmarshal(raw, &se); err == nil {
				sessionEnds[se.SessionID] = se
			}
			continue
		}

		var m MoveRecord
		if err := json.Unmarshal(raw, &m); err == nil && m.SessionID != "" {
			totalMovesCount++
			prev, exists := lastMoveBySession[m.SessionID]
			if !exists || m.MoveNumber > prev.MoveNumber {
				lastMoveBySession[m.SessionID] = m
			}
		}
	}

	fmt.Printf("Total logged moves: %d\n", totalMovesCount)
	fmt.Printf("Total unique sessions: %d\n", len(lastMoveBySession))
	fmt.Printf("Sessions with SESSION_END: %d\n\n", len(sessionEnds))

	// Analyze terminal states (losses)
	pieceLossCount := make(map[string]int)
	inCleanupAtDeath := 0
	notInCleanupAtDeath := 0
	totalDeathHeight := 0
	totalDeathHoles := 0
	totalDeathBumpiness := 0
	totalMovesAtDeath := 0

	var moveLengths []int
	var linesList []int

	colHeightsSum := make([]int, 10)
	heightDist := make(map[int]int)

	for sessID, lastM := range lastMoveBySession {
		pieceLossCount[lastM.PieceType]++
		if lastM.IsCleanupMode {
			inCleanupAtDeath++
		} else {
			notInCleanupAtDeath++
		}
		totalDeathHeight += lastM.MaxHeightAfter
		totalDeathHoles += lastM.HolesAfter
		totalDeathBumpiness += lastM.BumpinessAfter
		totalMovesAtDeath += lastM.MoveNumber
		moveLengths = append(moveLengths, lastM.MoveNumber)
		linesList = append(linesList, lastM.TotalLines)
		heightDist[lastM.MaxHeightAfter]++

		// Inspect BoardStateAfter
		if len(lastM.BoardStateAfter) == 20 {
			for x := 0; x < 10; x++ {
				h := 0
				for y := 0; y < 20; y++ {
					if lastM.BoardStateAfter[y][x] != '0' {
						h = 20 - y
						break
					}
				}
				colHeightsSum[x] += h
			}
		}
		_ = sessID
	}

	n := len(lastMoveBySession)
	if n == 0 {
		fmt.Println("No sessions found.")
		return
	}

	sort.Ints(moveLengths)
	sort.Ints(linesList)

	fmt.Println("================================================================")
	fmt.Println("           🔍 DIAGNÓSTICO PROFUNDO DAS DERROTAS (GAME OVER)     ")
	fmt.Println("================================================================")
	fmt.Printf("Partidas Analisadas:         %d\n", n)
	fmt.Printf("Média de Jogadas / Partida:  %.1f (Min: %d | Mediana: %d | Max: %d)\n",
		float64(totalMovesAtDeath)/float64(n), moveLengths[0], moveLengths[n/2], moveLengths[n-1])
	fmt.Printf("Média de Linhas / Partida:   %.1f (Min: %d | Mediana: %d | Max: %d)\n",
		float64(linesList[len(linesList)-1])/float64(n), linesList[0], linesList[n/2], linesList[n-1])
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("ESTADO DO TABULEIRO NA ÚLTIMA JOGADA ANTES DA MORTE:")
	fmt.Printf("   📏 Altura Máxima Média:      %.2f / 20 linhas\n", float64(totalDeathHeight)/float64(n))
	fmt.Printf("   🕳️  Buracos Médios na Morte:  %.2f buracos (vs 0.19 na média global!)\n", float64(totalDeathHoles)/float64(n))
	fmt.Printf("   〰️  Irregularidade (Bumpiness): %.2f (vs 9.90 na média global!)\n", float64(totalDeathBumpiness)/float64(n))
	fmt.Println("   📊 Média de Altura por Coluna (0 a 9):")
	for x := 0; x < 10; x++ {
		fmt.Printf("      Col %d: %.2f linhas\n", x, float64(colHeightsSum[x])/float64(n))
	}
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("DISTRIBUIÇÃO DA ALTURA MÁXIMA NA DERROTA:")
	var heights []int
	for h := range heightDist {
		heights = append(heights, h)
	}
	sort.Ints(heights)
	for _, h := range heights {
		cnt := heightDist[h]
		fmt.Printf("   Altura %2d: %3d partidas (%.1f%%)\n", h, cnt, (float64(cnt)/float64(n))*100.0)
	}
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("STATUS DO CLEANUP MODE NO MOMENTO DA DERROTA:")
	fmt.Printf("   🚨 Estava em Cleanup Mode (>= 65%%): %d partidas (%.1f%%)\n",
		inCleanupAtDeath, (float64(inCleanupAtDeath)/float64(n))*100.0)
	fmt.Printf("   🟢 Estava em Modo Normal (< 65%%):   %d partidas (%.1f%%)\n",
		notInCleanupAtDeath, (float64(notInCleanupAtDeath)/float64(n))*100.0)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("PEÇAS COLOCADAS NA JOGADA FINAL (TOP-OUT):")
	type pieceStat struct {
		piece string
		count int
	}
	var pStats []pieceStat
	for p, cnt := range pieceLossCount {
		pStats = append(pStats, pieceStat{p, cnt})
	}
	sort.Slice(pStats, func(i, j int) bool {
		return pStats[i].count > pStats[j].count
	})
	for _, ps := range pStats {
		pct := (float64(ps.count) / float64(n)) * 100.0
		fmt.Printf("   Peça %-2s: %4d vezes (%5.1f%%)\n", ps.piece, ps.count, pct)
	}
	fmt.Println("================================================================")
}
