package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// MCPServer represents an MCP (Model Context Protocol) server configuration
type MCPServer struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Name         string    `gorm:"size:255;not null"` // User-friendly name (e.g., "Atlassian Jira")
	Description  string    `gorm:"type:text"`
	ServerURL    string    `gorm:"size:500;not null"` // MCP server endpoint
	AuthType     string    `gorm:"size:50;not null"`  // 'none', 'api_key', 'oauth2'

	// API Key auth (if AuthType = 'api_key')
	EncryptedAPIKey string `gorm:"type:text"` // Encrypted API key

	// OAuth 2 configuration (if AuthType = 'oauth2')
	OAuth2Config datatypes.JSON `gorm:"type:jsonb"` // OAuth2 config (client_id, client_secret, auth_url, token_url, scopes)

	// OAuth 2 tokens (encrypted)
	EncryptedAccessToken  string     `gorm:"type:text"`
	EncryptedRefreshToken string     `gorm:"type:text"`
	TokenExpiresAt        *time.Time

	// MCP capabilities (cached from server)
	AvailableTools     datatypes.JSON `gorm:"type:jsonb"` // List of available tools from MCP server
	AvailableResources datatypes.JSON `gorm:"type:jsonb"` // List of available resources
	AvailablePrompts   datatypes.JSON `gorm:"type:jsonb"` // List of available prompts
	LastSyncedAt       *time.Time                         // When capabilities were last fetched

	// Status
	IsActive       bool       `gorm:"default:true"`
	LastTestedAt   *time.Time
	TestResult     *string    `gorm:"size:255"` // Success or error message
	LastUsedAt     *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Profile   *UserProfile    `gorm:"foreignKey:ProfileID"`
	ToolUsage []MCPToolUsage  `gorm:"foreignKey:MCPServerID"`
}

// BeforeCreate hook
func (m *MCPServer) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (MCPServer) TableName() string {
	return "mcp_servers"
}

// Auth types
const (
	MCPAuthTypeNone   = "none"
	MCPAuthTypeAPIKey = "api_key"
	MCPAuthTypeOAuth2 = "oauth2"
)

// OAuth2State stores temporary OAuth 2 flow state
type OAuth2State struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	State         string    `gorm:"size:64;not null;uniqueIndex"` // Random state parameter
	ProfileID     uuid.UUID `gorm:"type:uuid;not null"`
	MCPServerID   uuid.UUID `gorm:"type:uuid;not null"`
	CodeVerifier  string    `gorm:"size:128;not null"` // PKCE code verifier
	RedirectURL   string    `gorm:"size:500;not null"` // Where to redirect after OAuth
	ExpiresAt     time.Time `gorm:"not null"`          // State expires after 10 minutes
	CreatedAt     time.Time

	// Relations
	Profile   *UserProfile `gorm:"foreignKey:ProfileID"`
	MCPServer *MCPServer   `gorm:"foreignKey:MCPServerID"`
}

// BeforeCreate hook
func (o *OAuth2State) BeforeCreate(tx *gorm.DB) error {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	// Set expiry to 10 minutes if not set
	if o.ExpiresAt.IsZero() {
		o.ExpiresAt = time.Now().Add(10 * time.Minute)
	}
	return nil
}

// TableName specifies the table name
func (OAuth2State) TableName() string {
	return "oauth2_states"
}

// MCPToolUsage tracks MCP tool usage in conversations
type MCPToolUsage struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ConversationID uuid.UUID      `gorm:"type:uuid;not null;index"`
	MessageID      uuid.UUID      `gorm:"type:uuid;not null;index"`
	MCPServerID    uuid.UUID      `gorm:"type:uuid;not null;index"`
	ToolName       string         `gorm:"size:255;not null"` // Name of the tool that was called
	ToolInput      datatypes.JSON `gorm:"type:jsonb"`        // Input parameters
	ToolOutput     datatypes.JSON `gorm:"type:jsonb"`        // Tool response
	ErrorMessage   *string        `gorm:"type:text"`         // Error if tool call failed
	DurationMs     *int                                      // How long the tool call took
	CreatedAt      time.Time

	// Relations
	Conversation *ChatConversation `gorm:"foreignKey:ConversationID"`
	Message      *ChatMessage      `gorm:"foreignKey:MessageID"`
	MCPServer    *MCPServer        `gorm:"foreignKey:MCPServerID"`
}

// BeforeCreate hook
func (m *MCPToolUsage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (MCPToolUsage) TableName() string {
	return "mcp_tool_usage"
}

// MCP Tool/Resource/Prompt structures (for JSON storage)
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MCPResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MimeType    string `json:"mimeType"`
}

type MCPPrompt struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Arguments   []struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Required    bool   `json:"required"`
	} `json:"arguments"`
}

// OAuth2Config structure (for JSON storage)
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	AuthURL      string   `json:"auth_url"`
	TokenURL     string   `json:"token_url"`
	Scopes       []string `json:"scopes"`
}
