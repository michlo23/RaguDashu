package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/api/middleware"
	"github.com/michlo23/rag-dashboard/internal/models"
	"github.com/michlo23/rag-dashboard/internal/services"
)

type DocumentHandler struct {
	documentService *services.DocumentService
}

func NewDocumentHandler(documentService *services.DocumentService) *DocumentHandler {
	return &DocumentHandler{documentService: documentService}
}

type DocumentResponse struct {
	ID           string   `json:"id"`
	Filename     string   `json:"filename"`
	FileType     string   `json:"file_type"`
	FileSize     int64    `json:"file_size"`
	ChunkCount   int      `json:"chunk_count"`
	UploadStatus string   `json:"upload_status"`
	ErrorMessage *string  `json:"error_message"`
	CreatedAt    string   `json:"created_at"`
}

// UploadDocument handles document upload
func (h *DocumentHandler) UploadDocument(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	// Get index_id from form
	indexIDStr := c.PostForm("index_id")
	if indexIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "index_id is required"})
		return
	}

	indexID, err := uuid.Parse(indexIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid index_id"})
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	// Validate file size (max 100MB)
	const maxFileSize = 100 * 1024 * 1024 // 100MB
	if file.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File size exceeds maximum allowed size of 100MB",
			"max_size_mb": 100,
			"file_size_mb": float64(file.Size) / (1024 * 1024),
		})
		return
	}

	// Validate file size is not zero
	if file.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File cannot be empty"})
		return
	}

	// Open file
	fileContent, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer fileContent.Close()

	// Process document
	doc, err := h.documentService.UploadDocument(
		profileID,
		indexID,
		file.Filename,
		file.Header.Get("Content-Type"),
		file.Size,
		fileContent,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toDocumentResponse(doc))
}

// ListDocuments lists all documents for the user
func (h *DocumentHandler) ListDocuments(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	docs, err := h.documentService.ListDocuments(profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
		return
	}

	responses := make([]DocumentResponse, len(docs))
	for i, doc := range docs {
		responses[i] = toDocumentResponse(&doc)
	}

	c.JSON(http.StatusOK, responses)
}

// GetDocument retrieves a single document
func (h *DocumentHandler) GetDocument(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	doc, err := h.documentService.GetDocument(profileID, docID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, toDocumentResponse(doc))
}

// DeleteDocument deletes a document
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	docID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	if err := h.documentService.DeleteDocument(profileID, docID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

func toDocumentResponse(doc *models.Document) DocumentResponse {
	return DocumentResponse{
		ID:           doc.ID.String(),
		Filename:     doc.Filename,
		FileType:     doc.FileType,
		FileSize:     doc.FileSize,
		ChunkCount:   doc.ChunkCount,
		UploadStatus: doc.UploadStatus,
		ErrorMessage: doc.ErrorMessage,
		CreatedAt:    doc.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
