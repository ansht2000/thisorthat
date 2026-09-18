package database

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// in memory clients only have one connection, so a transaction that never lets go of it
// makes every later query wait forever. the timeout turns that into a failure in seconds
// instead of a hang until go test's own 10 minute limit
func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func newTestClient(t *testing.T) *Client {
	t.Helper()
	c, err := NewClient(":memory:")
	if err != nil {
		t.Fatalf("could not create test client: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return &c
}

func mustCreateList(t *testing.T, c *Client, name string) List {
	t.Helper()
	list, err := c.CreateList(testContext(t), CreateListParams{Name: name})
	if err != nil {
		t.Fatalf("could not create list %q: %v", name, err)
	}
	return list
}

func mustCreateCharacter(t *testing.T, c *Client, listID uuid.UUID, name string) Character {
	t.Helper()
	character, err := c.CreateCharacter(testContext(t), CreateCharacterParams{Name: name, ListID: listID})
	if err != nil {
		t.Fatalf("could not create character %q: %v", name, err)
	}
	return character
}

func TestMigrationsReachLatestVersion(t *testing.T) {
	c := newTestClient(t)
	version, err := c.schemaVersion(testContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if version != len(migrations) {
		t.Errorf("expected schema version %d, got %d", len(migrations), version)
	}
}

func TestMigrationsOnlyRunOnce(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	first, err := NewClient(path)
	if err != nil {
		t.Fatal(err)
	}
	list := mustCreateList(t, &first, "invincible")
	first.Close()

	// would fail on the ALTER TABLE if any migration ran a second time
	second, err := NewClient(path)
	if err != nil {
		t.Fatalf("reopening a migrated db failed: %v", err)
	}
	defer second.Close()

	lists, err := second.GetLists(testContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 1 || lists[0].ID != list.ID {
		t.Errorf("expected data to survive reopening, got %+v", lists)
	}
}

// databases made before versioning have lists and characters but no schema_migrations,
// no games_played, and possibly characters left behind by list deletes
func TestMigrateDatabaseFromBeforeVersioning(t *testing.T) {
	path := filepath.Join(t.TempDir(), "legacy.db")
	legacy, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	listID, keptID, orphanID := uuid.New(), uuid.New(), uuid.New()
	if _, err := legacy.Exec(`
		CREATE TABLE lists (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		);
		CREATE TABLE characters (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			picture_url TEXT NOT NULL,
			elo INTEGER NOT NULL DEFAULT 1200,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			list_id UUID NOT NULL,
			FOREIGN KEY (list_id) REFERENCES lists(id) ON DELETE CASCADE
		);
	`); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`
		INSERT INTO lists VALUES (?, 'invincible', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	`, listID); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`
		INSERT INTO characters VALUES
			(?, 'Mark Grayson', '', 1250, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?),
			(?, 'Orphan', '', 1200, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?);
	`, keptID, listID, orphanID, uuid.New()); err != nil {
		t.Fatal(err)
	}
	legacy.Close()

	c, err := NewClient(path)
	if err != nil {
		t.Fatalf("migrating legacy db failed: %v", err)
	}
	defer c.Close()
	ctx := testContext(t)

	kept, err := c.GetCharacterByID(ctx, keptID)
	if err != nil {
		t.Fatalf("character with a real list should survive: %v", err)
	}
	if kept.Elo != 1250 || kept.GamesPlayed != 0 || kept.ListID != listID {
		t.Errorf("character data changed during migration: %+v", kept)
	}
	if _, err := c.GetCharacterByID(ctx, orphanID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected orphaned character to be deleted, got err %v", err)
	}
}

func TestCreateCharacterRequiresExistingList(t *testing.T) {
	c := newTestClient(t)
	_, err := c.CreateCharacter(testContext(t), CreateCharacterParams{Name: "Nobody", ListID: uuid.New()})
	if err == nil {
		t.Fatal("expected a foreign key error for a list that doesn't exist")
	}
}

func TestDeleteListsCascadesToCharacters(t *testing.T) {
	c := newTestClient(t)
	ctx := testContext(t)
	list := mustCreateList(t, c, "invincible")
	character := mustCreateCharacter(t, c, list.ID, "Mark Grayson")

	if err := c.DeleteLists(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetCharacterByID(ctx, character.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected character to be deleted with its list, got err %v", err)
	}
}

func TestCharacters(t *testing.T) {
	c := newTestClient(t)
	ctx := testContext(t)
	list := mustCreateList(t, c, "invincible")
	other := mustCreateList(t, c, "the boys")

	nolan := mustCreateCharacter(t, c, list.ID, "Nolan Grayson")
	mark := mustCreateCharacter(t, c, list.ID, "Mark Grayson")
	mustCreateCharacter(t, c, other.ID, "Homelander")

	if mark.Elo != 1200 || mark.GamesPlayed != 0 || mark.ListID != list.ID {
		t.Errorf("unexpected defaults on new character: %+v", mark)
	}

	got, err := c.GetCharacterByID(ctx, mark.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got != mark {
		t.Errorf("expected %+v, got %+v", mark, got)
	}
	if _, err := c.GetCharacterByID(ctx, uuid.New()); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected sql.ErrNoRows for unknown id, got %v", err)
	}

	inList, err := c.GetCharactersByListID(ctx, list.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(inList) != 2 || inList[0].ID != mark.ID || inList[1].ID != nolan.ID {
		t.Errorf("expected [Mark, Nolan] sorted by name, got %+v", inList)
	}

	empty, err := c.GetCharactersByListID(ctx, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if empty == nil || len(empty) != 0 {
		t.Errorf("expected an empty non nil slice for an unknown list, got %#v", empty)
	}
}

func TestGetListsSortedByName(t *testing.T) {
	c := newTestClient(t)
	for _, name := range []string{"the boys", "avatar", "invincible"} {
		mustCreateList(t, c, name)
	}
	lists, err := c.GetLists(testContext(t))
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 3 || lists[0].Name != "avatar" || lists[1].Name != "invincible" || lists[2].Name != "the boys" {
		t.Errorf("expected lists sorted by name, got %+v", lists)
	}
}

func TestWithTxCommits(t *testing.T) {
	c := newTestClient(t)
	ctx := testContext(t)
	err := c.WithTx(ctx, func(tx *Client) error {
		_, err := tx.CreateList(ctx, CreateListParams{Name: "invincible"})
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	lists, err := c.GetLists(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 1 {
		t.Errorf("expected committed list, got %+v", lists)
	}
}

func TestWithTxRollsBackOnError(t *testing.T) {
	c := newTestClient(t)
	ctx := testContext(t)
	errBoom := errors.New("boom")
	err := c.WithTx(ctx, func(tx *Client) error {
		if _, err := tx.CreateList(ctx, CreateListParams{Name: "invincible"}); err != nil {
			return err
		}
		return errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected fn's error back, got %v", err)
	}
	lists, err := c.GetLists(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 0 {
		t.Errorf("expected rollback to discard the list, got %+v", lists)
	}
}

func TestWithTxRollsBackOnPanic(t *testing.T) {
	c := newTestClient(t)
	ctx := testContext(t)
	func() {
		defer func() { recover() }()
		c.WithTx(ctx, func(tx *Client) error {
			tx.CreateList(ctx, CreateListParams{Name: "invincible"})
			panic("boom")
		})
	}()

	// times out if the panicking transaction never let go of the connection
	lists, err := c.GetLists(ctx)
	if err != nil {
		t.Fatalf("connection wasn't released after panic: %v", err)
	}
	if len(lists) != 0 {
		t.Errorf("expected rollback to discard the list, got %+v", lists)
	}
}

func TestWithTxReusesOuterTransaction(t *testing.T) {
	c := newTestClient(t)
	// with one connection, starting a second transaction would block until the timeout
	ctx := testContext(t)

	errBoom := errors.New("boom")
	err := c.WithTx(ctx, func(outer *Client) error {
		if err := outer.WithTx(ctx, func(inner *Client) error {
			_, err := inner.CreateList(ctx, CreateListParams{Name: "invincible"})
			return err
		}); err != nil {
			return err
		}
		return errBoom
	})
	if !errors.Is(err, errBoom) {
		t.Fatalf("expected outer fn's error back, got %v", err)
	}
	lists, err := c.GetLists(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 0 {
		t.Errorf("expected the inner write to roll back with the outer transaction, got %+v", lists)
	}
}
