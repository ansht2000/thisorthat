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
	cfg.db.DeleteLists()
	c.IndentedJSON(http.StatusOK, returnMessageJSON("successfully reset db"))
}