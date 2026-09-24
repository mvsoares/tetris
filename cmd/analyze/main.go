package main

import (
	"flag"
	"fmt"
	"os"

	"tetris/internal/engine"
	"tetris/internal/logger"
)

func main() {
	logFile := flag.String("file", "logs/plays.jsonl", "Caminho do arquivo de log (.jsonl)")
	simulateGames := flag.Int("simulate", 0, "Número de partidas de IA para simular e gravar nos logs")
	clean := flag.Bool("clean", false, "Limpar/truncar arquivo de log (.jsonl)")
	flag.Parse()

	if *clean {
		if err := logger.CleanLog(*logFile); err != nil {
			fmt.Printf("❌ Erro ao limpar arquivo de log (%s): %v\n", *logFile, err)
			os.Exit(1)
		}
		fmt.Printf("🧹 Arquivo de log limpo com sucesso: %s\n", *logFile)
		if *simulateGames == 0 {
			return
		}
	}

	if *simulateGames > 0 {
		fmt.Printf("🤖 Simulando %d partida(s) de IA com gravação de logs...\n", *simulateGames)
		playLogger, err := logger.NewLogger(*logFile)
		if err != nil {
			fmt.Printf("❌ Erro ao inicializar logger: %v\n", err)
			os.Exit(1)
		}

		for gNum := 1; gNum <= *simulateGames; gNum++ {
			g := engine.NewGame()
			g.AutoPlay = true
			g.SetLogger(playLogger)

			// Step AI up to 300 moves per game
			for step := 0; step < 300 && g.State == engine.StatePlaying; step++ {
				g.StepAIImmediate()
			}
			_ = g.Close()
		}
		_ = playLogger.Close()
		fmt.Printf("✅ Simulação concluída com sucesso! Partidas registradas em %s\n\n", *logFile)
	}

	stats, err := logger.ReadLogStats(*logFile)
	if err != nil {
		fmt.Printf("❌ Erro ao abrir ou ler arquivo de log (%s): %v\n", *logFile, err)
		fmt.Println("Dica: jogue no ./tetris ou use ./train -workers 20 -games 20 para gerar dados!")
		os.Exit(1)
	}

	if stats.TotalMoves == 0 {
		fmt.Println("⚠️  Nenhum movimento registrado ainda no arquivo de log.")
		return
	}

	fmt.Println("================================================================")
	fmt.Println("            📊 RELATÓRIO DE APRENDIZADO & ANÁLISE DE JOGO       ")
	fmt.Println("================================================================")
	fmt.Printf("📁 Arquivo:            %s\n", *logFile)
	fmt.Printf("🎮 Sessões Logadas:    %d\n", stats.TotalSessions)
	fmt.Printf("🕹️  Total de Jogadas:   %d (🤖 IA: %d | 👤 Humano: %d)\n", stats.TotalMoves, stats.AIMoves, stats.HumanMoves)
	fmt.Printf("🏆 Maior Pontuação:    %d pts\n", stats.MaxScore)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("📈 EFICIÊNCIA DE LIMPEZA DE LINHAS:")
	fmt.Printf("   💥 Tetrises (4L):   %d  (%.1f%% do total de linhas limpas)\n", stats.Tetrises, stats.TetrisRate)
	fmt.Printf("   🟧 Triples  (3L):   %d\n", stats.Triples)
	fmt.Printf("   🟨 Doubles  (2L):   %d\n", stats.Doubles)
	fmt.Printf("   🟩 Singles  (1L):   %d\n", stats.Singles)
	fmt.Printf("   ✨ Total de Linhas: %d\n", stats.TotalLines)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("🧠 MÉTRICAS DA HEURÍSTICA:")
	fmt.Printf("   🕳️  Média de Buracos pós-drop:   %.2f buracos/jogada\n", stats.AvgHoles)
	fmt.Printf("   〰️  Média de Irregularidade:      %.2f desvios/jogada\n", stats.AvgBumpiness)
	fmt.Printf("   🚨 Jogadas em Modo Limpeza:       %d (%.1f%% das jogadas)\n", stats.CleanupMoves, (float64(stats.CleanupMoves)/float64(stats.TotalMoves))*100.0)
	fmt.Println("----------------------------------------------------------------")
	fmt.Println("💡 INSIGHTS PARA O ALGORITMO:")
	if stats.TetrisRate >= 70.0 {
		fmt.Println("   ⭐ Excelente taxa de Tetris (>70%)! O empilhamento 9-0 está funcionando com maestria.")
	} else if stats.TetrisRate >= 40.0 {
		fmt.Println("   👍 Boa taxa de Tetris (40-70%). A IA equilibra limpezas de emergência e Tetrises.")
	} else {
		fmt.Println("   ⚠️  Taxa de Tetris baixa (<40%). Muitas linhas parciais sendo limpas; considere subir o peso de Tetris.")
	}

	if stats.AvgHoles > 1.5 {
		fmt.Println("   ⚠️  Média de buracos elevada. Sugestão: Aumentar a penalidade por buracos (holes) na função de custo.")
	} else {
		fmt.Println("   ⭐ Pilha muito limpa (<1.5 buracos em média). Mínimo de buracos enterrados.")
	}
	fmt.Println("================================================================")
}
