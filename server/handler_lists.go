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

func (cfg *apiConfig) handlerCreateList(c *gin.Context) {
	var createListParams database.CreateListParams
	if err := c.ShouldBindJSON(&createListParams); err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("invalid request body: "+err.Error()))
		return
	}
	if strings.TrimSpace(createListParams.Name) == "" {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("name is required"))
		return
	}
	createdList, err := cfg.db.CreateList(c.Request.Context(), createListParams)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusCreated, createdList)
}

func (cfg *apiConfig) handlerGetLists(c *gin.Context) {
	lists, err := cfg.db.GetLists(c.Request.Context())
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, lists)
}

func (cfg *apiConfig) handlerGetList(c *gin.Context) {
	list, ok := cfg.lookupList(c)
	if !ok {
		return
	}
	c.IndentedJSON(http.StatusOK, list)
}

// loads the list named by the :id path param for the /lists/:id/... handlers.
// if that fails it writes the error response itself and returns false
func (cfg *apiConfig) lookupList(c *gin.Context) (database.List, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, returnErrJSON("invalid id provided"))
		return database.List{}, false
	}
	list, err := cfg.db.GetListByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.IndentedJSON(http.StatusNotFound, returnErrJSON("specified list not found"))
			return database.List{}, false
		}
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return database.List{}, false
	}
	return list, true
}
