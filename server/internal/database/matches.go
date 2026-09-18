package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/ansht2000/thisorthat/internal/elo"
	"github.com/google/uuid"
)

var (
	ErrSameCharacter  = errors.New("a character can't be matched against itself")
	ErrDifferentLists = errors.New("characters in a match must be from the same list")
)

type Match struct {
	ID              uuid.UUID `json:"id"`
	ListID          uuid.UUID `json:"list_id"`
	WinnerID        uuid.UUID `json:"winner_id"`
	LoserID         uuid.UUID `json:"loser_id"`
	WinnerEloBefore int       `json:"winner_elo_before"`
	WinnerEloAfter  int       `json:"winner_elo_after"`
	LoserEloBefore  int       `json:"loser_elo_before"`
	LoserEloAfter   int       `json:"loser_elo_after"`
	CreatedAt       time.Time `json:"created_at"`
}

// same order scanMatch reads them in
const matchColumns = `id, list_id, winner_id, loser_id, winner_elo_before, winner_elo_after, loser_elo_before, loser_elo_after, created_at`

func scanMatch(row rowScanner) (Match, error) {
	var match Match
	if err := row.Scan(
		&match.ID,
		&match.ListID,
		&match.WinnerID,
		&match.LoserID,
		&match.WinnerEloBefore,
		&match.WinnerEloAfter,
		&match.LoserEloBefore,
		&match.LoserEloAfter,
		&match.CreatedAt,
	); err != nil {
		return Match{}, err
	}
	return match, nil
}

// RecordMatch applies one vote: both characters are read, rerated, and the match logged
// in a single transaction, so votes that land at the same time queue up instead of
// reading the same old ratings and overwriting each other's results.
// returns sql.ErrNoRows if either character doesn't exist
func (c *Client) RecordMatch(ctx context.Context, winnerID uuid.UUID, loserID uuid.UUID) (Match, error) {
	if winnerID == loserID {
		return Match{}, ErrSameCharacter
	}

	var match Match
	err := c.WithTx(ctx, func(tx *Client) error {
		winner, err := tx.GetCharacterByID(ctx, winnerID)
		if err != nil {
			return err
		}
		loser, err := tx.GetCharacterByID(ctx, loserID)
		if err != nil {
			return err
		}
		if winner.ListID != loser.ListID {
			return ErrDifferentLists
		}

		newWinner, newLoser := elo.Update(
			elo.Rating{Elo: winner.Elo, GamesPlayed: winner.GamesPlayed},
			elo.Rating{Elo: loser.Elo, GamesPlayed: loser.GamesPlayed},
		)
		if err := tx.setRating(ctx, winner.ID, newWinner); err != nil {
			return err
		}
		if err := tx.setRating(ctx, loser.ID, newLoser); err != nil {
			return err
		}

		match, err = scanMatch(tx.q.QueryRowContext(ctx, `
			INSERT INTO matches (
				id,
				list_id,
				winner_id,
				loser_id,
				winner_elo_before,
				winner_elo_after,
				loser_elo_before,
				loser_elo_after,
				created_at
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
			RETURNING `+matchColumns+`;
		`,
			uuid.New(),
			winner.ListID,
			winner.ID,
			loser.ID,
			winner.Elo,
			newWinner.Elo,
			loser.Elo,
			newLoser.Elo,
		))
		return err
	})
	if err != nil {
		return Match{}, err
	}
	return match, nil
}

// returns sql.ErrNoRows if no character has that id
func (c *Client) setRating(ctx context.Context, id uuid.UUID, rating elo.Rating) error {
	result, err := c.q.ExecContext(ctx, `
		UPDATE characters
		SET elo = ?, games_played = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?;
	`, rating.Elo, rating.GamesPlayed, id)
	if err != nil {
		return err
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if updated == 0 {
		return sql.ErrNoRows
	}
	return nil
}
