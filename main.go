package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"w2w/internal/database"
	"w2w/internal/embeddings"
	"w2w/internal/handlers"
	"w2w/internal/llm"
	"w2w/internal/services"
)

// Config holds application configuration
type Config struct {
	Port          string
	DatabasePath  string
	OpenAIAPIKey  string
	EnableScraper bool
	ScrapeInterval time.Duration
}

func loadConfig() *Config {
	cfg := &Config{
		Port:           getEnv("PORT", "8080"),
		DatabasePath:   getEnv("DATABASE_PATH", "./vibe.db"),
		OpenAIAPIKey:   os.Getenv("OPENAI_API_KEY"),
		EnableScraper:  getEnv("ENABLE_SCRAPER", "false") == "true",
		ScrapeInterval: 1 * time.Hour,
	}

	if interval := os.Getenv("SCRAPE_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			cfg.ScrapeInterval = d
		}
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := loadConfig()

	// Validate required configuration
	if cfg.OpenAIAPIKey == "" {
		log.Println("WARNING: OPENAI_API_KEY not set. Using placeholder embedding provider.")
		log.Println("Set OPENAI_API_KEY environment variable for full functionality.")
	}

	// Initialize database
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize embedding provider
	var embedProvider embeddings.Provider
	if cfg.OpenAIAPIKey != "" {
		embedProvider = embeddings.NewOpenAIProvider(cfg.OpenAIAPIKey)
	} else {
		// Use placeholder provider for development
		embedProvider = &placeholderEmbedder{}
	}

	// Initialize LLM client
	var llmClient *llm.Client
	if cfg.OpenAIAPIKey != "" {
		llmClient = llm.NewClient(cfg.OpenAIAPIKey)
	}

	// Initialize vibe search service
	vibeSearch, err := services.NewVibeSearchService(db, embedProvider, llmClient)
	if err != nil {
		log.Fatalf("Failed to initialize vibe search: %v", err)
	}

	// Initialize Reddit scraper
	scraper := services.NewRedditScraper(db, llmClient)

	// Start background scraper if enabled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.EnableScraper {
		log.Printf("Starting Reddit scraper with interval: %v", cfg.ScrapeInterval)
		scraper.Start(ctx, cfg.ScrapeInterval)
	}

	// Initialize handlers
	h := handlers.NewHandler(db, vibeSearch, scraper)

	// Setup router
	r := gin.Default()

	// Health check
	r.GET("/health", h.GetHealth)

	// API routes with /api prefix (for production where frontend is served from same origin)
	api := r.Group("/api")
	{
		// Seen media endpoints (State Management)
		api.POST("/seen", h.PostSeen)
		api.GET("/seen", h.GetSeen)
		api.DELETE("/seen", h.DeleteSeen)

		// Recommendation endpoints (The Core)
		api.POST("/recommend", h.PostRecommend)
		api.GET("/vibe", h.GetRecommendSimple)
		api.GET("/similar/:media_id", h.GetSimilar)
		api.GET("/hidden-gems", h.GetHiddenGems)

		// Media management endpoints
		api.POST("/media", h.PostMedia)
		api.GET("/media/:id", h.GetMedia)
		api.POST("/media/:id/refresh", h.PostRefreshVibe)

		// Admin endpoints
		api.GET("/stats", h.GetStats)
		api.POST("/admin/scrape", h.PostScrapeNow)
	}

	// Legacy routes without /api prefix (for backwards compatibility)
	r.POST("/seen", h.PostSeen)
	r.GET("/seen", h.GetSeen)
	r.DELETE("/seen", h.DeleteSeen)
	r.POST("/recommend", h.PostRecommend)
	r.GET("/vibe", h.GetRecommendSimple)
	r.GET("/similar/:media_id", h.GetSimilar)
	r.GET("/hidden-gems", h.GetHiddenGems)
	r.POST("/media", h.PostMedia)
	r.GET("/media/:id", h.GetMedia)
	r.POST("/media/:id/refresh", h.PostRefreshVibe)
	r.GET("/stats", h.GetStats)
	r.POST("/admin/scrape", h.PostScrapeNow)

	// Serve static files from ./static directory (frontend build)
	r.Static("/assets", "./static/assets")
	r.StaticFile("/favicon.ico", "./static/favicon.ico")

	// SPA fallback: serve index.html for any unmatched routes
	r.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down...")
		cancel()
		scraper.Stop()
		os.Exit(0)
	}()

	// Start server
	fmt.Printf("\n")
	fmt.Println("========================================")
	fmt.Println("   Vibe-First Recommendation Engine    ")
	fmt.Println("========================================")
	fmt.Printf("  Server:    http://localhost:%s\n", cfg.Port)
	fmt.Printf("  Database:  %s\n", cfg.DatabasePath)
	fmt.Printf("  Scraper:   %v\n", cfg.EnableScraper)
	fmt.Printf("  OpenAI:    %v\n", cfg.OpenAIAPIKey != "")
	fmt.Println("========================================")
	fmt.Println("\nEndpoints:")
	fmt.Println("  POST /seen           - Mark media as watched")
	fmt.Println("  GET  /seen           - Get your watch history")
	fmt.Println("  POST /recommend      - Get vibe-based recommendations")
	fmt.Println("  GET  /vibe?q=...     - Quick vibe search")
	fmt.Println("  GET  /hidden-gems    - Discover quality hidden gems")
	fmt.Println("  POST /media          - Add new media to database")
	fmt.Println("  GET  /stats          - System statistics")
	fmt.Println("")

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// placeholderEmbedder provides dummy embeddings for development without API key
type placeholderEmbedder struct{}

func (p *placeholderEmbedder) Embed(text string) ([]float32, error) {
	// Generate a deterministic pseudo-embedding based on text hash
	// This is NOT suitable for production - just for testing the pipeline
	embedding := make([]float32, 1536) // Same dimension as text-embedding-3-small

	// Simple hash-based embedding (not semantically meaningful)
	for i, r := range text {
		idx := i % 1536
		embedding[idx] += float32(r) / 1000.0
	}

	// Normalize
	var sum float32
	for _, v := range embedding {
		sum += v * v
	}
	if sum > 0 {
		norm := float32(1.0) / float32(sum)
		for i := range embedding {
			embedding[i] *= norm
		}
	}

	return embedding, nil
}

func (p *placeholderEmbedder) ModelName() string {
	return "placeholder-dev"
}
