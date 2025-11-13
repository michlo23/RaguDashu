package services

import (
	"time"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// AnalyticsService handles usage metrics and analytics
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// TrackDocumentUpload increments document count
func (s *AnalyticsService) TrackDocumentUpload(profileID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)

	return s.db.Transaction(func(tx *gorm.DB) error {
		var metrics models.UsageMetrics
		err := tx.Where("profile_id = ? AND date = ?", profileID, today).
			FirstOrCreate(&metrics, models.UsageMetrics{
				ProfileID: profileID,
				Date:      today,
			}).Error
		if err != nil {
			return err
		}

		return tx.Model(&metrics).UpdateColumn("documents_count", gorm.Expr("documents_count + 1")).Error
	})
}

// TrackSearch increments search count
func (s *AnalyticsService) TrackSearch(profileID uuid.UUID) error {
	today := time.Now().Truncate(24 * time.Hour)

	return s.db.Transaction(func(tx *gorm.DB) error {
		var metrics models.UsageMetrics
		err := tx.Where("profile_id = ? AND date = ?", profileID, today).
			FirstOrCreate(&metrics, models.UsageMetrics{
				ProfileID: profileID,
				Date:      today,
			}).Error
		if err != nil {
			return err
		}

		return tx.Model(&metrics).UpdateColumn("search_count", gorm.Expr("search_count + 1")).Error
	})
}

// TrackChat increments chat count and token usage
func (s *AnalyticsService) TrackChat(profileID uuid.UUID, tokensUsed int) error {
	today := time.Now().Truncate(24 * time.Hour)

	return s.db.Transaction(func(tx *gorm.DB) error {
		var metrics models.UsageMetrics
		err := tx.Where("profile_id = ? AND date = ?", profileID, today).
			FirstOrCreate(&metrics, models.UsageMetrics{
				ProfileID: profileID,
				Date:      today,
			}).Error
		if err != nil {
			return err
		}

		return tx.Model(&metrics).Updates(map[string]interface{}{
			"chat_count":  gorm.Expr("chat_count + 1"),
			"tokens_used": gorm.Expr("tokens_used + ?", tokensUsed),
		}).Error
	})
}

// GetDashboardStats returns overview statistics
func (s *AnalyticsService) GetDashboardStats(profileID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total documents
	var totalDocs int64
	s.db.Model(&models.Document{}).Where("profile_id = ?", profileID).Count(&totalDocs)
	stats["total_documents"] = totalDocs

	// Total conversations
	var totalConvs int64
	s.db.Model(&models.ChatConversation{}).Where("profile_id = ?", profileID).Count(&totalConvs)
	stats["total_conversations"] = totalConvs

	// Today's usage
	today := time.Now().Truncate(24 * time.Hour)
	var todayMetrics models.UsageMetrics
	s.db.Where("profile_id = ? AND date = ?", profileID, today).First(&todayMetrics)

	stats["today_searches"] = todayMetrics.SearchCount
	stats["today_chats"] = todayMetrics.ChatCount
	stats["today_tokens"] = todayMetrics.TokensUsed

	// Last 7 days metrics
	sevenDaysAgo := today.AddDate(0, 0, -7)
	var weekMetrics []models.UsageMetrics
	s.db.Where("profile_id = ? AND date >= ?", profileID, sevenDaysAgo).
		Order("date ASC").
		Find(&weekMetrics)
	stats["week_metrics"] = weekMetrics

	// Last 30 days totals
	thirtyDaysAgo := today.AddDate(0, 0, -30)
	var monthTotals struct {
		TotalSearches int
		TotalChats    int
		TotalTokens   int
	}
	s.db.Model(&models.UsageMetrics{}).
		Select("SUM(search_count) as total_searches, SUM(chat_count) as total_chats, SUM(tokens_used) as total_tokens").
		Where("profile_id = ? AND date >= ?", profileID, thirtyDaysAgo).
		Scan(&monthTotals)

	stats["month_searches"] = monthTotals.TotalSearches
	stats["month_chats"] = monthTotals.TotalChats
	stats["month_tokens"] = monthTotals.TotalTokens

	// Estimate costs (OpenAI pricing)
	// gpt-4: $0.03/1K input tokens, $0.06/1K output tokens (avg ~$0.045/1K)
	// text-embedding-ada-002: $0.0001/1K tokens
	estimatedCost := float64(monthTotals.TotalTokens) / 1000.0 * 0.045
	stats["estimated_monthly_cost"] = estimatedCost

	return stats, nil
}

// GetUsageHistory returns usage metrics for a date range
func (s *AnalyticsService) GetUsageHistory(profileID uuid.UUID, startDate, endDate time.Time) ([]models.UsageMetrics, error) {
	var metrics []models.UsageMetrics
	err := s.db.Where("profile_id = ? AND date >= ? AND date <= ?", profileID, startDate, endDate).
		Order("date ASC").
		Find(&metrics).Error
	return metrics, err
}
