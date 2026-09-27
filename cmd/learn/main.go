package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"tetris/internal/engine"
	"tetris/internal/logger"
)

func main() {
	file := flag.String("file", "logs/learning-source.jsonl", "Log com snapshots ricos de decisões (logs antigos não bastam)")
	out := flag.String("out", "models/move-risk.json", "Modelo JSON para carregar na inicialização")
	decisions := flag.Int("decisions", 100, "Quantidade máxima de decisões amostradas uniformemente")
	horizon := flag.Int("horizon", 50, "Horizonte incluindo a colocação candidata (2..100)")
	rollouts := flag.Int("rollouts", 4, "Futuros amostrados por alternativa (1..64)")
	candidates := flag.Int("candidates", 4, "Alternativas por decisão (1..64)")
	workers := flag.Int("workers", 4, "Workers de coleta (1..16)")
	epochs := flag.Int("epochs", 300, "Épocas de ajuste logístico")
	seed := flag.Int64("seed", 20260927, "Seed do treino, independente dos futuros reais das partidas")
	flag.Parse()
	if err := run(*file, *out, *decisions, engine.RiskTrainingConfig{Horizon: *horizon, Rollouts: *rollouts, Candidates: *candidates, Workers: *workers, Epochs: *epochs, Seed: *seed}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(path, out string, limit int, cfg engine.RiskTrainingConfig) error {
	if limit < 10 || limit > 10000 {
		return fmt.Errorf("decisions deve estar entre 10 e 10000")
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 65536), 2<<20)
	rng := rand.New(rand.NewSource(cfg.Seed))
	var sample []*logger.DecisionLog
	seen, line := 0, 0
	sessions := map[string]bool{}
	sampleSessions := []string{}
	for scanner.Scan() {
		line++
		var header struct {
			Type      string `json:"type"`
			SessionID string `json:"session_id"`
		}
		if err = json.Unmarshal(scanner.Bytes(), &header); err != nil {
			return fmt.Errorf("linha %d: %w", line, err)
		}
		if header.Type == "SESSION_END" {
			var end logger.SessionEndLog
			if err = json.Unmarshal(scanner.Bytes(), &end); err != nil {
				return err
			}
			sessions[header.SessionID] = end.EndReason == engine.EndMoveLimit || end.EndReason == engine.EndTopOut || end.EndReason == engine.EndNoLegalMove
			continue
		}
		var move logger.MoveLog
		if err = json.Unmarshal(scanner.Bytes(), &move); err != nil {
			return err
		}
		if move.Decision == nil || !move.HadAIPlan || !move.AIPlanMatched {
			continue
		}
		seen++
		if len(sample) < limit {
			sample = append(sample, move.Decision)
			sampleSessions = append(sampleSessions, move.SessionID)
		} else {
			i := rng.Intn(seen)
			if i < limit {
				sample[i] = move.Decision
				sampleSessions[i] = move.SessionID
			}
		}
	}
	if err = scanner.Err(); err != nil {
		return err
	}
	complete := sample[:0]
	for i, d := range sample {
		if sessions[sampleSessions[i]] {
			complete = append(complete, d)
		}
	}
	sample = complete
	fmt.Printf("Decisões válidas: %d; amostradas de sessões completas: %d\n", seen, len(sample))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	examples, err := engine.SimulateRiskExamples(ctx, sample, cfg)
	if err != nil {
		return err
	}
	model, err := engine.FitMoveModel(examples, cfg)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(out), 0755); err != nil {
		return err
	}
	// Atomic publication prevents readers seeing a half-written model.
	tmp, err := os.CreateTemp(filepath.Dir(out), ".move-risk-*.json")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	encoder := json.NewEncoder(tmp)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(model); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Rename(name, out); err != nil {
		return err
	}
	fmt.Printf("Modelo salvo: %s | treino: %d rollouts | validação: %d | Brier: %.5f (prior: %.5f)\n", out, model.TrainingSamples, model.ValidationSamples, model.ValidationBrier, model.BaselineBrier)
	if model.ValidationBrier >= model.BaselineBrier {
		fmt.Println("Modelo não superou o prior: a política híbrida manterá o fallback heurístico.")
	}
	return nil
}
