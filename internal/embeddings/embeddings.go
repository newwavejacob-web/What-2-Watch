package embeddings

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sort"
	"time"
)

// Provider defines the interface for embedding generation
type Provider interface {
	Embed(text string) ([]float32, error)
	ModelName() string
}

// OpenAIProvider uses OpenAI's embedding API
type OpenAIProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewOpenAIProvider creates a new OpenAI embedding provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  "text-embedding-3-small",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// openAIEmbeddingRequest is the request body for OpenAI embeddings API
type openAIEmbeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

// openAIEmbeddingResponse is the response from OpenAI embeddings API
type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// Embed generates an embedding for the given text using OpenAI
func (p *OpenAIProvider) Embed(text string) ([]float32, error) {
	reqBody := openAIEmbeddingRequest{
		Input: text,
		Model: p.model,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var embResp openAIEmbeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if embResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", embResp.Error.Message)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	return embResp.Data[0].Embedding, nil
}

// ModelName returns the name of the model being used
func (p *OpenAIProvider) ModelName() string {
	return p.model
}

// ============================================================================
// In-Memory Vector Search (for when a full vector DB is too heavy)
// ============================================================================

// VectorStore provides in-memory vector similarity search
type VectorStore struct {
	vectors map[string][]float32 // mediaID -> embedding
}

// NewVectorStore creates an empty vector store
func NewVectorStore() *VectorStore {
	return &VectorStore{
		vectors: make(map[string][]float32),
	}
}

// LoadFromMap populates the store from a map of embeddings
func (vs *VectorStore) LoadFromMap(embeddings map[string][]float32) {
	vs.vectors = embeddings
}

// Add stores an embedding for a media ID
func (vs *VectorStore) Add(mediaID string, embedding []float32) {
	vs.vectors[mediaID] = embedding
}

// Remove deletes an embedding
func (vs *VectorStore) Remove(mediaID string) {
	delete(vs.vectors, mediaID)
}

// Size returns the number of vectors in the store
func (vs *VectorStore) Size() int {
	return len(vs.vectors)
}

// SearchResult represents a single search result with similarity score
type SearchResult struct {
	MediaID    string
	Similarity float64
}

// Search finds the top-k most similar vectors to the query
// excludeIDs allows filtering out specific media (for anti-join of seen items)
func (vs *VectorStore) Search(query []float32, topK int, excludeIDs map[string]bool) []SearchResult {
	if len(vs.vectors) == 0 {
		return nil
	}

	var results []SearchResult

	for mediaID, embedding := range vs.vectors {
		// Skip excluded IDs (anti-join for seen media)
		if excludeIDs != nil && excludeIDs[mediaID] {
			continue
		}

		similarity := CosineSimilarity(query, embedding)
		results = append(results, SearchResult{
			MediaID:    mediaID,
			Similarity: similarity,
		})
	}

	// Sort by similarity descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	// Return top K
	if len(results) > topK {
		results = results[:topK]
	}

	return results
}

// CosineSimilarity computes the cosine similarity between two vectors
func CosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// EuclideanDistance computes the Euclidean distance between two vectors
func EuclideanDistance(a, b []float32) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	var sum float64
	for i := range a {
		diff := float64(a[i]) - float64(b[i])
		sum += diff * diff
	}

	return math.Sqrt(sum)
}
