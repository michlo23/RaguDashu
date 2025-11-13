package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type Document struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID    uuid.UUID      `gorm:"type:uuid;not null"`
	IndexID      uuid.UUID      `gorm:"type:uuid;not null"`
	Filename     string         `gorm:"size:255;not null"`
	FileType     string         `gorm:"size:50;not null"`
	FileSize     int64          `gorm:"not null"`
	VectorIDs    pq.StringArray `gorm:"type:text[]"` // Array of Pinecone vector IDs
	ChunkCount   int            `gorm:"default:0"`
	UploadStatus string         `gorm:"size:50;default:'processing'"` // 'processing', 'completed', 'failed'
	ErrorMessage *string        `gorm:"type:text"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Relations
	Profile *UserProfile    `gorm:"foreignKey:ProfileID"`
	Index   *PineconeIndex  `gorm:"foreignKey:IndexID"`
	Chunks  []DocumentChunk `gorm:"foreignKey:DocumentID"`
}

// BeforeCreate hook
func (d *Document) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Document) TableName() string {
	return "documents"
}

// Document statuses
const (
	DocumentStatusProcessing = "processing"
	DocumentStatusCompleted  = "completed"
	DocumentStatusFailed     = "failed"
)

type DocumentChunk struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	DocumentID uuid.UUID `gorm:"type:uuid;not null"`
	ChunkIndex int       `gorm:"not null"`
	ChunkText  string    `gorm:"type:text;not null"`
	VectorID   string    `gorm:"size:255;not null"` // Pinecone vector ID
	TokenCount *int
	CreatedAt  time.Time

	// Relations
	Document *Document `gorm:"foreignKey:DocumentID"`
}

// BeforeCreate hook
func (c *DocumentChunk) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (DocumentChunk) TableName() string {
	return "document_chunks"
}
