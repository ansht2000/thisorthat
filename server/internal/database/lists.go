package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type List struct {
	ID uuid.UUID `json:"id"`
	Name string `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateListParams struct {
	Name string `json:"name"`
}

func (c *Client) CreateList(ctx context.Context, params CreateListParams) (List, error) {
	id := uuid.New()
	query, err := c.db.Prepare(`
		INSERT INTO lists (id, name, created_at, updated_at)
			VALUES (
				?,
				?,
				CURRENT_TIMESTAMP,
				CURRENT_TIMESTAMP
			)
		RETURNING *;
	`)
	if err != nil {
		return List{}, err
	}
	defer query.Close()

	var list List
	if err = query.QueryRowContext(ctx, id, params.Name).Scan(
		&list.ID,
		&list.Name,
		&list.CreatedAt,
		&list.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return List{}, err
		}
		return List{}, err
	}

	return list, nil
}

func (c *Client) GetLists(ctx context.Context) ([]List, error) {
	// preparing a query is conventional according
	// to the docs, but it feels kinda pointless
	// here so i might remove
	query, err := c.db.Prepare(`
		SELECT * FROM lists
		ORDER BY name;
	`)
	if err != nil {
		return nil, err
	}
	defer query.Close()
	rows, err := query.QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lists := []List{}
	// prepare each row to be read
	for rows.Next() {
		var list List
		if err := rows.Scan(
			&list.ID,
			&list.Name,
			&list.CreatedAt,
			&list.UpdatedAt,
		); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return lists, err
			}
			return nil, err
		}
		lists = append(lists, list)
	}
	return lists, nil
}

func (c *Client) GetListByID(ctx context.Context, id uuid.UUID) (List, error) {
	query, err := c.db.Prepare(`
		SELECT * FROM lists
		WHERE id = ?;
	`)
	if err != nil {
		return List{}, err
	}
	defer query.Close()
	
	var list List
	if err = query.QueryRowContext(ctx, id).Scan(
		&list.ID,
		&list.Name,
		&list.CreatedAt,
		&list.UpdatedAt,
	); err != nil {
		return List{}, err
	}

	return list, nil
}

func (c *Client) DeleteLists() error {
	query, err := c.db.Prepare(`
		DELETE FROM lists;
	`)
	if err != nil {
		return err
	}
	defer query.Close()

	_, err = query.Exec()
	return err
}