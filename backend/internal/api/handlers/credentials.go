package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/api/middleware"
	"github.com/michlo23/rag-dashboard/internal/models"
	"github.com/michlo23/rag-dashboard/internal/services"
)

type CredentialHandler struct {
	credentialService *services.CredentialService
}

func NewCredentialHandler(credentialService *services.CredentialService) *CredentialHandler {
	return &CredentialHandler{credentialService: credentialService}
}

type CreateCredentialRequest struct {
	CredentialType string `json:"credential_type" binding:"required"`
	Value          string `json:"value" binding:"required"`
}

type CredentialResponse struct {
	ID             string  `json:"id"`
	CredentialType string  `json:"credential_type"`
	IsActive       bool    `json:"is_active"`
	LastTestedAt   *string `json:"last_tested_at"`
	TestResult     *string `json:"test_result"`
	CreatedAt      string  `json:"created_at"`
}

// CreateCredential creates or updates a credential
func (h *CredentialHandler) CreateCredential(c *gin.Context) {
	profileID, _ := middleware.GetProfileID(c)

	var req CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	credential, err := h.credentialService.CreateCredential(profileID, req.CredentialType, req.Value)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, toCredentialResponse(credential))
}

// ListCredentials lists all credentials for the user
func (h *CredentialHandler) ListCredentials(c *gin.Context) {
	profileID, _ := middleware.GetProfileID(c)

	credentials, err := h.credentialService.ListCredentials(profileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch credentials"})
		return
	}

	responses := make([]CredentialResponse, len(credentials))
	for i, cred := range credentials {
		responses[i] = toCredentialResponse(&cred)
	}

	c.JSON(http.StatusOK, responses)
}

// DeleteCredential deletes a credential
func (h *CredentialHandler) DeleteCredential(c *gin.Context) {
	profileID, _ := middleware.GetProfileID(c)
	credentialID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID"})
		return
	}

	if err := h.credentialService.DeleteCredential(profileID, credentialID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Credential deleted successfully"})
}

// TestCredential tests a credential's validity
func (h *CredentialHandler) TestCredential(c *gin.Context) {
	profileID, _ := middleware.GetProfileID(c)
	credentialID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid credential ID"})
		return
	}

	valid, result, err := h.credentialService.TestCredential(profileID, credentialID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":  valid,
		"result": result,
	})
}

func toCredentialResponse(cred *models.UserCredential) CredentialResponse {
	var lastTested *string
	if cred.LastTestedAt != nil {
		ts := cred.LastTestedAt.Format("2006-01-02 15:04:05")
		lastTested = &ts
	}

	return CredentialResponse{
		ID:             cred.ID.String(),
		CredentialType: cred.CredentialType,
		IsActive:       cred.IsActive,
		LastTestedAt:   lastTested,
		TestResult:     cred.TestResult,
		CreatedAt:      cred.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
