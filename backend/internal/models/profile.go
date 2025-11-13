package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserProfile struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID            uuid.UUID `gorm:"type:uuid;not null"`
	Name              string    `gorm:"size:255;not null"`
	PineconeNamespace string    `gorm:"size:255;uniqueIndex;not null"` // Format: user_{user_id}
	IsDefault         bool      `gorm:"default:false"`
	CreatedAt         time.Time
	UpdatedAt         time.Time

	// Relations
	User         *User                `gorm:"foreignKey:UserID"`
	Credentials  []UserCredential     `gorm:"foreignKey:ProfileID"`
	Indexes      []PineconeIndex      `gorm:"foreignKey:ProfileID"`
	Documents    []Document           `gorm:"foreignKey:ProfileID"`
	SlackThreads []SlackThread        `gorm:"foreignKey:ProfileID"`
	ChatConfigs  []ChatConfiguration  `gorm:"foreignKey:ProfileID"`
	ChatConvs    []ChatConversation   `gorm:"foreignKey:ProfileID"`
}

// BeforeCreate hook
func (p *UserProfile) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	// Generate namespace if not set
	if p.PineconeNamespace == "" {
		p.PineconeNamespace = fmt.Sprintf("user_%s", p.UserID.String())
	}
	return nil
}

// TableName specifies the table name
func (UserProfile) TableName() string {
	return "user_profiles"
}
