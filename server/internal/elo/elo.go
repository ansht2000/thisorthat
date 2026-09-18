package elo

import (
	"math"

	"github.com/ansht2000/thisorthat/internal/utils"
)

const (
	// k controls how far one result moves a rating. it starts at maxK so new characters
	// find their spot quickly, then shrinks toward minK as they play more so an
	// established ranking doesn't swing on a single vote. compared to the old fixed
	// k of 30 in simulation, this places a new character faster and cuts the jitter
	// of a settled one by about a quarter (see TestDynamicKBeatsFixedK)
	maxK = 64.0
	minK = 16.0
	// games played at which k is halfway between maxK and minK
	kHalfwayGames = 15.0

	// ratings never drop below this, so a character on a long losing
	// streak can't sink so far that climbing back out takes forever
	ratingFloor = 100
)

type Rating struct {
	Elo         int
	GamesPlayed int
}

func KFactor(gamesPlayed int) float64 {
	return minK + (maxK-minK)*kHalfwayGames/(kHalfwayGames+float64(gamesPlayed))
}

// chance that ratingOne beats ratingTwo
func probability(ratingOne int, ratingTwo int) float64 {
	return 1.0 / (1.0 + math.Pow(10, (float64(ratingTwo-ratingOne)/400)))
}

func newRating(oldElo int, k float64, outcomeProb float64, isWinner bool) int {
	// int(isWinner) is 1 if true, 0, if false, then cast
	// to float to get a decimal to multiply by k
	change := k * (float64(utils.FastBoolToInt(isWinner)) - outcomeProb)
	// rounding, not truncating: truncation rounds toward zero, which shorts the winner's
	// gain and adds to the loser's loss, so about a point leaked out of the pool every vote
	return max(ratingFloor, oldElo+int(math.Round(change)))
}

// Update returns the winner's and loser's ratings after the winner beats the loser.
// each side moves by its own k, so when both have played the same number of games
// the winner gains exactly what the loser loses
func Update(winner Rating, loser Rating) (Rating, Rating) {
	winnerProb := probability(winner.Elo, loser.Elo)
	newWinner := Rating{
		Elo:         newRating(winner.Elo, KFactor(winner.GamesPlayed), winnerProb, true),
		GamesPlayed: winner.GamesPlayed + 1,
	}
	newLoser := Rating{
		Elo:         newRating(loser.Elo, KFactor(loser.GamesPlayed), 1-winnerProb, false),
		GamesPlayed: loser.GamesPlayed + 1,
	}
	return newWinner, newLoser
}
