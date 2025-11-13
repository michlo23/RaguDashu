package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ChatConfiguration struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID       uuid.UUID      `gorm:"type:uuid;not null"`
	Name            string         `gorm:"size:255;not null"`
	Model           string         `gorm:"size:100;not null"` // gpt-4, gpt-4-turbo-preview, etc.
	IndexID         *uuid.UUID     `gorm:"type:uuid"`         // NULL = no RAG
	SystemPrompt    string         `gorm:"type:text;not null"`
	Temperature     float32        `gorm:"type:decimal(2,1);default:0.7"`
	MaxTokens       int            `gorm:"default:1000"`
	TopK            int            `gorm:"default:5"` // Context chunks to retrieve
	EnabledMCPTools datatypes.JSON `gorm:"type:jsonb"` // Array of MCP server IDs to enable for this config
	IsDefault       bool           `gorm:"default:false"`
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// Relations
	Profile *UserProfile   `gorm:"foreignKey:ProfileID"`
	Index   *PineconeIndex `gorm:"foreignKey:IndexID"`
}

// BeforeCreate hook
func (c *ChatConfiguration) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (ChatConfiguration) TableName() string {
	return "chat_configurations"
}

type ChatConversation struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID     uuid.UUID  `gorm:"type:uuid;not null"`
	ConfigID      *uuid.UUID `gorm:"type:uuid"`
	Title         string     `gorm:"size:255;not null"`
	IsSaved       bool       `gorm:"default:false"` // false = 60-day expiry
	MessageCount  int        `gorm:"default:0"`
	LastMessageAt *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ExpiresAt     *time.Time // NULL if IsSaved=true, otherwise created_at + 60 days

	// Relations
	Profile  *UserProfile       `gorm:"foreignKey:ProfileID"`
	Config   *ChatConfiguration `gorm:"foreignKey:ConfigID"`
	Messages []ChatMessage      `gorm:"foreignKey:ConversationID"`
}

// BeforeCreate hook
func (c *ChatConversation) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	// Set expiry if not saved
	if !c.IsSaved && c.ExpiresAt == nil {
		expiryDate := time.Now().AddDate(0, 0, 60) // 60 days from now
		c.ExpiresAt = &expiryDate
	}
	return nil
}

// TableName specifies the table name
func (ChatConversation) TableName() string {
	return "chat_conversations"
}

type ChatMessage struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ConversationID   uuid.UUID      `gorm:"type:uuid;not null"`
	Role             string         `gorm:"size:20;not null"` // user, assistant, system
	Content          string         `gorm:"type:text;not null"`
	RetrievedContext datatypes.JSON `gorm:"type:jsonb"` // Store retrieved chunks
	TokenCount       *int
	CreatedAt        time.Time

	// Relations
	Conversation *ChatConversation `gorm:"foreignKey:ConversationID"`
}

// BeforeCreate hook
func (m *ChatMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// Message roles
const (
	MessageRoleUser      = "user"
	MessageRoleAssistant = "assistant"
	MessageRoleSystem    = "system"
)
