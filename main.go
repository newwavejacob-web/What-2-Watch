package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"w2w/internal/database"
	"w2w/internal/embeddings"
	"w2w/internal/handlers"
	"w2w/internal/llm"
	"w2w/internal/middleware"
	"w2w/internal/services"
)

// Config holds application configuration
type Config struct {
	Port               string
	DatabasePath       string
	OpenAIAPIKey       string
	EnableScraper      bool
	ScrapeInterval     time.Duration
	SessionSecret      string
	AdminSecret        string
	RateLimitPerMinute int
	CORSAllowedOrigins []string
}

func loadConfig() *Config {
	cfg := &Config{
		Port:               getEnv("PORT", "8080"),
		DatabasePath:       getEnv("DATABASE_PATH", "./vibe.db"),
		OpenAIAPIKey:       os.Getenv("OPENAI_API_KEY"),
		EnableScraper:      getEnv("ENABLE_SCRAPER", "false") == "true",
		ScrapeInterval:     1 * time.Hour,
		SessionSecret:      os.Getenv("SESSION_SECRET"),
		AdminSecret:        os.Getenv("ADMIN_SECRET"),
		RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 20),
		CORSAllowedOrigins: splitAndTrim(os.Getenv("CORS_ALLOWED_ORIGINS")),
	}

	if interval := os.Getenv("SCRAPE_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			cfg.ScrapeInterval = d
		}
	}

	// A session secret is required to sign cookies. If one isn't provided,
	// mint an ephemeral one so the app still runs — but sessions won't survive
	// a restart, so this should only happen in development.
	if cfg.SessionSecret == "" {
		cfg.SessionSecret = randomSecret()
		log.Println("WARNING: SESSION_SECRET not set. Using an ephemeral secret.")
		log.Println("Sessions will be invalidated on restart. Set SESSION_SECRET in production.")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvInt reads an integer env var, falling back on missing/invalid values.
func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallback
}

// splitAndTrim turns a comma-separated env value into a clean slice.
func splitAndTrim(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// randomSecret generates a 256-bit hex secret for ephemeral session signing.
func randomSecret() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "insecure-fallback-secret-change-me"
	}
	return hex.EncodeToString(b)
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

	// Setup router (release mode disables debug logging / route dumps)
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.TrustedPlatform = gin.PlatformCloudflare
	r.Use(gin.Logger(), gin.Recovery())

	// Global middleware: security headers, CORS, and anonymous sessions.
	r.Use(middleware.SecurityHeaders())
	if len(cfg.CORSAllowedOrigins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.CORSAllowedOrigins,
			AllowMethods:     []string{"GET", "POST", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Content-Type", "X-Admin-Secret"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}
	r.Use(middleware.Session(cfg.SessionSecret))

	// Rate limiter for OpenAI-billed endpoints (keyed by session, falls back to IP).
	rateLimit := middleware.RateLimit(cfg.RateLimitPerMinute)
	// Admin guard requiring the X-Admin-Secret header.
	adminAuth := middleware.AdminAuth(cfg.AdminSecret)

	// Health check
	r.GET("/health", h.GetHealth)

	// registerRoutes mounts the full API on a router group. Called for both the
	// /api-prefixed routes and the legacy un-prefixed ones.
	registerRoutes := func(rg *gin.RouterGroup) {
		// Seen media endpoints (State Management)
		rg.POST("/seen", h.PostSeen)
		rg.GET("/seen", h.GetSeen)
		rg.DELETE("/seen", h.DeleteSeen)

		// Recommendation endpoints (The Core) — rate-limited (OpenAI cost)
		rg.POST("/recommend", rateLimit, h.PostRecommend)
		rg.GET("/vibe", rateLimit, h.GetRecommendSimple)
		rg.GET("/similar/:media_id", h.GetSimilar)
		rg.GET("/hidden-gems", h.GetHiddenGems)

		// Media management endpoints — rate-limited (OpenAI cost)
		rg.POST("/media", rateLimit, h.PostMedia)
		rg.GET("/media/:id", h.GetMedia)
		rg.POST("/media/:id/refresh", rateLimit, h.PostRefreshVibe)

		// Admin endpoints — behind shared-secret auth
		rg.GET("/stats", adminAuth, h.GetStats)
		rg.POST("/admin/scrape", adminAuth, h.PostScrapeNow)
	}

	// API routes with /api prefix (for production where frontend is served from same origin)
	registerRoutes(r.Group("/api"))
	// Legacy routes without /api prefix (for backwards compatibility)
	registerRoutes(&r.RouterGroup)

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
