package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/Jack4Code/bedrock"
	"github.com/Jack4Code/bedrock/config"
	"github.com/Jack4Code/cardforge/internal/claude"
	"github.com/Jack4Code/cardforge/internal/handlers"
	"github.com/Jack4Code/cardforge/internal/storage"
)

//go:embed web/dist/*
var frontendFS embed.FS

// CardForgeApp implements the bedrock.App interface
type CardForgeApp struct {
	handler *handlers.Handler
	staticFS http.Handler
}

func (a *CardForgeApp) OnStart(ctx context.Context) error {
	log.Println("CardForge starting...")
	return nil
}

func (a *CardForgeApp) OnStop(ctx context.Context) error {
	log.Println("CardForge stopping...")
	return nil
}

func (a *CardForgeApp) Routes() []bedrock.Route {
	return []bedrock.Route{
		// API routes
		{
			Method: "POST",
			Path:   "/api/generate",
			Handler: adaptHandler(a.handler.GenerateCards),
		},
		{
			Method: "GET",
			Path:   "/api/cards",
			Handler: adaptHandler(a.handler.ListCards),
		},
		{
			Method: "POST",
			Path:   "/api/cards",
			Handler: adaptHandler(a.handler.SaveCards),
		},
		{
			Method: "PUT",
			Path:   "/api/cards/:id",
			Handler: adaptHandler(a.handler.UpdateCard),
		},
		{
			Method: "DELETE",
			Path:   "/api/cards/:id",
			Handler: adaptHandler(a.handler.DeleteCard),
		},
		{
			Method: "GET",
			Path:   "/api/sessions",
			Handler: adaptHandler(a.handler.ListSessions),
		},
		{
			Method: "GET",
			Path:   "/api/sessions/:id",
			Handler: adaptHandler(a.handler.GetSession),
		},
		{
			Method: "GET",
			Path:   "/api/export",
			Handler: adaptHandler(a.handler.ExportCards),
		},
		// Health check
		{
			Method: "GET",
			Path:   "/health",
			Handler: func(ctx context.Context, r *http.Request) bedrock.Response {
				return bedrock.JSON(http.StatusOK, map[string]string{"status": "ok"})
			},
		},
		// Static files (catch-all)
		{
			Method: "GET",
			Path:   "/*",
			Handler: func(ctx context.Context, r *http.Request) bedrock.Response {
				return &staticResponse{handler: a.staticFS, request: r}
			},
		},
	}
}

// adaptHandler converts http.HandlerFunc to bedrock.Handler
func adaptHandler(h http.HandlerFunc) bedrock.Handler {
	return func(ctx context.Context, r *http.Request) bedrock.Response {
		return &handlerResponse{handler: h, request: r}
	}
}

// handlerResponse wraps an http.HandlerFunc as a bedrock.Response
type handlerResponse struct {
	handler http.HandlerFunc
	request *http.Request
}

func (hr *handlerResponse) Write(ctx context.Context, w http.ResponseWriter) error {
	hr.handler(w, hr.request)
	return nil
}

// staticResponse wraps a static file handler
type staticResponse struct {
	handler http.Handler
	request *http.Request
}

func (sr *staticResponse) Write(ctx context.Context, w http.ResponseWriter) error {
	sr.handler.ServeHTTP(w, sr.request)
	return nil
}

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

	// Serve static frontend (embedded assets)
	staticFS, err := fs.Sub(frontendFS, "web/dist")
	if err != nil {
		log.Fatalf("Failed to get frontend filesystem: %v", err)
	}

	// Create Bedrock app
	app := &CardForgeApp{
		handler:  h,
		staticFS: http.FileServer(http.FS(staticFS)),
	}

	// Create config
	cfg := config.BaseConfig{
		Port: port,
	}

	// Start server with default CORS
	log.Printf("Starting CardForge server on :%s", port)
	if err := bedrock.Run(app, cfg); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
