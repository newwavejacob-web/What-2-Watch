# W2W - Vibe-First Recommendation Engine

**Domain:** chidaucf.win
**Stack:** Go (Gin) + React (Vite) + SQLite + OpenAI
**Purpose:** AI-powered movie/TV/anime recommendations based on "vibes" - aesthetic and emotional descriptors rather than genres or actors.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                            Caddy                                  │
│                (Reverse Proxy + SSL + Security Headers)          │
│                    chidaucf.win → movie-app:8080                 │
└──────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│                         Go Backend (Gin)                          │
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │  VibeSearch     │  │  Reddit         │  │  LLM Client     │  │
│  │  Service        │  │  Scraper        │  │  (GPT-4o-mini)  │  │
│  └────────┬────────┘  └────────┬────────┘  └────────┬────────┘  │
│           │                    │                     │           │
│           ▼                    ▼                     ▼           │
│  ┌─────────────────────────────────────────────────────────────┐ │
│  │                    SQLite Database                          │ │
│  │  Tables: media, users, seen_media, vibe_embeddings,        │ │
│  │          reddit_threads, reddit_mentions                    │ │
│  └─────────────────────────────────────────────────────────────┘ │
│           │                                                      │
│           ▼                                                      │
│  ┌─────────────────┐  ┌─────────────────┐                       │
│  │  Vector Store   │  │  OpenAI         │                       │
│  │  (In-Memory)    │  │  Embeddings     │                       │
│  │  Cosine Sim.    │  │  (1536 dims)    │                       │
│  └─────────────────┘  └─────────────────┘                       │
└──────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌──────────────────────────────────────────────────────────────────┐
│                       React Frontend (Vite)                       │
│   - SearchBar (vibe prompts)                                      │
│   - RecommendationCard (media display)                            │
│   - WatchHistory (seen media)                                     │
│   - ParticleBackground (animations)                               │
└──────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
w2w/
├── main.go                     # Entry point, routes, server config
├── cmd/
│   └── seed/main.go            # Database seeding script
├── internal/
│   ├── database/
│   │   └── database.go         # SQLite abstraction, migrations, queries
│   ├── models/
│   │   └── models.go           # Data models (Media, User, Recommendation)
│   ├── services/
│   │   ├── vibesearch.go       # Core recommendation logic
│   │   └── scraper.go          # Reddit scraper
│   ├── handlers/
│   │   └── handlers.go         # HTTP request handlers
│   ├── embeddings/
│   │   └── embeddings.go       # OpenAI embedding provider + vector store
│   └── llm/
│       └── llm.go              # GPT client for vibe profiles
├── frontend/
│   ├── src/
│   │   ├── App.jsx             # Main React component
│   │   ├── components/         # UI components
│   │   ├── hooks/              # Custom React hooks
│   │   └── lib/vibes.js        # Vibe prompt library
│   ├── package.json
│   └── vite.config.js
├── Dockerfile                  # Multi-stage build (Node + Go)
├── go.mod / go.sum
└── .env                        # Environment variables (OPENAI_API_KEY)
```

---

## Core Logic Explained

### 1. The "Vibe" Concept

Traditional recommendation systems use:
- Genres ("action", "comedy")
- Metadata (actors, directors, year)
- Collaborative filtering ("users like you also watched...")

**W2W uses "vibes"** - aesthetic and emotional descriptors:
- "Melancholic rainy-day atmosphere with existential undertones"
- "High-energy neon-lit cyberpunk with a killer synth soundtrack"
- "Cozy slice-of-life with found family themes"

**Why vibes work:**
- They capture the *feeling* of watching something, not just facts
- Users often know the mood they want, not the specific title
- Vibes transcend genre boundaries (a thriller and a drama can have the same vibe)

### 2. Data Models (`internal/models/models.go`)

**Media:**
```go
type Media struct {
    ID              string    // Deterministic: "{type}-{sanitized-title}"
    Title           string
    MediaType       string    // "movie", "tv", "anime"
    Year            int
    PlotSummary     string    // Brief synopsis
    VibeProfile     string    // LLM-generated aesthetic description
    QualityScore    float64   // Reddit quality signals
    PopularityScore float64   // How often mentioned
    SourceSubreddit string    // Where it was discovered
    ExternalID      string    // IMDB/TMDB ID if available
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

**Recommendation:**
```go
type Recommendation struct {
    Media       Media
    VibeScore   float64  // Cosine similarity to query (0-1)
    Explanation string   // LLM-generated reason for recommendation
    Rank        int      // Position in results
}
```

**SeenMedia:**
```go
type SeenMedia struct {
    ID        int64
    UserID    string
    MediaID   string
    Rating    *int      // Optional 1-10 rating
    WatchedAt time.Time
    CreatedAt time.Time
}
```

### 3. Vibe Search Pipeline (`internal/services/vibesearch.go`)

The recommendation system follows a 5-step pipeline:

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│ User Query  │────▶│  Embed      │────▶│  Vector     │
│ (natural    │     │  Query      │     │  Search     │
│  language)  │     │  (OpenAI)   │     │  (top-K)    │
└─────────────┘     └─────────────┘     └─────────────┘
                                              │
                                              ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  Rerank     │◀────│  Anti-Join  │
                    │  (LLM)      │     │  (exclude   │
                    │             │     │   seen)     │
                    └─────────────┘     └─────────────┘
                          │
                          ▼
                    ┌─────────────┐
                    │  Return     │
                    │  Results    │
                    └─────────────┘
```

**Step 1: Query Embedding**
```go
queryEmbedding, err := s.embedder.Embed(config.Query)
```
Convert the user's natural language query into a 1536-dimensional vector using OpenAI's `text-embedding-3-small` model.

**Step 2: Anti-Join (Exclude Seen)**
```go
seenIDs, err := s.db.GetSeenMediaIDs(config.UserID)
```
Fetch all media the user has already marked as seen, so we don't recommend things they've watched.

**Step 3: Vector Search**
```go
candidates := s.vectorStore.Search(queryEmbedding, config.TopK, seenIDs)
```
Find the top-K most similar media by cosine similarity:
```
cosine_similarity = (A · B) / (||A|| × ||B||)
```
The in-memory vector store iterates all embeddings and computes similarity scores.

**Step 4: Fetch Full Details**
Load full `Media` objects from SQLite for the candidate IDs.

**Step 5: LLM Reranking (Optional)**
```go
reranked, err := s.llmClient.RerankByVibe(config.Query, rerankCandidates)
```
Use GPT-4o-mini to:
- Verify vibe matches are actually relevant
- Generate human-readable explanations
- Potentially reorder based on deeper understanding

**Example Vibe Profile (Generated by LLM):**
```
"Cerebral and haunting digital noir. The pacing is deliberate and tense,
with a pervasive sense of existential dread. Visual aesthetic is sterile
and clinical with moments of visceral body horror. Themes of identity,
consciousness, and the nature of reality. Minimal but impactful
electronic score."
```

### 4. Embedding System (`internal/embeddings/embeddings.go`)

**OpenAI Provider:**
- Model: `text-embedding-3-small`
- Dimensions: 1536
- Called for: Query embedding, new media vibe profiles

**Vector Store (In-Memory):**
```go
type VectorStore struct {
    embeddings map[string][]float32  // mediaID → embedding
    mu         sync.RWMutex
}

func (vs *VectorStore) Search(query []float32, topK int, exclude map[string]bool) []SearchResult
```
- Loads all embeddings on startup from `vibe_embeddings` table
- Performs brute-force cosine similarity search
- Excludes IDs in the `exclude` map (for anti-join)

**Why In-Memory?**
- SQLite can't efficiently perform vector similarity search
- Dataset is small enough (~100s to low 1000s of media)
- Sub-millisecond search latency

### 5. LLM Client (`internal/llm/llm.go`)

**GenerateVibeProfile:**
```go
func (c *Client) GenerateVibeProfile(title, mediaType string, year int, synopsis string) (string, error)
```
Prompt template (conceptual):
```
You are an expert at describing the aesthetic and emotional qualities of media.
Given this {mediaType} "{title}" ({year}): {synopsis}

Create a vibe profile that captures:
- Visual aesthetic and cinematography style
- Emotional tone and atmosphere
- Pacing and rhythm
- Thematic elements
- Sound/music characteristics

Write 2-4 sentences, evocative and specific.
```

**RerankByVibe:**
```go
func (c *Client) RerankByVibe(query string, candidates []RerankCandidate) ([]RerankedResult, error)
```
Takes the query and candidates, returns explanations like:
```
"This matches your request for 'cozy supernatural mystery' because it features
a small-town setting with paranormal elements, warm autumn aesthetics, and an
ensemble cast solving mysteries together."
```

### 6. Reddit Scraper (`internal/services/scraper.go`)

**Target Subreddits:**
- `r/animesuggest`
- `r/MovieSuggestions`
- `r/televisionsuggestions`

**Thread Types Detected:**
- `similar_to` - "Looking for something like X"
- `hidden_gem` - Threads discussing underrated content
- `quality_discussion` - General quality recommendations

**Quality Boost Keywords:**
```go
qualityKeywords = []string{
    "masterpiece", "underrated", "hidden gem", "criminally underrated",
    "best I've ever seen", "life-changing", "perfect", "flawless"
}
```
Mentions near these keywords get a quality score boost.

**Scrape Flow:**
1. Fetch recent threads from subreddit (via Reddit JSON API)
2. Parse thread titles and bodies for media mentions
3. Extract mentions and surrounding context
4. Update quality/popularity scores for existing media
5. Queue new media for vibe profile generation

**Scheduling:**
```go
scraper.Start(ctx, cfg.ScrapeInterval)  // Default: 1 hour
```
Runs as a background goroutine when `ENABLE_SCRAPER=true`.

### 7. Database Schema (`internal/database/database.go`)

```sql
-- Core media table
CREATE TABLE media (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    media_type TEXT NOT NULL,  -- 'movie', 'tv', 'anime'
    year INTEGER,
    plot_summary TEXT,
    vibe_profile TEXT,
    quality_score REAL DEFAULT 0,
    popularity_score REAL DEFAULT 0,
    source_subreddit TEXT,
    external_id TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User watch history
CREATE TABLE seen_media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT NOT NULL,
    media_id TEXT NOT NULL,
    rating INTEGER,
    watched_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_id) REFERENCES media(id),
    UNIQUE(user_id, media_id)
);

-- Vector embeddings for similarity search
CREATE TABLE vibe_embeddings (
    media_id TEXT PRIMARY KEY,
    embedding BLOB NOT NULL,  -- Serialized []float32
    model TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (media_id) REFERENCES media(id)
);

-- Reddit thread data for quality scoring
CREATE TABLE reddit_threads (
    id TEXT PRIMARY KEY,
    subreddit TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    thread_type TEXT,  -- 'similar_to', 'hidden_gem', 'quality_discussion'
    reference_show TEXT,
    score INTEGER,
    num_comments INTEGER,
    scraped_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Individual media mentions in threads
CREATE TABLE reddit_mentions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    thread_id TEXT NOT NULL,
    media_id TEXT NOT NULL,
    mention_context TEXT,
    quality_boost REAL DEFAULT 0,
    FOREIGN KEY (thread_id) REFERENCES reddit_threads(id),
    FOREIGN KEY (media_id) REFERENCES media(id)
);

-- User accounts (simple)
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    username TEXT UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 8. API Endpoints (`main.go`)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| **Seen Media** |
| POST | `/api/seen` | Mark media as watched |
| GET | `/api/seen` | Get user's watch history |
| DELETE | `/api/seen` | Remove from watch history |
| **Recommendations** |
| POST | `/api/recommend` | Full vibe search with reranking |
| GET | `/api/vibe?q=...` | Quick vibe search (no reranking) |
| GET | `/api/similar/:media_id` | Find similar to specific media |
| GET | `/api/hidden-gems` | High-quality low-popularity media |
| **Media Management** |
| POST | `/api/media` | Add new media (generates vibe profile) |
| GET | `/api/media/:id` | Get media details |
| POST | `/api/media/:id/refresh` | Regenerate vibe profile |
| **Admin** |
| GET | `/api/stats` | System statistics |
| POST | `/api/admin/scrape` | Trigger manual Reddit scrape |

**Request/Response Examples:**

```bash
# Get recommendations
curl -X POST http://localhost:8080/api/recommend \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "query": "melancholic slow-burn drama with beautiful cinematography", "limit": 5}'

# Response
{
  "recommendations": [
    {
      "media": {
        "id": "movie-In-the-Mood-for-Love",
        "title": "In the Mood for Love",
        "media_type": "movie",
        "year": 2000,
        "vibe_profile": "Achingly romantic and melancholic..."
      },
      "vibe_score": 0.89,
      "explanation": "This matches your request for melancholic slow-burn...",
      "rank": 1
    }
  ],
  "query": "melancholic slow-burn drama with beautiful cinematography"
}
```

### 9. Frontend (`frontend/src/`)

**Main Components:**

- **SearchBar** - Vibe prompt input with placeholder suggestions
- **RecommendationCard** - Displays media with score, explanation, "mark as seen" button
- **WatchHistory** - Shows all media user has marked as seen
- **QuickActions** - Buttons for "Hidden Gems", "Random Vibe", "History"
- **ParticleBackground** - Animated particles using canvas

**State Management:**
- `user_id` stored in localStorage (anonymous users)
- `seenIDs` cached locally for instant UI updates
- API calls via custom `useApi` hook

**Styling:**
- Tailwind CSS for utility classes
- Dark theme with neon accent colors
- Framer Motion for smooth animations

---

## Running Locally

### Prerequisites
- Go 1.21+
- Node.js 18+
- OpenAI API key (optional, for full functionality)

### Development

```bash
cd w2w

# Install frontend dependencies
cd frontend && npm install && cd ..

# Set environment
export OPENAI_API_KEY="sk-..."
export DATABASE_PATH="./vibe.db"

# Run Go server (serves API + built frontend)
go run main.go
```

Without OpenAI key, uses placeholder embeddings (non-semantic, for testing pipeline only).

### Seed Database

```bash
go run cmd/seed/main.go
```

Adds sample media with pre-written vibe profiles.

### Docker

```bash
# From project root
docker compose up --build movie-app
```

---

## Ways to Push Forward

### AI & Recommendations
1. **Better embeddings:**
   - Try `text-embedding-3-large` for higher accuracy
   - Fine-tune embeddings on media descriptions
   - Hybrid approach: combine vibe embeddings with genre/year filters

2. **Improved vibe profiles:**
   - Include visual stills analysis (multimodal)
   - Incorporate trailer audio analysis for soundtrack vibes
   - User-contributed vibe tags

3. **Personalization:**
   - Learn from user's seen/rated history
   - "More like this" button that improves over time
   - Negative feedback ("not what I meant")

4. **Explanation quality:**
   - More specific explanations referencing exact vibe elements
   - Highlight which words in query matched which vibe aspects

### Data Acquisition
5. **Expand Reddit scraping:**
   - More subreddits (r/horror, r/kdramarecommends, r/documentaries)
   - Parse "What did you watch this week" threads
   - Track upvotes over time for quality signals

6. **External APIs:**
   - TMDB/IMDB for metadata (posters, cast, official descriptions)
   - JustWatch for streaming availability
   - Trakt.tv for community ratings

7. **User-generated content:**
   - Allow users to add media
   - Let users write/edit vibe profiles
   - Voting on vibe accuracy

### User Experience
8. **Search refinements:**
   - "Not like this" exclusion filters
   - Decade/era filter
   - Minimum rating threshold
   - Streaming service filter

9. **Social features:**
   - Share recommendation lists
   - "Vibe playlists" (curated collections)
   - Friend activity feed

10. **Better discovery:**
    - Daily/weekly vibe challenges
    - "Surprise me" with explanation
    - Mood-based landing page (sad, energetic, thoughtful)

### Technical Improvements
11. **Vector database:**
    - Replace in-memory store with pgvector or Pinecone
    - Support larger media catalogs (10K+)
    - Approximate nearest neighbor for speed

12. **Caching:**
    - Cache popular queries
    - Pre-compute embeddings for common vibe phrases
    - Redis for session data

13. **Background processing:**
    - Job queue for vibe profile generation
    - Batch embedding updates
    - Scheduled Reddit scraping with backoff

14. **Testing:**
    - Unit tests for vector search
    - Integration tests for recommendation pipeline
    - A/B testing framework for ranking experiments

15. **Monitoring:**
    - Track recommendation click-through rates
    - Monitor embedding model costs
    - Log query patterns for analysis

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DATABASE_PATH` | `./vibe.db` | SQLite database file |
| `OPENAI_API_KEY` | - | Required for full functionality |
| `ENABLE_SCRAPER` | `false` | Enable background Reddit scraping |
| `SCRAPE_INTERVAL` | `1h` | How often to scrape Reddit |

---

## Dependencies

**Go:**
- `github.com/gin-gonic/gin` - Web framework
- `github.com/mattn/go-sqlite3` - SQLite driver (CGO required)
- `github.com/joho/godotenv` - Environment file loading
- `github.com/sashabaranov/go-openai` - OpenAI client

**Frontend:**
- React 18 + Vite
- Tailwind CSS
- Framer Motion (animations)
- Lucide React (icons)

---

## Security Notes

- The `.env` file in the repo contains an API key - this should be in `.gitignore` and loaded from environment only
- User IDs are anonymous (localStorage-generated UUIDs)
- No authentication system - all users are anonymous
- Security headers set by Caddy: `X-Frame-Options: DENY`, `X-Content-Type-Options: nosniff`
