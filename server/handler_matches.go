package main

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/ansht2000/thisorthat/internal/elo"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	defaultMatchHistoryLimit = 20
	maxMatchHistoryLimit     = 100
	// caps the work one matchup request can ask for
	maxExcludedPairs = 20
)

func (cfg *apiConfig) handlerCreateMatch(c *gin.Context) {
	var createMatchParams createMatchParams
	if err := c.ShouldBindJSON(&createMatchParams); err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("invalid request body: "+err.Error()))
		return
	}

	winnerID := createMatchParams.WinnerID
	loserID := createMatchParams.LoserID
	if winnerID == uuid.Nil || loserID == uuid.Nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("winner_id and loser_id are required"))
		return
	}

	match, err := cfg.db.RecordMatch(c.Request.Context(), winnerID, loserID)
	if err != nil {
		switch {
		case errors.Is(err, database.ErrSameCharacter), errors.Is(err, database.ErrDifferentLists):
			c.IndentedJSON(http.StatusBadRequest, returnErrJSON(err.Error()))
		case errors.Is(err, sql.ErrNoRows):
			c.IndentedJSON(http.StatusNotFound, returnErrJSON("specified character not found"))
		default:
			c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		}
		return
	}
	c.IndentedJSON(http.StatusCreated, match)
}

// GET /lists/:id/matchup?exclude=<character id>,<character id>
//
// picks the next two characters to vote on, returned as a two item array.
// each exclude is a pair the voter saw recently, written as two character ids
// joined by a comma, and it can be repeated to pass several pairs. an excluded
// pair only comes back if the list has no other pairs left to show
func (cfg *apiConfig) handlerGetMatchup(c *gin.Context) {
	list, ok := cfg.lookupList(c)
	if !ok {
		return
	}
	excludedPairs, err := parseExcludedPairs(c.QueryArray("exclude"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON(err.Error()))
		return
	}
	characters, err := cfg.db.GetCharactersByListID(c.Request.Context(), list.ID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}

	ratings := make([]elo.Rating, len(characters))
	indexOf := make(map[uuid.UUID]int, len(characters))
	for i, character := range characters {
		ratings[i] = elo.Rating{Elo: character.Elo, GamesPlayed: character.GamesPlayed}
		indexOf[character.ID] = i
	}
	avoid := map[[2]int]bool{}
	for _, pair := range excludedPairs {
		i, iInList := indexOf[pair[0]]
		j, jInList := indexOf[pair[1]]
		// a pair from another list couldn't come up here anyway
		if iInList && jInList {
			avoid[[2]int{min(i, j), max(i, j)}] = true
		}
	}

	i, j, err := elo.PickPair(ratings, func(i, j int) bool { return avoid[[2]int{i, j}] }, rand.Float64)
	if err != nil {
		if errors.Is(err, elo.ErrNotEnoughContenders) {
			c.IndentedJSON(http.StatusConflict, returnErrJSON("list needs at least two characters for a matchup"))
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, []database.Character{characters[i], characters[j]})
}

func parseExcludedPairs(values []string) ([][2]uuid.UUID, error) {
	if len(values) > maxExcludedPairs {
		return nil, fmt.Errorf("at most %d exclude pairs are allowed", maxExcludedPairs)
	}
	pairs := make([][2]uuid.UUID, 0, len(values))
	for _, value := range values {
		ids := strings.Split(value, ",")
		if len(ids) != 2 {
			return nil, fmt.Errorf("exclude must be two character ids separated by a comma, got %q", value)
		}
		var pair [2]uuid.UUID
		for k, raw := range ids {
			id, err := uuid.Parse(strings.TrimSpace(raw))
			if err != nil {
				return nil, fmt.Errorf("invalid character id %q in exclude", raw)
			}
			pair[k] = id
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

// GET /lists/:id/matches?limit=<n>
//
// the list's most recent matches, newest first
func (cfg *apiConfig) handlerGetMatchesByListID(c *gin.Context) {
	list, ok := cfg.lookupList(c)
	if !ok {
		return
	}
	limit := defaultMatchHistoryLimit
	if raw, given := c.GetQuery("limit"); given {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			c.IndentedJSON(http.StatusBadRequest, returnErrJSON("limit must be a whole number above 0"))
			return
		}
		limit = min(parsed, maxMatchHistoryLimit)
	}
	matches, err := cfg.db.GetMatchesByListID(c.Request.Context(), list.ID, limit)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, matches)
}
