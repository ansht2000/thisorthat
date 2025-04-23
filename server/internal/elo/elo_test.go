package elo

import "testing"

func TestCalculateELO(t *testing.T) {
	cases := []struct{
		winnerELO int
		loserELO int
		expectedWinnerELO int
		expectedLoserELO int
	}{
		{
			winnerELO: 1200,
			loserELO: 1200,
			expectedWinnerELO: 1215,
			expectedLoserELO: 1185,
		},
		{
			winnerELO: 1200,
			loserELO: 1600,
			expectedWinnerELO: 1227,
			expectedLoserELO: 1572,
		},
		{
			winnerELO: 1600,
			loserELO: 1200,
			expectedWinnerELO: 1602,
			expectedLoserELO: 1197,
		},
		{
			winnerELO: 1200,
			loserELO: 1000,
			expectedWinnerELO: 1207,
			expectedLoserELO: 992,
		},
		{
			winnerELO: 1000,
			loserELO: 1200,
			expectedWinnerELO: 1022,
			expectedLoserELO: 1177,
		},
	}

	for _, c := range cases {
		actualWinnerELO, actualLoserELO := CalculateELO(c.winnerELO, c.loserELO)
		if actualWinnerELO != c.expectedWinnerELO || actualLoserELO != c.expectedLoserELO {
			t.Errorf(
				`expected winner elo: %v is not equal to actual winner elo: %v,
				or expected loser elo: %v is not equal to actual loser elo: %v`,
				c.expectedWinnerELO,
				actualWinnerELO,
				c.expectedLoserELO,
				actualLoserELO,
			)
		}
	}
}