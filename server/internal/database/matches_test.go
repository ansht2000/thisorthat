package database

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

func countMatches(t *testing.T, c *Client) int {
	t.Helper()
	var n int
	if err := c.q.QueryRowContext(testContext(t), `SELECT COUNT(*) FROM matches;`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRecordMatch(t *testing.T) {
	c := newTestClient(t)
	ctx := testContext(t)
	list := mustCreateList(t, c, "invincible")
	mark := mustCreateCharacter(t, c, list.ID, "Mark Grayson")
	nolan := mustCreateCharacter(t, c, list.ID, "Nolan Grayson")

	// CURRENT_TIMESTAMP only has second precision, so backdate instead of sleeping
	if _, err := c.q.ExecContext(ctx, `UPDATE characters SET updated_at = '2000-01-01 00:00:00';`); err != nil {
		t.Fatal(err)
	}

	match, err := c.RecordMatch(ctx, mark.ID, nolan.ID)
	if err != nil {
		t.Fatal(err)
	}
	// two new characters at 1200 move by the full starting k of 64
	expected := Match{
		ID:              match.ID,
		ListID:          list.ID,
		WinnerID:        mark.ID,
		LoserID:         nolan.ID,
		WinnerEloBefore: 1200,
		WinnerEloAfter:  1232,
		LoserEloBefore:  1200,
		LoserEloAfter:   1168,
		CreatedAt:       match.CreatedAt,
	}
	if match != expected {
		t.Errorf("expected match %+v, got %+v", expected, match)
	}
	if match.ID == uuid.Nil || match.CreatedAt.IsZero() {
		t.Errorf("expected match to get an id and timestamp, got %+v", match)
	}

	for _, want := range []struct {
		id  uuid.UUID
		elo int
	}{{mark.ID, 1232}, {nolan.ID, 1168}} {
		got, err := c.GetCharacterByID(ctx, want.id)
		if err != nil {
			t.Fatal(err)
		}
		if got.Elo != want.elo || got.GamesPlayed != 1 {
			t.Errorf("expected %s at elo %d with 1 game played, got elo %d with %d", got.Name, want.elo, got.Elo, got.GamesPlayed)
		}
		if !got.UpdatedAt.After(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)) {
			t.Errorf("expected %s updated_at to be bumped, got %v", got.Name, got.UpdatedAt)
		}
	}
}

func TestRecordMatchRejectsBadMatchups(t *testing.T) {
	c := newTestClient(t)
	ctx := testContext(t)
	list := mustCreateList(t, c, "invincible")
	other := mustCreateList(t, c, "the boys")
	mark := mustCreateCharacter(t, c, list.ID, "Mark Grayson")
	homelander := mustCreateCharacter(t, c, other.ID, "Homelander")

	cases := []struct {
		name          string
		winner, loser uuid.UUID
		expectedErr   error
	}{
		{"same character", mark.ID, mark.ID, ErrSameCharacter},
		{"different lists", mark.ID, homelander.ID, ErrDifferentLists},
		{"unknown winner", uuid.New(), mark.ID, sql.ErrNoRows},
		{"unknown loser", mark.ID, uuid.New(), sql.ErrNoRows},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := c.RecordMatch(ctx, tc.winner, tc.loser); !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected %v, got %v", tc.expectedErr, err)
			}
		})
	}

	// nothing from the rejected matchups should have stuck
	for _, character := range []Character{mark, homelander} {
		got, err := c.GetCharacterByID(ctx, character.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Elo != 1200 || got.GamesPlayed != 0 {
			t.Errorf("expected %s untouched, got elo %d with %d games", got.Name, got.Elo, got.GamesPlayed)
		}
	}
	if n := countMatches(t, c); n != 0 {
		t.Errorf("expected no matches recorded, got %d", n)
	}
}

// votes landing at the same time used to both read the same old ratings, so
// whichever wrote last erased the other. every vote should count now
func TestRecordMatchConcurrentVotesAreNotLost(t *testing.T) {
	// a file db so the pool can open many connections, unlike the in memory one
	c, err := NewClient(filepath.Join(t.TempDir(), "concurrent.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	ctx := testContext(t)
	list := mustCreateList(t, &c, "invincible")
	mark := mustCreateCharacter(t, &c, list.ID, "Mark Grayson")
	nolan := mustCreateCharacter(t, &c, list.ID, "Nolan Grayson")

	const votes = 50
	var wg sync.WaitGroup
	errs := make(chan error, votes)
	for i := range votes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// alternate winners so neither character drifts toward the floor
			winner, loser := mark.ID, nolan.ID
			if i%2 == 1 {
				winner, loser = loser, winner
			}
			if _, err := c.RecordMatch(ctx, winner, loser); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Errorf("vote failed: %v", err)
	}

	total := 0
	for _, id := range []uuid.UUID{mark.ID, nolan.ID} {
		got, err := c.GetCharacterByID(ctx, id)
		if err != nil {
			t.Fatal(err)
		}
		if got.GamesPlayed != votes {
			t.Errorf("expected %s to have played %d games, got %d", got.Name, votes, got.GamesPlayed)
		}
		total += got.Elo
	}
	// both always have the same games played, so the same k, so points are only ever moved between them
	if total != 2400 {
		t.Errorf("expected ratings to still total 2400, got %d", total)
	}
	if n := countMatches(t, &c); n != votes {
		t.Errorf("expected %d matches recorded, got %d", votes, n)
	}
}
