package services

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// AuditService handles audit logging
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// Log creates an audit log entry
func (s *AuditService) Log(userID, profileID uuid.UUID, action string, resourceID *uuid.UUID, details map[string]interface{}, c *gin.Context) error {
	detailsJSON, _ := json.Marshal(details)

	log := &models.AuditLog{
		UserID:     userID,
		ProfileID:  &profileID,
		Action:     action,
		ResourceID: resourceID,
		Details:    detailsJSON,
		IPAddress:  c.ClientIP(),
		UserAgent:  c.Request.UserAgent(),
	}

	return s.db.Create(log).Error
}

// LogWithoutContext creates an audit log without gin context
func (s *AuditService) LogWithoutContext(userID, profileID uuid.UUID, action string, resourceID *uuid.UUID, details map[string]interface{}) error {
	detailsJSON, _ := json.Marshal(details)

	log := &models.AuditLog{
		UserID:     userID,
		ProfileID:  &profileID,
		Action:     action,
		ResourceID: resourceID,
		Details:    detailsJSON,
	}

	return s.db.Create(log).Error
}

// GetUserLogs retrieves audit logs for a user
func (s *AuditService) GetUserLogs(userID uuid.UUID, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetProfileLogs retrieves audit logs for a profile
func (s *AuditService) GetProfileLogs(profileID uuid.UUID, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.Where("profile_id = ?", profileID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// GetActionLogs retrieves logs for specific action type
func (s *AuditService) GetActionLogs(action string, limit int) ([]models.AuditLog, error) {
	var logs []models.AuditLog
	err := s.db.Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// Audit action constants
const (
	ActionUserRegister       = "user.register"
	ActionUserLogin          = "user.login"
	ActionUserLogout         = "user.logout"
	ActionDocumentUpload     = "document.upload"
	ActionDocumentDelete     = "document.delete"
	ActionCredentialCreate   = "credential.create"
	ActionCredentialDelete   = "credential.delete"
	ActionSearchPerform      = "search.perform"
	ActionChatCreate         = "chat.create"
	ActionChatDelete         = "chat.delete"
	ActionConversationSave   = "conversation.save"
	ActionConversationDelete = "conversation.delete"
	ActionFolderCreate       = "folder.create"
	ActionFolderDelete       = "folder.delete"
	ActionWebhookCreate      = "webhook.create"
	ActionWebhookDelete      = "webhook.delete"
)
