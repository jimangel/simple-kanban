package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kanban-simple/internal/api"
	"github.com/kanban-simple/internal/database"
	"github.com/kanban-simple/internal/repository"
)

func main() {
	// Parse command line flags
	var (
		dbPath         = flag.String("db", getEnv("DATABASE_PATH", "./data/kanban.db"), "Database path")
		migrationsPath = flag.String("migrations", getEnv("MIGRATIONS_PATH", "./migrations"), "Migrations path")
		port          = flag.String("port", getEnv("PORT", "8080"), "Server port")
		mode          = flag.String("mode", getEnv("GIN_MODE", "debug"), "Gin mode (debug/release)")
	)
	flag.Parse()

	// Set Gin mode
	gin.SetMode(*mode)

	// Initialize database
	db, err := database.NewConnection(*dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(*migrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	repos := &api.Repositories{
		Board: repository.NewBoardRepository(db.DB),
		List:  repository.NewListRepository(db.DB),
		Card:  repository.NewCardRepository(db.DB),
		Label: repository.NewLabelRepository(db.DB),
	}

	// Initialize router
	router := api.NewRouter(repos)

	// Start server
	log.Printf("Starting server on port %s", *port)
	if err := router.Run(":" + *port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}