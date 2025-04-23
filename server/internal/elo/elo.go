package elo

import (
	"math"

	"github.com/ansht2000/thisorthat/internal/utils"
)

// constant factor to control elo adjustment, set to 30 as default
// TODO: balance
var kFactor = 30.0

func probability(ratingOne int, ratingTwo int) float64 {
	return 1.0 / (1.0 + math.Pow(10, ((float64(ratingTwo - ratingOne) / 400))))

}

func newRating(oldElo int, outcomeProb float64, isWinner bool) float64 {
	// int(isWinner) is 1 if true, 0, if false, then cast
	// to float to get a decimal to multiply by Kfactor
	return float64(oldElo) + kFactor * (float64(utils.FastBoolToInt(isWinner)) - outcomeProb)
}

func CalculateELO(winnerELO int, loserELO int) (int, int) {
	outcomeProb := probability(loserELO, winnerELO)
	newWinnerElo := newRating(winnerELO, 1 - outcomeProb, true)
	newLoserELO := newRating(loserELO, outcomeProb, false)
	return int(newWinnerElo), int(newLoserELO)
}