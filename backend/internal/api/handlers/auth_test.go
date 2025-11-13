package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/michlo23/rag-dashboard/internal/models"
	"github.com/michlo23/rag-dashboard/internal/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Migrate tables
	err = db.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.UserCredential{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	return db
}

func TestAuthHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	authService := services.NewAuthService(db, "test-secret-key-32-characters-long", 60, 30)
	handler := NewAuthHandler(authService)

	router := gin.New()
	router.POST("/register", handler.Register)

	tests := []struct {
		name           string
		payload        RegisterRequest
		expectedStatus int
	}{
		{
			name: "Valid registration",
			payload: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid email",
			payload: RegisterRequest{
				Email:    "invalid-email",
				Password: "password123",
				Name:     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Short password",
			payload: RegisterRequest{
				Email:    "test2@example.com",
				Password: "short",
				Name:     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Duplicate email",
			payload: RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := setupTestDB(t)

	authService := services.NewAuthService(db, "test-secret-key-32-characters-long", 60, 30)
	handler := NewAuthHandler(authService)

	// Create test user
	_, err := authService.Register("test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	router := gin.New()
	router.POST("/login", handler.Login)

	tests := []struct {
		name           string
		payload        LoginRequest
		expectedStatus int
	}{
		{
			name: "Valid login",
			payload: LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid password",
			payload: LoginRequest{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "Non-existent user",
			payload: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d. Body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			// Verify token structure if login successful
			if tt.expectedStatus == http.StatusOK {
				var response TokenResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if response.AccessToken == "" || response.RefreshToken == "" {
					t.Error("Missing tokens in response")
				}
			}
		})
	}
}
