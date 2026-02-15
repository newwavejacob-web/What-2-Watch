package tmdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const baseURL = "https://api.themoviedb.org/3"

// Client handles TMDB API requests with rate limiting.
type Client struct {
	apiKey     string
	httpClient *http.Client
	mu         sync.Mutex
	lastReq    time.Time
	minDelay   time.Duration // minimum delay between requests
}

// NewClient creates a TMDB API client with rate limiting.
// TMDB allows ~40 requests per 10 seconds; we target ~3.5 req/s to be safe.
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		minDelay: 285 * time.Millisecond,
	}
}

func (c *Client) doGet(url string) ([]byte, error) {
	c.mu.Lock()
	elapsed := time.Since(c.lastReq)
	if elapsed < c.minDelay {
		time.Sleep(c.minDelay - elapsed)
	}
	c.lastReq = time.Now()
	c.mu.Unlock()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		// Rate limited — back off and retry once
		retryAfter := 2 * time.Second
		time.Sleep(retryAfter)
		return c.doGet(url) // single retry
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("not found")
	}

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TMDB API error %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// DiscoverPage represents a page of discover results.
type DiscoverPage struct {
	Page         int             `json:"page"`
	TotalPages   int             `json:"total_pages"`
	TotalResults int             `json:"total_results"`
	Results      []DiscoverEntry `json:"results"`
}

// DiscoverEntry is a minimal entry from the discover endpoint.
type DiscoverEntry struct {
	ID               int     `json:"id"`
	Title            string  `json:"title"`             // movies
	Name             string  `json:"name"`              // TV
	OriginalLanguage string  `json:"original_language"`
	GenreIDs         []int   `json:"genre_ids"`
	Popularity       float64 `json:"popularity"`
	VoteAverage      float64 `json:"vote_average"`
	VoteCount        int     `json:"vote_count"`
	Overview         string  `json:"overview"`
}

// GetTitle returns the title (works for both movie and TV).
func (e DiscoverEntry) GetTitle() string {
	if e.Title != "" {
		return e.Title
	}
	return e.Name
}

// DiscoverMovies fetches a page of popular movies from TMDB.
func (c *Client) DiscoverMovies(page int) (*DiscoverPage, error) {
	url := fmt.Sprintf("%s/discover/movie?sort_by=popularity.desc&page=%d&vote_count.gte=50&language=en-US", baseURL, page)
	data, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	var result DiscoverPage
	return &result, json.Unmarshal(data, &result)
}

// DiscoverTV fetches a page of popular TV shows from TMDB.
func (c *Client) DiscoverTV(page int) (*DiscoverPage, error) {
	url := fmt.Sprintf("%s/discover/tv?sort_by=popularity.desc&page=%d&vote_count.gte=50&language=en-US", baseURL, page)
	data, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	var result DiscoverPage
	return &result, json.Unmarshal(data, &result)
}

// DiscoverAnime fetches a page of Japanese animated TV shows.
func (c *Client) DiscoverAnime(page int) (*DiscoverPage, error) {
	// Genre 16 = Animation, with_original_language=ja for Japanese anime
	url := fmt.Sprintf("%s/discover/tv?sort_by=popularity.desc&page=%d&with_genres=16&with_original_language=ja&vote_count.gte=20&language=en-US", baseURL, page)
	data, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	var result DiscoverPage
	return &result, json.Unmarshal(data, &result)
}

// DiscoverAnimeMovies fetches a page of Japanese animated movies.
func (c *Client) DiscoverAnimeMovies(page int) (*DiscoverPage, error) {
	url := fmt.Sprintf("%s/discover/movie?sort_by=popularity.desc&page=%d&with_genres=16&with_original_language=ja&vote_count.gte=20&language=en-US", baseURL, page)
	data, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	var result DiscoverPage
	return &result, json.Unmarshal(data, &result)
}

// MovieDetails represents detailed info about a movie.
type MovieDetails struct {
	ID               int      `json:"id"`
	Title            string   `json:"title"`
	OriginalTitle    string   `json:"original_title"`
	Overview         string   `json:"overview"`
	Tagline          string   `json:"tagline"`
	ReleaseDate      string   `json:"release_date"`
	Runtime          int      `json:"runtime"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	OriginalLanguage string   `json:"original_language"`
	Status           string   `json:"status"`
	Genres           []Genre  `json:"genres"`
	Credits          *Credits `json:"credits,omitempty"`
	Keywords         *struct {
		Keywords []Keyword `json:"keywords"`
	} `json:"keywords,omitempty"`
}

// TVDetails represents detailed info about a TV series.
type TVDetails struct {
	ID               int      `json:"id"`
	Name             string   `json:"name"`
	OriginalName     string   `json:"original_name"`
	Overview         string   `json:"overview"`
	Tagline          string   `json:"tagline"`
	FirstAirDate     string   `json:"first_air_date"`
	VoteAverage      float64  `json:"vote_average"`
	VoteCount        int      `json:"vote_count"`
	Popularity       float64  `json:"popularity"`
	PosterPath       string   `json:"poster_path"`
	OriginalLanguage string   `json:"original_language"`
	Status           string   `json:"status"`
	Genres           []Genre  `json:"genres"`
	Credits          *Credits `json:"credits,omitempty"`
	Keywords         *struct {
		Results []Keyword `json:"results"`
	} `json:"keywords,omitempty"`
	CreatedBy []struct {
		Name string `json:"name"`
	} `json:"created_by"`
	EpisodeRunTime []int `json:"episode_run_time"`
}

// Genre represents a TMDB genre.
type Genre struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Keyword represents a TMDB keyword.
type Keyword struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Credits contains cast and crew info.
type Credits struct {
	Cast []CastMember `json:"cast"`
	Crew []CrewMember `json:"crew"`
}

// CastMember is an actor in the credits.
type CastMember struct {
	Name      string `json:"name"`
	Character string `json:"character"`
	Order     int    `json:"order"`
}

// CrewMember is a crew member in the credits.
type CrewMember struct {
	Name string `json:"name"`
	Job  string `json:"job"`
}

// GetMovieDetails fetches full details for a movie.
func (c *Client) GetMovieDetails(id int) (*MovieDetails, error) {
	url := fmt.Sprintf("%s/movie/%d?append_to_response=credits,keywords&language=en-US", baseURL, id)
	data, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	var result MovieDetails
	return &result, json.Unmarshal(data, &result)
}

// GetTVDetails fetches full details for a TV series.
func (c *Client) GetTVDetails(id int) (*TVDetails, error) {
	url := fmt.Sprintf("%s/tv/%d?append_to_response=credits,keywords&language=en-US", baseURL, id)
	data, err := c.doGet(url)
	if err != nil {
		return nil, err
	}
	var result TVDetails
	return &result, json.Unmarshal(data, &result)
}

// BuildMovieVibeText creates the composite text for embedding a movie.
func BuildMovieVibeText(m *MovieDetails) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Title: %s\n", m.Title)
	fmt.Fprintf(&b, "Type: movie\n")
	if year := extractYear(m.ReleaseDate); year != "" {
		fmt.Fprintf(&b, "Year: %s\n", year)
	}

	if len(m.Genres) > 0 {
		names := make([]string, len(m.Genres))
		for i, g := range m.Genres {
			names[i] = g.Name
		}
		fmt.Fprintf(&b, "Genres: %s\n", strings.Join(names, ", "))
	}

	if m.Keywords != nil && len(m.Keywords.Keywords) > 0 {
		limit := 15
		if len(m.Keywords.Keywords) < limit {
			limit = len(m.Keywords.Keywords)
		}
		names := make([]string, limit)
		for i := 0; i < limit; i++ {
			names[i] = m.Keywords.Keywords[i].Name
		}
		fmt.Fprintf(&b, "Keywords: %s\n", strings.Join(names, ", "))
	}

	if m.Tagline != "" {
		fmt.Fprintf(&b, "Tagline: %s\n", m.Tagline)
	}

	if m.Overview != "" {
		fmt.Fprintf(&b, "Overview: %s\n", m.Overview)
	}

	if m.Credits != nil {
		// Top 5 cast
		castLimit := 5
		if len(m.Credits.Cast) < castLimit {
			castLimit = len(m.Credits.Cast)
		}
		if castLimit > 0 {
			names := make([]string, castLimit)
			for i := 0; i < castLimit; i++ {
				names[i] = m.Credits.Cast[i].Name
			}
			fmt.Fprintf(&b, "Cast: %s\n", strings.Join(names, ", "))
		}

		// Directors
		var directors []string
		for _, c := range m.Credits.Crew {
			if c.Job == "Director" {
				directors = append(directors, c.Name)
			}
		}
		if len(directors) > 0 {
			fmt.Fprintf(&b, "Director: %s\n", strings.Join(directors, ", "))
		}
	}

	return b.String()
}

// BuildTVVibeText creates the composite text for embedding a TV series.
func BuildTVVibeText(t *TVDetails) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Title: %s\n", t.Name)
	if isAnime(t) {
		fmt.Fprintf(&b, "Type: anime\n")
	} else {
		fmt.Fprintf(&b, "Type: tv series\n")
	}
	if year := extractYear(t.FirstAirDate); year != "" {
		fmt.Fprintf(&b, "Year: %s\n", year)
	}

	if len(t.Genres) > 0 {
		names := make([]string, len(t.Genres))
		for i, g := range t.Genres {
			names[i] = g.Name
		}
		fmt.Fprintf(&b, "Genres: %s\n", strings.Join(names, ", "))
	}

	if t.Keywords != nil && len(t.Keywords.Results) > 0 {
		limit := 15
		if len(t.Keywords.Results) < limit {
			limit = len(t.Keywords.Results)
		}
		names := make([]string, limit)
		for i := 0; i < limit; i++ {
			names[i] = t.Keywords.Results[i].Name
		}
		fmt.Fprintf(&b, "Keywords: %s\n", strings.Join(names, ", "))
	}

	if t.Tagline != "" {
		fmt.Fprintf(&b, "Tagline: %s\n", t.Tagline)
	}

	if t.Overview != "" {
		fmt.Fprintf(&b, "Overview: %s\n", t.Overview)
	}

	if t.Credits != nil {
		castLimit := 5
		if len(t.Credits.Cast) < castLimit {
			castLimit = len(t.Credits.Cast)
		}
		if castLimit > 0 {
			names := make([]string, castLimit)
			for i := 0; i < castLimit; i++ {
				names[i] = t.Credits.Cast[i].Name
			}
			fmt.Fprintf(&b, "Cast: %s\n", strings.Join(names, ", "))
		}
	}

	if len(t.CreatedBy) > 0 {
		names := make([]string, len(t.CreatedBy))
		for i, c := range t.CreatedBy {
			names[i] = c.Name
		}
		fmt.Fprintf(&b, "Creator: %s\n", strings.Join(names, ", "))
	}

	return b.String()
}

// isAnime checks if a TV show is likely anime based on language and genre.
func isAnime(t *TVDetails) bool {
	if t.OriginalLanguage != "ja" {
		return false
	}
	for _, g := range t.Genres {
		if g.ID == 16 { // Animation
			return true
		}
	}
	return false
}

// IsAnimeMovie checks if a movie is Japanese animation.
func IsAnimeMovie(m *MovieDetails) bool {
	if m.OriginalLanguage != "ja" {
		return false
	}
	for _, g := range m.Genres {
		if g.ID == 16 { // Animation
			return true
		}
	}
	return false
}

func extractYear(dateStr string) string {
	if len(dateStr) >= 4 {
		return dateStr[:4]
	}
	return ""
}
