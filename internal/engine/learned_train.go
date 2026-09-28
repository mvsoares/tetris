package engine

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"

	"tetris/internal/logger"
)

type RiskTrainingConfig struct {
	Horizon    int
	Rollouts   int
	Candidates int
	Workers    int
	Epochs     int
	Seed       int64
}

type RiskExample struct {
	SourceSeed int64     `json:"source_seed"`
	SourceMove int       `json:"source_move"`
	Features   []float64 `json:"features"`
	Survived   int       `json:"survived"`
	Trials     int       `json:"trials"`
}

func normalizeRiskTraining(cfg RiskTrainingConfig) (RiskTrainingConfig, error) {
	if cfg.Horizon == 0 {
		cfg.Horizon = 50
	}
	if cfg.Rollouts == 0 {
		cfg.Rollouts = 4
	}
	if cfg.Candidates == 0 {
		cfg.Candidates = 4
	}
	if cfg.Workers == 0 {
		cfg.Workers = 4
	}
	if cfg.Epochs == 0 {
		cfg.Epochs = 300
	}
	if cfg.Horizon < 2 || cfg.Horizon > 100 || cfg.Rollouts < 1 || cfg.Rollouts > 64 || cfg.Candidates < 1 || cfg.Candidates > 64 || cfg.Workers < 1 || cfg.Workers > 16 || cfg.Epochs < 1 || cfg.Epochs > 5000 {
		return cfg, fmt.Errorf("invalid training configuration")
	}
	return cfg, nil
}

// SimulateRiskExamples creates counterfactual labels, sharing each sampled bag
// stream across candidate actions. All future decisions use v2, never the model.
func SimulateRiskExamples(ctx context.Context, decisions []*logger.DecisionLog, cfg RiskTrainingConfig) ([]RiskExample, error) {
	cfg, err := normalizeRiskTraining(cfg)
	if err != nil {
		return nil, err
	}
	if len(decisions) == 0 {
		return nil, fmt.Errorf("no rich decision snapshots; historical logs cannot supply queue/hold context")
	}
	type result struct {
		examples []RiskExample
		err      error
	}
	results := make([]result, len(decisions))
	jobs := make(chan int)
	var wg sync.WaitGroup
	for worker := 0; worker < cfg.Workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				d := decisions[i]
				g, e := GameFromDecision(d, cfg.Seed+int64(i))
				if e != nil {
					results[i].err = e
					continue
				}
				cs := riskCandidates(g)
				sort.SliceStable(cs, func(a, b int) bool { return cs[a].placement.score > cs[b].placement.score })
				// Include the logged choice and the best immediate placement, then
				// sample alternatives rather than labeling unchosen actions as failures.
				var chosen []riskCandidate
				contains := func(c riskCandidate) bool {
					for _, v := range chosen {
						if v.useHold == c.useHold && v.piece == c.piece && v.placement.x == c.placement.x && v.placement.y == c.placement.y && v.placement.rotation == c.placement.rotation {
							return true
						}
					}
					return false
				}
				for _, c := range cs {
					for _, logged := range d.Candidates {
						if logged.Selected && logged.UseHold == c.useHold && logged.Piece == string(c.piece) && logged.X == c.placement.x && logged.Y == c.placement.y && logged.Rotation == c.placement.rotation {
							if !contains(c) {
								chosen = append(chosen, c)
							}
						}
					}
				}
				if len(cs) > 0 && !contains(cs[0]) {
					chosen = append(chosen, cs[0])
				}
				if len(cs) > 1 && len(chosen) < cfg.Candidates && !contains(cs[1]) {
					chosen = append(chosen, cs[1])
				}
				if len(cs) > 2 && len(chosen) < cfg.Candidates && !contains(cs[len(cs)-1]) {
					chosen = append(chosen, cs[len(cs)-1])
				}
				permutation := rand.New(rand.NewSource(cfg.Seed + int64(i)*7919)).Perm(len(cs))
				for _, j := range permutation {
					if len(chosen) >= cfg.Candidates {
						break
					}
					if !contains(cs[j]) {
						chosen = append(chosen, cs[j])
					}
				}
				if len(chosen) > cfg.Candidates {
					chosen = chosen[:cfg.Candidates]
				}
				for _, c := range chosen {
					example := RiskExample{SourceSeed: d.Seed, SourceMove: d.MoveNumber, Features: candidateFeatures(g, c), Trials: cfg.Rollouts}
					for trial := 0; trial < cfg.Rollouts; trial++ {
						if ctx.Err() != nil {
							results[i].err = ctx.Err()
							break
						}
						// Provenance/game RNG seeds are not used as unseen future order.
						r, e := GameFromDecision(d, cfg.Seed+int64(i)*100003+int64(trial)*104729)
						if e != nil {
							results[i].err = e
							break
						}
						if c.useHold && !r.Hold() {
							continue
						}
						r.CurrentPiece.X, r.CurrentPiece.Y, r.CurrentPiece.Rotation = c.placement.x, c.placement.y, c.placement.rotation
						if !r.Board.IsValidPosition(r.CurrentPiece) {
							results[i].err = fmt.Errorf("invalid counterfactual action")
							break
						}
						r.CurrentAIMove = candidateMove(c)
						r.hardDrop(NewPiece(r.CurrentPiece.Type).Y)
						r.CurrentAIMove = nil
						nextAI, nextGravity := AIActionInterval, r.TickInterval()
						for r.MoveCount < cfg.Horizon && r.State == StatePlaying {
							if ctx.Err() != nil {
								results[i].err = ctx.Err()
								break
							}
							if nextAI <= nextGravity {
								r.SetAIGravityRemaining(nextGravity - nextAI)
								r.StepAI()
								nextAI += AIActionInterval
							} else {
								r.Tick()
								nextGravity += r.TickInterval()
							}
						}
						if r.State == StatePlaying {
							example.Survived++
						}
					}
					results[i].examples = append(results[i].examples, example)
					if results[i].err != nil {
						break
					}
				}
			}
		}()
	}
	for i := range decisions {
		select {
		case jobs <- i:
		case <-ctx.Done():
			close(jobs)
			wg.Wait()
			return nil, ctx.Err()
		}
	}
	close(jobs)
	wg.Wait()
	var out []RiskExample
	for _, r := range results {
		if r.err != nil {
			return nil, r.err
		}
		out = append(out, r.examples...)
	}
	return out, nil
}

// FitMoveModel splits by complete source seed, including all related candidate
// rollouts in one partition. Normalization and bounds use training data only.
func FitMoveModel(examples []RiskExample, cfg RiskTrainingConfig) (*MoveModel, error) {
	cfg, err := normalizeRiskTraining(cfg)
	if err != nil {
		return nil, err
	}
	names := RiskFeatureNames()
	n := len(names)
	m := &MoveModel{Version: MoveModelVersion, Target: "survive-v2-placement-sampled-7bag", Horizon: cfg.Horizon, Features: names, Means: make([]float64, n), Scales: make([]float64, n), Min: make([]float64, n), Max: make([]float64, n), Weights: make([]float64, n), Calibration: make([]CalibrationBin, 10)}
	train, valid := []RiskExample{}, []RiskExample{}
	trainSeeds, validSeeds := map[int64]bool{}, map[int64]bool{}
	for _, e := range examples {
		if len(e.Features) != n || e.Trials < 1 || e.Survived < 0 || e.Survived > e.Trials {
			return nil, fmt.Errorf("invalid risk example")
		}
		for _, v := range e.Features {
			if !finite(v) {
				return nil, fmt.Errorf("nonfinite feature")
			}
		}
		if e.SourceSeed%5 == 0 {
			valid = append(valid, e)
			validSeeds[e.SourceSeed] = true
		} else {
			train = append(train, e)
			trainSeeds[e.SourceSeed] = true
		}
	}
	if len(train) < 10 || len(valid) < 5 || len(trainSeeds) < 2 || len(validSeeds) < 1 {
		return nil, fmt.Errorf("need at least 10 training examples from 2 seeds and 5 validation examples from seeds divisible by 5")
	}
	for j := 0; j < n; j++ {
		m.Min[j], m.Max[j] = math.Inf(1), math.Inf(-1)
		for _, e := range train {
			m.Means[j] += e.Features[j]
			m.Min[j] = math.Min(m.Min[j], e.Features[j])
			m.Max[j] = math.Max(m.Max[j], e.Features[j])
		}
		m.Means[j] /= float64(len(train))
		for _, e := range train {
			d := e.Features[j] - m.Means[j]
			m.Scales[j] += d * d
		}
		m.Scales[j] = math.Sqrt(m.Scales[j] / float64(len(train)))
		if m.Scales[j] < 0.01 {
			m.Scales[j] = 1
		}
	}
	positive, total := 0, 0
	for _, e := range train {
		positive += e.Survived
		total += e.Trials
	}
	if positive == 0 || positive == total {
		return nil, fmt.Errorf("training labels contain only one outcome; collect more varied/dangerous states")
	}
	prior := float64(positive+1) / float64(total+2)
	m.Bias = math.Log(prior / (1 - prior))
	for epoch := 0; epoch < cfg.Epochs; epoch++ {
		gradient := make([]float64, n)
		biasGradient := 0.0
		for _, e := range train {
			value := m.Bias
			for j, f := range e.Features {
				value += m.Weights[j] * (f - m.Means[j]) / m.Scales[j]
			}
			diff := logistic(value)*float64(e.Trials) - float64(e.Survived)
			biasGradient += diff
			for j, f := range e.Features {
				gradient[j] += diff * (f - m.Means[j]) / m.Scales[j]
			}
		}
		step := 0.10 / math.Sqrt(1+float64(epoch)/100)
		m.Bias -= step * biasGradient / float64(total)
		for j := range m.Weights {
			m.Weights[j] -= step * (gradient[j]/float64(total) + 0.015*m.Weights[j])
		}
	}
	validTrials := 0
	for _, e := range valid {
		p, _ := m.Predict(e.Features)
		p = math.Max(1e-9, math.Min(1-1e-9, p))
		yes, no := float64(e.Survived), float64(e.Trials-e.Survived)
		m.ValidationBrier += yes*(1-p)*(1-p) + no*p*p
		m.BaselineBrier += yes*(1-prior)*(1-prior) + no*prior*prior
		m.ValidationLogLoss -= yes*math.Log(p) + no*math.Log(1-p)
		bin := int(p * 10)
		if bin > 9 {
			bin = 9
		}
		m.Calibration[bin].Count += e.Trials
		m.Calibration[bin].Predicted += p * float64(e.Trials)
		m.Calibration[bin].Observed += yes
		validTrials += e.Trials
	}
	m.ValidationBrier /= float64(validTrials)
	m.BaselineBrier /= float64(validTrials)
	m.ValidationLogLoss /= float64(validTrials)
	for i := range m.Calibration {
		if m.Calibration[i].Count > 0 {
			count := float64(m.Calibration[i].Count)
			m.Calibration[i].Predicted /= count
			m.Calibration[i].Observed /= count
		}
	}
	for s := range trainSeeds {
		m.TrainingSeeds = append(m.TrainingSeeds, s)
	}
	for s := range validSeeds {
		m.ValidationSeeds = append(m.ValidationSeeds, s)
	}
	sort.Slice(m.TrainingSeeds, func(i, j int) bool { return m.TrainingSeeds[i] < m.TrainingSeeds[j] })
	sort.Slice(m.ValidationSeeds, func(i, j int) bool { return m.ValidationSeeds[i] < m.ValidationSeeds[j] })
	m.TrainingSamples = total
	m.ValidationSamples = validTrials
	m.ID = modelFingerprint(m)
	return m, m.Validate()
}
