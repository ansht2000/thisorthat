package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Client struct {
	db *sql.DB
}

func NewClient(pathToDB string) (Client, error) {
	db, err := sql.Open("sqlite3", pathToDB)
	if err != nil {
		return Client{}, fmt.Errorf("could not initialize sqlite driver: %v", err)
	}
	c := Client{db: db}
	if err = c.autoMigrate(); err != nil {
		return Client{}, fmt.Errorf("could not run db migrations: %v", err)
	}
	return c, nil
}

func (c *Client) autoMigrate() error {
	listTable := `
	CREATE TABLE IF NOT EXISTS lists (
    	id UUID PRIMARY KEY,
    	name TEXT NOT NULL,
    	created_at TIMESTAMP NOT NULL,
    	updated_at TIMESTAMP NOT NULL
	);
	`
	_, err := c.db.Exec(listTable)
	if err != nil {
		return fmt.Errorf("could not run list table creation query: %v", err)
	}

	characterTable := `
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
	`
	_, err = c.db.Exec(characterTable)
	if err != nil {
		return fmt.Errorf("could not run character table creation query: %v", err)
	}

	return nil
}

func (c *Client) Reset() error {
	if _, err := c.db.Exec("DELETE FROM lists;"); err != nil {
		return fmt.Errorf("failed to reset table lists: %v", err)
	}
	if _, err := c.db.Exec("DELETE FROM characters;"); err != nil {
		return fmt.Errorf("failed to reset table chracters: %v", err)
	}
	return nil
}
