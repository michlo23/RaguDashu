package database

import (
	"fmt"
	"log"

	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establishes a connection to the database
func Connect(databaseURL string, isDevelopment bool) (*gorm.DB, error) {
	// Configure logger
	logLevel := logger.Silent
	if isDevelopment {
		logLevel = logger.Info
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Successfully connected to database")

	// Run migrations
	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return db, nil
}

// Migrate runs auto-migrations for all models
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate all models in correct order (respecting foreign keys)
	err := db.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.UserCredential{},
		&models.PineconeIndex{},
		&models.Document{},
		&models.DocumentChunk{},
		&models.DocumentFolder{},
		&models.DocumentTag{},
		&models.SlackThread{},
		&models.SlackMessage{},
		&models.ChatConfiguration{},
		&models.ChatConversation{},
		&models.ChatMessage{},
		&models.UsageMetrics{},
		&models.AuditLog{},
		&models.Webhook{},
		&models.ConversationTemplate{},
		&models.UserSession{},
		&models.TwoFactorAuth{},
	)
	if err != nil {
		return fmt.Errorf("auto-migration failed: %w", err)
	}

	// Create indexes for performance
	if err := createIndexes(db); err != nil {
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createIndexes creates additional database indexes for performance
func createIndexes(db *gorm.DB) error {
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_user_profiles_user_id ON user_profiles(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_credentials_profile_id ON user_credentials(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_pinecone_indexes_profile_id ON pinecone_indexes(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_documents_profile_id ON documents(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_documents_index_id ON documents(index_id);",
		"CREATE INDEX IF NOT EXISTS idx_document_chunks_document_id ON document_chunks(document_id);",
		"CREATE INDEX IF NOT EXISTS idx_document_folders_profile_id ON document_folders(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_document_tags_profile_id ON document_tags(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_slack_threads_profile_id ON slack_threads(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_slack_messages_thread_id ON slack_messages(thread_id);",
		"CREATE INDEX IF NOT EXISTS idx_chat_configurations_profile_id ON chat_configurations(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_chat_conversations_profile_id ON chat_conversations(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_chat_conversations_expires_at ON chat_conversations(expires_at) WHERE expires_at IS NOT NULL;",
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_conversation_id ON chat_messages(conversation_id);",
		"CREATE INDEX IF NOT EXISTS idx_usage_metrics_profile_id ON usage_metrics(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_usage_metrics_date ON usage_metrics(date);",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_profile_id ON audit_logs(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);",
		"CREATE INDEX IF NOT EXISTS idx_webhooks_profile_id ON webhooks(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_conversation_templates_profile_id ON conversation_templates(profile_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);",
		"CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);",
		"CREATE INDEX IF NOT EXISTS idx_two_factor_auth_user_id ON two_factor_auth(user_id);",
	}

	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	return nil
}
