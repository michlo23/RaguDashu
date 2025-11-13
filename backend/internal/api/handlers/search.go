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
	profileID, _ := middleware.GetProfileID(c)

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

	if req.TopK == 0 {
		req.TopK = 10
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
