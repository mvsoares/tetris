package engine

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"tetris/internal/logger"
)

func testRiskModel(t *testing.T) *MoveModel {
	t.Helper()
	var examples []RiskExample
	for _, seed := range []int64{1, 2, 5} {
		for i := 0; i < 20; i++ {
			f := make([]float64, len(RiskFeatureNames()))
			f[0] = float64(i % 2)
			examples = append(examples, RiskExample{SourceSeed: seed, Features: f, Trials: 4, Survived: 4 * (i % 2)})
		}
	}
	m, err := FitMoveModel(examples, RiskTrainingConfig{Horizon: 10, Epochs: 400})
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestRiskFitAndGroupedValidation(t *testing.T) {
	m := testRiskModel(t)
	if !reflect.DeepEqual(m.TrainingSeeds, []int64{1, 2}) || !reflect.DeepEqual(m.ValidationSeeds, []int64{5}) || m.ValidationBrier >= m.BaselineBrier {
		t.Fatalf("bad seed split/fit: %+v", m)
	}
	f := make([]float64, len(RiskFeatureNames()))
	low, ok := m.Predict(f)
	if !ok || low >= 0.5 {
		t.Fatal("failed low-risk-class prediction")
	}
	f[0] = 1
	high, ok := m.Predict(f)
	if !ok || high <= 0.5 {
		t.Fatal("failed high-survival prediction")
	}
	f[0] = 2
	_, ok = m.Predict(f)
	if ok {
		t.Fatal("out-of-range state supported")
	}
	f[0] = math.NaN()
	if _, ok = m.Predict(f); ok {
		t.Fatal("nonfinite features accepted")
	}
	m.Scales = nil
	if _, ok = m.Predict(f); ok {
		t.Fatal("malformed model accepted")
	}
}

func TestRiskModelLoadAndReject(t *testing.T) {
	m := testRiskModel(t)
	path := filepath.Join(t.TempDir(), "model.json")
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadMoveModel(path)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(m.Weights, loaded.Weights) || loaded.ID != m.ID {
		t.Fatal("model changed during load")
	}
	for _, change := range []func(*MoveModel){func(v *MoveModel) { v.Version = 99 }, func(v *MoveModel) { v.Weights = nil }, func(v *MoveModel) { v.Features[0] = "wrong" }, func(v *MoveModel) { v.Scales[0] = 0 }, func(v *MoveModel) { v.ValidationSeeds = []int64{1} }, func(v *MoveModel) { v.ValidationBrier = 2 }, func(v *MoveModel) { v.Horizon = 0 }} {
		copyModel := testRiskModel(t)
		change(copyModel)
		data, _ = json.Marshal(copyModel)
		os.WriteFile(path, data, 0600)
		if _, err = LoadMoveModel(path); err == nil {
			t.Fatal("invalid model loaded")
		}
	}
	os.WriteFile(path, []byte("{"), 0600)
	if _, err = LoadMoveModel(path); err == nil {
		t.Fatal("truncated JSON loaded")
	}
	if _, err = LoadMoveModel(filepath.Join(t.TempDir(), "missing")); err == nil {
		t.Fatal("missing model loaded")
	}
}

func TestLearnedFallbackAndRestart(t *testing.T) {
	g := NewGameWithSeed(42)
	baseline := FindBestMove(g)
	g.UseLearned = true
	fallback := FindBestMove(g)
	if fallback.LearnedFallback != "no_model" || fallback.TargetX != baseline.TargetX || fallback.TargetRotation != baseline.TargetRotation || fallback.UseHold != baseline.UseHold {
		t.Fatal("missing model changed decision")
	}
	g.LearnedModel = &MoveModel{}
	if m := FindBestMove(g); m.LearnedFallback != "invalid_model" {
		t.Fatal("invalid model did not fall back")
	}
	g.LearnedModel = testRiskModel(t)
	if m := FindBestMove(g); m.LearnedUsed {
		t.Fatal("unsupported board used model")
	}
	model := g.LearnedModel
	g.Restart()
	if !g.UseLearned || g.LearnedModel != model {
		t.Fatal("restart lost model")
	}
}

func TestLearnedCanRerankOnlyLegalCandidates(t *testing.T) {
	g := NewGameWithSeed(42)
	base := FindBestMove(g)
	cs := riskCandidates(g)
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].placement.score > cs[j].placement.score })
	var baseFeatures []float64
	for _, c := range cs {
		if matchesCandidate(base, c) {
			baseFeatures = candidateFeatures(g, c)
		}
	}
	if baseFeatures == nil {
		t.Fatal("baseline not in legal candidates")
	}
	feature := -1
	for i, c := range cs {
		if i >= 6 {
			break
		}
		f := candidateFeatures(g, c)
		for j := 14; j < 24; j++ {
			if f[j] > baseFeatures[j]+0.01 {
				feature = j
				break
			}
		}
		if feature >= 0 {
			break
		}
	}
	if feature < 0 {
		t.Fatal("fixture lacks distinct alternative")
	}
	m := testRiskModel(t)
	for i := range m.Weights {
		m.Weights[i] = 0
		m.Means[i] = 0
		m.Scales[i] = 1
		m.Min[i] = -1e6
		m.Max[i] = 1e6
	}
	m.Weights[feature] = 200
	m.Bias = math.Log(0.7/0.3) - 200*baseFeatures[feature]
	m.ValidationBrier = 0.1
	m.BaselineBrier = 0.2
	g.UseLearned, g.LearnedModel = true, m
	move := FindBestMove(g)
	if !move.LearnedUsed || move.MoveProbability == nil || *move.MoveProbability < 0.9 {
		t.Fatalf("reranking inactive: %+v", move)
	}
	legal := false
	for _, c := range cs {
		if matchesCandidate(move, c) {
			legal = true
		}
	}
	if !legal {
		t.Fatal("model selected illegal placement")
	}
	m.ValidationBrier = 0.3
	if fallback := FindBestMove(g); fallback.LearnedUsed || fallback.LearnedFallback != "validation_not_better_than_prior" {
		t.Fatal("failed calibration gate")
	}
}

func TestRichDecisionRoundTripAndConditionalBag(t *testing.T) {
	g := NewGameWithSeed(42)
	g.CurrentAIMove = FindBestMove(g)
	d := captureDecision(g, g.CurrentAIMove)
	if len(d.Candidates) == 0 || d.Piece != string(g.CurrentPiece.Type) || len(d.Queue) != len(g.NextQueue) {
		t.Fatal("incomplete decision")
	}
	selected := 0
	for _, c := range d.Candidates {
		if c.Selected {
			selected++
		}
	}
	if selected != 1 {
		t.Fatalf("selected candidate count %d", selected)
	}
	data, _ := json.Marshal(d)
	var decoded logger.DecisionLog
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	r, err := GameFromDecision(&decoded, 123)
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < BoardHeight; y++ {
		if g.Board.RowMask(y) != r.Board.RowMask(y) {
			t.Fatal("board round trip changed")
		}
	}
	if !reflect.DeepEqual(g.NextQueue, r.NextQueue) || *g.CurrentPiece != *r.CurrentPiece || g.CanHold != r.CanHold || g.Level != r.Level {
		t.Fatal("context round trip changed")
	}
	bag := append([]TetrominoType(nil), r.Randomizer.bag...)
	seen := map[TetrominoType]bool{}
	for range bag {
		seen[r.Randomizer.Next()] = true
	}
	for _, p := range bag {
		if !seen[p] {
			t.Fatal("conditional bag lost a piece")
		}
	}
	if _, err = GameFromDecision(&logger.DecisionLog{}, 1); err == nil {
		t.Fatal("malformed snapshot accepted")
	}
}

func TestCounterfactualLabelsDeterministicAndDoNotMutate(t *testing.T) {
	var ds []*logger.DecisionLog
	for _, seed := range []int64{1, 2, 5} {
		g := NewGameWithSeed(seed)
		ds = append(ds, captureDecision(g, FindBestMove(g)))
	}
	before, _ := json.Marshal(ds)
	cfg := RiskTrainingConfig{Horizon: 3, Rollouts: 2, Candidates: 3, Workers: 1, Seed: 99}
	a, err := SimulateRiskExamples(context.Background(), ds, cfg)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Workers = 4
	b, err := SimulateRiskExamples(context.Background(), ds, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a, b) || len(a) != 9 {
		t.Fatal("worker count changed counterfactuals")
	}
	after, _ := json.Marshal(ds)
	if string(before) != string(after) {
		t.Fatal("simulation mutated source")
	}
	for _, e := range a {
		if e.Trials != 2 || e.Survived < 0 || e.Survived > 2 || len(e.Features) != len(RiskFeatureNames()) {
			t.Fatal("invalid labels")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err = SimulateRiskExamples(ctx, ds, cfg); err == nil {
		t.Fatal("cancel ignored")
	}
}

func BenchmarkLearnedProbability(b *testing.B) {
	n := len(RiskFeatureNames())
	m := &MoveModel{Means: make([]float64, n), Scales: make([]float64, n), Min: make([]float64, n), Max: make([]float64, n), Weights: make([]float64, n)}
	for i := range m.Scales {
		m.Scales[i] = 1
		m.Max[i] = 1
	}
	f := make([]float64, n)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Predict(f)
	}
}

func BenchmarkHybridDecision(b *testing.B) {
	model, err := LoadMoveModel("../../models/move-risk.json")
	if err != nil {
		b.Skip(err)
	}
	for _, learned := range []bool{false, true} {
		name := "v2"
		if learned {
			name = "hybrid"
		}
		b.Run(name, func(b *testing.B) {
			g := NewGameWithSeed(5001)
			g.UseLearned, g.LearnedModel = learned, model
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				FindBestMove(g)
			}
		})
	}
}

func TestRiskFeaturesIncludeLookaheadHoleDeltaAndTransitions(t *testing.T) {
	names := RiskFeatureNames()
	wantSuffix := []string{"next_lookahead", "hole_delta", "top_hole_blockades", "row_transitions", "col_transitions"}
	if len(names) < len(wantSuffix) {
		t.Fatalf("expected at least %d features, got %d", len(wantSuffix), len(names))
	}
	gotSuffix := names[len(names)-len(wantSuffix):]
	if !reflect.DeepEqual(gotSuffix, wantSuffix) {
		t.Fatalf("expected trailing features %v, got %v", wantSuffix, gotSuffix)
	}
}

func TestLearnedMoveNeverOverridesIntoHoleCreatingCandidate(t *testing.T) {
	g := NewGameWithSeed(42)
	// Create a step in column 1 so placing a piece bridging col 0..1 creates a hole in col 0.
	g.Board.Cells[19][1] = Cell{Filled: true}
	g.Board.Cells[18][1] = Cell{Filled: true}
	g.CurrentPiece = NewPiece(PieceT)
	g.CanHold = false

	base := FindBestMove(g)
	if base == nil {
		t.Fatal("expected baseline move")
	}
	cs := riskCandidates(g)
	var holeCand *riskCandidate
	for i := range cs {
		if cs[i].placement.bits.CountHoles() > g.Board.CountHoles() {
			holeCand = &cs[i]
			break
		}
	}
	if holeCand == nil {
		t.Fatal("expected at least one hole-creating candidate on stepped board")
	}

	// Construct a model that strongly favors hole_delta > 0
	m := testRiskModel(t)
	for i := range m.Weights {
		m.Weights[i] = 0
		m.Means[i] = 0
		m.Scales[i] = 1
		m.Min[i] = -1e6
		m.Max[i] = 1e6
	}
	holeDeltaIdx := -1
	for i, name := range RiskFeatureNames() {
		if name == "hole_delta" {
			holeDeltaIdx = i
			break
		}
	}
	if holeDeltaIdx < 0 {
		t.Fatal("missing hole_delta feature")
	}
	m.Weights[holeDeltaIdx] = 500
	m.Bias = math.Log(0.75 / 0.25)
	m.ValidationBrier = 0.01
	m.BaselineBrier = 0.10

	g.UseLearned, g.LearnedModel = true, m
	chosen := FindBestMove(g)
	if chosen == nil {
		t.Fatal("expected move")
	}
	// Verify chosen move does not create a hole
	sim := g.Board.Clone()
	p := NewPiece(g.CurrentPiece.Type)
	p.Rotation, p.X, p.Y = chosen.TargetRotation, chosen.TargetX, chosen.TargetY
	sim.LockPiece(p)
	sim.ClearLines()
	if sim.CountHoles() > g.Board.CountHoles() {
		t.Fatalf("learned model overrode baseline into a hole-creating placement: %+v", chosen)
	}
}

