package main

import (
	"fmt"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// handlers give the db c.Request.Context(), never c itself. gin hands the same *gin.Context
// to a later request once a handler returns, while database/sql can still be watching the
// context it was given from a background goroutine, which is a data race. c.Done() is also
// always nil unless ContextWithFallback is on, so a client hanging up would never cancel a query
func newRouter(cfg *apiConfig) (*gin.Engine, error) {
	// using default router with logging and recovery middleware attached
	router := gin.Default()
	// nil means trust no proxy, so ClientIP is the address that actually connected
	if err := router.SetTrustedProxies(cfg.trustedProxies); err != nil {
		return nil, fmt.Errorf("invalid trusted proxies: %w", err)
	}

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	// so the client can read how long to wait after a 429
	corsConfig.ExposeHeaders = []string{"Retry-After"}
	router.Use(cors.New(corsConfig))

	// currently making a group for the post endpoints so they cant be used in prod
	// will probably change later when a strategy to properly accept user created
	// lists and characters is implemented
	// TODO: figure out how to properly get well formatted lists from users
	devOnly := router.Group("/")
	devOnly.Use(devModeMiddleware(cfg.platform))
	{
		devOnly.POST("/lists", cfg.handlerCreateList)
		devOnly.POST("/reset", cfg.handlerReset)
		devOnly.POST("/characters", cfg.handlerCreateCharacter)
	}

	router.GET("/healthz", cfg.handlerReadiness)
	router.GET("/lists", cfg.handlerGetLists)
	router.GET("/lists/:id", cfg.handlerGetList)
	router.GET("/lists/:id/characters", cfg.handlerGetCharactersByListID)
	router.GET("/lists/:id/leaderboard", cfg.handlerGetLeaderboard)
	router.GET("/lists/:id/matchup", cfg.handlerGetMatchup)
	router.GET("/lists/:id/matches", cfg.handlerGetMatchesByListID)
	router.GET("/characters/:id", cfg.handlerGetCharacterByID)

	// there's no way to make these frontend only: any secret shipped in the
	// client bundle is public, so a per ip limit is what keeps one person
	// from flooding the rankings
	votes := router.Group("/")
	votes.Use(rateLimitMiddleware(cfg.voteLimiter))
	{
		votes.POST("/matches", cfg.handlerCreateMatch)
		// old path, kept until the client has moved over to POST /matches
		votes.POST("/characters/elo", cfg.handlerCreateMatch)
	}

	return router, nil
}
