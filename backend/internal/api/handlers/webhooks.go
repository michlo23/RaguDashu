package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/api/middleware"
	"github.com/michlo23/rag-dashboard/internal/services"
)

type WebhookHandler struct {
	webhookService *services.WebhookService
}

func NewWebhookHandler(webhookService *services.WebhookService) *WebhookHandler {
	return &WebhookHandler{webhookService: webhookService}
}

type CreateWebhookRequest struct {
	Name   string   `json:"name" binding:"required"`
	URL    string   `json:"url" binding:"required,url"`
	Events []string `json:"events" binding:"required"`
}

// CreateWebhook creates a new webhook
func (h *WebhookHandler) CreateWebhook(c *gin.Context) {
	profileID, _ := middleware.GetProfileID(c)

	var req CreateWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	webhook, err := h.webhookService.CreateWebhook(profileID, req.Name, req.URL, req.Events)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create webhook"})
		return
	}

	c.JSON(http.StatusCreated, webhook)
}

// ListWebhooks lists all webhooks for the user
func (h *WebhookHandler) ListWebhooks(c *gin.Context) {
	profileID, _ := middleware.GetProfileID(c)

	webhooks, err := h.webhookService.ListWebhooks(profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch webhooks"})
		return
	}

	c.JSON(http.StatusOK, webhooks)
}

// DeleteWebhook deletes a webhook
func (h *WebhookHandler) DeleteWebhook(c *gin.Context) {
	profileID, _ := middleware.GetProfileID(c)
	webhookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook ID"})
		return
	}

	if err := h.webhookService.DeleteWebhook(profileID, webhookID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete webhook"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook deleted successfully"})
}
