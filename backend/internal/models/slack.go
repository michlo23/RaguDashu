package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SlackThread struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID    uuid.UUID `gorm:"type:uuid;not null"`
	IndexID      uuid.UUID `gorm:"type:uuid;not null"`
	ThreadTS     string    `gorm:"size:255;not null"`
	ChannelID    string    `gorm:"size:255;not null"`
	ChannelName  *string   `gorm:"size:255"`
	MessageCount int       `gorm:"default:0"`
	CombinedText *string   `gorm:"type:text"`
	VectorID     *string   `gorm:"size:255"` // Pinecone vector ID for the combined thread
	TrustScore   *float32  `gorm:"type:decimal(3,2)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Relations
	Profile  *UserProfile   `gorm:"foreignKey:ProfileID"`
	Index    *PineconeIndex `gorm:"foreignKey:IndexID"`
	Messages []SlackMessage `gorm:"foreignKey:ThreadID"`
}

// BeforeCreate hook
func (t *SlackThread) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (SlackThread) TableName() string {
	return "slack_threads"
}

type SlackMessage struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ThreadID  uuid.UUID      `gorm:"type:uuid;not null"`
	MessageTS string         `gorm:"size:255;not null"`
	UserID    *string        `gorm:"size:255"`
	UserName  *string        `gorm:"size:255"`
	Text      *string        `gorm:"type:text"`
	Reactions datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt time.Time

	// Relations
	Thread *SlackThread `gorm:"foreignKey:ThreadID"`
}

// BeforeCreate hook
func (m *SlackMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (SlackMessage) TableName() string {
	return "slack_messages"
}
