package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// EmbeddingService handles OpenAI embeddings
type EmbeddingService struct{}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService() *EmbeddingService {
	return &EmbeddingService{}
}

type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
}

// GenerateEmbedding generates an embedding for the given text using OpenAI
func (s *EmbeddingService) GenerateEmbedding(text, apiKey string) ([]float32, error) {
	if text == "" {
		return nil, errors.New("text cannot be empty")
	}
	if apiKey == "" {
		return nil, errors.New("OpenAI API key not provided")
	}

	// Prepare request
	reqBody := embeddingRequest{
		Input: text,
		Model: "text-embedding-ada-002",
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	// Make API call
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 30 * 1000000000} // 30 seconds
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, errors.New("no embedding returned from API")
	}

	return embResp.Data[0].Embedding, nil
}

// GenerateBatchEmbeddings generates embeddings for multiple texts
func (s *EmbeddingService) GenerateBatchEmbeddings(texts []string, apiKey string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	// TODO: Implement batch API call for efficiency
	// For now, call individually
	for i, text := range texts {
		emb, err := s.GenerateEmbedding(text, apiKey)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = emb
	}

	return embeddings, nil
}
