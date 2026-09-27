package engine

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"tetris/internal/logger"
)

const MoveModelVersion = 1

// MoveModel is an immutable, portable logistic risk model. Probabilities target
// sampled 7-bag futures with v2 DIRECT-PLACEMENT continuation, not UI win rates.
type MoveModel struct {
	Version           int              `json:"version"`
	ID                string           `json:"id"`
	Target            string           `json:"target"`
	Horizon           int              `json:"horizon"`
	Features          []string         `json:"features"`
	Means             []float64        `json:"means"`
	Scales            []float64        `json:"scales"`
	Min               []float64        `json:"min"`
	Max               []float64        `json:"max"`
	Weights           []float64        `json:"weights"`
	Bias              float64          `json:"bias"`
	TrainingSeeds     []int64          `json:"training_seeds"`
	ValidationSeeds   []int64          `json:"validation_seeds"`
	TrainingSamples   int              `json:"training_samples"`
	ValidationSamples int              `json:"validation_samples"`
	ValidationBrier   float64          `json:"validation_brier"`
	BaselineBrier     float64          `json:"baseline_brier"`
	ValidationLogLoss float64          `json:"validation_log_loss"`
	Calibration       []CalibrationBin `json:"calibration"`
}

type CalibrationBin struct {
	Count     int     `json:"count"`
	Predicted float64 `json:"predicted"`
	Observed  float64 `json:"observed"`
}

func RiskFeatureNames() []string {
	names := []string{"before_height", "before_holes", "before_bump", "after_height", "after_center", "after_holes", "after_bump", "after_blocks", "lines_cleared", "heuristic", "cleanup", "use_hold", "level", "gravity_phase"}
	for x := 0; x < BoardWidth; x++ {
		names = append(names, fmt.Sprintf("column_%d", x))
	}
	for _, prefix := range []string{"piece_", "next_", "hold_"} {
		for _, p := range AllPieces {
			names = append(names, prefix+string(p))
		}
	}
	return names
}

func (m *MoveModel) Validate() error {
	if m == nil {
		return fmt.Errorf("missing move model")
	}
	if m.Version != MoveModelVersion || m.Target != "survive-v2-placement-sampled-7bag" || m.Horizon < 2 || m.Horizon > 100 {
		return fmt.Errorf("unsupported move model version/target/horizon")
	}
	names := RiskFeatureNames()
	if len(m.Features) != len(names) || len(m.Means) != len(names) || len(m.Scales) != len(names) || len(m.Min) != len(names) || len(m.Max) != len(names) || len(m.Weights) != len(names) {
		return fmt.Errorf("incompatible model dimensions")
	}
	if !finite(m.Bias) || m.TrainingSamples < 1 || m.ValidationSamples < 1 || !finite(m.ValidationBrier) || m.ValidationBrier < 0 || m.ValidationBrier > 1 || !finite(m.BaselineBrier) || m.BaselineBrier < 0 || m.BaselineBrier > 1 || !finite(m.ValidationLogLoss) || m.ValidationLogLoss < 0 {
		return fmt.Errorf("invalid model metadata")
	}
	seen := map[int64]bool{}
	if len(m.TrainingSeeds) == 0 || len(m.ValidationSeeds) == 0 {
		return fmt.Errorf("missing seed split")
	}
	for _, s := range m.TrainingSeeds {
		if seen[s] {
			return fmt.Errorf("duplicate training seed")
		}
		seen[s] = true
	}
	for _, s := range m.ValidationSeeds {
		if seen[s] {
			return fmt.Errorf("training/validation seed overlap")
		}
		seen[s] = true
	}
	for i, n := range names {
		if m.Features[i] != n || !finite(m.Means[i]) || !finite(m.Scales[i]) || m.Scales[i] <= 0 || !finite(m.Min[i]) || !finite(m.Max[i]) || m.Min[i] > m.Max[i] || !finite(m.Weights[i]) {
			return fmt.Errorf("invalid feature %s", n)
		}
	}
	for _, bin := range m.Calibration {
		if bin.Count < 0 || !finite(bin.Predicted) || !finite(bin.Observed) || bin.Predicted < 0 || bin.Predicted > 1 || bin.Observed < 0 || bin.Observed > 1 {
			return fmt.Errorf("invalid calibration")
		}
	}
	return nil
}

func LoadMoveModel(path string) (*MoveModel, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 1<<20+1))
	if err != nil {
		return nil, err
	}
	if len(data) > 1<<20 {
		return nil, fmt.Errorf("model larger than 1 MB")
	}
	var m MoveModel
	if err = json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if err = m.Validate(); err != nil {
		return nil, err
	}
	m.ID = modelFingerprint(&m)
	return &m, nil
}

func modelFingerprint(m *MoveModel) string {
	copyModel := *m
	copyModel.ID = ""
	data, _ := json.Marshal(copyModel)
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:6])
}

func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func logistic(v float64) float64 {
	if v >= 0 {
		return 1 / (1 + math.Exp(-v))
	}
	e := math.Exp(v)
	return e / (1 + e)
}

func (m *MoveModel) Predict(features []float64) (float64, bool) {
	if m == nil || len(features) != len(m.Weights) || len(m.Means) != len(features) || len(m.Scales) != len(features) || len(m.Min) != len(features) || len(m.Max) != len(features) {
		return 0, false
	}
	value := m.Bias
	supported := true
	for i, f := range features {
		if !finite(f) || !finite(m.Scales[i]) || m.Scales[i] <= 0 || !finite(m.Weights[i]) || !finite(m.Means[i]) || !finite(m.Min[i]) || !finite(m.Max[i]) {
			return 0, false
		}
		margin := 0.05 * math.Max(1, m.Max[i]-m.Min[i])
		if f < m.Min[i]-margin || f > m.Max[i]+margin {
			supported = false
		}
		value += m.Weights[i] * (f - m.Means[i]) / m.Scales[i]
	}
	if !finite(value) {
		return 0, false
	}
	return logistic(value), supported
}

type riskCandidate struct {
	placement candidatePlacement
	piece     TetrominoType
	useHold   bool
	next      int
	hold      TetrominoType
	cleanup   bool
}

func riskCandidates(g *Game) []riskCandidate {
	if g.Board == nil || g.CurrentPiece == nil {
		return nil
	}
	var out []riskCandidate
	for _, useHold := range []bool{false, true} {
		p := g.CurrentPiece
		next := 0
		var held TetrominoType
		if g.HoldPiece != nil {
			held = g.HoldPiece.Type
		}
		if useHold {
			if !g.CanHold {
				continue
			}
			if held != "" {
				p = NewPiece(held)
			} else {
				if len(g.NextQueue) == 0 {
					continue
				}
				p = NewPiece(g.NextQueue[0])
				next = 1
			}
			held = g.CurrentPiece.Type
		}
		context := *g
		context.CurrentPiece = p
		context.NextQueue = g.NextQueue[next:]
		context.HoldPiece = nil
		if held != "" {
			context.HoldPiece = NewPiece(held)
		}
		cleanup := IsCleanupMode(&context)
		for _, c := range reachablePlacements(g, p, cleanup) {
			out = append(out, riskCandidate{placement: c, piece: p.Type, useHold: useHold, next: next, hold: held, cleanup: cleanup})
		}
	}
	return out
}

func filledCells(b *Board) int {
	n := 0
	for _, row := range b.Cells {
		for _, cell := range row {
			if cell.Filled {
				n++
			}
		}
	}
	return n
}

func candidateFeatures(g *Game, c riskCandidate) []float64 {
	b := c.placement.board
	lines := (filledCells(g.Board) + 4 - filledCells(b)) / BoardWidth
	flag := func(v bool) float64 {
		if v {
			return 1
		}
		return 0
	}
	f := []float64{float64(g.Board.MaxHeight()) / 20, float64(g.Board.CountHoles()) / 20, float64(g.Board.Bumpiness()) / 40, float64(b.MaxHeight()) / 20, float64(b.CenterHeight()) / 20, float64(b.CountHoles()) / 20, float64(b.Bumpiness()) / 40, float64(filledCells(b)) / 200, float64(lines) / 4, c.placement.score / 100000, flag(c.cleanup), flag(c.useHold), float64(g.Level) / 25, float64(g.remainingGravity()) / float64(time.Second)}
	for _, h := range b.ColHeights() {
		f = append(f, float64(h)/20)
	}
	next := TetrominoType("")
	if c.next < len(g.NextQueue) {
		next = g.NextQueue[c.next]
	}
	for _, t := range []TetrominoType{c.piece, next, c.hold} {
		for _, p := range AllPieces {
			f = append(f, flag(t == p))
		}
	}
	return f
}

func candidateMove(c riskCandidate) *AIMove {
	return &AIMove{UseHold: c.useHold, TargetRotation: c.placement.rotation, TargetX: c.placement.x, TargetY: c.placement.y, HasTargetY: true, Score: c.placement.score, CleanupMode: c.cleanup, Actions: c.placement.actions, Expected: c.placement.expected}
}

func matchesCandidate(m *AIMove, c riskCandidate) bool {
	return m != nil && m.UseHold == c.useHold && m.TargetX == c.placement.x && m.TargetY == c.placement.y && m.TargetRotation == c.placement.rotation
}

func findLearnedMove(g *Game) *AIMove {
	base := *g
	base.UseLearned = false
	move := findPolicyMove(&base)
	if move == nil {
		return nil
	}
	move.LearnedFallback = "no_model"
	if g.LearnedModel == nil {
		return move
	}
	if err := g.LearnedModel.Validate(); err != nil {
		move.LearnedFallback = "invalid_model"
		return move
	}
	move.LearnedFallback = "validation_not_better_than_prior"
	if g.LearnedModel.ValidationBrier >= g.LearnedModel.BaselineBrier {
		return move
	}
	cs := riskCandidates(g)
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].placement.score > cs[j].placement.score })
	var baseline float64
	found := false
	for _, c := range cs {
		if matchesCandidate(move, c) {
			p, ok := g.LearnedModel.Predict(candidateFeatures(g, c))
			if ok {
				baseline = p
				found = true
				move.MoveProbability = &baseline
			}
			break
		}
	}
	move.LearnedFallback = "unsupported_state"
	if !found {
		return move
	}
	move.LearnedFallback = "small_probability_margin"
	best := baseline
	var chosen *AIMove
	for i, c := range cs {
		if i >= 6 {
			break
		}
		p, ok := g.LearnedModel.Predict(candidateFeatures(g, c))
		if ok && p >= 0.90 && p > baseline+0.05 && p > best {
			best = p
			chosen = candidateMove(c)
		}
	}
	if chosen == nil {
		return move
	}
	chosen.MoveProbability = &best
	chosen.LearnedUsed = true
	return chosen
}

func captureDecision(g *Game, selected *AIMove) *logger.DecisionLog {
	d := &logger.DecisionLog{Version: 1, Seed: g.Seed, MoveNumber: g.MoveCount + 1, Policy: g.policyVersion(), Board: g.Board.ToStringGrid(), Piece: string(g.CurrentPiece.Type), X: g.CurrentPiece.X, Y: g.CurrentPiece.Y, Rotation: g.CurrentPiece.Rotation, CanHold: g.CanHold, Level: g.Level, Lines: g.Lines, GravityRemainingNS: int64(g.remainingGravity()), ReserveWell: g.ReserveWell}
	if g.HoldPiece != nil {
		d.Hold = string(g.HoldPiece.Type)
	}
	for _, p := range g.NextQueue {
		d.Queue = append(d.Queue, string(p))
	}
	if g.Randomizer != nil {
		for _, p := range g.Randomizer.bag {
			d.RemainingBag = append(d.RemainingBag, string(p))
		}
		sort.Strings(d.RemainingBag)
	}
	modelValid := g.LearnedModel != nil && g.LearnedModel.Validate() == nil
	if modelValid {
		d.ModelID, d.RiskHorizon, d.RiskTarget = g.LearnedModel.ID, g.LearnedModel.Horizon, g.LearnedModel.Target
	}
	for _, c := range riskCandidates(g) {
		logged := logger.CandidateLog{Piece: string(c.piece), UseHold: c.useHold, X: c.placement.x, Y: c.placement.y, Rotation: c.placement.rotation, Heuristic: c.placement.score, Selected: matchesCandidate(selected, c)}
		if modelValid {
			if p, ok := g.LearnedModel.Predict(candidateFeatures(g, c)); ok {
				logged.SurvivalProbability = &p
			}
		}
		d.Candidates = append(d.Candidates, logged)
	}
	return d
}

// GameFromDecision samples the unordered, unobserved remainder of the current
// bag. Its real order/seed is NOT used to leak future pieces into training.
func GameFromDecision(d *logger.DecisionLog, sampleSeed int64) (*Game, error) {
	validPiece := func(s string) bool {
		for _, p := range AllPieces {
			if s == string(p) {
				return true
			}
		}
		return false
	}
	if d == nil || d.Version != 1 || len(d.Board) != BoardHeight || !validPiece(d.Piece) || len(d.Queue) < 1 || len(d.Queue) > 10 || d.Level < 1 || d.Level != d.Lines/10+1 || d.Lines < 0 || d.GravityRemainingNS < 0 || d.GravityRemainingNS > int64(time.Second) || d.Rotation < 0 || d.Rotation > 3 {
		return nil, fmt.Errorf("invalid decision snapshot")
	}
	g := NewGameWithSeed(sampleSeed)
	g.Board = NewBoard()
	for y, row := range d.Board {
		if len(row) != BoardWidth || strings.Trim(row, "01") != "" {
			return nil, fmt.Errorf("invalid decision board")
		}
		for x, v := range row {
			g.Board.Cells[y][x].Filled = v == '1'
		}
	}
	g.CurrentPiece = NewPiece(TetrominoType(d.Piece))
	g.CurrentPiece.X, g.CurrentPiece.Y, g.CurrentPiece.Rotation = d.X, d.Y, d.Rotation
	if !g.Board.IsValidPosition(g.CurrentPiece) {
		return nil, fmt.Errorf("invalid current pose")
	}
	g.HoldPiece = nil
	if d.Hold != "" {
		if !validPiece(d.Hold) {
			return nil, fmt.Errorf("invalid hold")
		}
		g.HoldPiece = NewPiece(TetrominoType(d.Hold))
	}
	g.NextQueue = nil
	for _, p := range d.Queue {
		if !validPiece(p) {
			return nil, fmt.Errorf("invalid preview")
		}
		g.NextQueue = append(g.NextQueue, TetrominoType(p))
	}
	r := &Randomizer{rng: rand.New(rand.NewSource(sampleSeed))}
	seen := map[string]bool{}
	for _, p := range d.RemainingBag {
		if !validPiece(p) || seen[p] {
			return nil, fmt.Errorf("invalid remaining bag")
		}
		seen[p] = true
		r.bag = append(r.bag, TetrominoType(p))
	}
	r.rng.Shuffle(len(r.bag), func(i, j int) { r.bag[i], r.bag[j] = r.bag[j], r.bag[i] })
	g.Randomizer = r
	g.Seed = d.Seed
	g.Level, g.Lines = d.Level, d.Lines
	g.CanHold = d.CanHold
	g.AutoPlay = true
	g.ReserveWell = d.ReserveWell
	g.SetAIGravityRemaining(time.Duration(d.GravityRemainingNS))
	return g, nil
}
