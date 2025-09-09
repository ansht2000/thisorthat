package main

import (
	"log"
	"os"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type apiConfig struct {
	db database.Client
}

func main() {
	godotenv.Load(".env")

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalln("DB_URL must be set")
	}
	dbQueries, err := database.NewClient(dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	apiCfg := apiConfig{
		db: dbQueries,
	}

	port := os.Getenv("PORT")
	if port == "" {
		// if not specified default port is 8080
		port = "8080"
	}
	// using default router with logging and recovery middleware attached
	router := gin.Default()
	router.Use(cors.Default())

	// currently making a group for the post endpoints so they cant be used in prod
	// will probably change later when a strategy to properly accept user created
	// lists and characters is implemented
	// TODO: figure out how to properly get well formatted lists from users
	devOnly := router.Group("/")
	devOnly.Use(devModeMiddleware())
	{
		devOnly.POST("/lists", apiCfg.handlerCreateList)
		devOnly.POST("/reset", apiCfg.handlerReset)
		devOnly.POST("/characters", apiCfg.handlerCreateCharacter)
	}

	router.GET("/healthz", apiCfg.handlerReadiness)
	router.GET("/lists", apiCfg.handlerGetLists)
	router.GET("/lists/random")
	router.GET("/lists/:id/characters", apiCfg.handlerGetCharactersByListID)
	router.GET("/characters/:id", apiCfg.handlerGetCharacterByID)

	// TODO: add authentication so only the frontend can call this
	router.POST("/characters/elo", apiCfg.handlerUpdateWinnerAndLoserELOs)

	router.Run(":" + port)
}
