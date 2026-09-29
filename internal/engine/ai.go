package engine

import (
	"math"
	"sort"

	"tetris/internal/logger"
)

// AIMove represents the computed optimal move for the current turn.
type AIMove struct {
	UseHold         bool
	TargetRotation  int
	TargetX         int
	Score           float64
	CleanupMode     bool
	TargetY         int
	HasTargetY      bool
	Actions         []AIAction
	Expected        []Piece
	SearchDepth     int
	SearchNodes     int
	Decision        *logger.DecisionLog
	MoveProbability *float64
	LearnedUsed     bool
	LearnedFallback string
}

// CleanUpHeightThreshold defines the traditional 65% baseline of the 20-row board (13 rows).
const CleanUpHeightThreshold = 13

// HighSpeedCleanupThreshold defines the 40% board height threshold (8 rows out of 20)
// used when high-speed awareness is locked in so pieces have ample vertical clearance
// to traverse to the left corridor (columns 0..2) without colliding or timing out.
const HighSpeedCleanupThreshold = 8

// HighSpeedAwarenessStartLevel is the level where the AI begins progressively
// lowering its stacking ceiling from 65% (13-14 rows) toward 40% (8 rows).
const HighSpeedAwarenessStartLevel = 50

// HighSpeedModeLevel is the optimal level (25 levels / ~250 lines before Level 100)
// where the 40% (8-row) threshold and simpler line-clearing play lock in, ensuring
// the stack is already low and flat before entering Level 100+.
const HighSpeedModeLevel = 75

// IsHighSpeedLevel returns true once the level reaches HighSpeedModeLevel.
func IsHighSpeedLevel(level int) bool {
	return level >= HighSpeedModeLevel
}

// IsHighSpeedMode returns true when the game is at or above HighSpeedModeLevel.
func IsHighSpeedMode(g *Game) bool {
	return g != nil && IsHighSpeedLevel(g.Level)
}

// MaxCleanupThresholdForLevel returns the maximum stacking height before cleanup mode
// engages for a given level:
//   - Level < 10: up to 14 rows (65-70% of board)
//   - Level 10..12: 13 rows as gravity accelerates toward 80ms
//   - Level 13..49: 12 rows (at 80ms max gravity, vertical I-pieces need >= 5 rows of top clearance to enter col 9)
//   - Level 50..74: smooth linear ramp from 12 rows down to 8 rows
//   - Level >= 75: 8 rows (40% of the 20-row board)
func MaxCleanupThresholdForLevel(level int) int {
	if level >= HighSpeedModeLevel {
		return HighSpeedCleanupThreshold
	}
	if level >= HighSpeedAwarenessStartLevel {
		span := HighSpeedModeLevel - HighSpeedAwarenessStartLevel
		progress := level - HighSpeedAwarenessStartLevel
		return 12 - (progress*4)/span
	}
	if level >= 13 {
		return 12
	}
	if level >= 10 {
		return 13
	}
	return 14
}

// CleanupThresholdPercent returns the active Tetris/cleanup height threshold as a
// board percentage (40% in high-speed mode, 65% at normal speed, or interpolated during ramp).
func CleanupThresholdPercent(g *Game) int {
	if g == nil {
		return 65
	}
	if IsHighSpeedMode(g) {
		return 40
	}
	if g.Level >= HighSpeedAwarenessStartLevel {
		return (MaxCleanupThresholdForLevel(g.Level) * 100) / BoardHeight
	}
	return 65
}

// GetMaxHeight calculates the highest stack height among all columns.
func GetMaxHeight(b *Board) int {
	if b == nil {
		return 0
	}
	return b.MaxHeight()
}

// GetCenterHeight returns the maximum height in the critical piece-spawn corridor (columns 3, 4, 5, 6).
// Keeping the spawn corridor clear is essential to prevent immediate top-outs at row 0.
func GetCenterHeight(b *Board) int {
	if b == nil {
		return 0
	}
	return b.CenterHeight()
}

// CountHoles counts empty cells covered by at least one filled cell in the same column.
func CountHoles(b *Board) int {
	if b == nil {
		return 0
	}
	return b.CountHoles()
}

// GetBumpiness returns the sum of absolute height differences between adjacent building columns (0..8).
func GetBumpiness(b *Board) int {
	if b == nil {
		return 0
	}
	colHeights := b.ColHeights()
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
// - A messy board with holes, spires, or high bumpiness triggers cleanup much earlier before danger escalates.
func GetDynamicCleanupThreshold(b *Board) int {
	return GetDynamicCleanupThresholdForLevel(b, 1)
}

// GetDynamicCleanupThresholdForLevel calculates the adaptive cleanup threshold capped by level speed awareness.
func GetDynamicCleanupThresholdForLevel(b *Board, level int) int {
	if b == nil {
		if IsHighSpeedLevel(level) {
			return HighSpeedCleanupThreshold
		}
		return CleanUpHeightThreshold
	}
	bb := b.ToBitBoard()
	return getDynamicCleanupThresholdBitBoardWithLevel(bb, level, bb.MaxHeight(), bb.CenterHeight(), bb.CountHoles(), bb.Bumpiness8())
}

// IsCleanupMode returns true if the board requires defensive cleanup lines clearing.
// Adaptively considers:
// 1. Dynamic threshold (holes, bumpiness, center spawn height, and level-aware 40% high-speed cap)
// 2. Critical spawn danger (center >= 14 or maxHeight >= 16)
// 3. Early bumpiness / hole spikes triggering defense before traps form
// 4. Distinguishes between immediate line piece (in hand/hold) vs waiting for future pieces in queue
func IsCleanupMode(g *Game) bool {
	if g == nil || g.Board == nil {
		return false
	}
	var currentType, holdType TetrominoType
	if g.CurrentPiece != nil {
		currentType = g.CurrentPiece.Type
	}
	if g.HoldPiece != nil {
		holdType = g.HoldPiece.Type
	}
	return isCleanupModeBitBoardWithLevel(g.Board.ToBitBoard(), g.Level, currentType, holdType, g.NextQueue)
}

func getDynamicCleanupThresholdBitBoard(bb BitBoard, maxH, centerH, holes, bump int) int {
	return getDynamicCleanupThresholdBitBoardWithLevel(bb, 1, maxH, centerH, holes, bump)
}

func getDynamicCleanupThresholdBitBoardWithLevel(bb BitBoard, level, maxH, centerH, holes, bump int) int {
	maxCap := MaxCleanupThresholdForLevel(level)
	threshold := maxCap
	threshold -= holes * 3
	if centerH >= 9 {
		threshold -= (centerH - 8)
	}
	if bump > 8 {
		threshold -= (bump - 7)
	}
	minClamp := 7
	if IsHighSpeedLevel(level) {
		minClamp = 6
	}
	if threshold < minClamp {
		threshold = minClamp
	}
	maxClamp := 15
	if level >= 10 {
		maxClamp = maxCap
	}
	if threshold > maxClamp {
		threshold = maxClamp
	}
	return threshold
}

func isCleanupModeBitBoard(bb BitBoard, currentType, holdType TetrominoType, nextQueue []TetrominoType) bool {
	return isCleanupModeBitBoardWithLevel(bb, 1, currentType, holdType, nextQueue)
}

func isCleanupModeBitBoardWithLevel(bb BitBoard, level int, currentType, holdType TetrominoType, nextQueue []TetrominoType) bool {
	maxH := bb.MaxHeight()
	centerH := bb.CenterHeight()
	holes := bb.CountHoles()
	bump := bb.Bumpiness8()
	highSpeed := IsHighSpeedLevel(level)
	colHeights := bb.ColHeights()

	if centerH >= 14 || maxH >= 16 {
		return true
	}

	minH08 := colHeights[0]
	rightCorridorMax := colHeights[5]
	for x := 1; x <= 8; x++ {
		if colHeights[x] < minH08 {
			minH08 = colHeights[x]
		}
		if x >= 5 && colHeights[x] > rightCorridorMax {
			rightCorridorMax = colHeights[x]
		}
	}

	if highSpeed {
		// At high speed (40% threshold = 8 rows), trigger cleanup earlier on holes, bumpiness, or left-corridor walls
		if holes >= 1 && maxH >= 6 {
			return true
		}
		if bump >= 10 && maxH >= 7 {
			return true
		}
		if (colHeights[1] >= HighSpeedCleanupThreshold || colHeights[2] >= HighSpeedCleanupThreshold) &&
			(colHeights[1] > colHeights[0]+1 || colHeights[2] > colHeights[0]+2) {
			return true
		}
	} else {
		if holes >= 1 && maxH >= 8 {
			return true
		}
		if bump >= 12 && maxH >= 8 {
			return true
		}
		// Double-well deadlock prevention: if col 0 (or any col 0..8) has a deep canyon while stack is elevated,
		// trigger cleanup before the surrounding walls grow too tall to clear.
		if maxH >= 11 && maxH-minH08 >= 5 && colHeights[9] == 0 {
			return true
		}
		// At Level >= 13 (80ms gravity), once cols 5..8 reach height 14+, vertical I-pieces can no longer
		// cross into column 9. Trigger cleanup at height 13 if a Tetris cannot be immediately scored.
		if level >= 13 && rightCorridorMax >= 13 && (minH08 < 4 || holes > 0) {
			return true
		}
	}

	threshold := getDynamicCleanupThresholdBitBoardWithLevel(bb, level, maxH, centerH, holes, bump)
	if maxH >= threshold {
		immediateI := currentType == PieceI || holdType == PieceI
		tetrisReady := holes == 0 && minH08 >= 3 && colHeights[9] == 0
		if highSpeed {
			// In 40% high-speed mode, only defer cleanup if an I-piece is immediately available
			// and the board is still within 1 row of the 40% threshold with 0 holes.
			if immediateI && tetrisReady && maxH <= HighSpeedCleanupThreshold+1 && centerH <= HighSpeedCleanupThreshold {
				return false
			}
			return true
		}
		if level >= 13 {
			if immediateI && tetrisReady && maxH <= 13 && centerH <= 12 && rightCorridorMax <= 12 {
				return false
			}
			return true
		}
		if immediateI && tetrisReady && maxH < 15 && centerH <= 13 {
			return false
		}
		comingI := immediateI || (len(nextQueue) > 0 && nextQueue[0] == PieceI) || (len(nextQueue) > 1 && nextQueue[1] == PieceI)
		if comingI && tetrisReady && maxH < 12 && centerH <= 10 && bump < 10 {
			return false
		}
		return true
	}

	return false
}

// FindBestMove ranks legal, gravity-aware routes for the current/held piece.
// Future-piece lookahead remains a faster placement approximation; each future
// piece is replanned from its actual position when it becomes active.
func FindBestMove(g *Game) *AIMove {
	move := findPolicyMove(g)
	if move != nil && g.Logger != nil {
		move.Decision = captureDecision(g, move)
	}
	return move
}

func findPolicyMove(g *Game) *AIMove {
	if g.CurrentPiece == nil || g.Board == nil {
		return nil
	}
	if g.UseLearned {
		return findLearnedMove(g)
	}
	if g.LookaheadDepth > 0 {
		return findBeamMove(g)
	}

	cleanupMode := IsCleanupMode(g)
	highSpeed := g.Level >= HighSpeedAwarenessStartLevel
	beforeHoles := CountHoles(g.Board)
	maxH := GetMaxHeight(g.Board)
	bump := GetBumpiness(g.Board)

	var nextType TetrominoType
	if len(g.NextQueue) > 0 {
		nextType = g.NextQueue[0]
	}
	var nextNextType TetrominoType
	if len(g.NextQueue) > 1 {
		nextNextType = g.NextQueue[1]
	}

	bestCandidate, bestScore := rankPlacements(reachablePlacements(g, g.CurrentPiece, cleanupMode), nextType, nextNextType, cleanupMode, !g.ReserveWell, highSpeed)
	bestRot, bestX := bestCandidate.rotation, bestCandidate.x
	var holdCandidate candidatePlacement

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
		var subsequentNextNext TetrominoType

		if g.HoldPiece != nil {
			candidateHoldType = g.HoldPiece.Type
			subsequentNext = nextType
			if len(g.NextQueue) > 1 {
				subsequentNextNext = g.NextQueue[1]
			}
		} else if len(g.NextQueue) > 0 {
			candidateHoldType = g.NextQueue[0]
			if len(g.NextQueue) > 1 {
				subsequentNext = g.NextQueue[1]
			}
			if len(g.NextQueue) > 2 {
				subsequentNextNext = g.NextQueue[2]
			}
		}

		if candidateHoldType != "" {
			holdPiece := NewPiece(candidateHoldType)
			var scoreH float64
			holdCandidate, scoreH = rankPlacements(reachablePlacements(g, holdPiece, cleanupMode), subsequentNext, subsequentNextNext, cleanupMode, !g.ReserveWell, highSpeed)
			rotH, xH := holdCandidate.rotation, holdCandidate.x

			// If current piece is I, but cannot clear 4 lines yet and board is safe,
			// save it in Hold for the upcoming Tetris provided the held piece doesn't create holes or ruin the stack!
			isSZ := g.CurrentPiece.Type == PieceS || g.CurrentPiece.Type == PieceZ
			candidateNotSZ := candidateHoldType != PieceS && candidateHoldType != PieceZ
			terrainStrained := maxH >= 9 || bump >= 8 || beforeHoles > 0
			if IsHighSpeedMode(g) {
				terrainStrained = maxH >= 6 || bump >= 6 || beforeHoles > 0
			}

			isO := g.CurrentPiece.Type == PieceO
			candidateNotO := candidateHoldType != PieceO && candidateHoldType != PieceS && candidateHoldType != PieceZ
			oStrained := isO && candidateNotO && (bump >= 12 || maxH >= 10)
			if IsHighSpeedMode(g) {
				oStrained = isO && candidateNotO && (bump >= 8 || maxH >= 7)
			}

			isLJ := g.CurrentPiece.Type == PieceL || g.CurrentPiece.Type == PieceJ
			ljStrained := isLJ && candidateNotO && candidateHoldType != PieceI && (bump >= 12 || maxH >= 10)
			if IsHighSpeedMode(g) {
				ljStrained = isLJ && candidateNotO && candidateHoldType != PieceI && (bump >= 8 || maxH >= 7)
			}

			beforeBits := g.Board.ToBitBoard()
			beforeInnerWell := maxInnerWell08(beforeBits)

			holdBits := holdCandidate.bits
			if !holdCandidate.hasBits && holdCandidate.board != nil {
				holdBits = holdCandidate.board.ToBitBoard()
			}
			currBits := bestCandidate.bits
			if !bestCandidate.hasBits && bestCandidate.board != nil {
				currBits = bestCandidate.board.ToBitBoard()
			}
			currHoles := currBits.CountHoles()
			holdHoles := holdBits.CountHoles()
			currInnerWell := maxInnerWell08(currBits)
			holdInnerWell := maxInnerWell08(holdBits)

			holdCreatesMoreHoles := holdHoles > currHoles
			holdAvoidsHoleCreation := currHoles > beforeHoles && holdHoles < currHoles && (candidateHoldType != PieceI || cleanupMode || maxH >= 7 || beforeInnerWell >= 3 || (IsHighSpeedMode(g) && maxH >= 6))

			maxSaveIHeight := 12
			if IsHighSpeedMode(g) {
				maxSaveIHeight = HighSpeedCleanupThreshold
			}
			currPlugsDeepCanyon := g.CurrentPiece.Type == PieceI && beforeInnerWell >= 4 && currInnerWell < beforeInnerWell && currHoles == beforeHoles && holdInnerWell > currInnerWell
			if scoreH != -math.MaxFloat64 && g.CurrentPiece.Type == PieceI && candidateHoldType != PieceI && maxH < maxSaveIHeight && !cleanupMode && !holdCreatesMoreHoles && !currPlugsDeepCanyon {
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH + 20000.0,
					CleanupMode:    cleanupMode,
				}
			} else if holdAvoidsHoleCreation && scoreH > bestScore-15000.0 {
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH + 15000.0,
					CleanupMode:    cleanupMode,
				}
			} else if isSZ && candidateNotSZ && terrainStrained && !holdCreatesMoreHoles && scoreH > bestScore-1000.0 {
				// S & Z pieces accounted for high top-out percentages.
				// Under terrain strain, swap S/Z into hold for a more accommodating piece!
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH + 12000.0,
					CleanupMode:    cleanupMode,
				}
			} else if oStrained && !holdCreatesMoreHoles && scoreH > bestScore-1000.0 {
				// O pieces require a 2x1 flat spot. Under high bumpiness, stash O for a flexible piece!
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH + 10000.0,
					CleanupMode:    cleanupMode,
				}
			} else if ljStrained && !holdCreatesMoreHoles && scoreH > bestScore-800.0 {
				// L & J pieces create overhangs on bumpy surfaces; swap for a flatter piece when strained.
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH + 8000.0,
					CleanupMode:    cleanupMode,
				}
			} else if cleanupMode && !holdCreatesMoreHoles && scoreH > bestScore {
				// In emergency cleanup mode, switch to hold piece if it scores higher
				bestMove = &AIMove{
					UseHold:        true,
					TargetRotation: rotH,
					TargetX:        xH,
					Score:          scoreH,
					CleanupMode:    cleanupMode,
				}
			} else if !holdCreatesMoreHoles && scoreH > bestScore+60.0 {
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

	if bestMove.Score == -math.MaxFloat64 {
		return nil
	}
	if bestMove.UseHold {
		bestCandidate = holdCandidate
	}
	bestMove.TargetY, bestMove.HasTargetY = bestCandidate.y, true
	bestMove.Actions, bestMove.Expected = bestCandidate.actions, bestCandidate.expected
	return bestMove
}

type candidatePlacement struct {
	rotation int
	x        int
	y        int
	board    *Board
	bits     BitBoard
	hasBits  bool
	score    float64
	actions  []AIAction
	expected []Piece
}

// findBestPlacementWithLookahead tests valid placements and evaluates the top candidates with next-piece lookahead.
func findBestPlacementWithLookahead(b *Board, p *Piece, nextType, nextNextType TetrominoType, cleanupMode bool, emergency ...bool) (int, int, float64) {
	cand, score := rankPlacements(getCandidatePlacements(b, p, cleanupMode, emergency...), nextType, nextNextType, cleanupMode, emergency...)
	return cand.rotation, cand.x, score
}

func rankPlacements(candidates []candidatePlacement, nextType, nextNextType TetrominoType, cleanupMode bool, emergency ...bool) (candidatePlacement, float64) {
	if len(candidates) == 0 {
		return candidatePlacement{x: 3}, -math.MaxFloat64
	}

	// Sort candidates by immediate score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	// If no next piece available, return top candidate directly
	if nextType == "" {
		return candidates[0], candidates[0].score
	}

	// Evaluate top 7 candidates with lookahead
	topLimit := 7
	if len(candidates) < topLimit {
		topLimit = len(candidates)
	}

	deep := nextNextType != ""

	bestCombinedScore := -math.MaxFloat64
	bestCandidate := candidates[0]

	for i := 0; i < topLimit; i++ {
		cand := candidates[i]
		candBits := cand.bits
		if !cand.hasBits && cand.board != nil {
			candBits = cand.board.ToBitBoard()
		}

		var nextScore float64
		if deep {
			_, _, nextScore = findBestBitPlacementWithLookahead(candBits, nextType, nextNextType, cleanupMode, emergency...)
		} else {
			_, _, nextScore = findBestBitPlacementSimple(candBits, nextType, cleanupMode, emergency...)
		}

		weight := 0.65
		if nextScore == -math.MaxFloat64 {
			nextScore = -1e12
		}
		if nextScore < -15000.0 {
			weight = 0.85
		}

		combined := cand.score + weight*nextScore
		if combined > bestCombinedScore {
			bestCombinedScore = combined
			bestCandidate = cand
		}
	}

	return bestCandidate, bestCombinedScore
}

func findBestBitPlacementWithLookahead(bb BitBoard, pieceType, nextType TetrominoType, cleanupMode bool, emergency ...bool) (int, int, float64) {
	candidates := getCandidateBitPlacements(bb, pieceType, cleanupMode, emergency...)
	if len(candidates) == 0 {
		return 0, 3, -math.MaxFloat64
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})
	if nextType == "" {
		return candidates[0].rotation, candidates[0].x, candidates[0].score
	}
	topLimit := 5
	if len(candidates) < topLimit {
		topLimit = len(candidates)
	}
	bestCombined := -math.MaxFloat64
	bestCand := candidates[0]
	for i := 0; i < topLimit; i++ {
		cand := candidates[i]
		_, _, nextScore := findBestBitPlacementSimple(cand.bits, nextType, cleanupMode, emergency...)
		weight := 0.65
		if nextScore == -math.MaxFloat64 {
			nextScore = -1e12
		}
		if nextScore < -15000.0 {
			weight = 0.85
		}
		combined := cand.score + weight*nextScore
		if combined > bestCombined {
			bestCombined = combined
			bestCand = cand
		}
	}
	return bestCand.rotation, bestCand.x, bestCombined
}

// isReachable checks if a piece can navigate horizontally from spawnX to targetX without floating above the board.
func isReachable(b *Board, p *Piece, targetRot, targetX int) bool {
	if b == nil || p == nil {
		return false
	}
	return isReachableBitBoard(b.ToBitBoard(), pieceTypeIndex(p.Type), p.Type, targetRot, targetX)
}

func isReachableBitBoard(bb BitBoard, pIdx int, pType TetrominoType, targetRot, targetX int) bool {
	spawnX := 3
	if pType == PieceO {
		spawnX = 4
	}
	if targetX == spawnX {
		return true
	}

	step := 1
	if targetX < spawnX {
		step = -1
	}

	highStack := bb.MaxHeight() >= 14

	// Try traversal at y=0 and y=-1 (1-row SRS wall kick clearance), but never y=-2
	// which would float 2-row pieces completely above row 0.
	// When stack height >= 14, also account for gravity drop during multi-step horizontal traversal.
	for _, tryY := range [2]int{0, -1} {
		canTraverse := true
		dist := 0
		for currX := spawnX; currX != targetX+step; currX += step {
			checkY := tryY
			if highStack && dist >= 3 {
				checkY = tryY + (dist-1)/2
			}
			if !bb.IsValidPosition(pIdx, targetRot, currX, checkY) {
				if !highStack || checkY == tryY || !bb.IsValidPosition(pIdx, targetRot, currX, checkY-1) {
					canTraverse = false
					break
				}
			}
			dist++
		}
		if canTraverse {
			return true
		}
	}
	return false
}

func getCandidateBitPlacements(bb BitBoard, pType TetrominoType, cleanupMode bool, emergency ...bool) []candidatePlacement {
	if !bb.CanSpawn(pType) {
		return nil
	}
	pIdx := pieceTypeIndex(pType)
	candidates := make([]candidatePlacement, 0, 40)

	for rot := 0; rot < 4; rot++ {
		if pType == PieceO && rot > 0 {
			continue
		}
		for x := -3; x < BoardWidth; x++ {
			startY := -1
			if !bb.IsValidPosition(pIdx, rot, x, startY) {
				startY = 0
				if !bb.IsValidPosition(pIdx, rot, x, startY) {
					continue
				}
			}
			if !isReachableBitBoard(bb, pIdx, pType, rot, x) {
				continue
			}
			ghostY := bb.GetGhostY(pIdx, rot, x, startY)
			if ghostY < 0 {
				continue
			}
			afterBB, linesCleared := bb.LockAndClear(pIdx, rot, x, ghostY)
			score := evaluateBitBoardWithSupport(bb, afterBB, pType, rot, x, ghostY, linesCleared, cleanupMode, emergency...)
			candidates = append(candidates, candidatePlacement{
				rotation: rot,
				x:        x,
				y:        ghostY,
				bits:     afterBB,
				hasBits:  true,
				score:    score,
			})
		}
	}
	return candidates
}

func getCandidatePlacements(b *Board, p *Piece, cleanupMode bool, emergency ...bool) []candidatePlacement {
	if b == nil || p == nil {
		return nil
	}
	bb := b.ToBitBoard()
	bitCands := getCandidateBitPlacements(bb, p.Type, cleanupMode, emergency...)
	for i := range bitCands {
		bitCands[i].board = bitCands[i].bits.ToBoard()
	}
	return bitCands
}

// findBestPlacementSimple is a fast 1-ply evaluator used in lookahead simulations.
func findBestPlacementSimple(b *Board, p *Piece, cleanupMode bool, emergency ...bool) (int, int, float64) {
	if b == nil || p == nil {
		return 0, 3, -math.MaxFloat64
	}
	return findBestBitPlacementSimple(b.ToBitBoard(), p.Type, cleanupMode, emergency...)
}

func findBestBitPlacementSimple(bb BitBoard, pType TetrominoType, cleanupMode bool, emergency ...bool) (int, int, float64) {
	bestScore := -math.MaxFloat64
	bestRot := 0
	bestX := 3
	if !bb.CanSpawn(pType) {
		return bestRot, bestX, bestScore
	}
	pIdx := pieceTypeIndex(pType)

	for rot := 0; rot < 4; rot++ {
		if pType == PieceO && rot > 0 {
			continue
		}
		for x := -3; x < BoardWidth; x++ {
			startY := -1
			if !bb.IsValidPosition(pIdx, rot, x, startY) {
				startY = 0
				if !bb.IsValidPosition(pIdx, rot, x, startY) {
					continue
				}
			}
			if !isReachableBitBoard(bb, pIdx, pType, rot, x) {
				continue
			}
			ghostY := bb.GetGhostY(pIdx, rot, x, startY)
			if ghostY < 0 {
				continue
			}
			afterBB, linesCleared := bb.LockAndClear(pIdx, rot, x, ghostY)
			score := evaluateBitBoardWithSupport(bb, afterBB, pType, rot, x, ghostY, linesCleared, cleanupMode, emergency...)
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
func evaluatePlacement(b *Board, p *Piece, linesCleared int, cleanupMode bool, emergency ...bool) float64 {
	if b == nil || p == nil {
		return -math.MaxFloat64
	}
	return evaluateBitBoard(b.ToBitBoard(), p.Type, p.Rotation, p.X, p.Y, linesCleared, cleanupMode, emergency...)
}

func evaluateBitBoard(bb BitBoard, pType TetrominoType, pRot, pX, pY, linesCleared int, cleanupMode bool, emergency ...bool) float64 {
	highSpeed := len(emergency) > 1 && emergency[1]
	colHeights := bb.ColHeights()
	maxHeight := 0
	aggHeight := 0
	centerMax := 0

	for x := 0; x < BoardWidth; x++ {
		h := colHeights[x]
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
			linesScore = 95000.0 // Huge reward for triple
		case 2:
			linesScore = 65000.0 // Great reward for double
		case 1:
			linesScore = 35000.0 // Good reward for single
		case 0:
			linesScore = -15000.0 // Penalize placing pieces without clearing lines
		}
	} else if highSpeed {
		// High-speed simpler mode (40% = 8-row Tetris threshold):
		// Keep 4-line Tetrises as the top reward when below 40% height,
		// but positively reward singles, doubles, and triples so the board stays
		// low and flat and pieces can always traverse to the left corridor (cols 0..2).
		switch linesCleared {
		case 4:
			linesScore = 180000.0
		case 3:
			if maxHeight >= 5 || centerMax >= 5 {
				linesScore = 25000.0
			} else {
				linesScore = 6000.0
			}
		case 2:
			if maxHeight >= 5 || centerMax >= 5 {
				linesScore = 12000.0
			} else {
				linesScore = 2000.0
			}
		case 1:
			if maxHeight >= 5 || centerMax >= 5 {
				linesScore = 4000.0
			} else {
				linesScore = -500.0
			}
		}

		// In 40% high-speed mode, only reward 9-0 stacking up to height 6 (leaving 2 rows of buffer below 8)
		if linesCleared == 0 && colHeights[9] == 0 && colHeights[8] >= 3 && colHeights[8] <= 6 {
			linesScore += 2500.0
		}
	} else {
		// Normal mode: Heavy priority on 4-line TETRIS
		switch linesCleared {
		case 4:
			linesScore = 180000.0 // Massive reward for TETRIS!
		case 3:
			if maxHeight >= 12 || centerMax >= 10 {
				linesScore = 2000.0
			} else {
				linesScore = -2000.0 // Discourage triple when safe
			}
		case 2:
			if maxHeight >= 12 || centerMax >= 10 {
				linesScore = 1000.0
			} else {
				linesScore = -3500.0 // Strongly discourage double when safe
			}
		case 1:
			if maxHeight >= 12 || centerMax >= 10 {
				linesScore = 500.0
			} else {
				linesScore = -5000.0 // Strongly discourage single when safe
			}
		}

		// Reward healthy 9-0 stacking ready for Tetris:
		// When columns 0..8 are building cleanly at height 3..8, column 9 is open, and 0 holes
		if linesCleared == 0 && colHeights[9] == 0 && colHeights[8] >= 3 && colHeights[8] <= 8 {
			linesScore += 2500.0
		}
	}

	// 2. Column 9 (Corner / Well) Reservation & Anti-Closure Protection
	wellPenalties := 0.0

	// Check if this placement put a vertical I in column 9 without slice allocation
	isVerticalIInCol9 := pType == PieceI && ((pRot == 1 && pX == 7) || (pRot == 3 && pX == 8))

	if isVerticalIInCol9 {
		if linesCleared == 4 {
			linesScore += 90000.0 // Super reward for clean 4-line Tetris in the corner well
		} else {
			wellPenalties += 120000.0
		}
	}

	if colHeights[9] > colHeights[8] {
		wellPenalties += float64(colHeights[9]-colHeights[8]) * 45000.0
	}

	if colHeights[9] > 0 {
		choked := false
		for y := BoardHeight - colHeights[9]; y < BoardHeight; y++ {
			if (bb[y] & 1) == 0 {
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
		if colHeights[9] > 0 {
			wellPenalties += float64(colHeights[9]) * 20000.0
		}

		wellDepth := colHeights[8] - colHeights[9]
		if wellDepth > 4 {
			wellPenalties += float64(wellDepth-4) * 10000.0
		}

		if colHeights[8] > colHeights[7]+2 {
			wellPenalties += float64(colHeights[8]-(colHeights[7]+2)) * 12000.0
		}

		if colHeights[7] > colHeights[6]+1 {
			wellPenalties += float64(colHeights[7]-(colHeights[6]+1)) * 3500.0
		}
		if colHeights[7] > colHeights[8]+1 {
			wellPenalties += float64(colHeights[7]-(colHeights[8]+1)) * 3500.0
		}
	} else {
		if isVerticalIInCol9 && linesCleared == 0 {
			wellPenalties += 80000.0
		}
		// Only penalize non-clearing blocks in Col 9 if Col 9 is already at or above Col 8;
		// when Col 9 < Col 8 and unchoked, filling Col 9 helps unlock multi-step cleanup line clears.
		if linesCleared == 0 && colHeights[9] > colHeights[8] {
			for y := BoardHeight - colHeights[9]; y < BoardHeight; y++ {
				if (bb[y] & 1) != 0 {
					wellPenalties += 800.0
				}
			}
		}
	}

	if cleanupMode && len(emergency) > 0 && emergency[0] {
		wellPenalties *= 0.08
	}

	// 3. Holes, Blockades & Bounded Shallow-Hole Digging Priority
	holes := 0
	blockades := 0
	shallowCover := 0
	for x := 0; x < BoardWidth; x++ {
		if colHeights[x] == 0 {
			continue
		}
		colBit := uint16(1 << (BoardWidth - 1 - x))
		blocksAbove := 0
		firstHoleRecorded := false
		for y := BoardHeight - colHeights[x]; y < BoardHeight; y++ {
			if (bb[y] & colBit) != 0 {
				blocksAbove++
			} else {
				holes++
				blockades += blocksAbove
				if !firstHoleRecorded {
					firstHoleRecorded = true
					if blocksAbove <= 2 {
						shallowCover += blocksAbove
					} else {
						shallowCover += 2
					}
				}
			}
		}
	}

	// 4. Bumpiness & Non-Linear Cliffs
	bumpinessLimit := 9
	bumpiness := 0
	cliffPenalty := 0.0

	for x := 0; x < bumpinessLimit-1; x++ {
		diff := colHeights[x] - colHeights[x+1]
		if diff < 0 {
			diff = -diff
		}
		bumpiness += diff

		if diff >= 3 {
			cliffPenalty += float64((diff-2)*(diff-2)) * 450.0
			if diff >= 4 {
				cliffPenalty += 2500.0
			}
		}
	}

	// Right-channel anti-spire protection in cleanup mode (where wellPenalties is scaled by 0.08):
	// Prevent Column 7 (or Column 8) from forming a spire that blocks pieces from reaching Column 9.
	if cleanupMode {
		if colHeights[7] > colHeights[6]+1 {
			cliffPenalty += float64(colHeights[7]-(colHeights[6]+1)) * 4200.0
		}
		if colHeights[7] > colHeights[8]+1 {
			cliffPenalty += float64(colHeights[7]-(colHeights[8]+1)) * 4200.0
		}
		if colHeights[8] > colHeights[7]+2 {
			cliffPenalty += float64(colHeights[8]-(colHeights[7]+2)) * 6000.0
		}
	}

	// Left-Flank & Left-Corridor Slope Protection:
	// Prevent columns 1 and 2 from forming a wall above column 0 that blocks pieces
	// spawning at x=3 from traversing left into column 0 at high speed, or creating a Double-Well Deadlock.
	leftSlopeWeight1 := 3200.0
	leftSlopeWeight2 := 2400.0
	leftSlopeWeight21 := 1900.0
	if highSpeed {
		leftSlopeWeight1 = 7500.0
		leftSlopeWeight2 = 6000.0
		leftSlopeWeight21 = 4500.0
	}
	if colHeights[1] > colHeights[0]+1 {
		diff01 := colHeights[1] - colHeights[0]
		cliffPenalty += float64(diff01-1) * leftSlopeWeight1
		if diff01 >= 3 {
			cliffPenalty += float64((diff01-2)*(diff01-2)) * 1800.0
		}
	} else if colHeights[0] >= colHeights[1]-1 && colHeights[0] <= colHeights[1]+1 {
		cliffPenalty -= 1200.0
	}
	if colHeights[2] > colHeights[0]+2 {
		cliffPenalty += float64(colHeights[2]-(colHeights[0]+2)) * leftSlopeWeight2
	}
	if colHeights[2] > colHeights[1]+1 {
		cliffPenalty += float64(colHeights[2]-(colHeights[1]+1)) * leftSlopeWeight21
	}

	// 3b. Piece-Aware Danger Handling: O/S/Z/L/J pieces on uneven or elevated terrain
	pieceRiskPenalty := 0.0
	if holes > 0 && maxHeight >= 10 {
		switch pType {
		case PieceO, PieceS, PieceZ:
			pieceRiskPenalty += float64(holes) * float64(maxHeight-9) * 4500.0
		case PieceL, PieceJ:
			pieceRiskPenalty += float64(holes) * float64(maxHeight-9) * 2500.0
		}
	}

	// Piece S & Z Alignment:
	// Horizontal S/Z is stable; vertical S/Z has an overhang tail that traps holes in uneven terrain.
	if pType == PieceS || pType == PieceZ {
		if pRot == 1 || pRot == 3 {
			if linesCleared == 0 && (bumpiness >= 6 || maxHeight >= 8) {
				pieceRiskPenalty += 3500.0
			}
		} else {
			if linesCleared == 0 && holes == 0 {
				pieceRiskPenalty -= 1500.0
			}
		}
	}

	// 5. Deep valleys in building zone (columns 0..8) & Double-Well Deadlock Prevention
	innerWells := 0
	innerWellQuadPenalty := 0.0
	doubleWellFactor := 750.0
	if colHeights[9] <= 2 {
		doubleWellFactor = 1150.0
	}
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
		depth := minAdjacent - colHeights[x]
		if depth >= 3 {
			innerWells += depth
			innerWellQuadPenalty += float64((depth-2)*(depth-2)) * doubleWellFactor
		}
	}

	// Landing height bonus
	landingBonus := float64(pY) * 12.0
	if cleanupMode {
		landingBonus = float64(pY) * 25.0
	}

	holeWeight := 28000.0
	blockadeWeight := 600.0
	bumpinessWeight := 160.0
	shallowCoverWeight := 0.0

	if cleanupMode {
		holeWeight = 55000.0
		blockadeWeight = 2200.0
		bumpinessWeight = 220.0
		shallowCoverWeight = 1200.0 // Penalize burying a 1-deep hole so AI keeps it open for digging
	}

	heightWeight := 45.0
	if highSpeed && !cleanupMode {
		heightWeight = 85.0
	}
	maxHeightPenalty := 0.0
	if cleanupMode {
		heightWeight = 140.0
		ceilLimit := 13
		if highSpeed {
			ceilLimit = HighSpeedCleanupThreshold
		}
		if maxHeight > ceilLimit {
			maxHeightPenalty = float64(maxHeight-ceilLimit) * float64(maxHeight-ceilLimit) * 250.0
		}
	}

	// 6. Spawn Corridor Clearance & Anti-Center Dome Penalty
	spawnChutePenalty := 0.0
	if centerMax >= 9 {
		spawnChutePenalty += float64((centerMax-8)*(centerMax-8)) * 400.0
		if centerMax >= 12 {
			spawnChutePenalty += float64(centerMax-11) * 15000.0
		}
	}

	flankLeftMax := colHeights[0]
	if colHeights[1] > flankLeftMax {
		flankLeftMax = colHeights[1]
	}
	if colHeights[2] > flankLeftMax {
		flankLeftMax = colHeights[2]
	}

	flankRightMax := colHeights[7]
	if colHeights[8] > flankRightMax {
		flankRightMax = colHeights[8]
	}

	if centerMax > flankLeftMax {
		spawnChutePenalty += float64(centerMax-flankLeftMax) * 1500.0
	}
	if centerMax > flankRightMax && flankRightMax < 12 {
		spawnChutePenalty += float64(centerMax-flankRightMax) * 1500.0
	}
	// Left & right corridor clearance: at Level >= 13 (80ms gravity), columns 5..8 reaching
	// height >= 14 block vertical I-pieces from entering column 9, and columns 1..8 reaching >= 14
	// block traversal to the flanks.
	for x := 1; x <= 8; x++ {
		if x >= 5 && colHeights[x] >= 13 {
			spawnChutePenalty += float64(colHeights[x]-12) * 8500.0
		}
		if colHeights[x] >= 14 {
			spawnChutePenalty += float64(colHeights[x]-13) * 16000.0
		}
		if highSpeed && colHeights[x] > HighSpeedCleanupThreshold {
			spawnChutePenalty += float64(colHeights[x]-HighSpeedCleanupThreshold) * 6500.0
		}
	}

	score := linesScore -
		wellPenalties -
		float64(holes)*holeWeight -
		float64(blockades)*blockadeWeight -
		float64(shallowCover)*shallowCoverWeight -
		float64(bumpiness)*bumpinessWeight -
		cliffPenalty -
		spawnChutePenalty -
		float64(aggHeight)*heightWeight -
		maxHeightPenalty -
		float64(innerWells)*500.0 -
		innerWellQuadPenalty -
		pieceRiskPenalty +
		landingBonus

	return score
}

func maxInnerWell08(bb BitBoard) int {
	colHeights := bb.ColHeights()
	maxDepth := 0
	for x := 0; x < 9; x++ {
		leftH := 20
		if x > 0 {
			leftH = colHeights[x-1]
		}
		rightH := 20
		if x < 8 {
			rightH = colHeights[x+1]
		}
		minAdj := leftH
		if rightH < minAdj {
			minAdj = rightH
		}
		if depth := minAdj - colHeights[x]; depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

// Measure the support on the input board: locking an O makes its top two
// heights equal even when one side was unsupported before placement.
func evaluatePlacementWithSupport(before, after *Board, p *Piece, lines int, cleanup bool, emergency ...bool) float64 {
	if before == nil || after == nil || p == nil {
		return -math.MaxFloat64
	}
	return evaluateBitBoardWithSupport(before.ToBitBoard(), after.ToBitBoard(), p.Type, p.Rotation, p.X, p.Y, lines, cleanup, emergency...)
}

func evaluateBitBoardWithSupport(before, after BitBoard, pType TetrominoType, pRot, pX, pY, lines int, cleanup bool, emergency ...bool) float64 {
	score := evaluateBitBoard(after, pType, pRot, pX, pY, lines, cleanup, emergency...)
	holeDelta := after.CountHoles() - before.CountHoles()
	if holeDelta > 0 {
		score -= float64(holeDelta) * 18000.0
		if pType == PieceL || pType == PieceJ {
			score -= float64(holeDelta) * 5500.0
		}
	} else if (pType == PieceL || pType == PieceJ) && lines == 0 && after.Bumpiness8() < before.Bumpiness8() {
		score += 1500.0
	}
	if holeDelta <= 0 {
		if beforeWell := maxInnerWell08(before); beforeWell >= 3 {
			if afterWell := maxInnerWell08(after); afterWell < beforeWell {
				score += float64(beforeWell-afterWell) * 4500.0
			}
		}
	}
	if pType == PieceO && pX >= 0 && pX+1 < BoardWidth {
		heights := before.ColHeights()
		diff := heights[pX] - heights[pX+1]
		if diff < 0 {
			diff = -diff
		}
		if diff == 0 && lines == 0 {
			score += 3500
		} else if diff >= 2 {
			score -= float64(diff) * 3500
		}
	}
	return score
}

