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

// ChatService handles AI chat with RAG and MCP tools
type ChatService struct {
	db                *gorm.DB
	searchService     *SearchService
	credentialService *CredentialService
	mcpService        *MCPService
}

// NewChatService creates a new chat service
func NewChatService(
	db *gorm.DB,
	searchService *SearchService,
	credentialService *CredentialService,
	mcpService *MCPService,
) *ChatService {
	return &ChatService{
		db:                db,
		searchService:     searchService,
		credentialService: credentialService,
		mcpService:        mcpService,
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

	// Get OpenAI API key
	openAIKey, err := s.credentialService.GetDecryptedCredential(profileID, models.CredentialTypeOpenAI)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API key not found")
	}

	// Check if MCP tools are enabled and fetch available tools
	var mcpTools []OpenAIFunction
	var enabledServerIDs []uuid.UUID
	if len(config.EnabledMCPTools) > 0 {
		if err := json.Unmarshal(config.EnabledMCPTools, &enabledServerIDs); err == nil && len(enabledServerIDs) > 0 {
			mcpTools, err = s.getMCPToolsAsOpenAIFunctions(profileID, enabledServerIDs)
			if err != nil {
				// Log error but continue without MCP tools
				fmt.Printf("Warning: Failed to fetch MCP tools: %v\n", err)
			}
		}
	}

	// Call OpenAI with MCP tools (may require multiple iterations for function calling)
	assistantReply, toolUsages, err := s.callOpenAIWithTools(config.Model, messages, config.Temperature, config.MaxTokens, openAIKey, mcpTools, profileID, enabledServerIDs, conversation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenAI: %w", err)
	}

	// Save messages to database
	userMessageID, err := s.saveMessagesWithTools(conversation.ID, req.Message, assistantReply, retrievedContext, toolUsages)
	if err != nil {
		return nil, err
	}

	// Save tool usage records
	if len(toolUsages) > 0 {
		s.saveToolUsageRecords(conversation.ID, userMessageID, toolUsages)
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

// OpenAIFunction represents an OpenAI function for function calling
type OpenAIFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	MCPServerID uuid.UUID              `json:"-"` // Not sent to OpenAI, used for routing
}

// ToolUsage represents a tool call during conversation
type ToolUsage struct {
	MCPServerID uuid.UUID
	ToolName    string
	Input       map[string]interface{}
	Output      interface{}
	Error       string
	DurationMs  int
}

// getMCPToolsAsOpenAIFunctions fetches MCP tools and converts them to OpenAI function format
func (s *ChatService) getMCPToolsAsOpenAIFunctions(profileID uuid.UUID, serverIDs []uuid.UUID) ([]OpenAIFunction, error) {
	var functions []OpenAIFunction

	for _, serverID := range serverIDs {
		server, err := s.mcpService.GetMCPServer(serverID, profileID)
		if err != nil || !server.IsActive {
			continue
		}

		// Parse available tools
		var tools []models.MCPTool
		if len(server.AvailableTools) > 0 {
			if err := json.Unmarshal(server.AvailableTools, &tools); err != nil {
				continue
			}
		}

		// Convert MCP tools to OpenAI function format
		for _, tool := range tools {
			fn := OpenAIFunction{
				Name:        fmt.Sprintf("%s_%s", serverID.String()[:8], tool.Name), // Prefix with server ID to avoid conflicts
				Description: tool.Description,
				Parameters:  tool.InputSchema,
				MCPServerID: serverID,
			}
			functions = append(functions, fn)
		}
	}

	return functions, nil
}

// callOpenAIWithTools calls OpenAI with function calling support
func (s *ChatService) callOpenAIWithTools(
	model string,
	messages []map[string]string,
	temperature float32,
	maxTokens int,
	apiKey string,
	functions []OpenAIFunction,
	profileID uuid.UUID,
	serverIDs []uuid.UUID,
	conversationID uuid.UUID,
) (string, []ToolUsage, error) {
	var toolUsages []ToolUsage
	maxIterations := 5 // Prevent infinite loops
	iteration := 0

	// Convert messages to interface{} format for OpenAI
	var openAIMessages []map[string]interface{}
	for _, msg := range messages {
		openAIMessages = append(openAIMessages, map[string]interface{}{
			"role":    msg["role"],
			"content": msg["content"],
		})
	}

	for iteration < maxIterations {
		iteration++

		// Build request body
		reqBody := map[string]interface{}{
			"model":       model,
			"messages":    openAIMessages,
			"temperature": temperature,
			"max_tokens":  maxTokens,
		}

		// Add functions if available
		if len(functions) > 0 {
			var openAIFunctions []map[string]interface{}
			for _, fn := range functions {
				openAIFunctions = append(openAIFunctions, map[string]interface{}{
					"name":        fn.Name,
					"description": fn.Description,
					"parameters":  fn.Parameters,
				})
			}
			reqBody["functions"] = openAIFunctions
			reqBody["function_call"] = "auto"
		}

		jsonData, err := json.Marshal(reqBody)
		if err != nil {
			return "", nil, err
		}

		req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
		if err != nil {
			return "", nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return "", nil, err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return "", nil, fmt.Errorf("OpenAI API error (status %d): %s", resp.StatusCode, string(body))
		}

		// Parse response
		var chatResp struct {
			Choices []struct {
				Message struct {
					Role         string                 `json:"role"`
					Content      string                 `json:"content"`
					FunctionCall *struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function_call"`
				} `json:"message"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal(body, &chatResp); err != nil {
			return "", nil, err
		}

		if len(chatResp.Choices) == 0 {
			return "", nil, fmt.Errorf("no response from OpenAI")
		}

		choice := chatResp.Choices[0]

		// Check if OpenAI wants to call a function
		if choice.Message.FunctionCall != nil {
			functionName := choice.Message.FunctionCall.Name
			var functionArgs map[string]interface{}
			if err := json.Unmarshal([]byte(choice.Message.FunctionCall.Arguments), &functionArgs); err != nil {
				return "", nil, fmt.Errorf("failed to parse function arguments: %w", err)
			}

			// Find the function definition to get MCP server ID
			var mcpServerID uuid.UUID
			var actualToolName string
			for _, fn := range functions {
				if fn.Name == functionName {
					mcpServerID = fn.MCPServerID
					// Extract actual tool name (remove server ID prefix)
					actualToolName = functionName[9:] // Remove "{serverID}_" prefix
					break
				}
			}

			if mcpServerID == uuid.Nil {
				return "", nil, fmt.Errorf("function not found: %s", functionName)
			}

			// Call MCP tool
			startTime := time.Now()
			toolResp, err := s.mcpService.CallTool(mcpServerID, profileID, actualToolName, functionArgs)
			duration := int(time.Since(startTime).Milliseconds())

			var toolResult interface{}
			var toolError string

			if err != nil || toolResp.IsError {
				if err != nil {
					toolError = err.Error()
				} else {
					toolError = fmt.Sprintf("%v", toolResp.Content)
				}
				toolResult = map[string]interface{}{"error": toolError}
			} else {
				toolResult = toolResp.Content
			}

			// Track tool usage
			toolUsages = append(toolUsages, ToolUsage{
				MCPServerID: mcpServerID,
				ToolName:    actualToolName,
				Input:       functionArgs,
				Output:      toolResult,
				Error:       toolError,
				DurationMs:  duration,
			})

			// Add function call and result to messages for next iteration
			openAIMessages = append(openAIMessages, map[string]interface{}{
				"role":          "assistant",
				"content":       nil,
				"function_call": choice.Message.FunctionCall,
			})
			openAIMessages = append(openAIMessages, map[string]interface{}{
				"role":    "function",
				"name":    functionName,
				"content": fmt.Sprintf("%v", toolResult),
			})

			// Continue to next iteration to get final response
			continue
		}

		// No function call, return the assistant's message
		return choice.Message.Content, toolUsages, nil
	}

	return "", nil, fmt.Errorf("maximum function call iterations reached")
}

// saveMessagesWithTools saves user and assistant messages with tool usage info
func (s *ChatService) saveMessagesWithTools(conversationID uuid.UUID, userMsg, assistantMsg string, context []models.SearchResult, toolUsages []ToolUsage) (uuid.UUID, error) {
	// Save user message
	contextJSON, _ := json.Marshal(context)
	userMessage := &models.ChatMessage{
		ConversationID:   conversationID,
		Role:             models.MessageRoleUser,
		Content:          userMsg,
		RetrievedContext: contextJSON,
	}
	if err := s.db.Create(userMessage).Error; err != nil {
		return uuid.Nil, err
	}

	// Save assistant message
	assistantMessage := &models.ChatMessage{
		ConversationID: conversationID,
		Role:           models.MessageRoleAssistant,
		Content:        assistantMsg,
	}
	if err := s.db.Create(assistantMessage).Error; err != nil {
		return uuid.Nil, err
	}

	return assistantMessage.ID, nil
}

// saveToolUsageRecords saves MCP tool usage records to database
func (s *ChatService) saveToolUsageRecords(conversationID, messageID uuid.UUID, toolUsages []ToolUsage) {
	for _, usage := range toolUsages {
		inputJSON, _ := json.Marshal(usage.Input)
		outputJSON, _ := json.Marshal(usage.Output)

		var errorMsg *string
		if usage.Error != "" {
			errorMsg = &usage.Error
		}

		record := &models.MCPToolUsage{
			ConversationID: conversationID,
			MessageID:      messageID,
			MCPServerID:    usage.MCPServerID,
			ToolName:       usage.ToolName,
			ToolInput:      inputJSON,
			ToolOutput:     outputJSON,
			ErrorMessage:   errorMsg,
			DurationMs:     &usage.DurationMs,
		}

		s.db.Create(record) // Ignore errors for analytics records
	}
}
