package seed

import (
	"context"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/ansht2000/thisorthat/internal/database"
)

func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func newTestClient(t *testing.T) *database.Client {
	t.Helper()
	c, err := database.NewClient(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return &c
}

func franchise(name string, characters ...string) Franchise {
	f := Franchise{Name: name}
	for _, character := range characters {
		f.Characters = append(f.Characters, Character{Name: character})
	}
	return f
}

func TestDefaultFranchisesAreValid(t *testing.T) {
	franchises, err := Default()
	if err != nil {
		t.Fatal(err)
	}
	if len(franchises) < 5 {
		t.Errorf("expected at least 5 built in franchises, got %d", len(franchises))
	}
	for _, f := range franchises {
		// the matchup endpoint needs two, and a leaderboard of two isn't much of one
		if len(f.Characters) < 8 {
			t.Errorf("expected %q to have at least 8 characters, got %d", f.Name, len(f.Characters))
		}
	}
}

func TestParseRejectsBadFiles(t *testing.T) {
	cases := map[string]string{
		"not json":               `{`,
		"unknown field":          `[{"name": "invincible", "characters": [{"name": "Mark", "pictureUrl": "x"}]}]`,
		"franchise without name": `[{"name": " ", "characters": []}]`,
		"character without name": `[{"name": "invincible", "characters": [{"name": ""}]}]`,
		"franchise twice":        `[{"name": "invincible"}, {"name": "Invincible "}]`,
		"character twice":        `[{"name": "invincible", "characters": [{"name": "Omni Man"}, {"name": "omni man"}]}]`,
	}
	for name, data := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse([]byte(data)); err == nil {
				t.Error("expected an error")
			}
		})
	}
}

func TestApplyOnlyAddsWhatsMissing(t *testing.T) {
	db := newTestClient(t)
	ctx := testContext(t)
	franchises := []Franchise{
		franchise("the boys", "Homelander", "Starlight", "Hughie Campbell"),
		franchise("invincible", "Omni Man", "Atom Eve"),
	}

	// already in the db under slightly different spelling, with a rating that has to survive
	existingList, err := db.CreateList(ctx, database.CreateListParams{Name: "The Boys"})
	if err != nil {
		t.Fatal(err)
	}
	homelander, err := db.CreateCharacter(ctx, database.CreateCharacterParams{Name: " homelander ", ListID: existingList.ID})
	if err != nil {
		t.Fatal(err)
	}
	starlight, err := db.CreateCharacter(ctx, database.CreateCharacterParams{Name: "Starlight", ListID: existingList.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.RecordMatch(ctx, homelander.ID, starlight.ID); err != nil {
		t.Fatal(err)
	}

	result, err := Apply(ctx, db, franchises)
	if err != nil {
		t.Fatal(err)
	}
	if result.ListsCreated != 1 || result.CharactersCreated != 3 {
		t.Errorf("expected invincible plus Hughie, Omni Man and Atom Eve to be added, got %+v", result)
	}

	again, err := Apply(ctx, db, franchises)
	if err != nil {
		t.Fatal(err)
	}
	if again != (Result{}) {
		t.Errorf("expected a second run to add nothing, got %+v", again)
	}

	lists, err := db.GetLists(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(lists) != 2 {
		t.Fatalf("expected 2 lists, got %+v", lists)
	}
	boys, err := db.GetCharactersByListID(ctx, existingList.ID)
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, character := range boys {
		names = append(names, strings.TrimSpace(character.Name))
		if character.ID == homelander.ID && (character.Elo == 1200 || character.GamesPlayed != 1) {
			t.Errorf("expected homelander's existing rating to be left alone, got %+v", character)
		}
	}
	if strings.Join(names, ", ") != "homelander, Hughie Campbell, Starlight" {
		t.Errorf("expected the existing two plus Hughie, got %v", names)
	}
}

func TestSimulateVotesFollowsTheOrder(t *testing.T) {
	db := newTestClient(t)
	ctx := testContext(t)
	order := []string{"First", "Second", "Third", "Fourth", "Fifth", "Sixth"}
	if _, err := Apply(ctx, db, []Franchise{franchise("test", order...)}); err != nil {
		t.Fatal(err)
	}
	lists, err := db.GetLists(ctx)
	if err != nil {
		t.Fatal(err)
	}

	const votes = 600
	if err := SimulateVotes(ctx, db, lists[0].ID, order, votes, rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Fatal(err)
	}

	matches, err := db.GetMatchesByListID(ctx, lists[0].ID, votes+1)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != votes {
		t.Errorf("expected %d matches recorded, got %d", votes, len(matches))
	}
	samePair := func(a, b database.Match) bool {
		return (a.WinnerID == b.WinnerID && a.LoserID == b.LoserID) || (a.WinnerID == b.LoserID && a.LoserID == b.WinnerID)
	}
	for i := 1; i < len(matches); i++ {
		if samePair(matches[i], matches[i-1]) {
			t.Fatalf("expected no pair twice in a row, got a repeat at match %d", i)
		}
	}

	leaderboard, err := db.GetLeaderboard(ctx, lists[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	eloOf := map[string]int{}
	for _, entry := range leaderboard {
		eloOf[entry.Name] = entry.Elo
	}
	top := eloOf["First"] + eloOf["Second"] + eloOf["Third"]
	bottom := eloOf["Fourth"] + eloOf["Fifth"] + eloOf["Sixth"]
	if eloOf["First"] <= eloOf["Sixth"] || top <= bottom {
		t.Errorf("expected the ratings to roughly follow the order, got %v", eloOf)
	}
}

func TestSimulateVotesNeedsTwoCharacters(t *testing.T) {
	db := newTestClient(t)
	ctx := testContext(t)
	if _, err := Apply(ctx, db, []Franchise{franchise("lonely", "Only One")}); err != nil {
		t.Fatal(err)
	}
	lists, err := db.GetLists(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := SimulateVotes(ctx, db, lists[0].ID, []string{"Only One"}, 10, rand.New(rand.NewPCG(1, 2))); err != nil {
		t.Errorf("expected nothing to happen, got %v", err)
	}
}
