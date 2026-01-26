package services

import (
	"fmt"
	"sort"

	"w2w/internal/database"
	"w2w/internal/embeddings"
	"w2w/internal/llm"
	"w2w/internal/models"
)

// VibeSearchService handles the core recommendation logic
type VibeSearchService struct {
	db          *database.DB
	embedder    embeddings.Provider
	llmClient   *llm.Client
	vectorStore *embeddings.VectorStore
}

// NewVibeSearchService creates a new vibe search service
func NewVibeSearchService(db *database.DB, embedder embeddings.Provider, llmClient *llm.Client) (*VibeSearchService, error) {
	svc := &VibeSearchService{
		db:          db,
		embedder:    embedder,
		llmClient:   llmClient,
		vectorStore: embeddings.NewVectorStore(),
	}

	// Load existing embeddings into memory
	if err := svc.LoadEmbeddings(); err != nil {
		return nil, fmt.Errorf("failed to load embeddings: %w", err)
	}

	return svc, nil
}

// LoadEmbeddings loads all embeddings from the database into the vector store
func (s *VibeSearchService) LoadEmbeddings() error {
	allEmbeddings, err := s.db.GetAllEmbeddings()
	if err != nil {
		return err
	}
	s.vectorStore.LoadFromMap(allEmbeddings)
	return nil
}

// IngestMedia adds a new media entry with its vibe profile and embedding
func (s *VibeSearchService) IngestMedia(req models.VibeProfileRequest) (*models.Media, error) {
	// Check if media already exists
	existing, err := s.db.GetMediaByTitle(req.Title)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing media: %w", err)
	}
	if existing != nil {
		return existing, nil
	}

	// Generate vibe profile using LLM
	vibeProfile, err := s.llmClient.GenerateVibeProfile(req.Title, req.MediaType, req.Year, req.Synopsis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate vibe profile: %w", err)
	}

	// Create media entry
	media := &models.Media{
		ID:          generateID(req.Title, req.MediaType),
		Title:       req.Title,
		MediaType:   req.MediaType,
		Year:        req.Year,
		PlotSummary: req.Synopsis,
		VibeProfile: vibeProfile,
	}

	if err := s.db.CreateMedia(media); err != nil {
		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	// Generate and store embedding for the vibe profile
	embedding, err := s.embedder.Embed(vibeProfile)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if err := s.db.StoreEmbedding(media.ID, embedding, s.embedder.ModelName()); err != nil {
		return nil, fmt.Errorf("failed to store embedding: %w", err)
	}

	// Add to in-memory vector store
	s.vectorStore.Add(media.ID, embedding)

	return media, nil
}

// SearchConfig holds configuration for a vibe search
type SearchConfig struct {
	UserID       string
	Query        string
	TopK         int  // Number of candidates to retrieve from vector search
	FinalResults int  // Number of final results after reranking
	UseReranking bool // Whether to use LLM reranking
}

// SearchResult holds the result of a vibe search
type SearchResult struct {
	Recommendations []models.Recommendation
	Query           string
	TotalCandidates int
	FilteredCount   int // How many were filtered due to being seen
}

// Search performs the full vibe search pipeline:
// 1. Convert query to vector
// 2. Find top candidates via vector similarity
// 3. Apply anti-join to filter seen media
// 4. Optionally rerank via LLM
func (s *VibeSearchService) Search(config SearchConfig) (*SearchResult, error) {
	// Set defaults
	if config.TopK <= 0 {
		config.TopK = 20
	}
	if config.FinalResults <= 0 {
		config.FinalResults = 10
	}

	// Step 1: Generate embedding for the query
	queryEmbedding, err := s.embedder.Embed(config.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to embed query: %w", err)
	}

	// Step 2: Get the user's seen media for filtering (anti-join)
	seenIDs, err := s.db.GetSeenMediaIDs(config.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seen media: %w", err)
	}

	// Step 3: Vector search with anti-join (exclude seen media)
	candidates := s.vectorStore.Search(queryEmbedding, config.TopK, seenIDs)

	if len(candidates) == 0 {
		return &SearchResult{
			Recommendations: []models.Recommendation{},
			Query:           config.Query,
			TotalCandidates: 0,
			FilteredCount:   len(seenIDs),
		}, nil
	}

	// Step 4: Fetch full media details for candidates
	var rerankCandidates []llm.RerankCandidate
	for _, c := range candidates {
		media, err := s.db.GetMedia(c.MediaID)
		if err != nil || media == nil {
			continue
		}
		rerankCandidates = append(rerankCandidates, llm.RerankCandidate{
			Media:     *media,
			VibeScore: c.Similarity,
		})
	}

	// Step 5: Optionally rerank using LLM
	var recommendations []models.Recommendation

	if config.UseReranking && s.llmClient != nil && len(rerankCandidates) > 0 {
		// Use LLM to rerank based on vibe match
		reranked, err := s.llmClient.RerankByVibe(config.Query, rerankCandidates)
		if err != nil {
			// Fall back to vector similarity ranking on error
			for i, c := range rerankCandidates {
				if i >= config.FinalResults {
					break
				}
				recommendations = append(recommendations, models.Recommendation{
					Media:       c.Media,
					VibeScore:   c.VibeScore,
					Explanation: fmt.Sprintf("Vibe match based on: %s", c.Media.VibeProfile),
					Rank:        i + 1,
				})
			}
		} else {
			// Build recommendations from reranked results
			mediaMap := make(map[string]llm.RerankCandidate)
			for _, c := range rerankCandidates {
				mediaMap[c.Media.ID] = c
			}

			for _, r := range reranked {
				if candidate, ok := mediaMap[r.MediaID]; ok {
					recommendations = append(recommendations, models.Recommendation{
						Media:       candidate.Media,
						VibeScore:   candidate.VibeScore,
						Explanation: r.Explanation,
						Rank:        r.Rank,
					})
				}
			}

			// Add remaining candidates if we need more results
			if len(recommendations) < config.FinalResults {
				rankedIDs := make(map[string]bool)
				for _, r := range recommendations {
					rankedIDs[r.Media.ID] = true
				}

				for _, c := range rerankCandidates {
					if len(recommendations) >= config.FinalResults {
						break
					}
					if !rankedIDs[c.Media.ID] {
						recommendations = append(recommendations, models.Recommendation{
							Media:       c.Media,
							VibeScore:   c.VibeScore,
							Explanation: fmt.Sprintf("Similar vibe: %s", c.Media.VibeProfile),
							Rank:        len(recommendations) + 1,
						})
					}
				}
			}
		}
	} else {
		// No reranking - just use vector similarity order
		for i, c := range rerankCandidates {
			if i >= config.FinalResults {
				break
			}
			recommendations = append(recommendations, models.Recommendation{
				Media:       c.Media,
				VibeScore:   c.VibeScore,
				Explanation: fmt.Sprintf("Vibe match: %s", c.Media.VibeProfile),
				Rank:        i + 1,
			})
		}
	}

	return &SearchResult{
		Recommendations: recommendations,
		Query:           config.Query,
		TotalCandidates: len(candidates),
		FilteredCount:   len(seenIDs),
	}, nil
}

// GetSimilarToMedia finds media similar to a specific title
func (s *VibeSearchService) GetSimilarToMedia(userID, mediaID string, limit int) ([]models.Recommendation, error) {
	// Get the source media's embedding
	sourceEmbedding, err := s.db.GetEmbedding(mediaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source embedding: %w", err)
	}
	if sourceEmbedding == nil {
		return nil, fmt.Errorf("no embedding found for media %s", mediaID)
	}

	// Get seen media for filtering
	seenIDs, err := s.db.GetSeenMediaIDs(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seen media: %w", err)
	}

	// Also exclude the source media itself
	seenIDs[mediaID] = true

	// Search for similar
	candidates := s.vectorStore.Search(sourceEmbedding, limit*2, seenIDs)

	var recommendations []models.Recommendation
	for i, c := range candidates {
		if i >= limit {
			break
		}
		media, err := s.db.GetMedia(c.MediaID)
		if err != nil || media == nil {
			continue
		}
		recommendations = append(recommendations, models.Recommendation{
			Media:       *media,
			VibeScore:   c.Similarity,
			Explanation: fmt.Sprintf("Similar vibe to source: %s", media.VibeProfile),
			Rank:        i + 1,
		})
	}

	return recommendations, nil
}

// GetHiddenGems finds high-quality but less popular media
func (s *VibeSearchService) GetHiddenGems(userID string, limit int) ([]models.Media, error) {
	// Get seen media for filtering
	seenIDs, err := s.db.GetSeenMediaIDs(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get seen media: %w", err)
	}

	// Query for high quality_score but lower popularity_score
	rows, err := s.db.Query(`
		SELECT m.id, m.title, m.media_type, m.year, m.plot_summary, m.vibe_profile,
		       m.quality_score, m.popularity_score, m.source_subreddit, m.external_id,
		       m.created_at, m.updated_at
		FROM media m
		LEFT JOIN seen_media sm ON m.id = sm.media_id AND sm.user_id = ?
		WHERE sm.media_id IS NULL
		AND m.quality_score > m.popularity_score * 0.5
		ORDER BY (m.quality_score - m.popularity_score * 0.3) DESC
		LIMIT ?
	`, userID, limit*2)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gems []models.Media
	for rows.Next() {
		var m models.Media
		if err := rows.Scan(&m.ID, &m.Title, &m.MediaType, &m.Year, &m.PlotSummary,
			&m.VibeProfile, &m.QualityScore, &m.PopularityScore,
			&m.SourceSubreddit, &m.ExternalID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			continue
		}
		if !seenIDs[m.ID] && len(gems) < limit {
			gems = append(gems, m)
		}
	}

	return gems, nil
}

// RefreshEmbedding regenerates the vibe profile and embedding for a media entry
func (s *VibeSearchService) RefreshEmbedding(mediaID string) error {
	media, err := s.db.GetMedia(mediaID)
	if err != nil {
		return fmt.Errorf("failed to get media: %w", err)
	}
	if media == nil {
		return fmt.Errorf("media not found: %s", mediaID)
	}

	// Generate new vibe profile
	vibeProfile, err := s.llmClient.GenerateVibeProfile(
		media.Title, media.MediaType, media.Year, media.PlotSummary,
	)
	if err != nil {
		return fmt.Errorf("failed to generate vibe profile: %w", err)
	}

	// Update media
	_, err = s.db.Exec(`UPDATE media SET vibe_profile = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`,
		vibeProfile, mediaID)
	if err != nil {
		return fmt.Errorf("failed to update media: %w", err)
	}

	// Generate and store new embedding
	embedding, err := s.embedder.Embed(vibeProfile)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if err := s.db.StoreEmbedding(mediaID, embedding, s.embedder.ModelName()); err != nil {
		return fmt.Errorf("failed to store embedding: %w", err)
	}

	// Update vector store
	s.vectorStore.Add(mediaID, embedding)

	return nil
}

// GetStats returns statistics about the vibe search index
func (s *VibeSearchService) GetStats() map[string]interface{} {
	var mediaCount, embeddingCount int
	s.db.QueryRow(`SELECT COUNT(*) FROM media`).Scan(&mediaCount)
	s.db.QueryRow(`SELECT COUNT(*) FROM vibe_embeddings`).Scan(&embeddingCount)

	return map[string]interface{}{
		"media_count":           mediaCount,
		"embedding_count":       embeddingCount,
		"vector_store_size":     s.vectorStore.Size(),
		"embedding_model":       s.embedder.ModelName(),
	}
}

// ByVibeScore implements sort.Interface for sorting recommendations
type ByVibeScore []models.Recommendation

func (a ByVibeScore) Len() int           { return len(a) }
func (a ByVibeScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByVibeScore) Less(i, j int) bool { return a[i].VibeScore > a[j].VibeScore }

// SortByVibeScore sorts recommendations by their vibe score
func SortByVibeScore(recs []models.Recommendation) {
	sort.Sort(ByVibeScore(recs))
}

// generateID creates a deterministic ID from title and type
func generateID(title, mediaType string) string {
	// Simple hash-like ID generation
	// In production, you might want to use UUID or a proper hash
	id := fmt.Sprintf("%s-%s", mediaType, title)
	// Sanitize for use as ID
	result := ""
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			result += string(r)
		} else if r == ' ' || r == '-' {
			result += "-"
		}
	}
	return result
}
