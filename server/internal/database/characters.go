package database

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
)

type Character struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	PictureURL string `json:"picture_url"`
	Elo int `json:"elo"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	ListID string `json:"list_id"`
}

type CreateCharacterParams struct {
	Name string `json:"name"`
	PictureURL string `json:"picture_url"`
	ListID uuid.UUID `json:"list_id"`
}

func (c *Client) CreateCharacter(ctx context.Context, params CreateCharacterParams) (Character, error) {
	id := uuid.New()
	query, err := c.db.Prepare(`
		INSERT INTO characters (
			id,
			name,
			picture_url,
			created_at,
			updated_at,
			list_id
		)
		VALUES (
			?,
			?,
			?,
			CURRENT_TIMESTAMP,
			CURRENT_TIMESTAMP,
			?
		)
		RETURNING *;
	`)
	if err != nil {
		log.Println(err)
		return Character{}, err
	}
	defer query.Close()

	var character Character
	if err = query.QueryRowContext(
		ctx,
		id,
		params.Name,
		params.PictureURL,
		params.ListID,
	).Scan(
		&character.ID,
		&character.Name,
		&character.PictureURL,
		&character.Elo,
		&character.CreatedAt,
		&character.UpdatedAt,
		&character.ListID,
	); err != nil {
		return Character{}, err
	}

	return character, nil
}

func (c *Client) GetCharacterByID(ctx context.Context, id uuid.UUID) (Character, error) {
	query, err := c.db.Prepare(`
		SELECT * FROM characters
		WHERE id = ?;
	`)
	if err != nil {
		return Character{}, err
	}
	defer query.Close()

	var character Character
	if err = query.QueryRowContext(ctx, id).Scan(
		&character.ID,
		&character.Name,
		&character.PictureURL,
		&character.Elo,
		&character.CreatedAt,
		&character.UpdatedAt,
		&character.ListID,
	); err != nil {
		return Character{}, err
	}

	return character, nil
}

func (c *Client) GetCharactersByListID(ctx context.Context, listID uuid.UUID) ([]Character, error) {
	query, err := c.db.Prepare(`
		SELECT * FROM characters
		WHERE list_id = ?
	`)
	if err != nil {
		return nil, err
	}
	defer query.Close()
	rows, err := query.QueryContext(ctx, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	characters := []Character{}
	for rows.Next() {
		var character Character
		if err := rows.Scan(
			&character.ID,
			&character.Name,
			&character.PictureURL,
			&character.Elo,
			&character.CreatedAt,
			&character.UpdatedAt,
			&character.ListID,
		); err != nil {
			return nil, err
		}
		characters = append(characters, character)
	}

	return characters, nil
}