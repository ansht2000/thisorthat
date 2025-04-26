package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (cfg *apiConfig) handlerReadiness(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, returnMessageJSON("api is responsive"))
}