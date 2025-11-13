package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

// PineconeService handles Pinecone vector operations
type PineconeService struct{}

// NewPineconeService creates a new Pinecone service
func NewPineconeService() *PineconeService {
	return &PineconeService{}
}

// Vector represents a Pinecone vector
type Vector struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata"`
}

// UpsertRequest represents a Pinecone upsert request
type UpsertRequest struct {
	Vectors   []Vector `json:"vectors"`
	Namespace string   `json:"namespace"`
}

// QueryRequest represents a Pinecone query request
type QueryRequest struct {
	Vector          []float32 `json:"vector"`
	TopK            int       `json:"topK"`
	Namespace       string    `json:"namespace"`
	IncludeMetadata bool      `json:"includeMetadata"`
}

// QueryResponse represents a Pinecone query response
type QueryResponse struct {
	Matches []struct {
		ID       string                 `json:"id"`
		Score    float32                `json:"score"`
		Metadata map[string]interface{} `json:"metadata"`
	} `json:"matches"`
}

// UpsertVectors uploads vectors to Pinecone
func (s *PineconeService) UpsertVectors(indexHost, apiKey, namespace string, vectors []Vector) error {
	if indexHost == "" || apiKey == "" {
		return errors.New("Pinecone credentials not provided")
	}

	reqBody := UpsertRequest{
		Vectors:   vectors,
		Namespace: namespace,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/vectors/upsert", indexHost)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Pinecone API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Pinecone API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// QueryVectors searches for similar vectors in Pinecone
func (s *PineconeService) QueryVectors(indexHost, apiKey, namespace string, vector []float32, topK int) (*QueryResponse, error) {
	if indexHost == "" || apiKey == "" {
		return nil, errors.New("Pinecone credentials not provided")
	}

	reqBody := QueryRequest{
		Vector:          vector,
		TopK:            topK,
		Namespace:       namespace,
		IncludeMetadata: true,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://%s/query", indexHost)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Pinecone API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Pinecone API error (status %d): %s", resp.StatusCode, string(body))
	}

	var queryResp QueryResponse
	if err := json.Unmarshal(body, &queryResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &queryResp, nil
}

// DeleteVectors deletes vectors from Pinecone
func (s *PineconeService) DeleteVectors(indexHost, apiKey, namespace string, ids []string) error {
	if indexHost == "" || apiKey == "" {
		return errors.New("Pinecone credentials not provided")
	}

	reqBody := map[string]interface{}{
		"ids":       ids,
		"namespace": namespace,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/vectors/delete", indexHost)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call Pinecone API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Pinecone API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// CreateIndex creates a new Pinecone index (management API)
// Note: This requires the management API, not the data plane API
func (s *PineconeService) CreateIndex(apiKey, indexName string, dimension int) error {
	// TODO: Implement Pinecone management API call
	// This requires calling the Pinecone management API (different from data API)
	return errors.New("index creation not yet implemented - create index manually in Pinecone console")
}

// GenerateVectorID generates a unique vector ID
func GenerateVectorID(prefix string) string {
	return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
}
