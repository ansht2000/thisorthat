package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Character struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	PictureURL  string    `json:"picture_url"`
	Elo         int       `json:"elo"`
	GamesPlayed int       `json:"games_played"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ListID      uuid.UUID `json:"list_id"`
}

type CreateCharacterParams struct {
	Name       string    `json:"name"`
	PictureURL string    `json:"picture_url"`
	ListID     uuid.UUID `json:"list_id"`
}

// same order scanCharacter reads them in
const characterColumns = `id, name, picture_url, elo, games_played, created_at, updated_at, list_id`

func scanCharacter(row rowScanner) (Character, error) {
	var character Character
	if err := row.Scan(
		&character.ID,
		&character.Name,
		&character.PictureURL,
		&character.Elo,
		&character.GamesPlayed,
		&character.CreatedAt,
		&character.UpdatedAt,
		&character.ListID,
	); err != nil {
		return Character{}, err
	}
	return character, nil
}

func (c *Client) CreateCharacter(ctx context.Context, params CreateCharacterParams) (Character, error) {
	return scanCharacter(c.q.QueryRowContext(ctx, `
		INSERT INTO characters (id, name, picture_url, created_at, updated_at, list_id)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		RETURNING `+characterColumns+`;
	`, uuid.New(), params.Name, params.PictureURL, params.ListID))
}

func (c *Client) GetCharacterByID(ctx context.Context, id uuid.UUID) (Character, error) {
	return scanCharacter(c.q.QueryRowContext(ctx, `
		SELECT `+characterColumns+` FROM characters
		WHERE id = ?;
	`, id))
}

func (c *Client) GetCharactersByListID(ctx context.Context, listID uuid.UUID) ([]Character, error) {
	return queryAll(ctx, c.q, scanCharacter, `
		SELECT `+characterColumns+` FROM characters
		WHERE list_id = ?
		ORDER BY name;
	`, listID)
}
