package seed

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"strings"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/ansht2000/thisorthat/internal/elo"
	"github.com/google/uuid"
)

// the franchises seeded when no file is given. characters are listed roughly
// strongest first, which only matters to SimulateVotes
//
//go:embed franchises.json
var defaultFranchises []byte

type Franchise struct {
	Name       string      `json:"name"`
	Characters []Character `json:"characters"`
}

type Character struct {
	Name       string `json:"name"`
	PictureURL string `json:"picture_url"`
}

type Result struct {
	ListsCreated      int
	CharactersCreated int
}

func Default() ([]Franchise, error) {
	return Parse(defaultFranchises)
}

// Parse reads a seed file: a json array of franchises, each with a name and its characters
func Parse(data []byte) ([]Franchise, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	// so a typo like "pictureUrl" is an error instead of a silently missing picture
	decoder.DisallowUnknownFields()
	var franchises []Franchise
	if err := decoder.Decode(&franchises); err != nil {
		return nil, fmt.Errorf("invalid seed file: %w", err)
	}

	seenFranchises := map[string]bool{}
	for _, franchise := range franchises {
		key := normalize(franchise.Name)
		if key == "" {
			return nil, fmt.Errorf("invalid seed file: a franchise is missing its name")
		}
		if seenFranchises[key] {
			return nil, fmt.Errorf("invalid seed file: franchise %q is listed twice", franchise.Name)
		}
		seenFranchises[key] = true

		seenCharacters := map[string]bool{}
		for _, character := range franchise.Characters {
			key := normalize(character.Name)
			if key == "" {
				return nil, fmt.Errorf("invalid seed file: a character in %q is missing its name", franchise.Name)
			}
			if seenCharacters[key] {
				return nil, fmt.Errorf("invalid seed file: %q is listed twice in %q", character.Name, franchise.Name)
			}
			seenCharacters[key] = true
		}
	}
	return franchises, nil
}

// Apply adds any franchises and characters from the file that the database doesn't have yet,
// all in one transaction. names match ignoring case and surrounding spaces, so running it
// again adds nothing, and anything already there, ratings included, is left alone
func Apply(ctx context.Context, db *database.Client, franchises []Franchise) (Result, error) {
	var result Result
	err := db.WithTx(ctx, func(tx *database.Client) error {
		lists, err := tx.GetLists(ctx)
		if err != nil {
			return err
		}
		listsByName := make(map[string]database.List, len(lists))
		for _, list := range lists {
			listsByName[normalize(list.Name)] = list
		}

		for _, franchise := range franchises {
			list, exists := listsByName[normalize(franchise.Name)]
			if !exists {
				list, err = tx.CreateList(ctx, database.CreateListParams{Name: franchise.Name})
				if err != nil {
					return fmt.Errorf("could not create %q: %w", franchise.Name, err)
				}
				listsByName[normalize(franchise.Name)] = list
				result.ListsCreated++
			}

			existing, err := tx.GetCharactersByListID(ctx, list.ID)
			if err != nil {
				return err
			}
			have := make(map[string]bool, len(existing))
			for _, character := range existing {
				have[normalize(character.Name)] = true
			}
			for _, character := range franchise.Characters {
				if have[normalize(character.Name)] {
					continue
				}
				if _, err := tx.CreateCharacter(ctx, database.CreateCharacterParams{
					Name:       character.Name,
					PictureURL: character.PictureURL,
					ListID:     list.ID,
				}); err != nil {
					return fmt.Errorf("could not create %q in %q: %w", character.Name, franchise.Name, err)
				}
				have[normalize(character.Name)] = true
				result.CharactersCreated++
			}
		}
		return nil
	})
	return result, err
}

const (
	// hidden strength of the first character in order, before the per place drop
	topStrength = 1500.0
	// hidden strength lost per place down the order. with 75, a character beats the one
	// just below it about 61% of the time and one ten places down about 99% of the time
	strengthStep = 75.0
)

// SimulateVotes casts votes on a list the way a crowd that mostly agrees with order would:
// characters earlier in order usually beat later ones, but upsets happen. pairs come from the
// same matchmaking the site uses and every vote goes through RecordMatch, so the ratings,
// games played and match history end up just like real voting would leave them.
// characters not in order are treated as middle of the pack. meant for demos and screenshots
func SimulateVotes(ctx context.Context, db *database.Client, listID uuid.UUID, order []string, votes int, rng *rand.Rand) error {
	characters, err := db.GetCharactersByListID(ctx, listID)
	if err != nil {
		return err
	}
	if len(characters) < 2 {
		return nil
	}

	place := make(map[string]int, len(order))
	for i, name := range order {
		place[normalize(name)] = i
	}
	ratings := make([]elo.Rating, len(characters))
	strengths := make([]float64, len(characters))
	for i, character := range characters {
		ratings[i] = elo.Rating{Elo: character.Elo, GamesPlayed: character.GamesPlayed}
		p, listed := place[normalize(character.Name)]
		if !listed {
			p = len(order) / 2
		}
		strengths[i] = topStrength - strengthStep*float64(p)
	}

	return db.WithTx(ctx, func(tx *database.Client) error {
		// like the site, don't serve the pair that was just voted on
		last := [2]int{-1, -1}
		for range votes {
			i, j, err := elo.PickPair(ratings, func(a, b int) bool { return a == last[0] && b == last[1] }, rng.Float64)
			if err != nil {
				return err
			}
			winner, loser := i, j
			if rng.Float64() >= winChance(strengths[i], strengths[j]) {
				winner, loser = j, i
			}

			match, err := tx.RecordMatch(ctx, characters[winner].ID, characters[loser].ID)
			if err != nil {
				return err
			}
			ratings[winner] = elo.Rating{Elo: match.WinnerEloAfter, GamesPlayed: ratings[winner].GamesPlayed + 1}
			ratings[loser] = elo.Rating{Elo: match.LoserEloAfter, GamesPlayed: ratings[loser].GamesPlayed + 1}
			last = [2]int{min(i, j), max(i, j)}
		}
		return nil
	})
}

// chance a character with strength a beats one with strength b, on the elo scale
func winChance(a float64, b float64) float64 {
	return 1 / (1 + math.Pow(10, (b-a)/400))
}

func normalize(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}
