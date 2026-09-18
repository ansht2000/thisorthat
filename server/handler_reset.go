package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func (cfg *apiConfig) handlerReset(c *gin.Context) {
	if os.Getenv("PLATFORM") != "dev" {
		c.IndentedJSON(http.StatusForbidden, returnErrJSON("unauthorized action"))
		return
	}
	if err := cfg.db.DeleteLists(c); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, returnErrJSON(err.Error()))
		return
	}
	c.IndentedJSON(http.StatusOK, returnMessageJSON("successfully reset db"))
}
