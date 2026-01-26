package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/gin-gonic/gin"
)

// Recommendation represents our final output
type Recommendation struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	VibeReason  string `json:"vibe_reason"`
	MediaType   string `json:"type"` // "Movie", "Book", or "TV"
}

func main() {
	r := gin.Default()

	r.GET("/vibe", func(c *gin.Context) {
		userQuery := c.Query("q") // Example: "cyberpunk but cozy"

		if userQuery == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tell me your vibe!"})
			return
		}

		// LOGIC STEP: In a real app, you'd send 'userQuery' to an LLM 
		// to extract "Semantic Keywords". For now, let's simulate the engine:
		results := vibeEngine(userQuery)

		c.JSON(http.StatusOK, gin.H{
			"input":           userQuery,
			"recommendations": results,
		})
	})

	fmt.Println("Vibe server starting on :8080")
	r.Run(":8080")
}

// This is where the magic happens
func vibeEngine(query string) []Recommendation {
	// Dummy logic: In production, you'd call TasteDive or TMDB API here
	// based on keywords extracted from the query.
	return []Recommendation{
		{
			Title:       "Spider-Man: Into the Spider-Verse",
			Description: "A teen becomes the Spider-Man of his universe.",
			VibeReason:  "Matches your request for 'Transformers-style spectacle but animated'.",
			MediaType:   "Movie",
		},
		{
			Title:       "Iron Giant",
			Description: "A boy befriends a giant robot from outer space.",
			VibeReason:  "Classic mechanical vibe with deep emotional resonance.",
			MediaType:   "Movie",
		},
	}
}
