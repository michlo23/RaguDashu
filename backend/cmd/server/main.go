package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/michlo23/rag-dashboard/internal/api"
	"github.com/michlo23/rag-dashboard/internal/config"
	"github.com/michlo23/rag-dashboard/internal/database"
	"github.com/michlo23/rag-dashboard/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.Connect(cfg.DatabaseURL, cfg.Environment == "development")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Database connection established successfully")

	// Initialize services
	encryptionService, err := services.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		log.Fatalf("Failed to initialize encryption service: %v", err)
	}

	authService := services.NewAuthService(
		db,
		cfg.JWTSecret,
		cfg.JWTAccessExpireMin,
		cfg.JWTRefreshExpireDays,
	)

	credentialService := services.NewCredentialService(db, encryptionService)
	embeddingService := services.NewEmbeddingService()
	pineconeService := services.NewPineconeService()

	// Initialize new services
	pdfService := services.NewPDFService()
	analyticsService := services.NewAnalyticsService(db)
	validationService := services.NewValidationService()
	auditService := services.NewAuditService(db)
	webhookService := services.NewWebhookService(db)
	cacheService := services.NewCacheService(cfg.RedisURL)
	oauth2Service := services.NewOAuth2Service(db, encryptionService)
	mcpService := services.NewMCPService(db, oauth2Service, encryptionService)

	documentService := services.NewDocumentService(
		db,
		embeddingService,
		pineconeService,
		credentialService,
		pdfService,
		analyticsService,
		webhookService,
	)

	searchService := services.NewSearchService(
		db,
		embeddingService,
		pineconeService,
		credentialService,
	)

	chatService := services.NewChatService(
		db,
		searchService,
		credentialService,
	)

	// Start cleanup service
	cleanupService := services.NewCleanupService(db)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanupService.Start(ctx, cfg.CleanupIntervalHours)

	// Setup router
	router := gin.Default()

	// Setup routes
	api.SetupRoutes(
		router,
		cfg,
		db,
		authService,
		credentialService,
		documentService,
		searchService,
		chatService,
		analyticsService,
		webhookService,
		validationService,
		auditService,
		cacheService,
		mcpService,
		oauth2Service,
	)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", cfg.Port)
		log.Printf("Environment: %s", cfg.Environment)
		log.Printf("Frontend URL: %s", cfg.FrontendURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
