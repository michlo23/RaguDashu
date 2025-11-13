package services

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// OAuth2Service handles OAuth 2.0 authorization flow with PKCE
type OAuth2Service struct {
	db                *gorm.DB
	encryptionService *EncryptionService
}

// NewOAuth2Service creates a new OAuth 2 service
func NewOAuth2Service(db *gorm.DB, encryptionService *EncryptionService) *OAuth2Service {
	return &OAuth2Service{
		db:                db,
		encryptionService: encryptionService,
	}
}

// GenerateAuthorizationURL generates the OAuth 2 authorization URL with PKCE
func (s *OAuth2Service) GenerateAuthorizationURL(
	profileID uuid.UUID,
	mcpServerID uuid.UUID,
	redirectURL string,
) (string, error) {
	// Get MCP server configuration
	var mcpServer models.MCPServer
	if err := s.db.Where("id = ? AND profile_id = ?", mcpServerID, profileID).
		First(&mcpServer).Error; err != nil {
		return "", fmt.Errorf("MCP server not found")
	}

	if mcpServer.AuthType != models.MCPAuthTypeOAuth2 {
		return "", fmt.Errorf("server does not use OAuth 2")
	}

	// Parse OAuth 2 config
	var oauth2Config models.OAuth2Config
	if err := json.Unmarshal(mcpServer.OAuth2Config, &oauth2Config); err != nil {
		return "", fmt.Errorf("invalid OAuth 2 configuration: %w", err)
	}

	// Generate PKCE code verifier and challenge
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return "", fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := generateCodeChallenge(codeVerifier)

	// Generate random state
	state, err := generateRandomState()
	if err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}

	// Store OAuth 2 state in database
	oauth2State := &models.OAuth2State{
		State:        state,
		ProfileID:    profileID,
		MCPServerID:  mcpServerID,
		CodeVerifier: codeVerifier,
		RedirectURL:  redirectURL,
	}

	if err := s.db.Create(oauth2State).Error; err != nil {
		return "", fmt.Errorf("failed to store OAuth 2 state: %w", err)
	}

	// Build authorization URL
	authURL, err := url.Parse(oauth2Config.AuthURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth URL: %w", err)
	}

	params := url.Values{}
	params.Set("client_id", oauth2Config.ClientID)
	params.Set("response_type", "code")
	params.Set("redirect_uri", redirectURL)
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	if len(oauth2Config.Scopes) > 0 {
		params.Set("scope", strings.Join(oauth2Config.Scopes, " "))
	}

	authURL.RawQuery = params.Encode()

	return authURL.String(), nil
}

// ExchangeCodeForToken exchanges authorization code for access token
func (s *OAuth2Service) ExchangeCodeForToken(
	state string,
	code string,
) (*models.MCPServer, error) {
	// Get OAuth 2 state from database
	var oauth2State models.OAuth2State
	if err := s.db.Where("state = ?", state).First(&oauth2State).Error; err != nil {
		return nil, fmt.Errorf("invalid or expired state")
	}

	// Check if state has expired
	if time.Now().After(oauth2State.ExpiresAt) {
		s.db.Delete(&oauth2State)
		return nil, fmt.Errorf("state has expired")
	}

	// Get MCP server
	var mcpServer models.MCPServer
	if err := s.db.Where("id = ?", oauth2State.MCPServerID).First(&mcpServer).Error; err != nil {
		return nil, fmt.Errorf("MCP server not found")
	}

	// Parse OAuth 2 config
	var oauth2Config models.OAuth2Config
	if err := json.Unmarshal(mcpServer.OAuth2Config, &oauth2Config); err != nil {
		return nil, fmt.Errorf("invalid OAuth 2 configuration: %w", err)
	}

	// Exchange authorization code for access token
	tokenResp, err := s.exchangeCode(
		oauth2Config,
		code,
		oauth2State.RedirectURL,
		oauth2State.CodeVerifier,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// Encrypt tokens
	encryptedAccessToken, err := s.encryptionService.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	var encryptedRefreshToken string
	if tokenResp.RefreshToken != "" {
		encryptedRefreshToken, err = s.encryptionService.Encrypt(tokenResp.RefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	// Update MCP server with tokens
	now := time.Now()
	expiresAt := now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	mcpServer.EncryptedAccessToken = encryptedAccessToken
	mcpServer.EncryptedRefreshToken = encryptedRefreshToken
	mcpServer.TokenExpiresAt = &expiresAt
	mcpServer.IsActive = true

	if err := s.db.Save(&mcpServer).Error; err != nil {
		return nil, fmt.Errorf("failed to save tokens: %w", err)
	}

	// Delete OAuth 2 state
	s.db.Delete(&oauth2State)

	return &mcpServer, nil
}

// RefreshAccessToken refreshes the access token using refresh token
func (s *OAuth2Service) RefreshAccessToken(mcpServerID uuid.UUID, profileID uuid.UUID) error {
	// Get MCP server
	var mcpServer models.MCPServer
	if err := s.db.Where("id = ? AND profile_id = ?", mcpServerID, profileID).
		First(&mcpServer).Error; err != nil {
		return fmt.Errorf("MCP server not found")
	}

	if mcpServer.EncryptedRefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	// Decrypt refresh token
	refreshToken, err := s.encryptionService.Decrypt(mcpServer.EncryptedRefreshToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt refresh token: %w", err)
	}

	// Parse OAuth 2 config
	var oauth2Config models.OAuth2Config
	if err := json.Unmarshal(mcpServer.OAuth2Config, &oauth2Config); err != nil {
		return fmt.Errorf("invalid OAuth 2 configuration: %w", err)
	}

	// Refresh token
	tokenResp, err := s.refreshToken(oauth2Config, refreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Encrypt new access token
	encryptedAccessToken, err := s.encryptionService.Encrypt(tokenResp.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	// Update MCP server
	now := time.Now()
	expiresAt := now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	mcpServer.EncryptedAccessToken = encryptedAccessToken
	mcpServer.TokenExpiresAt = &expiresAt

	// Update refresh token if provided
	if tokenResp.RefreshToken != "" {
		encryptedRefreshToken, err := s.encryptionService.Encrypt(tokenResp.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
		mcpServer.EncryptedRefreshToken = encryptedRefreshToken
	}

	if err := s.db.Save(&mcpServer).Error; err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	return nil
}

// GetValidAccessToken returns a valid access token, refreshing if necessary
func (s *OAuth2Service) GetValidAccessToken(mcpServerID uuid.UUID, profileID uuid.UUID) (string, error) {
	var mcpServer models.MCPServer
	if err := s.db.Where("id = ? AND profile_id = ?", mcpServerID, profileID).
		First(&mcpServer).Error; err != nil {
		return "", fmt.Errorf("MCP server not found")
	}

	// Check if token is expired or will expire in next 5 minutes
	if mcpServer.TokenExpiresAt != nil && time.Now().Add(5*time.Minute).After(*mcpServer.TokenExpiresAt) {
		// Token is expired or expiring soon, refresh it
		if err := s.RefreshAccessToken(mcpServerID, profileID); err != nil {
			return "", fmt.Errorf("failed to refresh token: %w", err)
		}

		// Re-fetch server to get new token
		if err := s.db.Where("id = ?", mcpServerID).First(&mcpServer).Error; err != nil {
			return "", err
		}
	}

	// Decrypt and return access token
	accessToken, err := s.encryptionService.Decrypt(mcpServer.EncryptedAccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt access token: %w", err)
	}

	return accessToken, nil
}

// TokenResponse represents OAuth 2 token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

// exchangeCode exchanges authorization code for tokens
func (s *OAuth2Service) exchangeCode(
	config models.OAuth2Config,
	code string,
	redirectURI string,
	codeVerifier string,
) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)
	data.Set("code_verifier", codeVerifier)

	return s.requestToken(config.TokenURL, data)
}

// refreshToken refreshes the access token
func (s *OAuth2Service) refreshToken(config models.OAuth2Config, refreshToken string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", config.ClientID)
	data.Set("client_secret", config.ClientSecret)

	return s.requestToken(config.TokenURL, data)
}

// requestToken makes a token request
func (s *OAuth2Service) requestToken(tokenURL string, data url.Values) (*TokenResponse, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// generateCodeVerifier generates a PKCE code verifier
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge generates a PKCE code challenge from verifier
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// generateRandomState generates a random state parameter
func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// CleanupExpiredStates removes expired OAuth 2 states
func (s *OAuth2Service) CleanupExpiredStates() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&models.OAuth2State{}).Error
}
