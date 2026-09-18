package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// applied to every connection in the pool:
// foreign keys are off by default in sqlite, so without them ON DELETE CASCADE does nothing,
// busy_timeout makes a second writer wait for the lock instead of failing with SQLITE_BUSY,
// WAL lets reads keep going while a write is happening,
// and txlock=immediate grabs the write lock at BEGIN so two read-then-write
// transactions can't both read and then deadlock trying to upgrade
const dsnParams = "_foreign_keys=on&_busy_timeout=5000&_journal_mode=WAL&_txlock=immediate"

// the methods *sql.DB and *sql.Tx have in common, so every query
// method works the same whether or not it's inside a transaction
type querier interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// satisfied by both *sql.Row and *sql.Rows, so one scan function
// per model handles single row and multi row queries
type rowScanner interface {
	Scan(dest ...any) error
}

type Client struct {
	db *sql.DB
	// either db itself or the transaction this client was handed by WithTx
	q querier
}

func NewClient(pathToDB string) (Client, error) {
	db, err := sql.Open("sqlite3", withDSNParams(pathToDB))
	if err != nil {
		return Client{}, fmt.Errorf("could not initialize sqlite driver: %w", err)
	}
	// every connection to an in memory db gets its own separate empty database,
	// so keep the pool at one connection or queries end up on different dbs
	if isInMemory(pathToDB) {
		db.SetMaxOpenConns(1)
	}

	c := Client{db: db, q: db}
	if err = c.migrate(context.Background()); err != nil {
		db.Close()
		return Client{}, fmt.Errorf("could not run db migrations: %w", err)
	}
	return c, nil
}

func (c *Client) Close() error {
	return c.db.Close()
}

// WithTx runs fn in a transaction, committing if fn returns nil and rolling back otherwise.
// fn has to use the client it's given, not the outer one, for its queries to be part of the transaction.
// calling WithTx on a client that's already in a transaction just reuses it
func (c *Client) WithTx(ctx context.Context, fn func(tx *Client) error) error {
	if _, inTx := c.q.(*sql.Tx); inTx {
		return fn(c)
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	// no-op after a successful commit, and still releases the lock if fn panics
	defer tx.Rollback()

	if err := fn(&Client{db: c.db, q: tx}); err != nil {
		return err
	}
	return tx.Commit()
}

// runs a query and scans every row it returns with scan
func queryAll[T any](ctx context.Context, q querier, scan func(rowScanner) (T, error), query string, args ...any) ([]T, error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// start non nil so an empty result encodes as [] instead of null
	items := []T{}
	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func withDSNParams(pathToDB string) string {
	sep := "?"
	if strings.Contains(pathToDB, "?") {
		sep = "&"
	}
	return pathToDB + sep + dsnParams
}

func isInMemory(pathToDB string) bool {
	return strings.Contains(pathToDB, ":memory:") || strings.Contains(pathToDB, "mode=memory")
}
