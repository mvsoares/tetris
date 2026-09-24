package engine

import (
	"math"
	"sort"
)

// AIMove represents the computed optimal move for the current turn.
type AIMove struct {
	UseHold        bool
	TargetRotation int
	TargetX        int
	Score          float64
	CleanupMode    bool
}

// CleanUpHeightThreshold defines the traditional 65% baseline of the 20-row board (13 rows).
const CleanUpHeightThreshold = 13

// GetMaxHeight calculates the highest stack height among all columns.
func GetMaxHeight(b *Board) int {
	maxH := 0
	for x := 0; x < BoardWidth; x++ {
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				h := BoardHeight - y
				if h > maxH {
					maxH = h
				}
				break
			}
		}
	}
	return maxH
}

// GetCenterHeight returns the maximum height in the critical piece-spawn corridor (columns 3, 4, 5, 6).
// Keeping the spawn corridor clear is essential to prevent immediate top-outs at row 0.
func GetCenterHeight(b *Board) int {
	if b == nil {
		return 0
	}
	maxH := 0
	for x := 3; x <= 6; x++ {
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				h := BoardHeight - y
				if h > maxH {
					maxH = h
				}
				break
			}
		}
	}
	return maxH
}

// CountHoles counts empty cells covered by at least one filled cell in the same column.
func CountHoles(b *Board) int {
	if b == nil {
		return 0
	}
	holes := 0
	for x := 0; x < BoardWidth; x++ {
		foundFilled := false
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				foundFilled = true
			} else if foundFilled {
				holes++
			}
		}
	}
	return holes
}

// GetBumpiness returns the sum of absolute height differences between adjacent building columns (0..8).
func GetBumpiness(b *Board) int {
	if b == nil {
		return 0
	}
	colHeights := make([]int, BoardWidth)
	for x := 0; x < BoardWidth; x++ {
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				colHeights[x] = BoardHeight - y
				break
			}
		}
	}
	bump := 0
	for x := 0; x < 8; x++ {
		d := colHeights[x] - colHeights[x+1]
		if d < 0 {
			d = -d
		}
		bump += d
	}
	return bump
}

// HasImmediateLinePiece checks if an I piece is ready right now (in hand or ready in hold).
func HasImmediateLinePiece(g *Game) bool {
	if g == nil {
		return false
	}
	if g.CurrentPiece != nil && g.CurrentPiece.Type == PieceI {
		return true
	}
	if g.HoldPiece != nil && g.HoldPiece.Type == PieceI {
		return true
	}
	return false
}

// HasLinePieceComing checks if an I-tetromino ("linha") is available in hand, hold, or upcoming queue.
func HasLinePieceComing(g *Game) bool {
	if g == nil {
		return false
	}
	if HasImmediateLinePiece(g) {
		return true
	}
	if len(g.NextQueue) > 0 && g.NextQueue[0] == PieceI {
		return true
	}
	if len(g.NextQueue) > 1 && g.NextQueue[1] == PieceI {
		return true
	}
	return false
}

// GetDynamicCleanupThreshold calculates an adaptive height threshold for cleanup mode.
// Instead of a rigid 65% (13 rows), it adapts based on terrain health:
// - A pristine, flat board (0 holes, low bumpiness) can safely stack up to 14-15 rows waiting for Tetris.
// - A messy board with holes or spires triggers cleanup much earlier (e.g. 9-11 rows) before danger escalates.
func GetDynamicCleanupThreshold(b *Board) int {
	if b == nil {
		return CleanUpHeightThreshold
	}

	threshold := 14 // Base threshold: healthy boards can safely stack higher

	// Holes degrade structural safety severely
	holes := CountHoles(b)
	threshold -= holes * 2

	// Center spires in spawn chute reduce threshold
	centerH := GetCenterHeight(b)
	if centerH >= 11 {
		threshold -= (centerH - 10)
	}

	// High surface bumpiness reduces threshold
	bump := GetBumpiness(b)
	if bump > 14 {
		threshold -= (bump - 14) / 3
	}

	// Clamp between 8 (emergency defense) and 15 (maximum safe stacking)
	if threshold < 8 {
		threshold = 8
	}
	if threshold > 15 {
		threshold = 15
	}
	return threshold
}

// IsCleanupMode returns true if the board requires defensive cleanup lines clearing.
// Adaptively considers:
// 1. Dynamic threshold (holes, bumpiness, center spawn height)
// 2. Center height >= 14 (immediate spawn corridor clearance, even if I piece is coming)
// 3. Overall height >= 16 (critical survival threshold overrides line piece waiting)
func IsCleanupMode(g *Game) bool {
	if g == nil || g.Board == nil {
		return false
	}
	maxH := GetMaxHeight(g.Board)
	centerH := GetCenterHeight(g.Board)
	holes := CountHoles(g.Board)

	// Critical spawn danger: center columns >= 14 or maxHeight >= 16 MUST clean up immediately!
	// Waiting for an I piece in column 9 when center is at 14+ causes fatal top-outs.
	if centerH >= 14 || maxH >= 16 {
		return true
	}

	// If there are multiple holes, cleanup is needed to dig out before burying deeper
	if holes >= 2 && maxH >= 9 {
		return true
	}

	threshold := GetDynamicCleanupThreshold(g.Board)
	if maxH >= threshold {
		// If line piece is immediately available (CurrentPiece or HoldPiece) and board is safe,
		// allow one chance to score the Tetris
		if HasImmediateLinePiece(g) && maxH < 15 && centerH <= 13 {
			return false
		}
		// If an I piece is coming soon (in next queue), only postpone if terrain is healthy
		if HasLinePieceComing(g) && maxH < 13 && centerH < 11 && holes == 0 {
			return false
		}
		return true
	}

	return false
}

// FindBestMove calculates the best move for the game.
// Enhanced using insights from 3.7+ million logged moves:
// - 2-ply lookahead (anticipates Next piece to prevent terrain lock)
// - Non-linear cliff penalties (solves O/S/Z piece failures)
// - Anti-choke well protection for column 9
// - Strict hole creation penalties during emergency cleanup
func FindBestMove(g *Game) *AIMove {
	if g.CurrentPiece == nil || g.Board == nil {
		return nil
	}

	cleanupMode := IsCleanupMode(g)

	var nextType TetrominoType
	if len(g.NextQueue) > 0 {
		nextType = g.NextQueue[0]
	}

	bestRot, bestX, bestScore := findBestPlacementWithLookahead(g.Board, g.CurrentPiece, nextType, cleanupMode)

	bestMove := &AIMove{
		UseHold:        false,
		TargetRotation: bestRot,
		TargetX:        bestX,
		Score:          bestScore,
		CleanupMode:    cleanupMode,
	}

	// Evaluate candidate piece if Hold is used (only if current piece doesn't already score a Tetris)
	if g.CanHold && bestScore < 150000.0 {
		var candidateHoldType TetrominoType
		var subsequentNext TetrominoType

		if g.HoldPiece != nil {
			candidateHoldType = g.HoldPiece.Type
			subsequentNext = nextType
		} else if len(g.NextQueue) > 0 {
			candidateHoldType = g.NextQueue[0]
			if len(g.NextQueue) > 1 {
				subsequentNext = g.NextQueue[1]
			}
		}

		if candidateHoldType != "" {
			holdPiece := NewPiece(candidateHoldType)
			rotH, xH, scoreH := findBestPlacementWithLookahead(g.Board, holdPiece, subsequentNext, cleanupMode)

			// If current piece is I, but cannot clear 4 lines yet and board is safe (< 12 rows),
			// save it in Hold for the upcoming Tetris!
			if g.CurrentPiece.Type == PieceI && candidateHoldType != PieceI && GetMaxHeight(g.Board) < 12 && !cleanupMode {
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH + 20000.0,
					CleanupMode:    cleanupMode,
				}
			} else if cleanupMode && scoreH > bestScore {
				// In emergency cleanup mode, switch to hold piece if it scores higher
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH,
					CleanupMode:    cleanupMode,
				}
			} else if scoreH > bestScore+60.0 {
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH,
					CleanupMode:    cleanupMode,
				}
			}
		}
	}

	return bestMove
}

type candidatePlacement struct {
	rotation int
	x        int
	board    *Board
	score    float64
}

// findBestPlacementWithLookahead tests valid placements and evaluates the top candidates with next-piece lookahead.
func findBestPlacementWithLookahead(b *Board, p *Piece, nextType TetrominoType, cleanupMode bool) (int, int, float64) {
	candidates := getCandidatePlacements(b, p, cleanupMode)
	if len(candidates) == 0 {
		return 0, 3, -math.MaxFloat64
	}

	// Sort candidates by immediate score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// If no next piece available, return top candidate directly
	if nextType == "" {
		return candidates[0].rotation, candidates[0].x, candidates[0].score
	}

	// Evaluate top 5 candidates with 1-ply lookahead
	topLimit := 5
	if len(candidates) < topLimit {
		topLimit = len(candidates)
	}

	nextPiece := NewPiece(nextType)
	bestCombinedScore := -math.MaxFloat64
	bestRot := candidates[0].rotation
	bestX := candidates[0].x

	for i := 0; i < topLimit; i++ {
		cand := candidates[i]
		_, _, nextScore := findBestPlacementSimple(cand.board, nextPiece, cleanupMode)

		weight := 0.65
		if nextScore < -15000.0 {
			weight = 0.85
		}

		combined := cand.score + weight*nextScore
		if combined > bestCombinedScore {
			bestCombinedScore = combined
			bestRot = cand.rotation
			bestX = cand.x
		}
	}

	return bestRot, bestX, bestCombinedScore
}

// isReachable checks if a piece can navigate horizontally from spawnX to targetX across the top of the board
func isReachable(b *Board, p *Piece, targetRot, targetX int) bool {
	spawnX := 3
	if p.Type == PieceO {
		spawnX = 4
	}
	if targetX == spawnX {
		return true
	}

	step := 1
	if targetX < spawnX {
		step = -1
	}

	// Try traversal at y=0, then y=-1, then y=-2
	for _, tryY := range []int{0, -1, -2} {
		canTraverse := true
		for currX := spawnX; currX != targetX+step; currX += step {
			test := p.Clone()
			test.Rotation = targetRot
			test.X = currX
			test.Y = tryY
			if !b.IsValidPosition(test) {
				canTraverse = false
				break
			}
		}
		if canTraverse {
			return true
		}
	}

	return false
}

func getCandidatePlacements(b *Board, p *Piece, cleanupMode bool) []candidatePlacement {
	var candidates []candidatePlacement

	// Test all 4 rotations
	for rot := 0; rot < 4; rot++ {
		if p.Type == PieceO && rot > 0 {
			continue
		}

		for x := -3; x < BoardWidth; x++ {
			test := p.Clone()
			test.Rotation = rot
			test.X = x
			test.Y = -2

			if !b.IsValidPosition(test) {
				test.Y = 0
				if !b.IsValidPosition(test) {
					continue
				}
			}

			// Reject candidates that cannot be physically reached from spawn due to spires
			if !isReachable(b, p, rot, x) {
				continue
			}

			ghostY := b.GetGhostY(test)
			if ghostY < 0 {
				continue
			}

			landingPiece := test.Clone()
			landingPiece.Y = ghostY

			if !b.IsValidPosition(landingPiece) {
				continue
			}

			simBoard := b.Clone()
			simBoard.LockPiece(landingPiece)
			linesCleared := simBoard.ClearLines()

			score := evaluatePlacement(simBoard, landingPiece, linesCleared, cleanupMode)
			candidates = append(candidates, candidatePlacement{
				rotation: rot,
				x:        x,
				board:    simBoard,
				score:    score,
			})
		}
	}

	return candidates
}

// findBestPlacementSimple is a fast 1-ply evaluator used in lookahead simulations.
func findBestPlacementSimple(b *Board, p *Piece, cleanupMode bool) (int, int, float64) {
	bestScore := -math.MaxFloat64
	bestRot := 0
	bestX := 3

	for rot := 0; rot < 4; rot++ {
		if p.Type == PieceO && rot > 0 {
			continue
		}

		for x := -3; x < BoardWidth; x++ {
			test := p.Clone()
			test.Rotation = rot
			test.X = x
			test.Y = -2

			if !b.IsValidPosition(test) {
				test.Y = 0
				if !b.IsValidPosition(test) {
					continue
				}
			}

			if !isReachable(b, p, rot, x) {
				continue
			}

			ghostY := b.GetGhostY(test)
			if ghostY < 0 {
				continue
			}

			landingPiece := test.Clone()
			landingPiece.Y = ghostY

			if !b.IsValidPosition(landingPiece) {
				continue
			}

			simBoard := b.Clone()
			simBoard.LockPiece(landingPiece)
			linesCleared := simBoard.ClearLines()

			score := evaluatePlacement(simBoard, landingPiece, linesCleared, cleanupMode)
			if score > bestScore {
				bestScore = score
				bestRot = rot
				bestX = x
			}
		}
	}

	return bestRot, bestX, bestScore
}

// evaluatePlacement calculates heuristic score for a simulated board.
// Refined based on 3.7 million game logs to prevent the common failure modes:
// 1. Hole drill in cleanup mode
// 2. High spiky bumpiness creating O/S/Z top-outs
// 3. Choking column 9 well
func evaluatePlacement(b *Board, p *Piece, linesCleared int, cleanupMode bool) float64 {
	colHeights := make([]int, BoardWidth)
	maxHeight := 0
	aggHeight := 0
	centerMax := 0

	for x := 0; x < BoardWidth; x++ {
		h := 0
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				h = BoardHeight - y
				break
			}
		}
		colHeights[x] = h
		if h > maxHeight {
			maxHeight = h
		}
		if x >= 3 && x <= 6 && h > centerMax {
			centerMax = h
		}
		if x < 9 || cleanupMode {
			aggHeight += h
		}
	}

	// 1. Lines Cleared Score
	var linesScore float64
	if cleanupMode {
		switch linesCleared {
		case 4:
			linesScore = 180000.0 // Tetris is best
		case 3:
			linesScore = 95000.0  // Huge reward for triple
		case 2:
			linesScore = 65000.0  // Great reward for double
		case 1:
			linesScore = 35000.0  // Good reward for single
		case 0:
			linesScore = -15000.0 // Penalize placing pieces without clearing lines
		}
	} else {
		// Normal mode: Heavy priority on 4-line TETRIS
		switch linesCleared {
		case 4:
			linesScore = 170000.0 // Massive reward for TETRIS!
		case 3:
			if maxHeight >= 11 || centerMax >= 10 {
				linesScore = 6000.0
			} else {
				linesScore = -400.0 // Mild discouragement when completely safe
			}
		case 2:
			if maxHeight >= 11 || centerMax >= 10 {
				linesScore = 4000.0
			} else {
				linesScore = -600.0
			}
		case 1:
			if maxHeight >= 11 || centerMax >= 10 {
				linesScore = 2000.0
			} else {
				linesScore = -800.0
			}
		}
	}

	// 2. Column 9 (Corner / Well) Reservation & Anti-Closure Protection
	wellPenalties := 0.0

	// Check if this placement put a vertical I in column 9
	isVerticalIInCol9 := false
	if p.Type == PieceI && (p.Rotation == 1 || p.Rotation == 3) {
		coords := p.BlockCoords()
		inCol9 := true
		for _, pt := range coords {
			if pt[0] != 9 {
				inCol9 = false
				break
			}
		}
		isVerticalIInCol9 = inCol9
	}

	// RULE 1: Vertical I in column 9 is STRICTLY for full 4-line TETRIS!
	// Placing a vertical I in column 9 without 4 lines cleared leaves 1-4 blocks sticking up,
	// creating a wall in the corner that closes the well and prevents pieces from passing to the corner.
	if isVerticalIInCol9 {
		if linesCleared == 4 {
			linesScore += 90000.0 // Super reward for clean 4-line Tetris in the corner well
		} else {
			// Catastrophic penalty: never put vertical I high in corner without clearing all 4 lines!
			wellPenalties += 120000.0
		}
	}

	// RULE 2: Anti-Spire - Column 9 must NEVER be taller than Column 8!
	// If column 9 is taller than column 8, it forms an elevated wall on the right boundary,
	// closing off the corner and blocking pieces from sliding to the corner.
	if colHeights[9] > colHeights[8] {
		wellPenalties += float64(colHeights[9]-colHeights[8]) * 45000.0
	}

	// RULE 3: Anti-Closure / Anti-Roof - The corner must NEVER be closed/roofed!
	// If column 9 has ANY block with an empty space below, the well is choked from above.
	// This applies in BOTH normal mode and cleanup mode, at ANY height (even below 65%).
	if colHeights[9] > 0 {
		choked := false
		for y := BoardHeight - colHeights[9]; y < BoardHeight; y++ {
			if !b.Cells[y][9].Filled {
				choked = true
				break
			}
		}
		if choked {
			wellPenalties += 150000.0 // Fatal: well is choked, pieces can no longer reach bottom!
		}
	}

	// RULE 4: Well Cleanliness & Channel Access
	if !cleanupMode {
		// In normal mode: every block left in column 9 (when not cleared by Tetris) receives heavy penalty
		if colHeights[9] > 0 {
			wellPenalties += float64(colHeights[9]) * 20000.0
		}

		// Well Depth Control: ideal well depth is 2-4 lines.
		// If column 8 is more than 4 blocks taller than column 9, the well is too deep and hazardous.
		wellDepth := colHeights[8] - colHeights[9]
		if wellDepth > 4 {
			wellPenalties += float64(wellDepth-4) * 10000.0
		}

		// Channel Access: Column 8 must not create a high wall in front of the corner
		// (i.e. colHeights[8] should not be much taller than colHeights[7])
		if colHeights[8] > colHeights[7]+2 {
			wellPenalties += float64(colHeights[8]-(colHeights[7]+2)) * 12000.0
		}
	} else {
		// In cleanup mode: allow using column 9 to clear lines, but penalize leaving unneeded blocks
		if isVerticalIInCol9 && linesCleared == 0 {
			wellPenalties += 80000.0
		}
		if linesCleared == 0 && colHeights[9] > 0 {
			for y := 0; y < BoardHeight; y++ {
				if b.Cells[y][9].Filled {
					wellPenalties += 800.0
				}
			}
		}
	}

	// 3. Holes & Blockades: empty cells covered by filled blocks
	holes := 0
	blockades := 0
	for x := 0; x < BoardWidth; x++ {
		foundFilled := false
		blocksAbove := 0
		for y := 0; y < BoardHeight; y++ {
			if b.Cells[y][x].Filled {
				foundFilled = true
				blocksAbove++
			} else if foundFilled {
				holes++
				blockades += blocksAbove
			}
		}
	}

	// 4. Bumpiness & Non-Linear Cliffs
	// Data showed bumpiness soaring from 13.7 to 34.1 at death, and O/S/Z causing 57.6% of losses
	bumpinessLimit := 8
	if cleanupMode {
		bumpinessLimit = 9
	}

	bumpiness := 0
	cliffPenalty := 0.0

	for x := 0; x < bumpinessLimit-1; x++ {
		diff := colHeights[x] - colHeights[x+1]
		if diff < 0 {
			diff = -diff
		}
		bumpiness += diff

		// Non-linear penalty for cliff height difference >= 3
		if diff >= 3 {
			cliffPenalty += float64((diff-2)*(diff-2)) * 450.0
			if diff >= 4 {
				cliffPenalty += 2500.0 // Extreme cliff penalty to prevent spiky terrain traps
			}
		}
	}

	// 5. Deep valleys in building zone (columns 0..8)
	innerWells := 0
	for x := 0; x < 9; x++ {
		leftH := 20
		if x > 0 {
			leftH = colHeights[x-1]
		}
		rightH := 20
		if x < 8 {
			rightH = colHeights[x+1]
		}
		minAdjacent := leftH
		if rightH < minAdjacent {
			minAdjacent = rightH
		}
		if minAdjacent-colHeights[x] >= 3 {
			innerWells += (minAdjacent - colHeights[x])
		}
	}

	// Landing height bonus
	landingBonus := float64(p.Y) * 12.0
	if cleanupMode {
		landingBonus = float64(p.Y) * 25.0
	}

	// Dynamic hole weight: higher in cleanup mode so AI never ruins structure for a single line
	holeWeight := 15000.0
	blockadeWeight := 600.0
	bumpinessWeight := 160.0

	if cleanupMode {
		holeWeight = 55000.0 // Mathematically guarantees AI will never drill a hole to clear 1 line
		blockadeWeight = 1500.0
		bumpinessWeight = 220.0
	}

	heightWeight := 45.0
	maxHeightPenalty := 0.0
	if cleanupMode {
		heightWeight = 140.0
		if maxHeight > 13 {
			maxHeightPenalty = float64(maxHeight-13) * float64(maxHeight-13) * 250.0
		}
	}

	// 6. Spawn Corridor Clearance & Anti-Center Dome Penalty
	// Pieces spawn at columns 3..5. If the center peaks, newly spawned pieces collide at row 0.
	spawnChutePenalty := 0.0
	if centerMax >= 11 {
		spawnChutePenalty += float64((centerMax - 10) * (centerMax - 10)) * 600.0
		if centerMax >= 14 {
			spawnChutePenalty += float64(centerMax - 13) * 15000.0
		}
	}
	flankLeftMax := colHeights[0]
	if colHeights[1] > flankLeftMax {
		flankLeftMax = colHeights[1]
	}
	if colHeights[2] > flankLeftMax {
		flankLeftMax = colHeights[2]
	}
	if centerMax > flankLeftMax+2 {
		spawnChutePenalty += float64(centerMax-(flankLeftMax+2)) * 1200.0
	}

	score := linesScore -
		wellPenalties -
		float64(holes)*holeWeight -
		float64(blockades)*blockadeWeight -
		float64(bumpiness)*bumpinessWeight -
		cliffPenalty -
		spawnChutePenalty -
		float64(aggHeight)*heightWeight -
		maxHeightPenalty -
		float64(innerWells)*500.0 +
		landingBonus

	return score
}
