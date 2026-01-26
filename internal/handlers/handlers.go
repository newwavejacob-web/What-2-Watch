package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"w2w/internal/database"
	"w2w/internal/models"
	"w2w/internal/services"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db            *database.DB
	vibeSearch    *services.VibeSearchService
	scraper       *services.RedditScraper
}

// NewHandler creates a new handler with dependencies
func NewHandler(db *database.DB, vibeSearch *services.VibeSearchService, scraper *services.RedditScraper) *Handler {
	return &Handler{
		db:         db,
		vibeSearch: vibeSearch,
		scraper:    scraper,
	}
}

// ============================================================================
// Seen Media Endpoints
// ============================================================================

// PostSeen marks a media item as seen by the user
// POST /seen
func (h *Handler) PostSeen(c *gin.Context) {
	var req models.SeenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Verify user exists (or create if using auto-create)
	user, err := h.db.GetUser(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if user == nil {
		// Auto-create user for simplicity
		user = &models.User{
			ID:        req.UserID,
			Username:  req.UserID,
			CreatedAt: time.Now(),
		}
		if err := h.db.CreateUser(user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
	}

	// Verify media exists
	media, err := h.db.GetMedia(req.MediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if media == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	// Mark as seen
	seen := &models.SeenMedia{
		UserID:    req.UserID,
		MediaID:   req.MediaID,
		Rating:    req.Rating,
		WatchedAt: time.Now(),
	}

	if err := h.db.MarkAsSeen(seen); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark as seen"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Marked as seen",
		"media":   media.Title,
		"user_id": req.UserID,
	})
}

// GetSeen retrieves the user's seen list
// GET /seen?user_id=xxx
func (h *Handler) GetSeen(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query parameter required"})
		return
	}

	seenMedia, err := h.db.GetSeenMediaWithDetails(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seen list"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id": userID,
		"count":   len(seenMedia),
		"seen":    seenMedia,
	})
}

// DeleteSeen removes a media from the seen list
// DELETE /seen
func (h *Handler) DeleteSeen(c *gin.Context) {
	var req models.SeenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	_, err := h.db.Exec(`DELETE FROM seen_media WHERE user_id = ? AND media_id = ?`,
		req.UserID, req.MediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Removed from seen list"})
}

// ============================================================================
// Recommendation Endpoints
// ============================================================================

// PostRecommend handles vibe-based recommendation requests
// POST /recommend
func (h *Handler) PostRecommend(c *gin.Context) {
	var req models.RecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// Set defaults
	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	// Perform vibe search with anti-join
	result, err := h.vibeSearch.Search(services.SearchConfig{
		UserID:       req.UserID,
		Query:        req.Query,
		TopK:         20,
		FinalResults: limit,
		UseReranking: true, // Use LLM reranking for best results
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":           result.Query,
		"total_candidates": result.TotalCandidates,
		"filtered_seen":   result.FilteredCount,
		"recommendations": result.Recommendations,
	})
}

// GetRecommendSimple handles simple GET-based recommendations
// GET /vibe?q=xxx&user_id=xxx
func (h *Handler) GetRecommendSimple(c *gin.Context) {
	query := c.Query("q")
	userID := c.DefaultQuery("user_id", "default")

	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tell me your vibe! Use ?q=your+vibe+description"})
		return
	}

	result, err := h.vibeSearch.Search(services.SearchConfig{
		UserID:       userID,
		Query:        query,
		TopK:         15,
		FinalResults: 5,
		UseReranking: true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"input":           query,
		"recommendations": result.Recommendations,
	})
}

// GetSimilar finds media similar to a specific title
// GET /similar/:media_id?user_id=xxx
func (h *Handler) GetSimilar(c *gin.Context) {
	mediaID := c.Param("media_id")
	userID := c.DefaultQuery("user_id", "default")

	recs, err := h.vibeSearch.GetSimilarToMedia(userID, mediaID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"source_id":       mediaID,
		"recommendations": recs,
	})
}

// GetHiddenGems returns high-quality but less popular recommendations
// GET /hidden-gems?user_id=xxx
func (h *Handler) GetHiddenGems(c *gin.Context) {
	userID := c.DefaultQuery("user_id", "default")

	gems, err := h.vibeSearch.GetHiddenGems(userID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"hidden_gems": gems,
	})
}

// ============================================================================
// Media Management Endpoints
// ============================================================================

// PostMedia ingests a new media entry with vibe profile generation
// POST /media
func (h *Handler) PostMedia(c *gin.Context) {
	var req models.VibeProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	media, err := h.vibeSearch.IngestMedia(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ingest media: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Media ingested successfully",
		"media":   media,
	})
}

// GetMedia retrieves a specific media entry
// GET /media/:id
func (h *Handler) GetMedia(c *gin.Context) {
	mediaID := c.Param("id")

	media, err := h.db.GetMedia(mediaID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	if media == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Media not found"})
		return
	}

	c.JSON(http.StatusOK, media)
}

// PostRefreshVibe regenerates the vibe profile for a media entry
// POST /media/:id/refresh
func (h *Handler) PostRefreshVibe(c *gin.Context) {
	mediaID := c.Param("id")

	if err := h.vibeSearch.RefreshEmbedding(mediaID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated media
	media, _ := h.db.GetMedia(mediaID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Vibe profile refreshed",
		"media":   media,
	})
}

// ============================================================================
// Admin/Stats Endpoints
// ============================================================================

// GetStats returns system statistics
// GET /stats
func (h *Handler) GetStats(c *gin.Context) {
	vibeStats := h.vibeSearch.GetStats()
	scraperStats := h.scraper.GetScrapingStats()

	c.JSON(http.StatusOK, gin.H{
		"vibe_search": vibeStats,
		"scraper":     scraperStats,
	})
}

// PostScrapeNow triggers an immediate Reddit scrape
// POST /admin/scrape
func (h *Handler) PostScrapeNow(c *gin.Context) {
	go func() {
		h.scraper.ScrapeNow()
	}()

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Scrape initiated in background",
	})
}

// GetHealth returns a health check
// GET /health
func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().Format(time.RFC3339),
	})
}
