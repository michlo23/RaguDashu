package services

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// SearchService handles semantic search
type SearchService struct {
	db                *gorm.DB
	embeddingService  *EmbeddingService
	pineconeService   *PineconeService
	credentialService *CredentialService
}

// NewSearchService creates a new search service
func NewSearchService(
	db *gorm.DB,
	embeddingService *EmbeddingService,
	pineconeService *PineconeService,
	credentialService *CredentialService,
) *SearchService {
	return &SearchService{
		db:                db,
		embeddingService:  embeddingService,
		pineconeService:   pineconeService,
		credentialService: credentialService,
	}
}

// Search performs semantic search across user's data
func (s *SearchService) Search(profileID, indexID uuid.UUID, query string, topK int) ([]models.SearchResult, error) {
	if topK == 0 {
		topK = 10
	}

	// Get credentials
	openAIKey, err := s.credentialService.GetDecryptedCredential(profileID, models.CredentialTypeOpenAI)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API key not found")
	}

	pineconeKey, err := s.credentialService.GetDecryptedCredential(profileID, models.CredentialTypePinecone)
	if err != nil {
		return nil, fmt.Errorf("Pinecone API key not found")
	}

	// Generate query embedding
	queryEmbedding, err := s.embeddingService.GenerateEmbedding(query, openAIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Get index info
	var index models.PineconeIndex
	if err := s.db.Where("id = ? AND profile_id = ?", indexID, profileID).
		First(&index).Error; err != nil {
		return nil, fmt.Errorf("index not found")
	}

	// Get profile for namespace
	var profile models.UserProfile
	if err := s.db.First(&profile, profileID).Error; err != nil {
		return nil, fmt.Errorf("profile not found")
	}

	// Query Pinecone
	indexHost := fmt.Sprintf("%s.svc.pinecone.io", index.IndexName)
	queryResp, err := s.pineconeService.QueryVectors(
		indexHost,
		pineconeKey,
		profile.PineconeNamespace,
		queryEmbedding,
		topK,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query Pinecone: %w", err)
	}

	// Convert to search results
	results := make([]models.SearchResult, len(queryResp.Matches))
	for i, match := range queryResp.Matches {
		source := "Unknown"
		text := ""

		// Extract metadata
		if filename, ok := match.Metadata["filename"].(string); ok {
			source = filename
		}
		if textVal, ok := match.Metadata["text"].(string); ok {
			text = textVal
		}

		results[i] = models.SearchResult{
			ID:       match.ID,
			Score:    match.Score,
			Text:     text,
			Source:   source,
			Metadata: match.Metadata,
		}
	}

	return results, nil
}
