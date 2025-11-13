package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ValidationService validates API keys by making real API calls
type ValidationService struct{}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidateOpenAIKey validates an OpenAI API key
func (s *ValidationService) ValidateOpenAIKey(apiKey string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return false, "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "Failed to connect to OpenAI API", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return true, "OpenAI API key is valid", nil
	case http.StatusUnauthorized:
		return false, "Invalid OpenAI API key", nil
	case http.StatusTooManyRequests:
		return false, "Rate limited. Key might be valid but cannot verify now", nil
	default:
		return false, fmt.Sprintf("OpenAI API returned status %d: %s", resp.StatusCode, string(body)), nil
	}
}

// ValidatePineconeKey validates a Pinecone API key
func (s *ValidationService) ValidatePineconeKey(apiKey, environment string) (bool, string, error) {
	if environment == "" {
		environment = "us-east1-gcp" // Default
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to list indexes
	url := fmt.Sprintf("https://controller.%s.pinecone.io/databases", environment)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, "", err
	}

	req.Header.Set("Api-Key", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "Failed to connect to Pinecone API", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	switch resp.StatusCode {
	case http.StatusOK:
		return true, "Pinecone API key is valid", nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return false, "Invalid Pinecone API key", nil
	default:
		return false, fmt.Sprintf("Pinecone API returned status %d: %s", resp.StatusCode, string(body)), nil
	}
}

// ValidateSlackToken validates a Slack OAuth token
func (s *ValidationService) ValidateSlackToken(token string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://slack.com/api/auth.test", nil)
	if err != nil {
		return false, "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "Failed to connect to Slack API", err
	}
	defer resp.Body.Close()

	var slackResp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
		User  string `json:"user"`
		Team  string `json:"team"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&slackResp); err != nil {
		return false, "Failed to parse Slack response", err
	}

	if slackResp.OK {
		return true, fmt.Sprintf("Slack token is valid for user %s in team %s", slackResp.User, slackResp.Team), nil
	}

	return false, fmt.Sprintf("Invalid Slack token: %s", slackResp.Error), nil
}

// TestEmbeddingAPI tests the OpenAI embedding API
func (s *ValidationService) TestEmbeddingAPI(apiKey string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reqBody := map[string]interface{}{
		"input": "test",
		"model": "text-embedding-ada-002",
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, "Failed to test embedding API", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, "Embedding API is working", nil
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Sprintf("Embedding API test failed: %s", string(body)), nil
}
