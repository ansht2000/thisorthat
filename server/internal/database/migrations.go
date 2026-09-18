package database

import (
	"context"
	"fmt"
)

type migration struct {
	name string
	up   string
}

// applied in order, each exactly once, and a database's schema version is the
// number of these it has run. only ever add to the end: changing an entry that
// has already run somewhere won't run it again there
var migrations = []migration{
	{
		// IF NOT EXISTS on the first two so databases made before versioning keep their data
		name: "create lists",
		up: `
		CREATE TABLE IF NOT EXISTS lists (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
		`,
	},
	{
		name: "create characters",
		up: `
		CREATE TABLE IF NOT EXISTS characters (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			picture_url TEXT NOT NULL,
			elo INTEGER NOT NULL DEFAULT 1200,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			list_id UUID NOT NULL,
			FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE
		);
		`,
	},
	{
		// foreign keys used to be off, so deleting a list didn't cascade to its characters
		name: "delete orphaned characters",
		up: `
		DELETE FROM characters
		WHERE list_id NOT IN (SELECT id FROM lists);
		`,
	},
	{
		name: "add characters.games_played",
		up: `
		ALTER TABLE characters
		ADD COLUMN games_played INTEGER NOT NULL DEFAULT 0;
		`,
	},
	{
		// one row per vote, with both ratings before and after so history can show the swing
		name: "create matches",
		up: `
		CREATE TABLE matches (
			id UUID PRIMARY KEY,
			list_id UUID NOT NULL,
			winner_id UUID NOT NULL,
			loser_id UUID NOT NULL,
			winner_elo_before INTEGER NOT NULL,
			winner_elo_after INTEGER NOT NULL,
			loser_elo_before INTEGER NOT NULL,
			loser_elo_after INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE,
			FOREIGN KEY (winner_id) REFERENCES characters(id) ON DELETE CASCADE,
			FOREIGN KEY (loser_id) REFERENCES characters(id) ON DELETE CASCADE,
			CHECK (winner_id <> loser_id)
		);
		CREATE INDEX idx_matches_list_created ON matches (list_id, created_at DESC);
		`,
	},
	{
		name: "index characters by list and elo",
		up: `
		CREATE INDEX idx_characters_list_elo ON characters (list_id, elo DESC);
		`,
	},
}

func (c *Client) migrate(ctx context.Context) error {
	if _, err := c.q.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP NOT NULL
		);
	`); err != nil {
		return fmt.Errorf("could not create schema_migrations table: %w", err)
	}

	current, err := c.schemaVersion(ctx)
	if err != nil {
		return err
	}
	if current > len(migrations) {
		return fmt.Errorf("db is at schema version %d but this build only knows about %d migrations", current, len(migrations))
	}

	for i := current; i < len(migrations); i++ {
		version, m := i+1, migrations[i]
		// the schema change and its version row commit together, so a failed
		// migration leaves the db exactly as it was and runs again next start
		err := c.WithTx(ctx, func(tx *Client) error {
			if _, err := tx.q.ExecContext(ctx, m.up); err != nil {
				return err
			}
			_, err := tx.q.ExecContext(ctx, `
				INSERT INTO schema_migrations (version, name, applied_at)
				VALUES (?, ?, CURRENT_TIMESTAMP);
			`, version, m.name)
			return err
		})
		if err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", version, m.name, err)
		}
	}
	return nil
}

func (c *Client) schemaVersion(ctx context.Context) (int, error) {
	var version int
	if err := c.q.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(version), 0) FROM schema_migrations;
	`).Scan(&version); err != nil {
		return 0, fmt.Errorf("could not read schema version: %w", err)
	}
	return version, nil
}
