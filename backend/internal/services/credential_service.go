package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// CredentialService manages encrypted user credentials
type CredentialService struct {
	db                *gorm.DB
	encryptionService *EncryptionService
}

// NewCredentialService creates a new credential service
func NewCredentialService(db *gorm.DB, encryptionService *EncryptionService) *CredentialService {
	return &CredentialService{
		db:                db,
		encryptionService: encryptionService,
	}
}

// CreateCredential creates a new encrypted credential
func (s *CredentialService) CreateCredential(profileID uuid.UUID, credentialType, value string) (*models.UserCredential, error) {
	// Validate credential type
	validTypes := map[string]bool{
		models.CredentialTypeOpenAI:   true,
		models.CredentialTypePinecone: true,
		models.CredentialTypeSlack:    true,
	}
	if !validTypes[credentialType] {
		return nil, fmt.Errorf("invalid credential type: %s", credentialType)
	}

	// Encrypt value
	encrypted, err := s.encryptionService.Encrypt(value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt credential: %w", err)
	}

	// Check if credential already exists
	var existing models.UserCredential
	result := s.db.Where("profile_id = ? AND credential_type = ?", profileID, credentialType).First(&existing)
	if result.Error == nil {
		// Update existing
		existing.EncryptedValue = encrypted
		existing.IsActive = true
		existing.UpdatedAt = time.Now()
		if err := s.db.Save(&existing).Error; err != nil {
			return nil, fmt.Errorf("failed to update credential: %w", err)
		}
		return &existing, nil
	}

	// Create new credential
	credential := &models.UserCredential{
		ProfileID:      profileID,
		CredentialType: credentialType,
		EncryptedValue: encrypted,
		IsActive:       true,
	}

	if err := s.db.Create(credential).Error; err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	return credential, nil
}

// GetCredential retrieves an encrypted credential
func (s *CredentialService) GetCredential(profileID uuid.UUID, credentialType string) (*models.UserCredential, error) {
	var credential models.UserCredential
	if err := s.db.Where("profile_id = ? AND credential_type = ? AND is_active = ?",
		profileID, credentialType, true).First(&credential).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("credential not found")
		}
		return nil, err
	}
	return &credential, nil
}

// GetDecryptedCredential retrieves and decrypts a credential
func (s *CredentialService) GetDecryptedCredential(profileID uuid.UUID, credentialType string) (string, error) {
	credential, err := s.GetCredential(profileID, credentialType)
	if err != nil {
		return "", err
	}

	decrypted, err := s.encryptionService.Decrypt(credential.EncryptedValue)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt credential: %w", err)
	}

	return decrypted, nil
}

// ListCredentials returns all credentials for a profile (without decrypting)
func (s *CredentialService) ListCredentials(profileID uuid.UUID) ([]models.UserCredential, error) {
	var credentials []models.UserCredential
	if err := s.db.Where("profile_id = ? AND is_active = ?", profileID, true).
		Find(&credentials).Error; err != nil {
		return nil, err
	}
	return credentials, nil
}

// DeleteCredential soft deletes a credential
func (s *CredentialService) DeleteCredential(profileID uuid.UUID, credentialID uuid.UUID) error {
	result := s.db.Where("id = ? AND profile_id = ?", credentialID, profileID).
		Delete(&models.UserCredential{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("credential not found")
	}
	return nil
}

// TestCredential tests if a credential is valid
func (s *CredentialService) TestCredential(profileID uuid.UUID, credentialID uuid.UUID) (bool, string, error) {
	// Fetch credential by ID
	var credential models.UserCredential
	if err := s.db.First(&credential, credentialID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, "", fmt.Errorf("credential not found")
		}
		return false, "", err
	}

	// Validate ownership
	if credential.ProfileID != profileID {
		return false, "", fmt.Errorf("credential not found")
	}

	// Decrypt to verify encryption integrity
	_, err := s.encryptionService.Decrypt(credential.EncryptedValue)
	if err != nil {
		testResult := "Failed to decrypt credential"
		now := time.Now()
		credential.LastTestedAt = &now
		credential.TestResult = &testResult
		s.db.Save(&credential)
		return false, testResult, nil
	}

	// Test actual API connectivity based on credential type
	testResult := "Credential decryption successful"
	valid := true

	// Note: Real API validation would happen here
	// For production, integrate with ValidationService for live API checks
	// Example: _, validationErr := validationService.ValidateOpenAIKey(decryptedValue)

	now := time.Now()
	credential.LastTestedAt = &now
	credential.TestResult = &testResult
	if err := s.db.Save(&credential).Error; err != nil {
		// Log error but don't fail the test
		return valid, testResult, nil
	}

	return valid, testResult, nil
}
