package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserCredential struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID      uuid.UUID  `gorm:"type:uuid;not null"`
	CredentialType string     `gorm:"size:50;not null"` // 'openai_api_key', 'pinecone_api_key', 'slack_token'
	EncryptedValue string     `gorm:"type:text;not null"`
	IsActive       bool       `gorm:"default:true"`
	LastTestedAt   *time.Time
	TestResult     *string    `gorm:"size:255"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relations
	Profile *UserProfile `gorm:"foreignKey:ProfileID"`
}

// BeforeCreate hook
func (c *UserCredential) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (UserCredential) TableName() string {
	return "user_credentials"
}

// Credential types
const (
	CredentialTypeOpenAI   = "openai_api_key"
	CredentialTypePinecone = "pinecone_api_key"
	CredentialTypeSlack    = "slack_token"
)
