package elo

import (
	"errors"
	"math"
	"math/rand/v2"
	"testing"
)

func seeded(seed uint64) func() float64 {
	return rand.New(rand.NewPCG(seed, 0)).Float64
}

// runs PickPair draws times and counts how often each unordered pair comes up
func countPairs(t *testing.T, ratings []Rating, avoid func(i, j int) bool, draws int) map[[2]int]int {
	t.Helper()
	randFloat := seeded(1)
	counts := map[[2]int]int{}
	for range draws {
		i, j, err := PickPair(ratings, avoid, randFloat)
		if err != nil {
			t.Fatal(err)
		}
		if i == j || i < 0 || j < 0 || i >= len(ratings) || j >= len(ratings) {
			t.Fatalf("got invalid pair (%d, %d) for %d ratings", i, j, len(ratings))
		}
		counts[[2]int{min(i, j), max(i, j)}]++
	}
	return counts
}

func TestPickPairNeedsTwoContenders(t *testing.T) {
	for _, ratings := range [][]Rating{nil, {{Elo: 1200}}} {
		if _, _, err := PickPair(ratings, nil, seeded(1)); !errors.Is(err, ErrNotEnoughContenders) {
			t.Errorf("expected ErrNotEnoughContenders for %d ratings, got %v", len(ratings), err)
		}
	}
}

func TestPickPairSkipsAvoidedPairs(t *testing.T) {
	ratings := []Rating{{Elo: 1200}, {Elo: 1200}, {Elo: 1200}}
	avoid := func(i, j int) bool {
		if i >= j {
			t.Fatalf("expected avoid to be called with i < j, got (%d, %d)", i, j)
		}
		return i == 0 && j == 1
	}
	counts := countPairs(t, ratings, avoid, 1000)
	if counts[[2]int{0, 1}] != 0 {
		t.Errorf("expected avoided pair to never come up, got it %d times", counts[[2]int{0, 1}])
	}
	if counts[[2]int{0, 2}] == 0 || counts[[2]int{1, 2}] == 0 {
		t.Errorf("expected every allowed pair to come up, got %v", counts)
	}
}

func TestPickPairRepeatsWhenEveryPairIsAvoided(t *testing.T) {
	ratings := []Rating{{Elo: 1200}, {Elo: 1200}}
	counts := countPairs(t, ratings, func(int, int) bool { return true }, 10)
	if counts[[2]int{0, 1}] != 10 {
		t.Errorf("expected the only pair every time, got %v", counts)
	}
}

func TestPickPairFavorsFewerGames(t *testing.T) {
	ratings := []Rating{
		{Elo: 1200, GamesPlayed: 0},
		{Elo: 1200, GamesPlayed: 100},
		{Elo: 1200, GamesPlayed: 100},
		{Elo: 1200, GamesPlayed: 100},
		{Elo: 1200, GamesPlayed: 100},
	}
	appearances := make([]int, len(ratings))
	for pair, n := range countPairs(t, ratings, nil, 10000) {
		appearances[pair[0]] += n
		appearances[pair[1]] += n
	}
	for i := 1; i < len(ratings); i++ {
		if appearances[0] < 2*appearances[i] {
			t.Errorf("expected the new character in at least twice as many matchups as established ones, got %v", appearances)
			break
		}
	}
}

func TestPickPairFavorsCloseRatings(t *testing.T) {
	ratings := []Rating{{Elo: 1200}, {Elo: 1220}, {Elo: 1800}}
	counts := countPairs(t, ratings, nil, 10000)
	evenCount, far1, far2 := counts[[2]int{0, 1}], counts[[2]int{0, 2}], counts[[2]int{1, 2}]
	if evenCount < 2*far1 || evenCount < 2*far2 {
		t.Errorf("expected 1200 vs 1220 at least twice as often as either matchup against 1800, got %v", counts)
	}
	if far1 == 0 || far2 == 0 {
		t.Errorf("expected lopsided matchups to still come up sometimes, got %v", counts)
	}
}

func TestPickPairRandomizesSides(t *testing.T) {
	// the new character is almost always the first pick, so without the
	// shuffle it would almost always come back as i
	ratings := []Rating{{Elo: 1200, GamesPlayed: 0}, {Elo: 1200, GamesPlayed: 10000}, {Elo: 1200, GamesPlayed: 10000}}
	randFloat := seeded(1)
	first, total := 0, 0
	for range 10000 {
		i, j, err := PickPair(ratings, nil, randFloat)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first++
		}
		if i == 0 || j == 0 {
			total++
		}
	}
	if share := float64(first) / float64(total); share < 0.45 || share > 0.55 {
		t.Errorf("expected the new character on either side about half the time, got %.2f", share)
	}
}

func TestWeightedPick(t *testing.T) {
	if i := weightedPick([]float64{0, 0}, seeded(1)); i != -1 {
		t.Errorf("expected -1 when every weight is 0, got %d", i)
	}
	if i := weightedPick([]float64{0, 3, 0}, seeded(1)); i != 1 {
		t.Errorf("expected the only weighted index, got %d", i)
	}
	// the largest float below 1 can overshoot the total by a rounding error,
	// which should land on the last weighted index, not the zero weight after it
	almostOne := func() float64 { return math.Nextafter(1, 0) }
	if i := weightedPick([]float64{0.1, 0.2, 0}, almostOne); i != 1 {
		t.Errorf("expected the last weighted index, got %d", i)
	}
}
