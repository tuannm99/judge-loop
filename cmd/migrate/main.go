// migrate runs goose migrations against the judge-loop database.
// Run this manually before starting any application service.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/tuannm99/judge-loop/internal/infrastructure/postgres"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := postgres.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := postgres.RunMigrations(ctx, db.SQL); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	log.Println("migrations applied successfully")
}
