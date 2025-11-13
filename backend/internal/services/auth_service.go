package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/michlo23/rag-dashboard/internal/models"
	"gorm.io/gorm"
)

// AuthService handles authentication logic
type AuthService struct {
	db        *gorm.DB
	jwtSecret string
	accessExp time.Duration
	refreshExp time.Duration
}

// NewAuthService creates a new auth service
func NewAuthService(db *gorm.DB, jwtSecret string, accessExpMin, refreshExpDays int) *AuthService {
	return &AuthService{
		db:         db,
		jwtSecret:  jwtSecret,
		accessExp:  time.Duration(accessExpMin) * time.Minute,
		refreshExp: time.Duration(refreshExpDays) * 24 * time.Hour,
	}
}

// TokenClaims represents JWT claims
type TokenClaims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	ProfileID uuid.UUID `json:"profile_id"`
	IsAdmin   bool      `json:"is_admin"`
	jwt.RegisteredClaims
}

// Register creates a new user account
func (s *AuthService) Register(email, password, name string) (*models.User, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Create user
	user := &models.User{
		Email:    email,
		IsActive: true,
		IsAdmin:  false,
	}
	if err := user.HashPassword(password); err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create user
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default profile
	profile := &models.UserProfile{
		UserID:            user.ID,
		Name:              name,
		PineconeNamespace: fmt.Sprintf("user_%s", user.ID.String()),
		IsDefault:         true,
	}
	if err := tx.Create(profile).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return user, nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(email, password string) (accessToken, refreshToken string, err error) {
	// Find user
	var user models.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", err
	}

	// Check if user is active
	if !user.IsActive {
		return "", "", errors.New("account is inactive")
	}

	// Verify password
	if !user.CheckPassword(password) {
		return "", "", errors.New("invalid credentials")
	}

	// Get default profile
	var profile models.UserProfile
	if err := s.db.Where("user_id = ? AND is_default = ?", user.ID, true).First(&profile).Error; err != nil {
		return "", "", fmt.Errorf("no default profile found: %w", err)
	}

	// Generate tokens
	accessToken, err = s.generateToken(user, profile, s.accessExp)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.generateToken(user, profile, s.refreshExp)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *AuthService) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*TokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates new tokens from a refresh token
func (s *AuthService) RefreshToken(refreshToken string) (accessToken, newRefreshToken string, err error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	// Get user and profile
	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return "", "", errors.New("user not found")
	}

	var profile models.UserProfile
	if err := s.db.First(&profile, claims.ProfileID).Error; err != nil {
		return "", "", errors.New("profile not found")
	}

	// Check if user is still active
	if !user.IsActive {
		return "", "", errors.New("account is inactive")
	}

	// Generate new tokens
	accessToken, err = s.generateToken(user, profile, s.accessExp)
	if err != nil {
		return "", "", err
	}

	newRefreshToken, err = s.generateToken(user, profile, s.refreshExp)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

// generateToken creates a JWT token
func (s *AuthService) generateToken(user models.User, profile models.UserProfile, expiry time.Duration) (string, error) {
	claims := TokenClaims{
		UserID:    user.ID,
		Email:     user.Email,
		ProfileID: profile.ID,
		IsAdmin:   user.IsAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID uuid.UUID) (*models.User, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
