package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/api/middleware"
	"github.com/michlo23/rag-dashboard/internal/models"
	"github.com/michlo23/rag-dashboard/internal/services"
	"gorm.io/gorm"
)

type ChatHandler struct {
	chatService *services.ChatService
	db          *gorm.DB
}

func NewChatHandler(chatService *services.ChatService, db *gorm.DB) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		db:          db,
	}
}

// SendMessage handles chat message
func (h *ChatHandler) SendMessage(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}


	var req services.ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.chatService.Send(profileID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// CreateConfiguration creates a chat configuration
func (h *ChatHandler) CreateConfiguration(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}


	var config models.ChatConfiguration
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.ProfileID = profileID
	if err := h.db.Create(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
		return
	}

	c.JSON(http.StatusCreated, config)
}

// ListConfigurations lists all configurations
func (h *ChatHandler) ListConfigurations(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}


	var configs []models.ChatConfiguration
	if err := h.db.Where("profile_id = ?", profileID).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch configurations"})
		return
	}

	c.JSON(http.StatusOK, configs)
}

// GetConversation retrieves a conversation with messages
func (h *ChatHandler) GetConversation(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	conv, err := h.chatService.GetConversation(profileID, convID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conv)
}

// ListConversations lists all conversations
func (h *ChatHandler) ListConversations(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}


	convs, err := h.chatService.ListConversations(profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	c.JSON(http.StatusOK, convs)
}

// SaveConversation marks a conversation as saved
func (h *ChatHandler) SaveConversation(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	if err := h.chatService.SaveConversation(profileID, convID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation saved successfully"})
}

// DeleteConversation deletes a conversation
func (h *ChatHandler) DeleteConversation(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	convID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	if err := h.chatService.DeleteConversation(profileID, convID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted successfully"})
}
