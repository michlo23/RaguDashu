package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"github.com/michlo23/rag-dashboard/internal/utils"
	"gorm.io/gorm"
)

// MCPService handles communication with MCP servers
type MCPService struct {
	db                *gorm.DB
	oauth2Service     *OAuth2Service
	encryptionService *EncryptionService
}

// NewMCPService creates a new MCP service
func NewMCPService(
	db *gorm.DB,
	oauth2Service *OAuth2Service,
	encryptionService *EncryptionService,
) *MCPService {
	return &MCPService{
		db:                db,
		oauth2Service:     oauth2Service,
		encryptionService: encryptionService,
	}
}

// CreateMCPServer creates a new MCP server configuration
func (s *MCPService) CreateMCPServer(profileID uuid.UUID, req *CreateMCPServerRequest) (*models.MCPServer, error) {
	server := &models.MCPServer{
		ProfileID:   profileID,
		Name:        req.Name,
		Description: req.Description,
		ServerURL:   req.ServerURL,
		AuthType:    req.AuthType,
		IsActive:    true,
	}

	// Handle different auth types
	switch req.AuthType {
	case models.MCPAuthTypeAPIKey:
		if req.APIKey == "" {
			return nil, fmt.Errorf("API key is required for api_key auth type")
		}
		encrypted, err := s.encryptionService.Encrypt(req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		server.EncryptedAPIKey = encrypted

	case models.MCPAuthTypeOAuth2:
		if req.OAuth2Config == nil {
			return nil, fmt.Errorf("OAuth 2 config is required for oauth2 auth type")
		}
		oauth2JSON, err := json.Marshal(req.OAuth2Config)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal OAuth 2 config: %w", err)
		}
		server.OAuth2Config = oauth2JSON
	}

	if err := s.db.Create(server).Error; err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}

	// Sync capabilities if auth is not OAuth 2 (OAuth 2 requires authorization first)
	if req.AuthType != models.MCPAuthTypeOAuth2 {
		if err := s.SyncCapabilities(server.ID, profileID); err != nil {
			// Log error but don't fail creation
			errorMsg := utils.SanitizeErrorMessage(err)
			server.TestResult = &errorMsg
			s.db.Save(server)
		}
	}

	return server, nil
}

// UpdateMCPServer updates an MCP server configuration
func (s *MCPService) UpdateMCPServer(serverID, profileID uuid.UUID, req *UpdateMCPServerRequest) (*models.MCPServer, error) {
	var server models.MCPServer
	if err := s.db.Where("id = ? AND profile_id = ?", serverID, profileID).First(&server).Error; err != nil {
		return nil, fmt.Errorf("MCP server not found")
	}

	// Update fields
	if req.Name != "" {
		server.Name = req.Name
	}
	if req.Description != "" {
		server.Description = req.Description
	}
	if req.IsActive != nil {
		server.IsActive = *req.IsActive
	}

	// Update API key if provided
	if req.APIKey != "" && server.AuthType == models.MCPAuthTypeAPIKey {
		encrypted, err := s.encryptionService.Encrypt(req.APIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt API key: %w", err)
		}
		server.EncryptedAPIKey = encrypted
	}

	if err := s.db.Save(&server).Error; err != nil {
		return nil, fmt.Errorf("failed to update MCP server: %w", err)
	}

	return &server, nil
}

// DeleteMCPServer deletes an MCP server
func (s *MCPService) DeleteMCPServer(serverID, profileID uuid.UUID) error {
	result := s.db.Where("id = ? AND profile_id = ?", serverID, profileID).Delete(&models.MCPServer{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete MCP server: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("MCP server not found")
	}
	return nil
}

// ListMCPServers lists all MCP servers for a profile
func (s *MCPService) ListMCPServers(profileID uuid.UUID) ([]models.MCPServer, error) {
	var servers []models.MCPServer
	err := s.db.Where("profile_id = ?", profileID).Order("created_at DESC").Find(&servers).Error
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// GetMCPServer gets an MCP server by ID
func (s *MCPService) GetMCPServer(serverID, profileID uuid.UUID) (*models.MCPServer, error) {
	var server models.MCPServer
	if err := s.db.Where("id = ? AND profile_id = ?", serverID, profileID).First(&server).Error; err != nil {
		return nil, fmt.Errorf("MCP server not found")
	}
	return &server, nil
}

// SyncCapabilities fetches and syncs capabilities from MCP server
func (s *MCPService) SyncCapabilities(serverID, profileID uuid.UUID) error {
	server, err := s.GetMCPServer(serverID, profileID)
	if err != nil {
		return err
	}

	// Fetch capabilities
	capabilities, err := s.fetchCapabilities(server)
	if err != nil {
		errorMsg := utils.SanitizeErrorMessage(err)
		now := time.Now()
		server.LastTestedAt = &now
		server.TestResult = &errorMsg
		s.db.Save(server)
		return fmt.Errorf("failed to fetch capabilities: %w", err)
	}

	// Update server with capabilities
	toolsJSON, _ := json.Marshal(capabilities.Tools)
	resourcesJSON, _ := json.Marshal(capabilities.Resources)
	promptsJSON, _ := json.Marshal(capabilities.Prompts)

	now := time.Now()
	successMsg := "Success"
	server.AvailableTools = toolsJSON
	server.AvailableResources = resourcesJSON
	server.AvailablePrompts = promptsJSON
	server.LastSyncedAt = &now
	server.LastTestedAt = &now
	server.TestResult = &successMsg

	if err := s.db.Save(server).Error; err != nil {
		return fmt.Errorf("failed to save capabilities: %w", err)
	}

	return nil
}

// CallTool calls an MCP tool
func (s *MCPService) CallTool(
	serverID, profileID uuid.UUID,
	toolName string,
	input map[string]interface{},
) (*MCPToolResponse, error) {
	server, err := s.GetMCPServer(serverID, profileID)
	if err != nil {
		return nil, err
	}

	startTime := time.Now()

	// Make MCP tool call request
	response, err := s.makeToolCall(server, toolName, input)
	if err != nil {
		return nil, fmt.Errorf("tool call failed: %w", err)
	}

	duration := int(time.Since(startTime).Milliseconds())
	response.DurationMs = duration

	// Update last used
	now := time.Now()
	server.LastUsedAt = &now
	s.db.Save(server)

	return response, nil
}

// MCPCapabilities represents MCP server capabilities
type MCPCapabilities struct {
	Tools     []models.MCPTool     `json:"tools"`
	Resources []models.MCPResource `json:"resources"`
	Prompts   []models.MCPPrompt   `json:"prompts"`
}

// MCPToolResponse represents a tool call response
type MCPToolResponse struct {
	Content    interface{} `json:"content"`
	IsError    bool        `json:"isError"`
	DurationMs int         `json:"durationMs"`
}

// fetchCapabilities fetches capabilities from MCP server
func (s *MCPService) fetchCapabilities(server *models.MCPServer) (*MCPCapabilities, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params":  map[string]interface{}{},
	}

	response, err := s.makeRequest(server, req)
	if err != nil {
		return nil, err
	}

	// Parse response
	var capabilities MCPCapabilities
	if result, ok := response["result"].(map[string]interface{}); ok {
		if tools, ok := result["tools"].([]interface{}); ok {
			toolsJSON, _ := json.Marshal(tools)
			json.Unmarshal(toolsJSON, &capabilities.Tools)
		}
		if resources, ok := result["resources"].([]interface{}); ok {
			resourcesJSON, _ := json.Marshal(resources)
			json.Unmarshal(resourcesJSON, &capabilities.Resources)
		}
		if prompts, ok := result["prompts"].([]interface{}); ok {
			promptsJSON, _ := json.Marshal(prompts)
			json.Unmarshal(promptsJSON, &capabilities.Prompts)
		}
	}

	return &capabilities, nil
}

// makeToolCall makes a tool call to MCP server
func (s *MCPService) makeToolCall(
	server *models.MCPServer,
	toolName string,
	input map[string]interface{},
) (*MCPToolResponse, error) {
	req := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": input,
		},
	}

	response, err := s.makeRequest(server, req)
	if err != nil {
		return &MCPToolResponse{
			Content: err.Error(),
			IsError: true,
		}, err
	}

	// Check for JSON-RPC error
	if errObj, ok := response["error"].(map[string]interface{}); ok {
		return &MCPToolResponse{
			Content: errObj,
			IsError: true,
		}, fmt.Errorf("tool call error: %v", errObj)
	}

	// Extract result
	result := response["result"]

	return &MCPToolResponse{
		Content: result,
		IsError: false,
	}, nil
}

// makeRequest makes an HTTP request to MCP server
func (s *MCPService) makeRequest(server *models.MCPServer, payload interface{}) (map[string]interface{}, error) {
	// Prepare request body
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", server.ServerURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add authentication
	if err := s.addAuthentication(server, httpReq); err != nil {
		return nil, fmt.Errorf("failed to add authentication: %w", err)
	}

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if httpResp.StatusCode < 200 || httpResp.StatusCode >= 300 {
		return nil, fmt.Errorf("server returned status %d: %s", httpResp.StatusCode, string(respBody))
	}

	// Parse JSON-RPC response
	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response, nil
}

// addAuthentication adds authentication headers to request
func (s *MCPService) addAuthentication(server *models.MCPServer, req *http.Request) error {
	switch server.AuthType {
	case models.MCPAuthTypeAPIKey:
		if server.EncryptedAPIKey == "" {
			return fmt.Errorf("API key not configured")
		}
		apiKey, err := s.encryptionService.Decrypt(server.EncryptedAPIKey)
		if err != nil {
			return fmt.Errorf("failed to decrypt API key: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+apiKey)

	case models.MCPAuthTypeOAuth2:
		accessToken, err := s.oauth2Service.GetValidAccessToken(server.ID, server.ProfileID)
		if err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

	case models.MCPAuthTypeNone:
		// No authentication needed
	}

	return nil
}

// Request/Response types
type CreateMCPServerRequest struct {
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	ServerURL    string                `json:"server_url"`
	AuthType     string                `json:"auth_type"`
	APIKey       string                `json:"api_key,omitempty"`
	OAuth2Config *models.OAuth2Config `json:"oauth2_config,omitempty"`
}

type UpdateMCPServerRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	APIKey      string `json:"api_key,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}
