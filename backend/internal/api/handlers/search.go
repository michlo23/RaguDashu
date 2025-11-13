package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/api/middleware"
	"github.com/michlo23/rag-dashboard/internal/services"
)

type SearchHandler struct {
	searchService *services.SearchService
}

func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

type SearchRequest struct {
	Query   string `json:"query" binding:"required"`
	IndexID string `json:"index_id" binding:"required"`
	TopK    int    `json:"top_k"`
}

// Search performs semantic search
func (h *SearchHandler) Search(c *gin.Context) {
	profileID, ok := middleware.MustGetProfileID(c)
	if !ok {
		return
	}

	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	indexID, err := uuid.Parse(req.IndexID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid index_id"})
		return
	}

	// Set default TopK if not provided
	if req.TopK == 0 {
		req.TopK = 10
	}

	// Validate TopK bounds (1-100) to prevent DoS attacks
	if req.TopK < 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "top_k must be at least 1",
			"provided": req.TopK,
		})
		return
	}
	if req.TopK > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "top_k cannot exceed 100",
			"provided": req.TopK,
			"maximum": 100,
		})
		return
	}

	results, err := h.searchService.Search(profileID, indexID, req.Query, req.TopK)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results": results,
		"count":   len(results),
	})
}
