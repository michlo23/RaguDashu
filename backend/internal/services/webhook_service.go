package services

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// WebhookService handles webhook delivery
type WebhookService struct {
	db *gorm.DB
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB) *WebhookService {
	return &WebhookService{db: db}
}

// WebhookPayload represents a webhook payload
type WebhookPayload struct {
	Event     string                 `json:"event"`
	Timestamp string                 `json:"timestamp"`
	ProfileID string                 `json:"profile_id"`
	Data      map[string]interface{} `json:"data"`
}

// Trigger sends webhooks for an event
func (s *WebhookService) Trigger(profileID uuid.UUID, event string, data map[string]interface{}) {
	// Get all active webhooks for this profile and event
	var webhooks []models.Webhook
	s.db.Where("profile_id = ? AND is_active = ? AND ? = ANY(events)", profileID, true, event).
		Find(&webhooks)

	// Send webhooks asynchronously
	for _, webhook := range webhooks {
		go s.sendWebhook(webhook, event, profileID, data)
	}
}

// sendWebhook sends a single webhook
func (s *WebhookService) sendWebhook(webhook models.Webhook, event string, profileID uuid.UUID, data map[string]interface{}) {
	payload := WebhookPayload{
		Event:     event,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		ProfileID: profileID.String(),
		Data:      data,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	req, err := http.NewRequest("POST", webhook.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "RAG-Dashboard-Webhook/1.0")

	// Add HMAC signature if secret is configured
	if webhook.Secret != "" {
		signature := s.generateSignature(jsonData, webhook.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send request (ignore response for now)
	client.Do(req)
}

// generateSignature creates HMAC SHA256 signature
func (s *WebhookService) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	return hex.EncodeToString(h.Sum(nil))
}

// CreateWebhook creates a new webhook
func (s *WebhookService) CreateWebhook(profileID uuid.UUID, name, url string, events []string) (*models.Webhook, error) {
	// Generate secret
	secretBytes := make([]byte, 32)
	// Use timestamp as simple secret for now
	secret := fmt.Sprintf("%d", time.Now().UnixNano())

	webhook := &models.Webhook{
		ProfileID: profileID,
		Name:      name,
		URL:       url,
		Events:    events,
		Secret:    secret,
		IsActive:  true,
	}

	if err := s.db.Create(webhook).Error; err != nil {
		return nil, err
	}

	return webhook, nil
}

// ListWebhooks returns all webhooks for a profile
func (s *WebhookService) ListWebhooks(profileID uuid.UUID) ([]models.Webhook, error) {
	var webhooks []models.Webhook
	err := s.db.Where("profile_id = ?", profileID).Find(&webhooks).Error
	return webhooks, err
}

// DeleteWebhook deletes a webhook
func (s *WebhookService) DeleteWebhook(profileID, webhookID uuid.UUID) error {
	return s.db.Where("id = ? AND profile_id = ?", webhookID, profileID).
		Delete(&models.Webhook{}).Error
}

// Webhook event constants
const (
	WebhookEventDocumentUploaded   = "document.uploaded"
	WebhookEventDocumentProcessed  = "document.processed"
	WebhookEventDocumentDeleted    = "document.deleted"
	WebhookEventChatCreated        = "chat.created"
	WebhookEventConversationSaved  = "conversation.saved"
	WebhookEventSearchPerformed    = "search.performed"
)
