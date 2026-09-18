package elo

import (
	"errors"
	"math"
)

var ErrNotEnoughContenders = errors.New("need at least two characters to make a matchup")

// the least weight any allowed opponent gets, so matchups lopsided enough
// to round to a sure thing still come up once in a while instead of never
const minOpponentWeight = 1e-3

// PickPair chooses two different indexes into ratings to put in front of the next voter.
//
// the first character is more likely the fewer games it has played, so new characters
// get placed quickly. its opponent is more likely the closer their ratings are, because
// a vote between a 1200 and an 1800 tells you almost nothing new while an even matchup
// could go either way. the pair comes back in random order so the favored pick doesn't
// always land on the same side of the screen.
//
// avoid is called with i < j and should return true for pairs to skip, like ones the voter
// just saw. nil avoids nothing. if every pair is avoided the avoid rule is dropped, since
// a repeat beats no matchup at all. randFloat returns a number in [0, 1), like rand.Float64
func PickPair(ratings []Rating, avoid func(i, j int) bool, randFloat func() float64) (int, int, error) {
	if len(ratings) < 2 {
		return -1, -1, ErrNotEnoughContenders
	}

	skip := func(i, j int) bool {
		return avoid != nil && avoid(min(i, j), max(i, j))
	}
	i, j, ok := pickPair(ratings, skip, randFloat)
	if !ok {
		i, j, _ = pickPair(ratings, func(int, int) bool { return false }, randFloat)
	}

	if randFloat() < 0.5 {
		i, j = j, i
	}
	return i, j, nil
}

func pickPair(ratings []Rating, skip func(i, j int) bool, randFloat func() float64) (int, int, bool) {
	firstWeights := make([]float64, len(ratings))
	for i, r := range ratings {
		if hasOpponent(i, len(ratings), skip) {
			// square root so a brand new character is favored
			// without showing up in nearly every matchup
			firstWeights[i] = 1 / math.Sqrt(float64(1+r.GamesPlayed))
		}
	}
	i := weightedPick(firstWeights, randFloat)
	if i < 0 {
		return -1, -1, false
	}

	opponentWeights := make([]float64, len(ratings))
	for j, r := range ratings {
		if j == i || skip(i, j) {
			continue
		}
		// p(1-p) is the variance of the result: highest for an
		// even matchup, near zero for a foregone conclusion
		p := probability(ratings[i].Elo, r.Elo)
		opponentWeights[j] = max(p*(1-p), minOpponentWeight)
	}
	// can't be -1: hasOpponent guaranteed at least one allowed opponent
	j := weightedPick(opponentWeights, randFloat)
	return i, j, true
}

func hasOpponent(i int, n int, skip func(i, j int) bool) bool {
	for j := range n {
		if j != i && !skip(i, j) {
			return true
		}
	}
	return false
}

// returns index i with probability weights[i] / sum(weights), or -1 if every weight is 0
func weightedPick(weights []float64, randFloat func() float64) int {
	total := 0.0
	for _, w := range weights {
		total += w
	}
	if total == 0 {
		return -1
	}

	target := randFloat() * total
	last := -1
	for i, w := range weights {
		if w == 0 {
			continue
		}
		if target < w {
			return i
		}
		target -= w
		last = i
	}
	// float rounding can leave target a hair past the final weight
	return last
}
