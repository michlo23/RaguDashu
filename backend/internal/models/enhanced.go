package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// DocumentFolder represents a folder for organizing documents
type DocumentFolder struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"size:255;not null"`
	ParentID  *uuid.UUID `gorm:"type:uuid"` // For nested folders
	Color     string    `gorm:"size:50"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Profile  *UserProfile     `gorm:"foreignKey:ProfileID"`
	Parent   *DocumentFolder  `gorm:"foreignKey:ParentID"`
	Children []DocumentFolder `gorm:"foreignKey:ParentID"`
}

// BeforeCreate hook
func (f *DocumentFolder) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (DocumentFolder) TableName() string {
	return "document_folders"
}

// DocumentTag represents a tag for categorizing documents
type DocumentTag struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID uuid.UUID `gorm:"type:uuid;not null"`
	Name      string    `gorm:"size:100;not null"`
	Color     string    `gorm:"size:50"`
	CreatedAt time.Time

	// Relations
	Profile *UserProfile `gorm:"foreignKey:ProfileID"`
}

// BeforeCreate hook
func (t *DocumentTag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (DocumentTag) TableName() string {
	return "document_tags"
}

// Add to Document model (update existing)
type DocumentMetadata struct {
	Tags       pq.StringArray `gorm:"type:text[]"`
	FolderID   *uuid.UUID     `gorm:"type:uuid"`
	IsFavorite bool           `gorm:"default:false"`
	Metadata   datatypes.JSON `gorm:"type:jsonb"` // PDF metadata, etc.
}

// UsageMetrics tracks user usage for analytics
type UsageMetrics struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_usage_profile_date"`
	Date           time.Time `gorm:"type:date;not null;uniqueIndex:idx_usage_profile_date"`
	DocumentsCount int       `gorm:"default:0"`
	SearchCount    int       `gorm:"default:0"`
	ChatCount      int       `gorm:"default:0"`
	TokensUsed     int       `gorm:"default:0"`
	StorageBytes   int64     `gorm:"default:0"`
	APICallsCount  int       `gorm:"default:0"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Relations
	Profile *UserProfile `gorm:"foreignKey:ProfileID"`
}

// BeforeCreate hook
func (m *UsageMetrics) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (UsageMetrics) TableName() string {
	return "usage_metrics"
}

// AuditLog tracks all user actions for compliance
type AuditLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID     uuid.UUID      `gorm:"type:uuid;not null"`
	ProfileID  *uuid.UUID     `gorm:"type:uuid"`
	Action     string         `gorm:"size:100;not null"` // document.upload, chat.create, etc.
	ResourceID *uuid.UUID     `gorm:"type:uuid"`
	Details    datatypes.JSON `gorm:"type:jsonb"`
	IPAddress  string         `gorm:"size:50"`
	UserAgent  string         `gorm:"size:500"`
	CreatedAt  time.Time

	// Relations
	User    *User        `gorm:"foreignKey:UserID"`
	Profile *UserProfile `gorm:"foreignKey:ProfileID"`
}

// BeforeCreate hook
func (l *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if l.ID == uuid.Nil {
		l.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (AuditLog) TableName() string {
	return "audit_logs"
}

// Webhook represents a user-configured webhook
type Webhook struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID uuid.UUID      `gorm:"type:uuid;not null"`
	Name      string         `gorm:"size:255;not null"`
	URL       string         `gorm:"size:1000;not null"`
	Events    pq.StringArray `gorm:"type:text[]"` // document.uploaded, chat.created
	Secret    string         `gorm:"size:500"`    // For signature verification
	IsActive  bool           `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	Profile *UserProfile `gorm:"foreignKey:ProfileID"`
}

// BeforeCreate hook
func (w *Webhook) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (Webhook) TableName() string {
	return "webhooks"
}

// ConversationTemplate represents pre-configured chat templates
type ConversationTemplate struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	ProfileID   *uuid.UUID `gorm:"type:uuid"` // NULL = system template
	Name        string    `gorm:"size:255;not null"`
	Description string    `gorm:"type:text"`
	Icon        string    `gorm:"size:50"`
	Model       string    `gorm:"size:100;not null"`
	SystemPrompt string   `gorm:"type:text;not null"`
	Temperature  float32  `gorm:"type:decimal(2,1);default:0.7"`
	MaxTokens    int      `gorm:"default:1000"`
	TopK         int      `gorm:"default:5"`
	IsPublic     bool     `gorm:"default:false"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// Relations
	Profile *UserProfile `gorm:"foreignKey:ProfileID"`
}

// BeforeCreate hook
func (t *ConversationTemplate) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (ConversationTemplate) TableName() string {
	return "conversation_templates"
}

// UserSession tracks user sessions for security
type UserSession struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Token     string    `gorm:"size:500;uniqueIndex;not null"` // Hashed session token
	IPAddress string    `gorm:"size:50"`
	UserAgent string    `gorm:"size:500"`
	ExpiresAt time.Time `gorm:"not null"`
	CreatedAt time.Time
	LastUsedAt time.Time

	// Relations
	User *User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook
func (s *UserSession) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (UserSession) TableName() string {
	return "user_sessions"
}

// TwoFactorAuth stores 2FA secrets
type TwoFactorAuth struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;uniqueIndex;not null"`
	Secret    string    `gorm:"size:500;not null"` // Encrypted TOTP secret
	Enabled   bool      `gorm:"default:false"`
	BackupCodes pq.StringArray `gorm:"type:text[]"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// Relations
	User *User `gorm:"foreignKey:UserID"`
}

// BeforeCreate hook
func (t *TwoFactorAuth) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

// TableName specifies the table name
func (TwoFactorAuth) TableName() string {
	return "two_factor_auth"
}
