package main

import (
	"net/http"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/gin-gonic/gin"
)

func (cfg *apiConfig) handlerCreateList(c *gin.Context) {
	var createListParams database.CreateListParams
	if err := c.BindJSON(&createListParams); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	createdList, err := cfg.db.CreateList(c, createListParams)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusCreated, createdList)
}

func (cfg *apiConfig) handlerGetLists(c *gin.Context) {
	lists, err := cfg.db.GetLists(c)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	if len(lists) == 0 {
		c.IndentedJSON(http.StatusNotFound, returnErrJSON("no lists found"))
		return
	}
	c.IndentedJSON(http.StatusOK, lists)
}
