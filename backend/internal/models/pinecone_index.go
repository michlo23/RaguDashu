package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PineconeIndex struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID     uuid.UUID `gorm:"type:uuid;not null"`
	IndexName     string    `gorm:"size:255;not null"` // Format: rag-{user_id}-{timestamp}
	DisplayName   string    `gorm:"size:255;not null"`
	Dimension     int       `gorm:"default:1536"` // OpenAI embedding dimension
	Metric        string    `gorm:"size:50;default:'cosine'"`
	IsActive      bool      `gorm:"default:true"`
	DocumentCount int       `gorm:"default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	// Relations
	Profile   *UserProfile   `gorm:"foreignKey:ProfileID"`
	Documents []Document     `gorm:"foreignKey:IndexID"`
	Threads   []SlackThread  `gorm:"foreignKey:IndexID"`
}

// BeforeCreate hook
func (i *PineconeIndex) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// GenerateIndexName creates a unique index name
func GenerateIndexName(userID uuid.UUID) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("rag-%s-%d", userID.String()[:8], timestamp)
}

// TableName specifies the table name
func (PineconeIndex) TableName() string {
	return "pinecone_indexes"
}
