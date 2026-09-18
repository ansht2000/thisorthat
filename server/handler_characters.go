package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateCharacter(c *gin.Context) {
	var createCharacterParams database.CreateCharacterParams
	if err := c.ShouldBindJSON(&createCharacterParams); err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("invalid request body: "+err.Error()))
		return
	}
	if strings.TrimSpace(createCharacterParams.Name) == "" || createCharacterParams.ListID == uuid.Nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("name and list_id are required"))
		return
	}
	createdCharacter, err := cfg.db.CreateCharacter(c.Request.Context(), createCharacterParams)
	if err != nil {
		if errors.Is(err, database.ErrListNotFound) {
			c.IndentedJSON(http.StatusNotFound, returnErrJSON("specified list not found"))
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusCreated, createdCharacter)
}

func (cfg *apiConfig) handlerGetCharacterByID(c *gin.Context) {
	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("invalid id provided"))
		return
	}
	character, err := cfg.db.GetCharacterByID(c.Request.Context(), uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, returnErrJSON("specified character not found"))
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, character)
}

func (cfg *apiConfig) handlerGetCharactersByListID(c *gin.Context) {
	list, ok := cfg.lookupList(c)
	if !ok {
		return
	}
	characters, err := cfg.db.GetCharactersByListID(c.Request.Context(), list.ID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, characters)
}

func (cfg *apiConfig) handlerGetLeaderboard(c *gin.Context) {
	list, ok := cfg.lookupList(c)
	if !ok {
		return
	}
	leaderboard, err := cfg.db.GetLeaderboard(c.Request.Context(), list.ID)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, leaderboard)
}
