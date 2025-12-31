package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/Jack4Code/bedrock"
	"github.com/Jack4Code/cardforge/internal/claude"
	"github.com/Jack4Code/cardforge/internal/handlers"
	"github.com/Jack4Code/cardforge/internal/storage"
)

//go:embed web/dist/*
var frontendFS embed.FS

func main() {
	// Load configuration from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable is required")
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./cardforge.db"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize database
	db, err := storage.NewSQLite(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	log.Printf("Database initialized at: %s", dbPath)

	// Initialize Claude client
	claudeClient := claude.NewClient(apiKey)
	log.Println("Claude API client initialized")

	// Initialize handlers
	h := handlers.NewHandler(db, claudeClient)

	// Initialize Bedrock app
	app := bedrock.New()

	// Add middleware
	app.Use(bedrock.CORS())
	app.Use(bedrock.Logger())
	app.Use(bedrock.Recovery())

	// API routes
	app.POST("/api/generate", h.GenerateCards)
	app.GET("/api/cards", h.ListCards)
	app.POST("/api/cards", h.SaveCards)
	app.PUT("/api/cards/:id", h.UpdateCard)
	app.DELETE("/api/cards/:id", h.DeleteCard)
	app.GET("/api/sessions", h.ListSessions)
	app.GET("/api/sessions/:id", h.GetSession)
	app.GET("/api/export", h.ExportCards)

	// Health check endpoint
	app.GET("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Serve static frontend (embedded assets)
	staticFS, err := fs.Sub(frontendFS, "web/dist")
	if err != nil {
		log.Fatalf("Failed to get frontend filesystem: %v", err)
	}
	app.Static("/", http.FileServer(http.FS(staticFS)))

	// Start server
	log.Printf("Starting CardForge server on :%s", port)
	if err := app.Run(":" + port); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
