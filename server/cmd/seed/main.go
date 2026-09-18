// seed fills the database with franchises and characters, and can simulate votes so the
// leaderboard and history have something to show. run it from the server directory:
//
//	go run ./cmd/seed                  add the built in franchises
//	go run ./cmd/seed -votes 300       and cast 300 simulated votes in each of them
//	go run ./cmd/seed -file mine.json  add franchises from your own file instead
//
// it's safe to run again: franchises and characters that already exist are left alone.
// the database comes from DB_URL, read from .env like the server does
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strings"
	"time"

	"github.com/ansht2000/thisorthat/internal/database"
	"github.com/ansht2000/thisorthat/internal/seed"
	"github.com/joho/godotenv"
)

func main() {
	file := flag.String("file", "", "json file of franchises to add, instead of the built in ones")
	votes := flag.Int("votes", 0, "simulated votes to cast in each franchise being seeded")
	flag.Parse()

	godotenv.Load(".env")
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatalln("DB_URL must be set")
	}

	franchises, err := loadFranchises(*file)
	if err != nil {
		log.Fatalln(err)
	}

	db, err := database.NewClient(dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}
	defer db.Close()
	ctx := context.Background()

	result, err := seed.Apply(ctx, &db, franchises)
	if err != nil {
		log.Fatalf("seeding failed, nothing was added: %v", err)
	}
	fmt.Printf("added %d franchises and %d characters to %s\n", result.ListsCreated, result.CharactersCreated, dbURL)

	if *votes <= 0 {
		return
	}
	lists, err := db.GetLists(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	rng := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 0))
	for _, franchise := range franchises {
		for _, list := range lists {
			if !strings.EqualFold(strings.TrimSpace(list.Name), strings.TrimSpace(franchise.Name)) {
				continue
			}
			order := make([]string, len(franchise.Characters))
			for i, character := range franchise.Characters {
				order[i] = character.Name
			}
			if err := seed.SimulateVotes(ctx, &db, list.ID, order, *votes, rng); err != nil {
				log.Fatalf("simulating votes for %q failed: %v", franchise.Name, err)
			}
			fmt.Printf("cast %d simulated votes in %s\n", *votes, franchise.Name)
		}
	}
}

func loadFranchises(path string) ([]seed.Franchise, error) {
	if path == "" {
		return seed.Default()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read seed file: %w", err)
	}
	return seed.Parse(data)
}
