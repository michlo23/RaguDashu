package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/api/middleware"
	"github.com/michlo23/rag-dashboard/internal/services"
)

// MCPHandler handles MCP server management requests
type MCPHandler struct {
	mcpService    *services.MCPService
	oauth2Service *services.OAuth2Service
}

// NewMCPHandler creates a new MCP handler
func NewMCPHandler(mcpService *services.MCPService, oauth2Service *services.OAuth2Service) *MCPHandler {
	return &MCPHandler{
		mcpService:    mcpService,
		oauth2Service: oauth2Service,
	}
}

// CreateMCPServer creates a new MCP server configuration
func (h *MCPHandler) CreateMCPServer(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	var req services.CreateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "details": err.Error()})
		return
	}

	// Validation
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	if req.ServerURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Server URL is required"})
		return
	}
	if req.AuthType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Auth type is required"})
		return
	}

	server, err := h.mcpService.CreateMCPServer(profileID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create MCP server", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, server)
}

// ListMCPServers lists all MCP servers
func (h *MCPHandler) ListMCPServers(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	servers, err := h.mcpService.ListMCPServers(profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list MCP servers"})
		return
	}

	c.JSON(http.StatusOK, servers)
}

// GetMCPServer gets an MCP server by ID
func (h *MCPHandler) GetMCPServer(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	server, err := h.mcpService.GetMCPServer(serverID, profileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "MCP server not found"})
		return
	}

	c.JSON(http.StatusOK, server)
}

// UpdateMCPServer updates an MCP server
func (h *MCPHandler) UpdateMCPServer(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req services.UpdateMCPServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	server, err := h.mcpService.UpdateMCPServer(serverID, profileID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update MCP server", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, server)
}

// DeleteMCPServer deletes an MCP server
func (h *MCPHandler) DeleteMCPServer(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	if err := h.mcpService.DeleteMCPServer(serverID, profileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete MCP server", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "MCP server deleted successfully"})
}

// SyncCapabilities syncs capabilities from an MCP server
func (h *MCPHandler) SyncCapabilities(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	if err := h.mcpService.SyncCapabilities(serverID, profileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync capabilities", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Capabilities synced successfully"})
}

// CallTool calls an MCP tool
func (h *MCPHandler) CallTool(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req struct {
		ToolName string                 `json:"tool_name"`
		Input    map[string]interface{} `json:"input"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.ToolName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tool name is required"})
		return
	}

	response, err := h.mcpService.CallTool(serverID, profileID, req.ToolName, req.Input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tool call failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// InitiateOAuth2 initiates OAuth 2 authorization flow
func (h *MCPHandler) InitiateOAuth2(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req struct {
		RedirectURL string `json:"redirect_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.RedirectURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Redirect URL is required"})
		return
	}

	authURL, err := h.oauth2Service.GenerateAuthorizationURL(profileID, serverID, req.RedirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate authorization URL", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"authorization_url": authURL})
}

// HandleOAuth2Callback handles OAuth 2 callback
func (h *MCPHandler) HandleOAuth2Callback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")
	errorParam := c.Query("error")

	if errorParam != "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "OAuth 2 authorization failed",
			"error_description": c.Query("error_description"),
		})
		return
	}

	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing state or code parameter"})
		return
	}

	server, err := h.oauth2Service.ExchangeCodeForToken(state, code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Authorization successful",
		"server":  server,
	})
}

// RefreshToken manually refreshes the OAuth 2 token
func (h *MCPHandler) RefreshToken(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	serverID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	if err := h.oauth2Service.RefreshAccessToken(serverID, profileID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to refresh token", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token refreshed successfully"})
}
