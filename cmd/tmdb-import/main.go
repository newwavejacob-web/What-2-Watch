package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"w2w/internal/database"
	"w2w/internal/embeddings"
	"w2w/internal/models"
	"w2w/internal/tmdb"
)

func main() {
	godotenv.Load()

	tmdbKey := os.Getenv("TMDB_API_KEY")
	openaiKey := os.Getenv("OPENAI_API_KEY")
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./vibe.db"
	}

	if tmdbKey == "" {
		log.Fatal("TMDB_API_KEY is required. Get one free at https://www.themoviedb.org/settings/api and add it to .env")
	}
	if openaiKey == "" {
		log.Fatal("OPENAI_API_KEY is required for embedding generation.")
	}

	// Parse targets from args or use defaults
	moviePages := 100  // ~2000 movies (20 per page)
	tvPages := 50      // ~1000 TV shows
	animePages := 50   // ~1000 anime TV shows
	animeMoviePages := 25 // ~500 anime movies
	skipEmbeddings := false

	for _, arg := range os.Args[1:] {
		switch {
		case strings.HasPrefix(arg, "--movies="):
			moviePages, _ = strconv.Atoi(strings.TrimPrefix(arg, "--movies="))
		case strings.HasPrefix(arg, "--tv="):
			tvPages, _ = strconv.Atoi(strings.TrimPrefix(arg, "--tv="))
		case strings.HasPrefix(arg, "--anime="):
			animePages, _ = strconv.Atoi(strings.TrimPrefix(arg, "--anime="))
		case strings.HasPrefix(arg, "--anime-movies="):
			animeMoviePages, _ = strconv.Atoi(strings.TrimPrefix(arg, "--anime-movies="))
		case arg == "--skip-embeddings":
			skipEmbeddings = true
		case arg == "--small":
			moviePages = 10
			tvPages = 5
			animePages = 5
			animeMoviePages = 3
		case arg == "--help":
			fmt.Println("Usage: tmdb-import [flags]")
			fmt.Println("  --movies=N         Number of pages of movies to import (20 per page, default 100)")
			fmt.Println("  --tv=N             Number of pages of TV shows (default 50)")
			fmt.Println("  --anime=N          Number of pages of anime TV (default 50)")
			fmt.Println("  --anime-movies=N   Number of pages of anime movies (default 25)")
			fmt.Println("  --skip-embeddings  Import metadata only, skip embedding generation")
			fmt.Println("  --small            Quick test: 10/5/5/3 pages")
			os.Exit(0)
		}
	}

	fmt.Println("========================================")
	fmt.Println("  TMDB Import for What 2 Watch")
	fmt.Println("========================================")
	fmt.Printf("  Database:      %s\n", dbPath)
	fmt.Printf("  Movies:        %d pages (~%d)\n", moviePages, moviePages*20)
	fmt.Printf("  TV shows:      %d pages (~%d)\n", tvPages, tvPages*20)
	fmt.Printf("  Anime TV:      %d pages (~%d)\n", animePages, animePages*20)
	fmt.Printf("  Anime movies:  %d pages (~%d)\n", animeMoviePages, animeMoviePages*20)
	fmt.Printf("  Embeddings:    %v\n", !skipEmbeddings)
	fmt.Println("========================================")
	fmt.Println()

	// Init database
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatalf("Database error: %v", err)
	}
	defer db.Close()

	// Init TMDB client
	tmdbClient := tmdb.NewClient(tmdbKey)

	// Init embedding provider
	var embedder embeddings.Provider
	if !skipEmbeddings {
		embedder = embeddings.NewOpenAIProvider(openaiKey)
	}

	// Ensure default user exists
	db.CreateUser(&models.User{
		ID:        "default",
		Username:  "default",
		CreatedAt: time.Now(),
	})

	stats := &importStats{}
	startTime := time.Now()

	// Phase 1: Import movies
	fmt.Println("[PHASE 1] Importing movies...")
	importMovies(tmdbClient, db, embedder, moviePages, stats)

	// Phase 2: Import TV shows (non-anime)
	fmt.Println("\n[PHASE 2] Importing TV shows...")
	importTV(tmdbClient, db, embedder, tvPages, false, stats)

	// Phase 3: Import anime TV
	fmt.Println("\n[PHASE 3] Importing anime TV shows...")
	importAnimeTV(tmdbClient, db, embedder, animePages, stats)

	// Phase 4: Import anime movies
	fmt.Println("\n[PHASE 4] Importing anime movies...")
	importAnimeMovies(tmdbClient, db, embedder, animeMoviePages, stats)

	// Phase 5: Generate embeddings for any entries missing them
	if !skipEmbeddings {
		fmt.Println("\n[PHASE 5] Backfilling missing embeddings...")
		backfillEmbeddings(db, embedder, stats)
	}

	elapsed := time.Since(startTime)
	fmt.Println("\n========================================")
	fmt.Println("  Import Complete!")
	fmt.Println("========================================")
	fmt.Printf("  New media added:    %d\n", stats.added)
	fmt.Printf("  Already existed:    %d\n", stats.skipped)
	fmt.Printf("  Embeddings created: %d\n", stats.embedded)
	fmt.Printf("  Errors:             %d\n", stats.errors)
	fmt.Printf("  Duration:           %s\n", elapsed.Round(time.Second))
	fmt.Println("========================================")

	// Print total counts
	var totalMedia, totalEmbed int
	db.QueryRow("SELECT COUNT(*) FROM media").Scan(&totalMedia)
	db.QueryRow("SELECT COUNT(*) FROM vibe_embeddings").Scan(&totalEmbed)
	fmt.Printf("\n  Total media in DB:     %d\n", totalMedia)
	fmt.Printf("  Total with embeddings: %d\n", totalEmbed)
}

type importStats struct {
	added    int
	skipped  int
	embedded int
	errors   int
}

func importMovies(client *tmdb.Client, db *database.DB, embedder embeddings.Provider, pages int, stats *importStats) {
	for page := 1; page <= pages; page++ {
		disc, err := client.DiscoverMovies(page)
		if err != nil {
			fmt.Printf("  [page %d] discover error: %v\n", page, err)
			stats.errors++
			continue
		}

		for _, entry := range disc.Results {
			mediaID := fmt.Sprintf("tmdb-movie-%d", entry.ID)

			// Check if already exists
			existing, _ := db.GetMedia(mediaID)
			if existing != nil {
				stats.skipped++
				continue
			}

			// Skip entries without overview (can't build good vibe text)
			if entry.Overview == "" {
				continue
			}

			// Get full details
			details, err := client.GetMovieDetails(entry.ID)
			if err != nil {
				stats.errors++
				continue
			}

			// Determine media type
			mediaType := "movie"
			if tmdb.IsAnimeMovie(details) {
				mediaType = "anime"
			}

			// Build vibe text
			vibeText := tmdb.BuildMovieVibeText(details)

			// Extract year
			year := 0
			if y := extractYear(details.ReleaseDate); y > 0 {
				year = y
			}

			// Quality & popularity
			qualityScore := details.VoteAverage / 10.0
			popularityScore := normalizePopularity(details.Popularity)

			media := &models.Media{
				ID:              mediaID,
				Title:           details.Title,
				MediaType:       mediaType,
				Year:            year,
				PlotSummary:     details.Overview,
				VibeProfile:     vibeText,
				QualityScore:    qualityScore,
				PopularityScore: popularityScore,
				ExternalID:      fmt.Sprintf("tmdb:%d", details.ID),
			}

			if err := db.CreateMedia(media); err != nil {
				stats.errors++
				continue
			}
			stats.added++

			// Generate embedding
			if embedder != nil {
				embedding, err := embedder.Embed(vibeText)
				if err != nil {
					fmt.Printf("  [%d/%d] %s (%d) — embed error: %v\n",
						(page-1)*20+1, pages*20, details.Title, year, err)
					stats.errors++
					continue
				}
				if err := db.StoreEmbedding(mediaID, embedding, embedder.ModelName()); err != nil {
					stats.errors++
					continue
				}
				stats.embedded++
			}

			fmt.Printf("  [p%d] + %s (%d) [%s]\n", page, details.Title, year, mediaType)
		}

		if page%10 == 0 {
			fmt.Printf("  ... %d/%d pages done (%d added, %d skipped)\n", page, pages, stats.added, stats.skipped)
		}
	}
}

func importTV(client *tmdb.Client, db *database.DB, embedder embeddings.Provider, pages int, isAnime bool, stats *importStats) {
	for page := 1; page <= pages; page++ {
		disc, err := client.DiscoverTV(page)
		if err != nil {
			fmt.Printf("  [page %d] discover error: %v\n", page, err)
			stats.errors++
			continue
		}

		for _, entry := range disc.Results {
			mediaID := fmt.Sprintf("tmdb-tv-%d", entry.ID)

			existing, _ := db.GetMedia(mediaID)
			if existing != nil {
				stats.skipped++
				continue
			}

			if entry.Overview == "" {
				continue
			}

			details, err := client.GetTVDetails(entry.ID)
			if err != nil {
				stats.errors++
				continue
			}

			// Skip anime from regular TV import (they get their own phase)
			if !isAnime && details.OriginalLanguage == "ja" {
				hasAnimation := false
				for _, g := range details.Genres {
					if g.ID == 16 {
						hasAnimation = true
						break
					}
				}
				if hasAnimation {
					continue
				}
			}

			mediaType := "tv"
			vibeText := tmdb.BuildTVVibeText(details)

			year := 0
			if y := extractYear(details.FirstAirDate); y > 0 {
				year = y
			}

			qualityScore := details.VoteAverage / 10.0
			popularityScore := normalizePopularity(details.Popularity)

			media := &models.Media{
				ID:              mediaID,
				Title:           details.Name,
				MediaType:       mediaType,
				Year:            year,
				PlotSummary:     details.Overview,
				VibeProfile:     vibeText,
				QualityScore:    qualityScore,
				PopularityScore: popularityScore,
				ExternalID:      fmt.Sprintf("tmdb:%d", details.ID),
			}

			if err := db.CreateMedia(media); err != nil {
				stats.errors++
				continue
			}
			stats.added++

			if embedder != nil {
				embedding, err := embedder.Embed(vibeText)
				if err != nil {
					stats.errors++
					continue
				}
				if err := db.StoreEmbedding(mediaID, embedding, embedder.ModelName()); err != nil {
					stats.errors++
					continue
				}
				stats.embedded++
			}

			fmt.Printf("  [p%d] + %s (%d) [tv]\n", page, details.Name, year)
		}

		if page%10 == 0 {
			fmt.Printf("  ... %d/%d pages done\n", page, pages)
		}
	}
}

func importAnimeTV(client *tmdb.Client, db *database.DB, embedder embeddings.Provider, pages int, stats *importStats) {
	for page := 1; page <= pages; page++ {
		disc, err := client.DiscoverAnime(page)
		if err != nil {
			fmt.Printf("  [page %d] discover error: %v\n", page, err)
			stats.errors++
			continue
		}

		for _, entry := range disc.Results {
			mediaID := fmt.Sprintf("tmdb-tv-%d", entry.ID)

			existing, _ := db.GetMedia(mediaID)
			if existing != nil {
				stats.skipped++
				continue
			}

			if entry.Overview == "" {
				continue
			}

			details, err := client.GetTVDetails(entry.ID)
			if err != nil {
				stats.errors++
				continue
			}

			vibeText := tmdb.BuildTVVibeText(details) // isAnime returns "anime" for type

			year := 0
			if y := extractYear(details.FirstAirDate); y > 0 {
				year = y
			}

			qualityScore := details.VoteAverage / 10.0
			popularityScore := normalizePopularity(details.Popularity)

			media := &models.Media{
				ID:              mediaID,
				Title:           details.Name,
				MediaType:       "anime",
				Year:            year,
				PlotSummary:     details.Overview,
				VibeProfile:     vibeText,
				QualityScore:    qualityScore,
				PopularityScore: popularityScore,
				ExternalID:      fmt.Sprintf("tmdb:%d", details.ID),
			}

			if err := db.CreateMedia(media); err != nil {
				stats.errors++
				continue
			}
			stats.added++

			if embedder != nil {
				embedding, err := embedder.Embed(vibeText)
				if err != nil {
					stats.errors++
					continue
				}
				if err := db.StoreEmbedding(mediaID, embedding, embedder.ModelName()); err != nil {
					stats.errors++
					continue
				}
				stats.embedded++
			}

			fmt.Printf("  [p%d] + %s (%d) [anime]\n", page, details.Name, year)
		}

		if page%10 == 0 {
			fmt.Printf("  ... %d/%d pages done\n", page, pages)
		}
	}
}

func importAnimeMovies(client *tmdb.Client, db *database.DB, embedder embeddings.Provider, pages int, stats *importStats) {
	for page := 1; page <= pages; page++ {
		disc, err := client.DiscoverAnimeMovies(page)
		if err != nil {
			fmt.Printf("  [page %d] discover error: %v\n", page, err)
			stats.errors++
			continue
		}

		for _, entry := range disc.Results {
			mediaID := fmt.Sprintf("tmdb-movie-%d", entry.ID)

			existing, _ := db.GetMedia(mediaID)
			if existing != nil {
				stats.skipped++
				continue
			}

			if entry.Overview == "" {
				continue
			}

			details, err := client.GetMovieDetails(entry.ID)
			if err != nil {
				stats.errors++
				continue
			}

			vibeText := tmdb.BuildMovieVibeText(details)

			year := 0
			if y := extractYear(details.ReleaseDate); y > 0 {
				year = y
			}

			qualityScore := details.VoteAverage / 10.0
			popularityScore := normalizePopularity(details.Popularity)

			media := &models.Media{
				ID:              mediaID,
				Title:           details.Title,
				MediaType:       "anime",
				Year:            year,
				PlotSummary:     details.Overview,
				VibeProfile:     vibeText,
				QualityScore:    qualityScore,
				PopularityScore: popularityScore,
				ExternalID:      fmt.Sprintf("tmdb:%d", details.ID),
			}

			if err := db.CreateMedia(media); err != nil {
				stats.errors++
				continue
			}
			stats.added++

			if embedder != nil {
				embedding, err := embedder.Embed(vibeText)
				if err != nil {
					stats.errors++
					continue
				}
				if err := db.StoreEmbedding(mediaID, embedding, embedder.ModelName()); err != nil {
					stats.errors++
					continue
				}
				stats.embedded++
			}

			fmt.Printf("  [p%d] + %s (%d) [anime movie]\n", page, details.Title, year)
		}

		if page%10 == 0 {
			fmt.Printf("  ... %d/%d pages done\n", page, pages)
		}
	}
}

func backfillEmbeddings(db *database.DB, embedder embeddings.Provider, stats *importStats) {
	// Find media entries without embeddings
	rows, err := db.Query(`
		SELECT m.id, m.vibe_profile
		FROM media m
		LEFT JOIN vibe_embeddings ve ON m.id = ve.media_id
		WHERE ve.media_id IS NULL AND m.vibe_profile != ''
	`)
	if err != nil {
		fmt.Printf("  Backfill query error: %v\n", err)
		return
	}
	defer rows.Close()

	type entry struct {
		id          string
		vibeProfile string
	}
	var missing []entry
	for rows.Next() {
		var e entry
		if err := rows.Scan(&e.id, &e.vibeProfile); err != nil {
			continue
		}
		missing = append(missing, e)
	}

	if len(missing) == 0 {
		fmt.Println("  No missing embeddings found.")
		return
	}

	fmt.Printf("  Found %d entries missing embeddings, generating...\n", len(missing))

	for i, e := range missing {
		embedding, err := embedder.Embed(e.vibeProfile)
		if err != nil {
			fmt.Printf("  [%d/%d] %s — error: %v\n", i+1, len(missing), e.id, err)
			stats.errors++
			continue
		}
		if err := db.StoreEmbedding(e.id, embedding, embedder.ModelName()); err != nil {
			stats.errors++
			continue
		}
		stats.embedded++

		if (i+1)%100 == 0 {
			fmt.Printf("  ... %d/%d embeddings generated\n", i+1, len(missing))
		}
	}
}

func extractYear(dateStr string) int {
	if len(dateStr) >= 4 {
		y, _ := strconv.Atoi(dateStr[:4])
		return y
	}
	return 0
}

func normalizePopularity(pop float64) float64 {
	// TMDB popularity ranges widely (0 to ~1000+).
	// Simple linear normalization capped at 1.0.
	if pop <= 0 {
		return 0
	}
	if pop > 1000 {
		return 1.0
	}
	return pop / 1000.0
}
