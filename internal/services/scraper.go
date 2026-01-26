package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"w2w/internal/database"
	"w2w/internal/llm"
	"w2w/internal/models"
)

// RedditScraper handles scraping recommendation subreddits
type RedditScraper struct {
	db         *database.DB
	llmClient  *llm.Client
	httpClient *http.Client
	subreddits []string
	mu         sync.Mutex
	running    bool
	stopCh     chan struct{}
}

// NewRedditScraper creates a new Reddit scraper
func NewRedditScraper(db *database.DB, llmClient *llm.Client) *RedditScraper {
	return &RedditScraper{
		db:        db,
		llmClient: llmClient,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		subreddits: []string{
			"animesuggest",
			"MovieSuggestions",
			"televisionsuggestions",
		},
	}
}

// redditListing represents the Reddit API response structure
type redditListing struct {
	Data struct {
		Children []struct {
			Data struct {
				ID          string  `json:"id"`
				Title       string  `json:"title"`
				Selftext    string  `json:"selftext"`
				Score       int     `json:"score"`
				NumComments int     `json:"num_comments"`
				Created     float64 `json:"created_utc"`
				Subreddit   string  `json:"subreddit"`
			} `json:"data"`
		} `json:"children"`
		After string `json:"after"`
	} `json:"data"`
}

// qualityKeywords that boost the quality_score when found in thread context
var qualityKeywords = map[string]float64{
	"masterpiece":        0.5,
	"beautifully":        0.3,
	"incredible writing": 0.5,
	"unique art":         0.4,
	"underrated":         0.3,
	"hidden gem":         0.4,
	"best written":       0.5,
	"emotional":          0.2,
	"thought-provoking":  0.4,
	"art style":          0.3,
	"cinematography":     0.3,
	"atmospheric":        0.3,
	"character study":    0.3,
}

// Start begins the background scraping worker
func (s *RedditScraper) Start(ctx context.Context, interval time.Duration) {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Run immediately on start
		s.scrapeAll()

		for {
			select {
			case <-ticker.C:
				s.scrapeAll()
			case <-s.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop halts the background scraper
func (s *RedditScraper) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		close(s.stopCh)
		s.running = false
	}
}

// scrapeAll scrapes all configured subreddits
func (s *RedditScraper) scrapeAll() {
	for _, subreddit := range s.subreddits {
		if err := s.scrapeSubreddit(subreddit); err != nil {
			log.Printf("Error scraping r/%s: %v", subreddit, err)
		}
		// Rate limiting - Reddit API is strict
		time.Sleep(2 * time.Second)
	}
}

// scrapeSubreddit fetches and processes posts from a subreddit
func (s *RedditScraper) scrapeSubreddit(subreddit string) error {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/hot.json?limit=50", subreddit)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Reddit requires a custom User-Agent
	req.Header.Set("User-Agent", "VibeRecommender/1.0 (educational project)")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read body: %w", err)
	}

	var listing redditListing
	if err := json.Unmarshal(body, &listing); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Process each thread
	for _, child := range listing.Data.Children {
		post := child.Data

		// Skip low-engagement posts
		if post.Score < 5 {
			continue
		}

		thread := &models.RedditThread{
			ID:          post.ID,
			Subreddit:   post.Subreddit,
			Title:       post.Title,
			Body:        post.Selftext,
			Score:       post.Score,
			NumComments: post.NumComments,
			ScrapedAt:   time.Now(),
		}

		// Classify the thread type
		if s.llmClient != nil {
			threadType, refShow, err := s.llmClient.ClassifyThreadType(post.Title, post.Selftext)
			if err == nil {
				thread.ThreadType = threadType
				thread.ReferenceShow = refShow
			}
		} else {
			// Fallback to keyword-based classification
			thread.ThreadType = classifyByKeywords(post.Title, post.Selftext)
			thread.ReferenceShow = extractReferenceShow(post.Title)
		}

		// Store thread
		if err := s.db.CreateRedditThread(thread); err != nil {
			log.Printf("Failed to store thread %s: %v", thread.ID, err)
			continue
		}

		// Process mentions in the thread
		if err := s.processMentions(thread); err != nil {
			log.Printf("Failed to process mentions for %s: %v", thread.ID, err)
		}
	}

	return nil
}

// processMentions extracts and stores show mentions from a thread
func (s *RedditScraper) processMentions(thread *models.RedditThread) error {
	// Combine title and body for extraction
	fullText := thread.Title + "\n" + thread.Body

	var mentions []string
	if s.llmClient != nil {
		var err error
		mentions, err = s.llmClient.ExtractMentions(fullText)
		if err != nil {
			log.Printf("LLM extraction failed, using fallback: %v", err)
			mentions = extractMentionsByPattern(fullText)
		}
	} else {
		mentions = extractMentionsByPattern(fullText)
	}

	// Calculate quality boost based on thread type and keywords
	qualityBoost := calculateQualityBoost(thread, fullText)

	for _, title := range mentions {
		// Try to find existing media
		media, err := s.db.GetMediaByTitle(title)
		if err != nil {
			continue
		}

		if media != nil {
			// Create mention record
			mention := &models.RedditMention{
				ThreadID:       thread.ID,
				MediaID:        media.ID,
				MentionContext: extractContext(fullText, title),
				QualityBoost:   qualityBoost,
			}
			s.db.CreateRedditMention(mention)

			// Update media quality score
			s.db.UpdateQualityScore(media.ID, qualityBoost)
		}
	}

	return nil
}

// classifyByKeywords does simple keyword-based thread classification
func classifyByKeywords(title, body string) string {
	text := strings.ToLower(title + " " + body)

	if strings.Contains(text, "similar to") ||
		strings.Contains(text, "like ") ||
		strings.Contains(text, "if you liked") {
		return "similar_to"
	}

	if strings.Contains(text, "hidden gem") ||
		strings.Contains(text, "underrated") ||
		strings.Contains(text, "unknown") ||
		strings.Contains(text, "overlooked") {
		return "hidden_gem"
	}

	if strings.Contains(text, "best written") ||
		strings.Contains(text, "unique art") ||
		strings.Contains(text, "masterpiece") ||
		strings.Contains(text, "quality") {
		return "quality_discussion"
	}

	return "other"
}

// extractReferenceShow tries to extract a show name from "similar to X" patterns
func extractReferenceShow(title string) string {
	patterns := []string{
		"similar to ",
		"like ",
		"if you liked ",
		"shows like ",
		"movies like ",
		"anime like ",
	}

	lower := strings.ToLower(title)
	for _, pattern := range patterns {
		if idx := strings.Index(lower, pattern); idx != -1 {
			// Extract the part after the pattern
			rest := title[idx+len(pattern):]
			// Take until common delimiters
			for _, delim := range []string{",", "?", "!", " but", " and", " or"} {
				if delimIdx := strings.Index(rest, delim); delimIdx != -1 {
					rest = rest[:delimIdx]
				}
			}
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// extractMentionsByPattern uses simple patterns to find potential titles
func extractMentionsByPattern(text string) []string {
	// This is a simplified extraction - the LLM version is much better
	// Look for capitalized phrases that might be titles
	var mentions []string

	// Split into sentences/phrases
	phrases := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '.' || r == '!' || r == '?' || r == '\n'
	})

	for _, phrase := range phrases {
		phrase = strings.TrimSpace(phrase)
		words := strings.Fields(phrase)

		// Look for sequences of capitalized words
		var currentTitle []string
		for _, word := range words {
			if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
				currentTitle = append(currentTitle, word)
			} else if len(currentTitle) >= 2 {
				// End of a potential title
				title := strings.Join(currentTitle, " ")
				if len(title) > 3 && !isCommonWord(title) {
					mentions = append(mentions, title)
				}
				currentTitle = nil
			}
		}

		// Check remaining
		if len(currentTitle) >= 2 {
			title := strings.Join(currentTitle, " ")
			if len(title) > 3 && !isCommonWord(title) {
				mentions = append(mentions, title)
			}
		}
	}

	return mentions
}

// isCommonWord filters out common phrases that aren't titles
func isCommonWord(s string) bool {
	common := map[string]bool{
		"I":     true,
		"The":   true,
		"A":     true,
		"It":    true,
		"This":  true,
		"That":  true,
		"My":    true,
		"Your":  true,
		"Their": true,
	}
	return common[s]
}

// calculateQualityBoost determines how much to boost quality score
func calculateQualityBoost(thread *models.RedditThread, text string) float64 {
	boost := 0.0
	lowerText := strings.ToLower(text)

	// Boost for thread type
	switch thread.ThreadType {
	case "hidden_gem":
		boost += 0.5
	case "quality_discussion":
		boost += 0.3
	}

	// Boost for quality keywords
	for keyword, value := range qualityKeywords {
		if strings.Contains(lowerText, keyword) {
			boost += value
		}
	}

	// Scale by engagement (logarithmic to avoid huge boosts)
	if thread.Score > 100 {
		boost *= 1.5
	} else if thread.Score > 50 {
		boost *= 1.2
	}

	// Cap the maximum boost
	if boost > 2.0 {
		boost = 2.0
	}

	return boost
}

// extractContext gets the surrounding text where a title was mentioned
func extractContext(text, title string) string {
	idx := strings.Index(strings.ToLower(text), strings.ToLower(title))
	if idx == -1 {
		return ""
	}

	start := idx - 50
	if start < 0 {
		start = 0
	}

	end := idx + len(title) + 50
	if end > len(text) {
		end = len(text)
	}

	return "..." + text[start:end] + "..."
}

// ScrapeNow triggers an immediate scrape (for manual invocation)
func (s *RedditScraper) ScrapeNow() error {
	s.scrapeAll()
	return nil
}

// GetScrapingStats returns statistics about scraped data
func (s *RedditScraper) GetScrapingStats() map[string]interface{} {
	var threadCount, mentionCount int
	s.db.QueryRow(`SELECT COUNT(*) FROM reddit_threads`).Scan(&threadCount)
	s.db.QueryRow(`SELECT COUNT(*) FROM reddit_mentions`).Scan(&mentionCount)

	var bySubreddit []struct {
		Subreddit string `json:"subreddit"`
		Count     int    `json:"count"`
	}

	rows, _ := s.db.Query(`SELECT subreddit, COUNT(*) as cnt FROM reddit_threads GROUP BY subreddit`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var sub string
			var cnt int
			if rows.Scan(&sub, &cnt) == nil {
				bySubreddit = append(bySubreddit, struct {
					Subreddit string `json:"subreddit"`
					Count     int    `json:"count"`
				}{sub, cnt})
			}
		}
	}

	return map[string]interface{}{
		"total_threads":  threadCount,
		"total_mentions": mentionCount,
		"by_subreddit":   bySubreddit,
		"subreddits":     s.subreddits,
	}
}
