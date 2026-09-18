package main

import (
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func devModeMiddleware(platform string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if platform != "dev" {
			log.Println("attempted use of unauthorized endpoint")
			c.IndentedJSON(http.StatusForbidden, returnErrJSON("unauthorized endpoint"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func rateLimitMiddleware(limiter *ipRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// only as trustworthy as the router's trusted proxies setting: trusting
		// X-Forwarded-For from anyone would let a client pick a fresh ip per request
		allowed, wait := limiter.take(c.ClientIP())
		if !allowed {
			c.Header("Retry-After", strconv.Itoa(int(math.Ceil(wait.Seconds()))))
			c.IndentedJSON(http.StatusTooManyRequests, returnErrJSON("too many votes, slow down"))
			c.Abort()
			return
		}
		c.Next()
	}
}
