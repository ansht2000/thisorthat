package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func devModeMiddleware() gin.HandlerFunc {
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatalln("please set PLATFORM env variable")
	}
	return func(c *gin.Context) {
		if platform != "dev" {
			log.Println("attempted use of unauthorized endpoint")
			c.IndentedJSON(http.StatusUnauthorized, returnErrJSON("unauthorized endpoint"))
			c.Abort()
			return
		}
		c.Next()
	}
}