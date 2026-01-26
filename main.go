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

	// Seen media endpoints (State Management)
	r.POST("/seen", h.PostSeen)
	r.GET("/seen", h.GetSeen)
	r.DELETE("/seen", h.DeleteSeen)

	// Recommendation endpoints (The Core)
	r.POST("/recommend", h.PostRecommend)      // Full vibe search with reranking
	r.GET("/vibe", h.GetRecommendSimple)       // Simple GET endpoint (backwards compatible)
	r.GET("/similar/:media_id", h.GetSimilar)  // Find similar to specific media
	r.GET("/hidden-gems", h.GetHiddenGems)     // Surface quality hidden gems

	// Media management endpoints
	r.POST("/media", h.PostMedia)              // Ingest new media with vibe generation
	r.GET("/media/:id", h.GetMedia)            // Get media details
	r.POST("/media/:id/refresh", h.PostRefreshVibe) // Regenerate vibe profile

	// Admin endpoints
	r.GET("/stats", h.GetStats)                // System statistics
	r.POST("/admin/scrape", h.PostScrapeNow)   // Trigger immediate scrape

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
