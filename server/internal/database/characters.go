package database

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrListNotFound = errors.New("list not found")

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

// a character with its place in its list, where ties share a rank
type LeaderboardEntry struct {
	Rank int `json:"rank"`
	Character
}

// same order characterFields lists them in
const characterColumns = `id, name, picture_url, elo, games_played, created_at, updated_at, list_id`

// scan destinations for characterColumns, split out so queries that
// select extra columns alongside a character can still reuse them
func characterFields(character *Character) []any {
	return []any{
		&character.ID,
		&character.Name,
		&character.PictureURL,
		&character.Elo,
		&character.GamesPlayed,
		&character.CreatedAt,
		&character.UpdatedAt,
		&character.ListID,
	}
}

func scanCharacter(row rowScanner) (Character, error) {
	var character Character
	if err := row.Scan(characterFields(&character)...); err != nil {
		return Character{}, err
	}
	return character, nil
}

func scanLeaderboardEntry(row rowScanner) (LeaderboardEntry, error) {
	var entry LeaderboardEntry
	if err := row.Scan(append([]any{&entry.Rank}, characterFields(&entry.Character)...)...); err != nil {
		return LeaderboardEntry{}, err
	}
	return entry, nil
}

// returns ErrListNotFound if params.ListID doesn't match a list
func (c *Client) CreateCharacter(ctx context.Context, params CreateCharacterParams) (Character, error) {
	character, err := scanCharacter(c.q.QueryRowContext(ctx, `
		INSERT INTO characters (id, name, picture_url, created_at, updated_at, list_id)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, ?)
		RETURNING `+characterColumns+`;
	`, uuid.New(), params.Name, params.PictureURL, params.ListID))
	if err != nil {
		if isForeignKeyError(err) {
			return Character{}, ErrListNotFound
		}
		return Character{}, err
	}
	return character, nil
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

// highest elo first. RANK() gives tied characters the same rank
// and skips the ranks they took up, so ties go 1, 2, 2, 4
func (c *Client) GetLeaderboard(ctx context.Context, listID uuid.UUID) ([]LeaderboardEntry, error) {
	return queryAll(ctx, c.q, scanLeaderboardEntry, `
		SELECT RANK() OVER (ORDER BY elo DESC), `+characterColumns+` FROM characters
		WHERE list_id = ?
		ORDER BY elo DESC, name;
	`, listID)
}
