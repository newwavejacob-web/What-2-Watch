package llm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"w2w/internal/models"
)

// Client handles LLM API calls for vibe profile generation and reranking
type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new LLM client (defaults to OpenAI)
func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:  apiKey,
		model:   "gpt-4o-mini", // Cost-effective for our use case
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// NewClientWithModel creates a client with a specific model
func NewClientWithModel(apiKey, model string) *Client {
	c := NewClient(apiKey)
	c.model = model
	return c
}

// chatMessage represents a message in the chat format
type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// chatRequest is the request body for chat completions
type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// chatResponse is the response from chat completions
type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// complete sends a chat completion request
func (c *Client) complete(systemPrompt, userPrompt string, temperature float64) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
		MaxTokens:   1500,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GenerateVibeProfile creates a vibe profile for a media entry
// This is the core "style over substance" description
func (c *Client) GenerateVibeProfile(title, mediaType string, year int, synopsis string) (string, error) {
	systemPrompt := `You are a film/TV critic who specializes in describing the AESTHETIC and FEELING of media,
not the plot. You focus on style, pacing, visual language, emotional texture, and "vibe."

Your descriptions should be evocative and specific, using terms like:
- Visual style: "neon-noir", "pastel dreamscape", "gritty realism", "hyperkinetic animation"
- Pacing: "meditative slowburn", "frenetic energy", "deliberate tension"
- Emotional texture: "existential dread", "cozy melancholy", "manic joy", "contemplative silence"
- Atmosphere: "rain-soaked streets", "sun-drenched nostalgia", "clinical coldness"

DO NOT summarize the plot. Focus ONLY on how it FEELS to watch.
Keep the response to 2-3 sentences maximum.`

	userPrompt := fmt.Sprintf(`Describe the aesthetic, pacing, and emotional "vibe" of %s (%d) [%s].
%s

Remember: Focus on STYLE, not story. How does it FEEL to watch?`,
		title, year, mediaType, synopsis)

	return c.complete(systemPrompt, userPrompt, 0.7)
}

// RerankCandidate represents a candidate for reranking
type RerankCandidate struct {
	Media     models.Media
	VibeScore float64 // Initial similarity score
}

// RerankResult represents a reranked recommendation
type RerankResult struct {
	MediaID     string
	Rank        int
	Explanation string
}

// RerankByVibe uses an LLM to rerank candidates based on vibe match
// This implements the "Curator Agent" logic
func (c *Client) RerankByVibe(query string, candidates []RerankCandidate) ([]RerankResult, error) {
	if len(candidates) == 0 {
		return nil, nil
	}

	systemPrompt := `You are a recommendation curator who understands VIBES, not just genres.
When a user asks for something "like X but focused on Y," you understand the FEELING they're chasing.

Your job is to rank candidates based on how well they capture the specific VIBE the user wants.
Genre similarity is secondary to emotional/aesthetic similarity.

Respond in this exact JSON format:
{
  "rankings": [
    {"media_id": "...", "rank": 1, "explanation": "..."},
    {"media_id": "...", "rank": 2, "explanation": "..."},
    {"media_id": "...", "rank": 3, "explanation": "..."}
  ]
}

Only include the top 3 best matches. Be specific in explanations about WHY each matches the vibe.`

	// Build the candidate list for the prompt
	var candidateList strings.Builder
	for i, c := range candidates {
		candidateList.WriteString(fmt.Sprintf(
			"%d. [ID: %s] %s (%d) - Vibe: %s\n",
			i+1, c.Media.ID, c.Media.Title, c.Media.Year, c.Media.VibeProfile,
		))
	}

	userPrompt := fmt.Sprintf(`User's vibe request: "%s"

Candidates to rank (with their vibe profiles):
%s

Rank the TOP 3 that best capture the user's requested vibe. Explain why each matches.`,
		query, candidateList.String())

	response, err := c.complete(systemPrompt, userPrompt, 0.3)
	if err != nil {
		return nil, fmt.Errorf("rerank request failed: %w", err)
	}

	// Parse the JSON response
	// First, try to extract JSON from the response (it might be wrapped in markdown)
	jsonStr := response
	if idx := strings.Index(response, "{"); idx != -1 {
		jsonStr = response[idx:]
		if endIdx := strings.LastIndex(jsonStr, "}"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx+1]
		}
	}

	var result struct {
		Rankings []struct {
			MediaID     string `json:"media_id"`
			Rank        int    `json:"rank"`
			Explanation string `json:"explanation"`
		} `json:"rankings"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		// If JSON parsing fails, return candidates in original order with generic explanations
		var fallback []RerankResult
		for i, c := range candidates {
			if i >= 3 {
				break
			}
			fallback = append(fallback, RerankResult{
				MediaID:     c.Media.ID,
				Rank:        i + 1,
				Explanation: fmt.Sprintf("Matches your vibe based on: %s", c.Media.VibeProfile),
			})
		}
		return fallback, nil
	}

	var results []RerankResult
	for _, r := range result.Rankings {
		results = append(results, RerankResult{
			MediaID:     r.MediaID,
			Rank:        r.Rank,
			Explanation: r.Explanation,
		})
	}

	return results, nil
}

// ClassifyThreadType analyzes a Reddit thread title to determine its type
func (c *Client) ClassifyThreadType(title, body string) (string, string, error) {
	systemPrompt := `You analyze Reddit recommendation thread titles and bodies.
Classify each thread and extract the reference show if mentioned.

Respond in JSON format:
{
  "thread_type": "similar_to|hidden_gem|quality_discussion|other",
  "reference_show": "Show Name or null"
}

Thread types:
- similar_to: Asking for shows similar to a specific title (e.g., "Shows like Breaking Bad")
- hidden_gem: Asking for underrated/unknown recommendations (e.g., "Hidden gems", "Underrated anime")
- quality_discussion: Discussing quality aspects (e.g., "Best written shows", "Unique art styles")
- other: General recommendations or doesn't fit above`

	userPrompt := fmt.Sprintf("Title: %s\nBody: %s", title, body)

	response, err := c.complete(systemPrompt, userPrompt, 0.1)
	if err != nil {
		return "other", "", err
	}

	// Parse JSON
	jsonStr := response
	if idx := strings.Index(response, "{"); idx != -1 {
		jsonStr = response[idx:]
		if endIdx := strings.LastIndex(jsonStr, "}"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx+1]
		}
	}

	var result struct {
		ThreadType    string  `json:"thread_type"`
		ReferenceShow *string `json:"reference_show"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return "other", "", nil
	}

	refShow := ""
	if result.ReferenceShow != nil {
		refShow = *result.ReferenceShow
	}

	return result.ThreadType, refShow, nil
}

// ExtractMentions extracts show/movie mentions from text
func (c *Client) ExtractMentions(text string) ([]string, error) {
	systemPrompt := `Extract all movie, TV show, and anime titles mentioned in the text.
Return ONLY a JSON array of title strings. Be precise with titles.
Example: ["Breaking Bad", "Better Call Saul", "Ozark"]`

	response, err := c.complete(systemPrompt, text, 0.1)
	if err != nil {
		return nil, err
	}

	// Parse JSON array
	jsonStr := response
	if idx := strings.Index(response, "["); idx != -1 {
		jsonStr = response[idx:]
		if endIdx := strings.LastIndex(jsonStr, "]"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx+1]
		}
	}

	var titles []string
	if err := json.Unmarshal([]byte(jsonStr), &titles); err != nil {
		return nil, nil // Return empty on parse failure
	}

	return titles, nil
}
