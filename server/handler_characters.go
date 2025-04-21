package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerCreateCharacter(c *gin.Context) {
	var createCharacterParams database.CreateCharacterParams
	if err := c.BindJSON(&createCharacterParams); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	createdCharacter, err := cfg.db.CreateCharacter(c, createCharacterParams)
	if err != nil {
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
	character, err := cfg.db.GetCharacterByID(c, uuid)
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
	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("invalid id provided"))
		return
	}
	characters, err := cfg.db.GetCharactersByListID(c, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, []database.Character{})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, characters)
}

func (cfg *apiConfig) handlerGetTwoRandomCharactersByListID(c *gin.Context) {
	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("invalid id provided"))
		return
	}
	characters, err := cfg.db.GetTwoRandomCharactersFromListID(c, uuid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, []database.Character{})
			return
		}
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, characters)
}