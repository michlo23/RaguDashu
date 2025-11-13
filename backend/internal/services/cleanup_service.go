package services

import (
	"context"
	"log"
	"time"

	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// CleanupService handles automatic cleanup of expired data
type CleanupService struct {
	db *gorm.DB
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(db *gorm.DB) *CleanupService {
	return &CleanupService{db: db}
}

// Start runs cleanup in background
func (s *CleanupService) Start(ctx context.Context, intervalHours int) {
	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)

	// Run immediately on start
	s.cleanupExpiredConversations()

	go func() {
		for {
			select {
			case <-ticker.C:
				s.cleanupExpiredConversations()
			case <-ctx.Done():
				ticker.Stop()
				log.Println("Cleanup service stopped")
				return
			}
		}
	}()

	log.Printf("Cleanup service started (running every %d hours)", intervalHours)
}

// cleanupExpiredConversations deletes conversations past their expiry date
func (s *CleanupService) cleanupExpiredConversations() {
	log.Println("Starting cleanup of expired conversations...")

	// Delete expired conversations (IsSaved=false AND ExpiresAt < now)
	result := s.db.
		Where("expires_at < ? AND is_saved = ?", time.Now(), false).
		Delete(&models.ChatConversation{})

	if result.Error != nil {
		log.Printf("Error cleaning up conversations: %v", result.Error)
		return
	}

	if result.RowsAffected > 0 {
		log.Printf("Cleaned up %d expired conversations", result.RowsAffected)
	} else {
		log.Println("No expired conversations to clean up")
	}
}
