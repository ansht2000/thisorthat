package main

import (
	"log"
	"os"
	"strings"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/joho/godotenv"
)

const (
	// a person reading two names and clicking one won't sustain more than about
	// one vote a second, and the burst leaves room for a few quick clicks in a row
	votesPerSecond = 1
	voteBurst      = 10
)

type apiConfig struct {
	db          database.Client
	platform    string
	voteLimiter *ipRateLimiter
	// proxies allowed to report the client's real ip through X-Forwarded-For
	trustedProxies []string
}

func main() {
	godotenv.Load(".env")

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalln("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatalln("please set PLATFORM env variable")
	}
	dbQueries, err := database.NewClient(dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	apiCfg := apiConfig{
		db:          dbQueries,
		platform:    platform,
		voteLimiter: newIPRateLimiter(votesPerSecond, voteBurst),
	}
	// comma separated ips or cidrs, only needed when running behind a reverse proxy or load balancer
	if proxies := os.Getenv("TRUSTED_PROXIES"); proxies != "" {
		apiCfg.trustedProxies = strings.Split(proxies, ",")
	}

	port := os.Getenv("PORT")
	if port == "" {
		// if not specified default port is 8080
		port = "8080"
	}
	router, err := newRouter(&apiCfg)
	if err != nil {
		log.Fatalf("failed to set up router: %v", err)
	}

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
