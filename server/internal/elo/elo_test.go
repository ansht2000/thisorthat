package elo

import (
	"math"
	"math/rand/v2"
	"testing"
)

func TestKFactor(t *testing.T) {
	if k := KFactor(0); k != maxK {
		t.Errorf("expected a new character to have k %v, got %v", maxK, k)
	}
	if k := KFactor(kHalfwayGames); k != (maxK+minK)/2 {
		t.Errorf("expected k halfway between %v and %v after %v games, got %v", maxK, minK, kHalfwayGames, k)
	}
	previous := KFactor(0)
	for games := 1; games <= 10000; games++ {
		k := KFactor(games)
		if k >= previous || k <= minK {
			t.Fatalf("expected k to shrink toward %v without reaching it, got %v after %v games (%v before)", minK, k, games, previous)
		}
		previous = k
	}
}

func TestUpdate(t *testing.T) {
	// 15 games played puts k at exactly 40, which keeps the math below easy to check by hand
	cases := []struct {
		name                string
		winner, loser       Rating
		expectedWinnerAfter int
		expectedLoserAfter  int
	}{
		{
			name:                "even matchup",
			winner:              Rating{Elo: 1200, GamesPlayed: 15},
			loser:               Rating{Elo: 1200, GamesPlayed: 15},
			expectedWinnerAfter: 1220,
			expectedLoserAfter:  1180,
		},
		{
			// 40 * 10/11 = 36.36, truncating used to give the loser 1563 and lose a point
			name:                "upset rounds instead of truncating",
			winner:              Rating{Elo: 1200, GamesPlayed: 15},
			loser:               Rating{Elo: 1600, GamesPlayed: 15},
			expectedWinnerAfter: 1236,
			expectedLoserAfter:  1564,
		},
		{
			name:                "favorite wins and barely moves",
			winner:              Rating{Elo: 1600, GamesPlayed: 15},
			loser:               Rating{Elo: 1200, GamesPlayed: 15},
			expectedWinnerAfter: 1604,
			expectedLoserAfter:  1196,
		},
		{
			// the new character moves by k=64, the established one by k~16.7
			name:                "new character moves further than an established one",
			winner:              Rating{Elo: 1200, GamesPlayed: 0},
			loser:               Rating{Elo: 1200, GamesPlayed: 1000},
			expectedWinnerAfter: 1232,
			expectedLoserAfter:  1192,
		},
		{
			name:                "loser stops at the floor",
			winner:              Rating{Elo: 110, GamesPlayed: 15},
			loser:               Rating{Elo: 110, GamesPlayed: 15},
			expectedWinnerAfter: 130,
			expectedLoserAfter:  ratingFloor,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			winner, loser := Update(c.winner, c.loser)
			if winner.Elo != c.expectedWinnerAfter || loser.Elo != c.expectedLoserAfter {
				t.Errorf("expected winner %d and loser %d, got winner %d and loser %d",
					c.expectedWinnerAfter, c.expectedLoserAfter, winner.Elo, loser.Elo)
			}
			if winner.GamesPlayed != c.winner.GamesPlayed+1 || loser.GamesPlayed != c.loser.GamesPlayed+1 {
				t.Errorf("expected both games played to go up by one, got winner %d and loser %d",
					winner.GamesPlayed, loser.GamesPlayed)
			}
		})
	}
}

func TestUpdateConservesPointsWithEqualGamesPlayed(t *testing.T) {
	// starting at 200 keeps every loser far enough above the floor that it can't kick in
	for _, games := range []int{0, 5, 15, 100, 1000} {
		for winnerElo := 200; winnerElo <= 2800; winnerElo += 50 {
			for loserElo := 200; loserElo <= 2800; loserElo += 50 {
				winner, loser := Update(Rating{winnerElo, games}, Rating{loserElo, games})
				if winner.Elo+loser.Elo != winnerElo+loserElo {
					t.Fatalf("%d beating %d after %d games: total went from %d to %d",
						winnerElo, loserElo, games, winnerElo+loserElo, winner.Elo+loser.Elo)
				}
				if winner.Elo < winnerElo || loser.Elo > loserElo {
					t.Fatalf("%d beating %d after %d games: winner went to %d and loser to %d",
						winnerElo, loserElo, games, winner.Elo, loser.Elo)
				}
			}
		}
	}
}

// a character whose true strength is 1600 plays random opponents from an established pool.
// returns the mean and standard deviation of its final rating across runs seeded 0 to runs-1
func simulateRatings(kFactor func(gamesPlayed int) float64, start Rating, games int, runs int) (float64, float64) {
	const trueElo = 1600
	opponents := []int{1000, 1100, 1200, 1300, 1400, 1500, 1600, 1700, 1800, 1900, 2000}

	finals := make([]float64, runs)
	for run := range runs {
		rng := rand.New(rand.NewPCG(uint64(run), 0))
		r := start
		for range games {
			opponent := opponents[rng.IntN(len(opponents))]
			won := rng.Float64() < probability(trueElo, opponent)
			r.Elo = newRating(r.Elo, kFactor(r.GamesPlayed), probability(r.Elo, opponent), won)
			r.GamesPlayed++
		}
		finals[run] = float64(r.Elo)
	}

	mean := 0.0
	for _, f := range finals {
		mean += f
	}
	mean /= float64(runs)
	variance := 0.0
	for _, f := range finals {
		variance += (f - mean) * (f - mean)
	}
	return mean, math.Sqrt(variance / float64(runs))
}

func TestDynamicKBeatsFixedK(t *testing.T) {
	oldFixedK := func(int) float64 { return 30 }

	// placement: a new character that starts 400 below where it belongs
	dynamicPlaced, _ := simulateRatings(KFactor, Rating{Elo: 1200}, 30, 2000)
	fixedPlaced, _ := simulateRatings(oldFixedK, Rating{Elo: 1200}, 30, 2000)
	if dynamicPlaced <= fixedPlaced {
		t.Errorf("expected dynamic k to place a new character faster: after 30 games dynamic averaged %.0f, fixed k averaged %.0f (true rating 1600)",
			dynamicPlaced, fixedPlaced)
	}
	if dynamicPlaced < 1400 {
		t.Errorf("expected a new character to close at least half of a 400 point gap in 30 games, got %.0f", dynamicPlaced)
	}

	// stability: a settled character already at its true rating should just jitter around it
	_, dynamicJitter := simulateRatings(KFactor, Rating{Elo: 1600, GamesPlayed: 200}, 50, 2000)
	_, fixedJitter := simulateRatings(oldFixedK, Rating{Elo: 1600, GamesPlayed: 200}, 50, 2000)
	if dynamicJitter >= fixedJitter {
		t.Errorf("expected dynamic k to keep a settled rating steadier: standard deviation %.1f vs %.1f with fixed k",
			dynamicJitter, fixedJitter)
	}
}
