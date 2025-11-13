package services

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"github.com/michlo23/rag-dashboard/internal/utils"
	"gorm.io/gorm"
)

// DocumentService handles document processing and storage
type DocumentService struct {
	db                *gorm.DB
	embeddingService  *EmbeddingService
	pineconeService   *PineconeService
	credentialService *CredentialService
	pdfService        *PDFService
	analyticsService  *AnalyticsService
	webhookService    *WebhookService
}

// NewDocumentService creates a new document service
func NewDocumentService(
	db *gorm.DB,
	embeddingService *EmbeddingService,
	pineconeService *PineconeService,
	credentialService *CredentialService,
	pdfService *PDFService,
	analyticsService *AnalyticsService,
	webhookService *WebhookService,
) *DocumentService {
	return &DocumentService{
		db:                db,
		embeddingService:  embeddingService,
		pineconeService:   pineconeService,
		credentialService: credentialService,
		pdfService:        pdfService,
		analyticsService:  analyticsService,
		webhookService:    webhookService,
	}
}

// UploadDocument processes and stores a document
func (s *DocumentService) UploadDocument(
	profileID, indexID uuid.UUID,
	filename, fileType string,
	fileSize int64,
	content io.Reader,
) (*models.Document, error) {
	// Read file content
	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var text string
	var metadata map[string]interface{}

	// Extract text based on file type
	if fileType == "application/pdf" || strings.HasSuffix(filename, ".pdf") {
		// Extract text from PDF
		if s.pdfService != nil {
			text, err = s.pdfService.ExtractText(contentBytes)
			if err != nil {
				text = string(contentBytes) // Fallback to raw content
			} else {
				// Extract PDF metadata
				if pdfMeta, err := s.pdfService.ExtractMetadata(contentBytes); err == nil {
					metadata = make(map[string]interface{})
					for k, v := range pdfMeta {
						metadata[k] = v
					}
				}
			}
		} else {
			text = string(contentBytes)
		}

	} else {
		// Plain text file
		text = string(contentBytes)
	}

	// Create document record with metadata
	var metadataJSON []byte
	if metadata != nil {
		metadataJSON, _ = json.Marshal(metadata)
	}

	doc := &models.Document{
		ProfileID:    profileID,
		IndexID:      indexID,
		Filename:     filename,
		FileType:     fileType,
		FileSize:     fileSize,
		UploadStatus: models.DocumentStatusProcessing,
		Metadata:     metadataJSON,
	}

	if err := s.db.Create(doc).Error; err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Track analytics
	if s.analyticsService != nil {
		s.analyticsService.TrackDocumentUpload(profileID)
	}

	// Trigger webhook
	if s.webhookService != nil {
		webhookData := map[string]interface{}{
			"document_id": doc.ID.String(),
			"filename":    filename,
			"file_size":   fileSize,
		}
		s.webhookService.Trigger(profileID, "document.uploaded", webhookData)
	}

	// Process asynchronously (in a real app, use a queue)
	go s.processDocument(doc.ID, profileID, text)

	return doc, nil
}

// processDocument chunks, embeds, and uploads to Pinecone
func (s *DocumentService) processDocument(docID, profileID uuid.UUID, text string) {
	var doc models.Document
	if err := s.db.First(&doc, docID).Error; err != nil {
		return
	}

	// Chunk text
	chunks := utils.ChunkText(text, 500, 50)
	if len(chunks) == 0 {
		errMsg := "No content to process"
		doc.ErrorMessage = &errMsg
		doc.UploadStatus = models.DocumentStatusFailed
		s.db.Save(&doc)
		return
	}

	// Get credentials
	openAIKey, err := s.credentialService.GetDecryptedCredential(profileID, models.CredentialTypeOpenAI)
	if err != nil {
		errMsg := "OpenAI API key not found"
		doc.ErrorMessage = &errMsg
		doc.UploadStatus = models.DocumentStatusFailed
		s.db.Save(&doc)
		return
	}

	pineconeKey, err := s.credentialService.GetDecryptedCredential(profileID, models.CredentialTypePinecone)
	if err != nil {
		errMsg := "Pinecone API key not found"
		doc.ErrorMessage = &errMsg
		doc.UploadStatus = models.DocumentStatusFailed
		s.db.Save(&doc)
		return
	}

	// Generate embeddings
	embeddings, err := s.embeddingService.GenerateBatchEmbeddings(chunks, openAIKey)
	if err != nil {
		errMsg := utils.GetGenericErrorMessage(err, "Embedding generation failed")
		doc.ErrorMessage = &errMsg
		doc.UploadStatus = models.DocumentStatusFailed
		s.db.Save(&doc)
		return
	}

	// Prepare vectors for Pinecone
	vectors := make([]Vector, len(chunks))
	vectorIDs := make([]string, len(chunks))
	for i, chunk := range chunks {
		vectorID := GenerateVectorID(fmt.Sprintf("doc_%s", docID.String()))
		vectorIDs[i] = vectorID
		vectors[i] = Vector{
			ID:     vectorID,
			Values: embeddings[i],
			Metadata: map[string]interface{}{
				"document_id": docID.String(),
				"filename":    doc.Filename,
				"chunk_index": i,
				"text":        chunk,
				"source_type": "document",
			},
		}

		// Save chunk to database
		dbChunk := &models.DocumentChunk{
			DocumentID: docID,
			ChunkIndex: i,
			ChunkText:  chunk,
			VectorID:   vectorID,
		}
		s.db.Create(dbChunk)
	}

	// Get index info
	var index models.PineconeIndex
	if err := s.db.First(&index, doc.IndexID).Error; err != nil {
		errMsg := "Index not found"
		doc.ErrorMessage = &errMsg
		doc.UploadStatus = models.DocumentStatusFailed
		s.db.Save(&doc)
		return
	}

	// Get profile for namespace
	var profile models.UserProfile
	if err := s.db.First(&profile, profileID).Error; err != nil {
		errMsg := "Profile not found"
		doc.ErrorMessage = &errMsg
		doc.UploadStatus = models.DocumentStatusFailed
		s.db.Save(&doc)
		return
	}

	// Upload to Pinecone (TODO: Use actual index host from config)
	indexHost := fmt.Sprintf("%s.svc.pinecone.io", index.IndexName)
	if err := s.pineconeService.UpsertVectors(indexHost, pineconeKey, profile.PineconeNamespace, vectors); err != nil {
		errMsg := utils.GetGenericErrorMessage(err, "Vector upload failed")
		doc.ErrorMessage = &errMsg
		doc.UploadStatus = models.DocumentStatusFailed
		s.db.Save(&doc)
		return
	}

	// Update document status
	doc.VectorIDs = vectorIDs
	doc.ChunkCount = len(chunks)
	doc.UploadStatus = models.DocumentStatusCompleted
	s.db.Save(&doc)

	// Update index document count
	s.db.Model(&index).Update("document_count", gorm.Expr("document_count + 1"))
}

// GetDocument retrieves a document by ID
func (s *DocumentService) GetDocument(profileID, docID uuid.UUID) (*models.Document, error) {
	var doc models.Document
	if err := s.db.Where("id = ? AND profile_id = ?", docID, profileID).
		First(&doc).Error; err != nil {
		return nil, err
	}
	return &doc, nil
}

// ListDocuments returns all documents for a profile
func (s *DocumentService) ListDocuments(profileID uuid.UUID) ([]models.Document, error) {
	var docs []models.Document
	if err := s.db.Where("profile_id = ?", profileID).
		Order("created_at DESC").
		Find(&docs).Error; err != nil {
		return nil, err
	}
	return docs, nil
}

// DeleteDocument deletes a document and its vectors
func (s *DocumentService) DeleteDocument(profileID, docID uuid.UUID) error {
	var doc models.Document
	if err := s.db.Where("id = ? AND profile_id = ?", docID, profileID).
		First(&doc).Error; err != nil {
		return err
	}

	// Delete from Pinecone if vectors exist
	if len(doc.VectorIDs) > 0 {
		pineconeKey, err := s.credentialService.GetDecryptedCredential(profileID, models.CredentialTypePinecone)
		if err == nil {
			var index models.PineconeIndex
			if err := s.db.First(&index, doc.IndexID).Error; err == nil {
				var profile models.UserProfile
				if err := s.db.First(&profile, profileID).Error; err == nil {
					indexHost := fmt.Sprintf("%s.svc.pinecone.io", index.IndexName)
					s.pineconeService.DeleteVectors(indexHost, pineconeKey, profile.PineconeNamespace, doc.VectorIDs)
				}
			}
		}
	}

	// Delete from database
	return s.db.Delete(&doc).Error
}

// ExtractTextFromPDF extracts text from PDF (placeholder)
func ExtractTextFromPDF(content []byte) (string, error) {
	// TODO: Implement PDF text extraction
	// Could use libraries like pdfcpu or call external service
	return "", fmt.Errorf("PDF extraction not yet implemented")
}
