package api

import (
	"github.com/gin-gonic/gin"
	"github.com/michlo23/rag-dashboard/internal/api/handlers"
	"github.com/michlo23/rag-dashboard/internal/api/middleware"
	"github.com/michlo23/rag-dashboard/internal/config"
	"github.com/michlo23/rag-dashboard/internal/services"
	"gorm.io/gorm"
)

// SetupRoutes configures all API routes
func SetupRoutes(
	router *gin.Engine,
	cfg *config.Config,
	db *gorm.DB,
	authService *services.AuthService,
	credentialService *services.CredentialService,
	documentService *services.DocumentService,
	searchService *services.SearchService,
	chatService *services.ChatService,
	analyticsService *services.AnalyticsService,
	webhookService *services.WebhookService,
	validationService *services.ValidationService,
	auditService *services.AuditService,
	cacheService *services.CacheService,
	mcpService *services.MCPService,
	oauth2Service *services.OAuth2Service,
) {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	credentialHandler := handlers.NewCredentialHandler(credentialService)
	documentHandler := handlers.NewDocumentHandler(documentService)
	searchHandler := handlers.NewSearchHandler(searchService)
	chatHandler := handlers.NewChatHandler(chatService, db)
	systemHandler := handlers.NewSystemHandler(db)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	webhookHandler := handlers.NewWebhookHandler(webhookService)
	mcpHandler := handlers.NewMCPHandler(mcpService, oauth2Service)

	// Apply middlewares
	router.Use(middleware.CORSMiddleware(cfg.FrontendURL))
	router.Use(middleware.RateLimitMiddleware(cfg.RateLimitRequests, cfg.RateLimitWindowMinutes))

	// Public routes
	api := router.Group("/api")
	{
		// System (public health check only)
		api.GET("/health", systemHandler.Health)

		// Auth
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
		}
	}

	// Protected routes
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// System (protected endpoints)
		protected.GET("/status", systemHandler.Status)
		protected.GET("/admin/status", systemHandler.DetailedStatus)

		// Auth
		authGroup := protected.Group("/auth")
		{
			authGroup.GET("/me", authHandler.GetMe)
			authGroup.POST("/logout", authHandler.Logout)
		}

		// Credentials
		credentials := protected.Group("/credentials")
		{
			credentials.POST("", credentialHandler.CreateCredential)
			credentials.GET("", credentialHandler.ListCredentials)
			credentials.DELETE("/:id", credentialHandler.DeleteCredential)
			credentials.POST("/:id/test", credentialHandler.TestCredential)
		}

		// Documents
		documents := protected.Group("/documents")
		{
			documents.POST("/upload", documentHandler.UploadDocument)
			documents.GET("", documentHandler.ListDocuments)
			documents.GET("/:id", documentHandler.GetDocument)
			documents.DELETE("/:id", documentHandler.DeleteDocument)
		}

		// Search
		search := protected.Group("/search")
		{
			search.POST("", searchHandler.Search)
		}

		// Chat
		chat := protected.Group("/chat")
		{
			chat.POST("/send", chatHandler.SendMessage)
			chat.POST("/configurations", chatHandler.CreateConfiguration)
			chat.GET("/configurations", chatHandler.ListConfigurations)
			chat.GET("/conversations", chatHandler.ListConversations)
			chat.GET("/conversations/:id", chatHandler.GetConversation)
			chat.POST("/conversations/:id/save", chatHandler.SaveConversation)
			chat.DELETE("/conversations/:id", chatHandler.DeleteConversation)
		}

		// Analytics
		analytics := protected.Group("/analytics")
		{
			analytics.GET("/dashboard", analyticsHandler.GetDashboardStats)
			analytics.GET("/usage-history", analyticsHandler.GetUsageHistory)
		}

		// Webhooks
		webhooks := protected.Group("/webhooks")
		{
			webhooks.POST("", webhookHandler.CreateWebhook)
			webhooks.GET("", webhookHandler.ListWebhooks)
			webhooks.DELETE("/:id", webhookHandler.DeleteWebhook)
		}

		// MCP (Model Context Protocol) Servers
		mcp := protected.Group("/mcp")
		{
			// Server management
			mcp.POST("/servers", mcpHandler.CreateMCPServer)
			mcp.GET("/servers", mcpHandler.ListMCPServers)
			mcp.GET("/servers/:id", mcpHandler.GetMCPServer)
			mcp.PUT("/servers/:id", mcpHandler.UpdateMCPServer)
			mcp.DELETE("/servers/:id", mcpHandler.DeleteMCPServer)

			// Capabilities and tools
			mcp.POST("/servers/:id/sync", mcpHandler.SyncCapabilities)
			mcp.POST("/servers/:id/tools/call", mcpHandler.CallTool)

			// OAuth 2 flow
			mcp.POST("/servers/:id/oauth2/initiate", mcpHandler.InitiateOAuth2)
			mcp.POST("/servers/:id/oauth2/refresh", mcpHandler.RefreshToken)
		}
	}

	// Public OAuth 2 callback (doesn't require authentication)
	api.GET("/mcp/oauth2/callback", mcpHandler.HandleOAuth2Callback)
}
