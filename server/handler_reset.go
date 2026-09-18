package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (cfg *apiConfig) handlerReset(c *gin.Context) {
	if cfg.platform != "dev" {
		c.IndentedJSON(http.StatusForbidden, returnErrJSON("unauthorized action"))
		return
	}
	if err := cfg.db.DeleteLists(c.Request.Context()); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, returnMessageJSON("successfully reset db"))
}
