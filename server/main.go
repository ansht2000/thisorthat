package main

import (
	"log"
	"os"

	"github.com/ansht2000/thisorthat/internal/database"
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
	router := gin.Default()
	router.POST("/lists", apiCfg.handlerCreateList)
	router.POST("/reset", apiCfg.handlerReset)
	router.POST("/characters", apiCfg.handlerCreateCharacter)

	router.GET("/lists", apiCfg.handlerGetLists)
	router.GET("/lists/:id", apiCfg.handlerGetListByID)
	router.GET("/characters/:id", apiCfg.handlerGetCharacterByID)
	router.GET("/characters/list/:id", apiCfg.handlerGetCharactersByListID)

	router.Run("localhost:" + port)
}