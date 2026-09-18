package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	// keeps the logger middleware from printing every test request
	gin.DefaultWriter = io.Discard
	os.Exit(m.Run())
}

// fails a stuck query in seconds instead of hanging until go test's 10 minute limit
func testContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

type testAPI struct {
	t      *testing.T
	db     *database.Client
	router *gin.Engine
}

// a dev server on a fresh in memory db with no vote limit
func newTestAPI(t *testing.T) *testAPI {
	return newTestAPIWith(t, "dev", newIPRateLimiter(rate.Inf, 1))
}

func newTestAPIWith(t *testing.T, platform string, voteLimiter *ipRateLimiter) *testAPI {
	t.Helper()
	db, err := database.NewClient(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	cfg := &apiConfig{db: db, platform: platform, voteLimiter: voteLimiter}
	router, err := newRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return &testAPI{t: t, db: &cfg.db, router: router}
}

func (api *testAPI) serve(req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	api.router.ServeHTTP(rec, req)
	return rec
}

// a string body is sent as is so tests can send broken json, anything else is encoded
func newRequest(t *testing.T, method string, path string, body any) *http.Request {
	t.Helper()
	var reader io.Reader
	switch b := body.(type) {
	case nil:
	case string:
		reader = strings.NewReader(b)
	default:
		encoded, err := json.Marshal(b)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(encoded)
	}
	req := httptest.NewRequest(method, path, reader)
	if reader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func (api *testAPI) request(method string, path string, body any) *httptest.ResponseRecorder {
	api.t.Helper()
	return api.serve(newRequest(api.t, method, path, body))
}

func (api *testAPI) createList(name string) database.List {
	api.t.Helper()
	list, err := api.db.CreateList(testContext(api.t), database.CreateListParams{Name: name})
	if err != nil {
		api.t.Fatal(err)
	}
	return list
}

func (api *testAPI) createCharacter(listID uuid.UUID, name string) database.Character {
	api.t.Helper()
	character, err := api.db.CreateCharacter(testContext(api.t), database.CreateCharacterParams{Name: name, ListID: listID})
	if err != nil {
		api.t.Fatal(err)
	}
	return character
}

func expectStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("expected status %d, got %d with body %s", want, rec.Code, rec.Body.String())
	}
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatalf("could not decode response %q: %v", rec.Body.String(), err)
	}
	return v
}

func vote(winner uuid.UUID, loser uuid.UUID) map[string]uuid.UUID {
	return map[string]uuid.UUID{"winner_id": winner, "loser_id": loser}
}

func TestGetLists(t *testing.T) {
	api := newTestAPI(t)

	rec := api.request(http.MethodGet, "/lists", nil)
	expectStatus(t, rec, http.StatusOK)
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("expected an empty array when there are no lists, got %s", body)
	}

	api.createList("the boys")
	api.createList("invincible")
	rec = api.request(http.MethodGet, "/lists", nil)
	expectStatus(t, rec, http.StatusOK)
	lists := decode[[]database.List](t, rec)
	if len(lists) != 2 || lists[0].Name != "invincible" || lists[1].Name != "the boys" {
		t.Errorf("expected both lists sorted by name, got %+v", lists)
	}
}

func TestGetList(t *testing.T) {
	api := newTestAPI(t)
	list := api.createList("invincible")

	rec := api.request(http.MethodGet, "/lists/"+list.ID.String(), nil)
	expectStatus(t, rec, http.StatusOK)
	if got := decode[database.List](t, rec); got.ID != list.ID || got.Name != list.Name {
		t.Errorf("expected %+v, got %+v", list, got)
	}
}

// every /lists/:id route goes through lookupList, so they should all agree on these
func TestListRoutesRejectBadAndUnknownIDs(t *testing.T) {
	api := newTestAPI(t)
	for _, suffix := range []string{"", "/characters", "/leaderboard", "/matchup", "/matches"} {
		t.Run("GET /lists/:id"+suffix, func(t *testing.T) {
			expectStatus(t, api.request(http.MethodGet, "/lists/not-a-uuid"+suffix, nil), http.StatusBadRequest)
			expectStatus(t, api.request(http.MethodGet, "/lists/"+uuid.NewString()+suffix, nil), http.StatusNotFound)
		})
	}
}

func TestGetCharactersByListID(t *testing.T) {
	api := newTestAPI(t)
	list := api.createList("invincible")

	rec := api.request(http.MethodGet, "/lists/"+list.ID.String()+"/characters", nil)
	expectStatus(t, rec, http.StatusOK)
	if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
		t.Errorf("expected an empty array for a list with no characters, got %s", body)
	}

	api.createCharacter(list.ID, "Nolan Grayson")
	api.createCharacter(list.ID, "Mark Grayson")
	rec = api.request(http.MethodGet, "/lists/"+list.ID.String()+"/characters", nil)
	expectStatus(t, rec, http.StatusOK)
	characters := decode[[]database.Character](t, rec)
	if len(characters) != 2 || characters[0].Name != "Mark Grayson" || characters[1].Name != "Nolan Grayson" {
		t.Errorf("expected both characters sorted by name, got %+v", characters)
	}
}

func TestCreateList(t *testing.T) {
	api := newTestAPI(t)
	cases := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{"broken json", `{"name":`, http.StatusBadRequest},
		{"missing name", map[string]string{}, http.StatusBadRequest},
		{"blank name", map[string]string{"name": "  "}, http.StatusBadRequest},
		{"valid", map[string]string{"name": "invincible"}, http.StatusCreated},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectStatus(t, api.request(http.MethodPost, "/lists", tc.body), tc.wantStatus)
		})
	}
}

func TestCreateCharacter(t *testing.T) {
	api := newTestAPI(t)
	list := api.createList("invincible")
	cases := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{"broken json", `{"name":`, http.StatusBadRequest},
		{"missing name", map[string]any{"list_id": list.ID}, http.StatusBadRequest},
		{"missing list", map[string]any{"name": "Mark Grayson"}, http.StatusBadRequest},
		{"invalid list id", map[string]any{"name": "Mark Grayson", "list_id": "nope"}, http.StatusBadRequest},
		{"unknown list", map[string]any{"name": "Mark Grayson", "list_id": uuid.New()}, http.StatusNotFound},
		{"valid", map[string]any{"name": "Mark Grayson", "list_id": list.ID}, http.StatusCreated},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectStatus(t, api.request(http.MethodPost, "/characters", tc.body), tc.wantStatus)
		})
	}
}

func TestDevOnlyRoutesAreBlockedOutsideDev(t *testing.T) {
	api := newTestAPIWith(t, "prod", newIPRateLimiter(rate.Inf, 1))
	list := api.createList("invincible")

	expectStatus(t, api.request(http.MethodPost, "/lists", map[string]string{"name": "the boys"}), http.StatusForbidden)
	expectStatus(t, api.request(http.MethodPost, "/characters", map[string]any{"name": "Mark", "list_id": list.ID}), http.StatusForbidden)
	expectStatus(t, api.request(http.MethodPost, "/reset", nil), http.StatusForbidden)

	lists := decode[[]database.List](t, api.request(http.MethodGet, "/lists", nil))
	if len(lists) != 1 {
		t.Errorf("expected blocked requests to change nothing, got lists %+v", lists)
	}
}

func TestCreateMatch(t *testing.T) {
	for _, path := range []string{"/matches", "/characters/elo"} {
		t.Run("POST "+path, func(t *testing.T) {
			api := newTestAPI(t)
			list := api.createList("invincible")
			mark := api.createCharacter(list.ID, "Mark Grayson")
			nolan := api.createCharacter(list.ID, "Nolan Grayson")

			rec := api.request(http.MethodPost, path, vote(mark.ID, nolan.ID))
			expectStatus(t, rec, http.StatusCreated)
			match := decode[database.Match](t, rec)
			if match.WinnerID != mark.ID || match.LoserID != nolan.ID || match.WinnerEloAfter != 1232 || match.LoserEloAfter != 1168 {
				t.Errorf("unexpected match %+v", match)
			}
		})
	}
}

func TestCreateMatchRejectsBadVotes(t *testing.T) {
	api := newTestAPI(t)
	list := api.createList("invincible")
	other := api.createList("the boys")
	mark := api.createCharacter(list.ID, "Mark Grayson")
	homelander := api.createCharacter(other.ID, "Homelander")

	cases := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{"broken json", `{"winner_id":`, http.StatusBadRequest},
		{"invalid id", map[string]string{"winner_id": "nope", "loser_id": mark.ID.String()}, http.StatusBadRequest},
		{"missing loser", map[string]any{"winner_id": mark.ID}, http.StatusBadRequest},
		{"same character", vote(mark.ID, mark.ID), http.StatusBadRequest},
		{"different lists", vote(mark.ID, homelander.ID), http.StatusBadRequest},
		{"unknown character", vote(mark.ID, uuid.New()), http.StatusNotFound},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			expectStatus(t, api.request(http.MethodPost, "/matches", tc.body), tc.wantStatus)
		})
	}
}

// builds a matchup url with each pair as its own exclude param
func matchupPath(listID uuid.UUID, excluded ...[2]uuid.UUID) string {
	query := url.Values{}
	for _, pair := range excluded {
		query.Add("exclude", pair[0].String()+","+pair[1].String())
	}
	return "/lists/" + listID.String() + "/matchup?" + query.Encode()
}

// asks for a matchup draws times and counts each unordered pair of names that comes back
func (api *testAPI) countMatchups(path string, draws int) map[[2]string]int {
	api.t.Helper()
	counts := map[[2]string]int{}
	for range draws {
		rec := api.request(http.MethodGet, path, nil)
		expectStatus(api.t, rec, http.StatusOK)
		pair := decode[[]database.Character](api.t, rec)
		if len(pair) != 2 || pair[0].ID == pair[1].ID {
			api.t.Fatalf("expected two different characters, got %+v", pair)
		}
		names := [2]string{pair[0].Name, pair[1].Name}
		if names[0] > names[1] {
			names[0], names[1] = names[1], names[0]
		}
		counts[names]++
	}
	return counts
}

func TestGetMatchup(t *testing.T) {
	api := newTestAPI(t)
	list := api.createList("invincible")
	other := api.createList("the boys")
	a := api.createCharacter(list.ID, "A")
	b := api.createCharacter(list.ID, "B")
	c := api.createCharacter(list.ID, "C")
	homelander := api.createCharacter(other.ID, "Homelander")

	t.Run("no exclude", func(t *testing.T) {
		counts := api.countMatchups(matchupPath(list.ID), 200)
		if len(counts) != 3 {
			t.Errorf("expected all 3 pairs to come up, got %v", counts)
		}
	})
	t.Run("excluded pair never comes back", func(t *testing.T) {
		counts := api.countMatchups(matchupPath(list.ID, [2]uuid.UUID{b.ID, a.ID}), 200)
		if counts[[2]string{"A", "B"}] != 0 {
			t.Errorf("expected A vs B to be skipped, got %v", counts)
		}
	})
	t.Run("several excluded pairs", func(t *testing.T) {
		counts := api.countMatchups(matchupPath(list.ID, [2]uuid.UUID{a.ID, b.ID}, [2]uuid.UUID{a.ID, c.ID}), 50)
		if counts[[2]string{"B", "C"}] != 50 {
			t.Errorf("expected only B vs C to be left, got %v", counts)
		}
	})
	t.Run("repeats when every pair is excluded", func(t *testing.T) {
		counts := api.countMatchups(matchupPath(list.ID, [2]uuid.UUID{a.ID, b.ID}, [2]uuid.UUID{a.ID, c.ID}, [2]uuid.UUID{b.ID, c.ID}), 20)
		if len(counts) == 0 {
			t.Error("expected a matchup anyway")
		}
	})
	t.Run("pairs from another list are ignored", func(t *testing.T) {
		api.countMatchups(matchupPath(list.ID, [2]uuid.UUID{a.ID, homelander.ID}), 10)
	})

	t.Run("malformed exclude", func(t *testing.T) {
		base := "/lists/" + list.ID.String() + "/matchup?exclude="
		tooMany := make([][2]uuid.UUID, maxExcludedPairs+1)
		for _, path := range []string{
			base + a.ID.String(),
			base + url.QueryEscape(a.ID.String()+","+b.ID.String()+","+c.ID.String()),
			base + url.QueryEscape(a.ID.String()+",nope"),
			matchupPath(list.ID, tooMany...),
		} {
			expectStatus(t, api.request(http.MethodGet, path, nil), http.StatusBadRequest)
		}
	})
}

func TestGetMatchupNeedsTwoCharacters(t *testing.T) {
	api := newTestAPI(t)
	empty := api.createList("empty")
	single := api.createList("single")
	api.createCharacter(single.ID, "Mark Grayson")

	for _, list := range []database.List{empty, single} {
		expectStatus(t, api.request(http.MethodGet, matchupPath(list.ID), nil), http.StatusConflict)
	}
}

func TestGetLeaderboard(t *testing.T) {
	api := newTestAPI(t)
	list := api.createList("invincible")
	mark := api.createCharacter(list.ID, "Mark Grayson")
	nolan := api.createCharacter(list.ID, "Nolan Grayson")
	api.createCharacter(list.ID, "Atom Eve")
	expectStatus(t, api.request(http.MethodPost, "/matches", vote(mark.ID, nolan.ID)), http.StatusCreated)

	rec := api.request(http.MethodGet, "/lists/"+list.ID.String()+"/leaderboard", nil)
	expectStatus(t, rec, http.StatusOK)
	leaderboard := decode[[]database.LeaderboardEntry](t, rec)
	got := []string{}
	for _, entry := range leaderboard {
		got = append(got, fmt.Sprintf("%d %s %d", entry.Rank, entry.Name, entry.Elo))
	}
	want := []string{"1 Mark Grayson 1232", "2 Atom Eve 1200", "3 Nolan Grayson 1168"}
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		t.Errorf("expected %v, got %v", want, got)
	}
}

func TestGetMatchesByListID(t *testing.T) {
	api := newTestAPI(t)
	list := api.createList("invincible")
	mark := api.createCharacter(list.ID, "Mark Grayson")
	nolan := api.createCharacter(list.ID, "Nolan Grayson")
	for range 3 {
		expectStatus(t, api.request(http.MethodPost, "/matches", vote(mark.ID, nolan.ID)), http.StatusCreated)
	}
	path := "/lists/" + list.ID.String() + "/matches"

	all := decode[[]database.Match](t, api.request(http.MethodGet, path, nil))
	if len(all) != 3 || all[0].WinnerEloBefore <= all[1].WinnerEloBefore {
		t.Errorf("expected all 3 matches newest first, got %+v", all)
	}
	if limited := decode[[]database.Match](t, api.request(http.MethodGet, path+"?limit=2", nil)); len(limited) != 2 || limited[0] != all[0] {
		t.Errorf("expected the 2 newest matches, got %+v", limited)
	}
	if capped := decode[[]database.Match](t, api.request(http.MethodGet, path+"?limit=100000", nil)); len(capped) != 3 {
		t.Errorf("expected a huge limit to be capped, not rejected, got %+v", capped)
	}
	for _, limit := range []string{"0", "-1", "abc"} {
		expectStatus(t, api.request(http.MethodGet, path+"?limit="+limit, nil), http.StatusBadRequest)
	}
}

func TestVotesAreRateLimitedPerIP(t *testing.T) {
	limiter, _ := newTestLimiter(rate.Every(time.Minute), 2)
	api := newTestAPIWith(t, "dev", limiter)
	list := api.createList("invincible")
	mark := api.createCharacter(list.ID, "Mark Grayson")
	nolan := api.createCharacter(list.ID, "Nolan Grayson")

	voteFrom := func(ip string, path string, forwardedFor string) *httptest.ResponseRecorder {
		req := newRequest(t, http.MethodPost, path, vote(mark.ID, nolan.ID))
		req.RemoteAddr = ip + ":1234"
		if forwardedFor != "" {
			req.Header.Set("X-Forwarded-For", forwardedFor)
		}
		return api.serve(req)
	}

	expectStatus(t, voteFrom("1.1.1.1", "/matches", ""), http.StatusCreated)
	expectStatus(t, voteFrom("1.1.1.1", "/matches", ""), http.StatusCreated)

	rec := voteFrom("1.1.1.1", "/matches", "")
	expectStatus(t, rec, http.StatusTooManyRequests)
	if retry := rec.Header().Get("Retry-After"); retry != "60" {
		t.Errorf("expected Retry-After of 60 seconds, got %q", retry)
	}
	// the old path shares the same budget
	expectStatus(t, voteFrom("1.1.1.1", "/characters/elo", ""), http.StatusTooManyRequests)
	// no proxies are trusted, so a made up X-Forwarded-For can't buy a fresh bucket
	expectStatus(t, voteFrom("1.1.1.1", "/matches", "9.9.9.9"), http.StatusTooManyRequests)
	expectStatus(t, voteFrom("2.2.2.2", "/matches", ""), http.StatusCreated)

	matches := decode[[]database.Match](t, api.request(http.MethodGet, "/lists/"+list.ID.String()+"/matches", nil))
	if len(matches) != 3 {
		t.Errorf("expected only the 3 allowed votes to count, got %d", len(matches))
	}
}
