package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tetris/internal/engine"
	"tetris/internal/simulator"
	"tetris/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	trainMode := flag.Bool("train", false, "Executa simulações em background para gerar logs")
	workers := flag.Int("workers", 0, "Número de partidas simultâneas em background (1 a 50)")
	games := flag.Int("games", 100, "Total de partidas no modo de treino em background")
	logFile := flag.String("file", "logs/plays.jsonl", "Caminho do arquivo de logs (.jsonl)")
	lookahead := flag.Int("lookahead", 0, "Experimental lookahead: 0 (v2) or 1..10 placements")
	beamWidth := flag.Int("beam-width", 4, "Experimental beam width (1..64)")
	learned := flag.Bool("learned", false, "Ativar IA híbrida experimental")
	modelFile := flag.String("model", "models/move-risk.json", "Modelo carregado uma vez na inicialização")
	flag.Parse()
	var model *engine.MoveModel
	if loaded, err := engine.LoadMoveModel(*modelFile); err == nil {
		model = loaded
	} else if *learned || !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Modelo indisponível; usando fallback: %v\n", err)
	}
	if *lookahead < 0 || *lookahead > 10 || *beamWidth < 1 || *beamWidth > 64 {
		fmt.Fprintln(os.Stderr, "lookahead must be 0..10 and beam-width 1..64")
		os.Exit(1)
	}

	if *trainMode || *workers > 0 {
		w := *workers
		if w <= 0 {
			w = 50
		}
		if w > 50 {
			w = 50
		}
		runBackgroundTrain(w, *games, *logFile, *lookahead, *beamWidth, model, *learned)
		return
	}

	m := ui.NewModelWithLearned(*lookahead, *beamWidth, model, *learned)
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Track mouse if needed
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao executar Tetris: %v\n", err)
		os.Exit(1)
	}
}

func runBackgroundTrain(workers int, totalGames int, logPath string, lookahead, beamWidth int, model *engine.MoveModel, learned bool) {
	cfg := simulator.Config{
		Workers:         workers,
		TotalGames:      totalGames,
		MaxMovesPerGame: 10000,
		LogPath:         logPath,
		LookaheadDepth:  lookahead,
		BeamWidth:       beamWidth,
		LearnedModel:    model,
		UseLearned:      learned,
	}

	sim, err := simulator.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro ao inicializar simulador: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n🛑 Interrompendo simulação e finalizando gravação de logs...")
		cancel()
	}()

	fmt.Println("================================================================")
	fmt.Printf("🚀 INICIANDO SIMULAÇÃO EM BACKGROUND (%d WORKERS SIMULTÂNEOS)\n", workers)
	fmt.Printf("🎯 Meta: %d partidas | 📁 Destino: %s\n", totalGames, logPath)
	fmt.Println("================================================================")

	lastLines := 0
	progressCallback := func(p simulator.Progress) {
		tetrisRate := 0.0
		if p.TotalLines > 0 {
			tetrisRate = (float64(p.TotalTetrises*4) / float64(p.TotalLines)) * 100.0
		}

		percent := (float64(p.CompletedGames) / float64(totalGames)) * 100.0
		status := fmt.Sprintf(
			"\r⚡ STATUS ATUAL:\n"+
				"   🎯 Progresso:        %d / %d partidas (%.1f%%)\n"+
				"   🕹️  Jogadas Salvas:   %d jogadas\n"+
				"   ⚡ Velocidade:       %.0f jogadas/segundo\n"+
				"   💥 Taxa de Tetris:   %.1f%% (Tetrises: %d | Linhas: %d)\n"+
				"   🏆 Maior Score:      %d pts\n"+
				"   ⏱️  Tempo Decorrido:  %.1fs\n",
			p.CompletedGames, totalGames, percent,
			p.TotalMoves,
			p.MovesPerSec,
			tetrisRate,
			p.TotalTetrises,
			p.TotalLines,
			p.HighScore,
			p.Elapsed.Seconds(),
		)

		if lastLines > 0 {
			fmt.Printf("\033[%dA", lastLines)
		}
		fmt.Print(status)
		lastLines = 7
	}

	startTime := time.Now()
	if err := sim.Run(ctx, progressCallback); err != nil {
		fmt.Fprintf(os.Stderr, "Simulação interrompida ou falhou: %v\n", err)
		os.Exit(1)
	}
	elapsed := time.Since(startTime)

	finalProg := sim.GetProgress(workers)
	tetrisRate := 0.0
	if finalProg.TotalLines > 0 {
		tetrisRate = (float64(finalProg.TotalTetrises*4) / float64(finalProg.TotalLines)) * 100.0
	}

	fmt.Println("\n================================================================")
	fmt.Println("✅ SIMULAÇÃO EM BACKGROUND CONCLUÍDA COM SUCESSO!")
	fmt.Println("================================================================")
	fmt.Printf("⏱️  Tempo Total:       %.2fs\n", elapsed.Seconds())
	fmt.Printf("🎮 Partidas Jogadas:  %d\n", finalProg.CompletedGames)
	fmt.Printf("🕹️  Jogadas Geradas:   %d\n", finalProg.TotalMoves)
	fmt.Printf("⚡ Média Throughput:  %.0f jogadas/seg\n", float64(finalProg.TotalMoves)/elapsed.Seconds())
	fmt.Printf("💥 Taxa Geral Tetris: %.1f%%\n", tetrisRate)
	fmt.Printf("🏆 Maior Pontuação:   %d pts\n", finalProg.HighScore)
	fmt.Printf("📁 Dataset Salvo Em:  %s\n", logPath)
	fmt.Println("----------------------------------------------------------------")
	fmt.Printf("💡 Para analisar o dataset gerado: ./analyze -file %s\n", logPath)
	fmt.Println("================================================================")
}
