package engine

import (
	"math/rand"
	"time"
)

// Randomizer implements the standard Tetris 7-bag randomizer.
type Randomizer struct {
	bag []TetrominoType
	rng *rand.Rand
}

func NewRandomizer() *Randomizer {
	r := &Randomizer{
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	r.refill()
	return r
}

func (r *Randomizer) refill() {
	r.bag = make([]TetrominoType, len(AllPieces))
	copy(r.bag, AllPieces)
	r.rng.Shuffle(len(r.bag), func(i, j int) {
		r.bag[i], r.bag[j] = r.bag[j], r.bag[i]
	})
}

// Next draws the next piece from the 7-bag.
func (r *Randomizer) Next() TetrominoType {
	if len(r.bag) == 0 {
		r.refill()
	}
	piece := r.bag[0]
	r.bag = r.bag[1:]
	return piece
}
