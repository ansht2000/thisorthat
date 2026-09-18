package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type List struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateListParams struct {
	Name string `json:"name"`
}

// same order scanList reads them in
const listColumns = `id, name, created_at, updated_at`

func scanList(row rowScanner) (List, error) {
	var list List
	if err := row.Scan(
		&list.ID,
		&list.Name,
		&list.CreatedAt,
		&list.UpdatedAt,
	); err != nil {
		return List{}, err
	}
	return list, nil
}

func (c *Client) CreateList(ctx context.Context, params CreateListParams) (List, error) {
	return scanList(c.q.QueryRowContext(ctx, `
		INSERT INTO lists (id, name, created_at, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING `+listColumns+`;
	`, uuid.New(), params.Name))
}

func (c *Client) GetLists(ctx context.Context) ([]List, error) {
	return queryAll(ctx, c.q, scanList, `
		SELECT `+listColumns+` FROM lists
		ORDER BY name;
	`)
}

func (c *Client) GetListByID(ctx context.Context, id uuid.UUID) (List, error) {
	return scanList(c.q.QueryRowContext(ctx, `
		SELECT `+listColumns+` FROM lists
		WHERE id = ?;
	`, id))
}

// characters and matches go with their lists through ON DELETE CASCADE
func (c *Client) DeleteLists(ctx context.Context) error {
	_, err := c.q.ExecContext(ctx, `DELETE FROM lists;`)
	return err
}
