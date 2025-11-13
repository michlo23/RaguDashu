package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// ChatService handles AI chat with RAG
type ChatService struct {
	db                *gorm.DB
	searchService     *SearchService
	credentialService *CredentialService
}

// NewChatService creates a new chat service
func NewChatService(
	db *gorm.DB,
	searchService *SearchService,
	credentialService *CredentialService,
) *ChatService {
	return &ChatService{
		db:                db,
		searchService:     searchService,
		credentialService: credentialService,
	}
}

// ChatRequest represents a chat request
type ChatRequest struct {
	ConversationID *uuid.UUID `json:"conversation_id"`
	ConfigID       uuid.UUID  `json:"config_id"`
	Message        string     `json:"message"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	ConversationID uuid.UUID               `json:"conversation_id"`
	Message        string                  `json:"message"`
	Context        []models.SearchResult   `json:"context,omitempty"`
}

// Send processes a chat message
func (s *ChatService) Send(profileID uuid.UUID, req ChatRequest) (*ChatResponse, error) {
	// Get or create conversation
	conversation, err := s.getOrCreateConversation(profileID, req.ConversationID, req.ConfigID)
	if err != nil {
		return nil, err
	}

	// Get configuration
	var config models.ChatConfiguration
	if err := s.db.Where("id = ? AND profile_id = ?", req.ConfigID, profileID).
		First(&config).Error; err != nil {
		return nil, fmt.Errorf("configuration not found")
	}

	// Retrieve context if RAG enabled
	var retrievedContext []models.SearchResult
	systemPrompt := config.SystemPrompt

	if config.IndexID != nil {
		// Perform semantic search
		results, err := s.searchService.Search(profileID, *config.IndexID, req.Message, config.TopK)
		if err == nil && len(results) > 0 {
			retrievedContext = results

			// Inject context into system prompt
			contextText := formatContext(retrievedContext)
			systemPrompt = fmt.Sprintf("%s\n\n=== Retrieved Context ===\n%s", config.SystemPrompt, contextText)
		}
	}

	// Build messages for OpenAI
	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
	}

	// Add conversation history (last 10 messages)
	historyMessages, err := s.getConversationHistory(conversation.ID, 10)
	if err == nil {
		for _, msg := range historyMessages {
			messages = append(messages, map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			})
		}
	}

	// Add current user message
	messages = append(messages, map[string]string{
		"role":    "user",
		"content": req.Message,
	})

	// Call OpenAI API
	openAIKey, err := s.credentialService.GetDecryptedCredential(profileID, models.CredentialTypeOpenAI)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API key not found")
	}

	assistantReply, err := s.callOpenAI(config.Model, messages, config.Temperature, config.MaxTokens, openAIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	// Save messages to database
	if err := s.saveMessages(conversation.ID, req.Message, assistantReply, retrievedContext); err != nil {
		return nil, err
	}

	// Update conversation
	now := time.Now()
	conversation.LastMessageAt = &now
	conversation.MessageCount += 2
	s.db.Save(&conversation)

	return &ChatResponse{
		ConversationID: conversation.ID,
		Message:        assistantReply,
		Context:        retrievedContext,
	}, nil
}

// getOrCreateConversation gets existing or creates new conversation
func (s *ChatService) getOrCreateConversation(profileID uuid.UUID, convID *uuid.UUID, configID uuid.UUID) (*models.ChatConversation, error) {
	if convID != nil {
		var conv models.ChatConversation
		if err := s.db.Where("id = ? AND profile_id = ?", *convID, profileID).
			First(&conv).Error; err == nil {
			return &conv, nil
		}
	}

	// Create new conversation
	conv := &models.ChatConversation{
		ProfileID: profileID,
		ConfigID:  &configID,
		Title:     "New Conversation",
		IsSaved:   false,
	}

	if err := s.db.Create(conv).Error; err != nil {
		return nil, err
	}

	return conv, nil
}

// getConversationHistory retrieves recent messages
func (s *ChatService) getConversationHistory(conversationID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	err := s.db.Where("conversation_id = ?", conversationID).
		Order("created_at DESC").
		Limit(limit).
		Find(&messages).Error
	if err != nil {
		return nil, err
	}

	// Reverse to chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// saveMessages saves user and assistant messages
func (s *ChatService) saveMessages(conversationID uuid.UUID, userMsg, assistantMsg string, context []models.SearchResult) error {
	// Save user message
	contextJSON, _ := json.Marshal(context)
	userMessage := &models.ChatMessage{
		ConversationID:   conversationID,
		Role:             models.MessageRoleUser,
		Content:          userMsg,
		RetrievedContext: contextJSON,
	}
	if err := s.db.Create(userMessage).Error; err != nil {
		return err
	}

	// Save assistant message
	assistantMessage := &models.ChatMessage{
		ConversationID: conversationID,
		Role:           models.MessageRoleAssistant,
		Content:        assistantMsg,
	}
	return s.db.Create(assistantMessage).Error
}

// callOpenAI makes a request to OpenAI Chat Completions API
func (s *ChatService) callOpenAI(model string, messages []map[string]string, temperature float32, maxTokens int, apiKey string) (string, error) {
	reqBody := map[string]interface{}{
		"model":       model,
		"messages":    messages,
		"temperature": temperature,
		"max_tokens":  maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	client := &http.Client{Timeout: 60 * 1000000000} // 60 seconds
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// formatContext formats search results for prompt
func formatContext(results []models.SearchResult) string {
	var formatted string
	for i, result := range results {
		formatted += fmt.Sprintf("\n[Context %d from %s (relevance: %.2f)]\n%s\n",
			i+1, result.Source, result.Score, result.Text)
	}
	return formatted
}

// GetConversation retrieves a conversation with messages
func (s *ChatService) GetConversation(profileID, conversationID uuid.UUID) (*models.ChatConversation, error) {
	var conv models.ChatConversation
	if err := s.db.Where("id = ? AND profile_id = ?", conversationID, profileID).
		Preload("Messages").
		First(&conv).Error; err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListConversations lists all conversations for a profile
func (s *ChatService) ListConversations(profileID uuid.UUID) ([]models.ChatConversation, error) {
	var convs []models.ChatConversation
	if err := s.db.Where("profile_id = ?", profileID).
		Order("last_message_at DESC NULLS LAST, created_at DESC").
		Find(&convs).Error; err != nil {
		return nil, err
	}
	return convs, nil
}

// SaveConversation marks a conversation as saved (removes expiry)
func (s *ChatService) SaveConversation(profileID, conversationID uuid.UUID) error {
	return s.db.Model(&models.ChatConversation{}).
		Where("id = ? AND profile_id = ?", conversationID, profileID).
		Updates(map[string]interface{}{
			"is_saved":   true,
			"expires_at": nil,
		}).Error
}

// DeleteConversation deletes a conversation
func (s *ChatService) DeleteConversation(profileID, conversationID uuid.UUID) error {
	return s.db.Where("id = ? AND profile_id = ?", conversationID, profileID).
		Delete(&models.ChatConversation{}).Error
}
