package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"w2w/internal/models"
)

// DB wraps the SQL database connection
type DB struct {
	*sql.DB
}

// New creates a new database connection and runs migrations
func New(dbPath string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{sqlDB}

	// Run migrations
	if err := db.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// migrate creates all necessary tables
func (db *DB) migrate() error {
	migrations := []string{
		// Users table
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Media table - stores movies, TV shows, anime with vibe profiles
		`CREATE TABLE IF NOT EXISTS media (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			media_type TEXT NOT NULL CHECK(media_type IN ('movie', 'tv', 'anime')),
			year INTEGER,
			plot_summary TEXT,
			vibe_profile TEXT NOT NULL,
			quality_score REAL DEFAULT 0.0,
			popularity_score REAL DEFAULT 0.0,
			source_subreddit TEXT,
			external_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Index for media lookup
		`CREATE INDEX IF NOT EXISTS idx_media_title ON media(title)`,
		`CREATE INDEX IF NOT EXISTS idx_media_type ON media(media_type)`,
		`CREATE INDEX IF NOT EXISTS idx_media_external_id ON media(external_id)`,

		// Seen media table - tracks what users have watched
		`CREATE TABLE IF NOT EXISTS seen_media (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			media_id TEXT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			rating REAL CHECK(rating IS NULL OR (rating >= 1 AND rating <= 10)),
			watched_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, media_id)
		)`,

		// Index for efficient anti-join queries
		`CREATE INDEX IF NOT EXISTS idx_seen_user_id ON seen_media(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_seen_media_id ON seen_media(media_id)`,

		// Vibe embeddings table - stores vector representations
		`CREATE TABLE IF NOT EXISTS vibe_embeddings (
			media_id TEXT PRIMARY KEY REFERENCES media(id) ON DELETE CASCADE,
			embedding BLOB NOT NULL,
			model TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Reddit threads table
		`CREATE TABLE IF NOT EXISTS reddit_threads (
			id TEXT PRIMARY KEY,
			subreddit TEXT NOT NULL,
			title TEXT NOT NULL,
			body TEXT,
			thread_type TEXT CHECK(thread_type IN ('similar_to', 'hidden_gem', 'quality_discussion', 'other')),
			reference_show TEXT,
			score INTEGER DEFAULT 0,
			num_comments INTEGER DEFAULT 0,
			scraped_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		`CREATE INDEX IF NOT EXISTS idx_threads_subreddit ON reddit_threads(subreddit)`,
		`CREATE INDEX IF NOT EXISTS idx_threads_type ON reddit_threads(thread_type)`,

		// Reddit mentions table - tracks show mentions in threads
		`CREATE TABLE IF NOT EXISTS reddit_mentions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			thread_id TEXT NOT NULL REFERENCES reddit_threads(id) ON DELETE CASCADE,
			media_id TEXT NOT NULL REFERENCES media(id) ON DELETE CASCADE,
			mention_context TEXT,
			quality_boost REAL DEFAULT 0.0,
			UNIQUE(thread_id, media_id)
		)`,

		`CREATE INDEX IF NOT EXISTS idx_mentions_media ON reddit_mentions(media_id)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}

	return nil
}

// ============================================================================
// User Operations
// ============================================================================

// CreateUser creates a new user
func (db *DB) CreateUser(user *models.User) error {
	_, err := db.Exec(
		`INSERT INTO users (id, username, created_at) VALUES (?, ?, ?)`,
		user.ID, user.Username, time.Now(),
	)
	return err
}

// GetUser retrieves a user by ID
func (db *DB) GetUser(id string) (*models.User, error) {
	user := &models.User{}
	err := db.QueryRow(
		`SELECT id, username, created_at FROM users WHERE id = ?`,
		id,
	).Scan(&user.ID, &user.Username, &user.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// ============================================================================
// Media Operations
// ============================================================================

// CreateMedia inserts a new media entry
func (db *DB) CreateMedia(media *models.Media) error {
	now := time.Now()
	_, err := db.Exec(
		`INSERT INTO media (id, title, media_type, year, plot_summary, vibe_profile,
		quality_score, popularity_score, source_subreddit, external_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		media.ID, media.Title, media.MediaType, media.Year, media.PlotSummary,
		media.VibeProfile, media.QualityScore, media.PopularityScore,
		media.SourceSubreddit, media.ExternalID, now, now,
	)
	return err
}

// GetMedia retrieves a media entry by ID
func (db *DB) GetMedia(id string) (*models.Media, error) {
	media := &models.Media{}
	err := db.QueryRow(
		`SELECT id, title, media_type, year, plot_summary, vibe_profile,
		quality_score, popularity_score, source_subreddit, external_id, created_at, updated_at
		FROM media WHERE id = ?`,
		id,
	).Scan(&media.ID, &media.Title, &media.MediaType, &media.Year, &media.PlotSummary,
		&media.VibeProfile, &media.QualityScore, &media.PopularityScore,
		&media.SourceSubreddit, &media.ExternalID, &media.CreatedAt, &media.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return media, err
}

// GetMediaByTitle finds media by exact title match
func (db *DB) GetMediaByTitle(title string) (*models.Media, error) {
	media := &models.Media{}
	err := db.QueryRow(
		`SELECT id, title, media_type, year, plot_summary, vibe_profile,
		quality_score, popularity_score, source_subreddit, external_id, created_at, updated_at
		FROM media WHERE title = ? COLLATE NOCASE`,
		title,
	).Scan(&media.ID, &media.Title, &media.MediaType, &media.Year, &media.PlotSummary,
		&media.VibeProfile, &media.QualityScore, &media.PopularityScore,
		&media.SourceSubreddit, &media.ExternalID, &media.CreatedAt, &media.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return media, err
}

// UpdateQualityScore updates the quality score for a media entry
func (db *DB) UpdateQualityScore(mediaID string, boost float64) error {
	_, err := db.Exec(
		`UPDATE media SET quality_score = quality_score + ?, updated_at = ? WHERE id = ?`,
		boost, time.Now(), mediaID,
	)
	return err
}

// ============================================================================
// Seen Media Operations (with Anti-Join support)
// ============================================================================

// MarkAsSeen adds a media to user's seen list
func (db *DB) MarkAsSeen(seen *models.SeenMedia) error {
	now := time.Now()
	_, err := db.Exec(
		`INSERT OR REPLACE INTO seen_media (user_id, media_id, rating, watched_at, created_at)
		VALUES (?, ?, ?, ?, ?)`,
		seen.UserID, seen.MediaID, seen.Rating, now, now,
	)
	return err
}

// GetSeenMedia retrieves all seen media for a user
func (db *DB) GetSeenMedia(userID string) ([]models.SeenMedia, error) {
	rows, err := db.Query(
		`SELECT id, user_id, media_id, rating, watched_at, created_at
		FROM seen_media WHERE user_id = ? ORDER BY watched_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var seen []models.SeenMedia
	for rows.Next() {
		var s models.SeenMedia
		if err := rows.Scan(&s.ID, &s.UserID, &s.MediaID, &s.Rating, &s.WatchedAt, &s.CreatedAt); err != nil {
			return nil, err
		}
		seen = append(seen, s)
	}
	return seen, rows.Err()
}

// GetSeenMediaWithDetails retrieves seen media with full media details
func (db *DB) GetSeenMediaWithDetails(userID string) ([]models.Media, error) {
	rows, err := db.Query(
		`SELECT m.id, m.title, m.media_type, m.year, m.plot_summary, m.vibe_profile,
		m.quality_score, m.popularity_score, m.source_subreddit, m.external_id, m.created_at, m.updated_at
		FROM media m
		INNER JOIN seen_media s ON m.id = s.media_id
		WHERE s.user_id = ?
		ORDER BY s.watched_at DESC`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var media []models.Media
	for rows.Next() {
		var m models.Media
		if err := rows.Scan(&m.ID, &m.Title, &m.MediaType, &m.Year, &m.PlotSummary,
			&m.VibeProfile, &m.QualityScore, &m.PopularityScore,
			&m.SourceSubreddit, &m.ExternalID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		media = append(media, m)
	}
	return media, rows.Err()
}

// IsMediaSeen checks if a user has seen a specific media
func (db *DB) IsMediaSeen(userID, mediaID string) (bool, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM seen_media WHERE user_id = ? AND media_id = ?`,
		userID, mediaID,
	).Scan(&count)
	return count > 0, err
}

// GetSeenMediaIDs returns just the IDs of seen media for efficient filtering
func (db *DB) GetSeenMediaIDs(userID string) (map[string]bool, error) {
	rows, err := db.Query(`SELECT media_id FROM seen_media WHERE user_id = ?`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	seen := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		seen[id] = true
	}
	return seen, rows.Err()
}

// ============================================================================
// Embedding Operations
// ============================================================================

// StoreEmbedding saves a vector embedding for a media entry
func (db *DB) StoreEmbedding(mediaID string, embedding []float32, model string) error {
	// Serialize embedding to JSON blob
	embBytes, err := json.Marshal(embedding)
	if err != nil {
		return fmt.Errorf("failed to serialize embedding: %w", err)
	}

	_, err = db.Exec(
		`INSERT OR REPLACE INTO vibe_embeddings (media_id, embedding, model, created_at)
		VALUES (?, ?, ?, ?)`,
		mediaID, embBytes, model, time.Now(),
	)
	return err
}

// GetEmbedding retrieves the embedding for a media entry
func (db *DB) GetEmbedding(mediaID string) ([]float32, error) {
	var embBytes []byte
	err := db.QueryRow(
		`SELECT embedding FROM vibe_embeddings WHERE media_id = ?`,
		mediaID,
	).Scan(&embBytes)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var embedding []float32
	if err := json.Unmarshal(embBytes, &embedding); err != nil {
		return nil, fmt.Errorf("failed to deserialize embedding: %w", err)
	}
	return embedding, nil
}

// GetAllEmbeddings retrieves all embeddings for vector search
// Returns a map of mediaID -> embedding
func (db *DB) GetAllEmbeddings() (map[string][]float32, error) {
	rows, err := db.Query(`SELECT media_id, embedding FROM vibe_embeddings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	embeddings := make(map[string][]float32)
	for rows.Next() {
		var mediaID string
		var embBytes []byte
		if err := rows.Scan(&mediaID, &embBytes); err != nil {
			return nil, err
		}

		var embedding []float32
		if err := json.Unmarshal(embBytes, &embedding); err != nil {
			return nil, fmt.Errorf("failed to deserialize embedding for %s: %w", mediaID, err)
		}
		embeddings[mediaID] = embedding
	}
	return embeddings, rows.Err()
}

// GetAllEmbeddingsExcludingSeen retrieves embeddings with ANTI-JOIN to exclude seen media
// This is the crucial query that filters out what the user has already watched
func (db *DB) GetAllEmbeddingsExcludingSeen(userID string) (map[string][]float32, error) {
	rows, err := db.Query(
		`SELECT ve.media_id, ve.embedding
		FROM vibe_embeddings ve
		LEFT JOIN seen_media sm ON ve.media_id = sm.media_id AND sm.user_id = ?
		WHERE sm.media_id IS NULL`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	embeddings := make(map[string][]float32)
	for rows.Next() {
		var mediaID string
		var embBytes []byte
		if err := rows.Scan(&mediaID, &embBytes); err != nil {
			return nil, err
		}

		var embedding []float32
		if err := json.Unmarshal(embBytes, &embedding); err != nil {
			return nil, fmt.Errorf("failed to deserialize embedding for %s: %w", mediaID, err)
		}
		embeddings[mediaID] = embedding
	}
	return embeddings, rows.Err()
}

// ============================================================================
// Reddit Scraping Operations
// ============================================================================

// CreateRedditThread stores a scraped thread
func (db *DB) CreateRedditThread(thread *models.RedditThread) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO reddit_threads
		(id, subreddit, title, body, thread_type, reference_show, score, num_comments, scraped_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		thread.ID, thread.Subreddit, thread.Title, thread.Body,
		thread.ThreadType, thread.ReferenceShow, thread.Score, thread.NumComments, thread.ScrapedAt,
	)
	return err
}

// CreateRedditMention records a media mention in a thread
func (db *DB) CreateRedditMention(mention *models.RedditMention) error {
	_, err := db.Exec(
		`INSERT OR IGNORE INTO reddit_mentions
		(thread_id, media_id, mention_context, quality_boost)
		VALUES (?, ?, ?, ?)`,
		mention.ThreadID, mention.MediaID, mention.MentionContext, mention.QualityBoost,
	)
	return err
}

// GetMentionCountForMedia returns how many times a media has been mentioned
func (db *DB) GetMentionCountForMedia(mediaID string) (int, error) {
	var count int
	err := db.QueryRow(
		`SELECT COUNT(*) FROM reddit_mentions WHERE media_id = ?`,
		mediaID,
	).Scan(&count)
	return count, err
}
