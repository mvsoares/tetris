package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"tetris/internal/simulator"
)

func main() {
	workers := flag.Int("workers", 50, "Número de partidas simultâneas em background (1 a 50)")
	totalGames := flag.Int("games", 100, "Total de partidas a simular (0 para contínuo até Ctrl+C)")
	maxMoves := flag.Int("max-moves", 10000, "Limite máximo de jogadas/linhas por partida")
	logFile := flag.String("file", "logs/plays.jsonl", "Caminho do arquivo de log (.jsonl)")
	clean := flag.Bool("clean", false, "Limpar/truncar arquivo de log antes de iniciar a simulação")
	bufKB := flag.Int("buf-kb", 102400, "Tamanho do buffer de escrita de log em KB (padrão: 102400 KB = 100 MB)")
	flag.Parse()

	if *workers < 1 {
		*workers = 1
	}
	if *workers > 50 {
		fmt.Printf("⚠️  Limite máximo é 50 workers simultâneos. Ajustando para 50.\n")
		*workers = 50
	}

	cfg := simulator.Config{
		Workers:         *workers,
		TotalGames:      *totalGames,
		MaxMovesPerGame: *maxMoves,
		LogPath:         *logFile,
		Clean:           *clean,
		BufferSize:      *bufKB * 1024,
	}

	sim, err := simulator.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Erro ao inicializar simulador: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown on Ctrl+C / SIGTERM
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n🛑 Sinal de interrupção recebido! Finalizando workers e salvando logs...")
		cancel()
	}()

	fmt.Println("================================================================")
	fmt.Printf("🚀 INICIANDO SIMULAÇÃO EM BACKGROUND (%d WORKERS SIMULTÂNEOS)\n", *workers)
	if *clean {
		fmt.Printf("🧹 Arquivo de log limpo/truncado: %s\n", *logFile)
	}
	if *totalGames > 0 {
		fmt.Printf("🎯 Meta: %d partidas | 📁 Destino: %s (Buffer: %d KB)\n", *totalGames, *logFile, *bufKB)
	} else {
		fmt.Printf("🎯 Modo contínuo (pressione Ctrl+C para parar) | 📁 Destino: %s (Buffer: %d KB)\n", *logFile, *bufKB)
	}
	fmt.Println("================================================================")

	lastLines := 0
	progressCallback := func(p simulator.Progress) {
		// Clean lines if previously printed
		tetrisRate := 0.0
		if p.TotalLines > 0 {
			tetrisRate = (float64(p.TotalTetrises*4) / float64(p.TotalLines)) * 100.0
		}

		progressStr := ""
		if *totalGames > 0 {
			percent := (float64(p.CompletedGames) / float64(*totalGames)) * 100.0
			progressStr = fmt.Sprintf("   🎯 Progresso:        %d / %d partidas (%.1f%%)\n", p.CompletedGames, *totalGames, percent)
		} else {
			progressStr = fmt.Sprintf("   🎯 Partidas:         %d partidas concluídas\n", p.CompletedGames)
		}

		status := fmt.Sprintf(
			"\r⚡ STATUS ATUAL:\n"+
				"%s"+
				"   🕹️  Jogadas Salvas:   %d jogadas\n"+
				"   ⚡ Velocidade:       %.0f jogadas/segundo\n"+
				"   💥 Taxa de Tetris:   %.1f%% (Tetrises: %d | Linhas: %d)\n"+
				"   🏆 Maior Score:      %d pts\n"+
				"   ⏱️  Tempo Decorrido:  %.1fs\n",
			progressStr,
			p.TotalMoves,
			p.MovesPerSec,
			tetrisRate,
			p.TotalTetrises,
			p.TotalLines,
			p.HighScore,
			p.Elapsed.Seconds(),
		)

		// Print or overwrite
		if lastLines > 0 {
			// Move cursor up by lastLines
			fmt.Printf("\033[%dA", lastLines)
		}
		fmt.Print(status)
		lastLines = 7
	}

	startTime := time.Now()
	err = sim.Run(ctx, progressCallback)
	elapsed := time.Since(startTime)

	finalProg := sim.GetProgress(*workers)
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
	fmt.Printf("📁 Dataset Salvo Em:  %s\n", *logFile)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("💡 Para analisar o dataset completo gerado, execute:")
	fmt.Printf("   ./analyze -file %s\n", *logFile)
	fmt.Println("================================================================")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Aviso: %v\n", err)
	}
}
