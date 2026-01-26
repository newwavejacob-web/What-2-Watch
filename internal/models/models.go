package models

import (
	"time"
)

// User represents a user of the recommendation system
type User struct {
	ID        string    `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Media represents a movie, TV show, or anime with its vibe profile
type Media struct {
	ID              string    `json:"id" db:"id"`
	Title           string    `json:"title" db:"title"`
	MediaType       string    `json:"media_type" db:"media_type"` // "movie", "tv", "anime"
	Year            int       `json:"year,omitempty" db:"year"`
	PlotSummary     string    `json:"plot_summary,omitempty" db:"plot_summary"`
	VibeProfile     string    `json:"vibe_profile" db:"vibe_profile"` // LLM-generated aesthetic description
	QualityScore    float64   `json:"quality_score" db:"quality_score"`
	PopularityScore float64   `json:"popularity_score" db:"popularity_score"`
	SourceSubreddit string    `json:"source_subreddit,omitempty" db:"source_subreddit"`
	ExternalID      string    `json:"external_id,omitempty" db:"external_id"` // TMDB/IMDB ID
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// SeenMedia tracks what a user has already watched
type SeenMedia struct {
	ID        int64     `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	MediaID   string    `json:"media_id" db:"media_id"`
	Rating    *float64  `json:"rating,omitempty" db:"rating"` // Optional user rating 1-10
	WatchedAt time.Time `json:"watched_at" db:"watched_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// VibeEmbedding stores the vector representation of a media's vibe profile
type VibeEmbedding struct {
	MediaID   string    `json:"media_id" db:"media_id"`
	Embedding []float32 `json:"embedding" db:"embedding"` // Vector from embedding model
	Model     string    `json:"model" db:"model"`         // e.g., "text-embedding-3-small"
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// RedditThread represents scraped data from recommendation subreddits
type RedditThread struct {
	ID            string    `json:"id" db:"id"`
	Subreddit     string    `json:"subreddit" db:"subreddit"`
	Title         string    `json:"title" db:"title"`
	Body          string    `json:"body" db:"body"`
	ThreadType    string    `json:"thread_type" db:"thread_type"` // "similar_to", "hidden_gem", "quality_discussion"
	ReferenceShow string    `json:"reference_show,omitempty" db:"reference_show"`
	Score         int       `json:"score" db:"score"`
	NumComments   int       `json:"num_comments" db:"num_comments"`
	ScrapedAt     time.Time `json:"scraped_at" db:"scraped_at"`
}

// RedditMention tracks when a show is mentioned in a thread
type RedditMention struct {
	ID             int64   `json:"id" db:"id"`
	ThreadID       string  `json:"thread_id" db:"thread_id"`
	MediaID        string  `json:"media_id" db:"media_id"`
	MentionContext string  `json:"mention_context" db:"mention_context"` // The surrounding text
	QualityBoost   float64 `json:"quality_boost" db:"quality_boost"`     // Boost from quality-related threads
}

// Recommendation is the output format for the API
type Recommendation struct {
	Media       Media   `json:"media"`
	VibeScore   float64 `json:"vibe_score"`   // Cosine similarity to query
	Explanation string  `json:"explanation"`  // LLM-generated reason for recommendation
	Rank        int     `json:"rank"`
}

// RecommendRequest is the input for the recommend endpoint
type RecommendRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Query  string `json:"query" binding:"required"` // Natural language vibe query
	Limit  int    `json:"limit,omitempty"`          // Max results (default 10)
}

// SeenRequest is the input for marking media as seen
type SeenRequest struct {
	UserID  string   `json:"user_id" binding:"required"`
	MediaID string   `json:"media_id" binding:"required"`
	Rating  *float64 `json:"rating,omitempty"` // Optional 1-10 rating
}

// VibeProfileRequest is used when generating a vibe profile for new media
type VibeProfileRequest struct {
	Title     string `json:"title" binding:"required"`
	MediaType string `json:"media_type" binding:"required"`
	Year      int    `json:"year,omitempty"`
	Synopsis  string `json:"synopsis,omitempty"`
}
