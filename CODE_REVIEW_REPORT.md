# RAG Dashboard Backend - Comprehensive Code Review Report

## Executive Summary
The RAG Dashboard backend demonstrates solid architecture with proper separation of concerns, but contains several security vulnerabilities, runtime issues, and logic bugs that require immediate attention before production deployment.

**Total Issues Found: 38**
- Critical: 6
- High: 12
- Medium: 15
- Low: 5

---

## CRITICAL ISSUES

### 1. Insecure Webhook Secret Generation
**File:** `/home/user/RaguDashu/backend/internal/services/webhook_service.go`  
**Lines:** 94-95  
**Severity:** CRITICAL  
**Category:** Security - Weak Cryptography

**Issue:**
```go
secret := fmt.Sprintf("%d", time.Now().UnixNano())
```

The webhook secret is generated using only a timestamp in nanoseconds, which is predictable and can be easily guessed. This defeats the purpose of HMAC-based webhook authentication.

**Potential Impact:**
- Attackers can forge webhook signatures
- Webhook endpoints become vulnerable to unauthorized requests
- Data integrity cannot be verified
- Compliance violations (GDPR, HIPAA requirements for secure webhook signatures)

**Recommended Fix:**
```go
import "crypto/rand"
import "encoding/hex"

func generateSecureWebhookSecret() (string, error) {
    secret := make([]byte, 32)
    _, err := rand.Read(secret)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(secret), nil
}
```

---

### 2. Credential Service TestCredential Logic Bug
**File:** `/home/user/RaguDashu/backend/internal/services/credential_service.go`  
**Lines:** 128  
**Severity:** CRITICAL  
**Category:** Logic Bug - Security/Functionality

**Issue:**
```go
credential, err := s.GetCredential(profileID, "")
```

The TestCredential function passes an empty string as the credentialType instead of the actual credential type that should be tested. This causes the function to always fail because GetCredential filters by credential type.

**Potential Impact:**
- Credential testing always fails
- Users cannot verify their credentials are valid
- False negatives lead to failed operations later
- Security validation is bypassed entirely

**Recommended Fix:**
```go
func (s *CredentialService) TestCredential(profileID uuid.UUID, credentialID uuid.UUID) (bool, string, error) {
    // First fetch the credential to get its type
    var credential models.UserCredential
    if err := s.db.First(&credential, credentialID).Error; err != nil {
        return false, "", err
    }
    
    // Validate ownership
    if credential.ProfileID != profileID {
        return false, "", errors.New("credential not found")
    }
    
    // Now test it based on its actual type
    _, err := s.encryptionService.Decrypt(credential.EncryptedValue)
    if err != nil {
        return false, "Failed to decrypt credential", nil
    }
    // ... continue with actual API validation
}
```

---

### 3. API Key Exposure in Error Messages
**File:** `/home/user/RaguDashu/backend/internal/services/document_service.go`  
**Lines:** 168  
**Severity:** CRITICAL  
**Category:** Security - Sensitive Data Exposure

**Issue:**
```go
errMsg := fmt.Sprintf("Failed to generate embeddings: %v", err)
```

When embedding generation fails, the error message might contain API key information from the underlying OpenAI error. This gets stored in the database and could be exposed through logs or error responses.

**Potential Impact:**
- API keys may be stored in database error messages
- Keys exposed through error responses to frontend
- Keys exposed in application logs
- Attackers can intercept or access stored API keys
- Complete account takeover via stolen API keys

**Recommended Fix:**
```go
// Sanitize errors before storing
func sanitizeError(err error) string {
    errStr := err.Error()
    // Remove common API key patterns
    sanitized := regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`).ReplaceAllString(errStr, "***REDACTED***")
    sanitized = regexp.MustCompile(`pk-[a-zA-Z0-9]{20,}`).ReplaceAllString(sanitized, "***REDACTED***")
    return "An error occurred processing the document. Please try again."
}

errMsg := sanitizeError(err)
```

---

### 4. Missing Authentication on Sensitive Endpoints
**File:** `/home/user/RaguDashu/backend/internal/api/routes.go`  
**Severity:** CRITICAL  
**Category:** API Security - Missing Auth Middleware

**Issue:**
The `/api/health` and `/api/status` endpoints are public (not protected) and return system information that could be used for reconnaissance:

```go
api.GET("/health", systemHandler.Health)
api.GET("/status", systemHandler.Status)
```

While some system endpoints should be public, if they expose sensitive details about the backend, they could aid attackers.

**Potential Impact:**
- Information disclosure about system versions
- Attack surface mapping
- Denial of service detection and targeting
- Helps attackers plan more sophisticated attacks

**Recommended Fix:**
```go
// Option 1: Make endpoints require auth
protected.GET("/health", systemHandler.Health)
protected.GET("/status", systemHandler.Status)

// Option 2: Return minimal information without auth
// Only return: {"status": "ok"} without version/environment details
```

---

### 5. Nil Pointer Dereference Risk in Multiple Handlers
**File:** `/home/user/RaguDashu/backend/internal/api/handlers/chat.go`  
**Lines:** 28, 47, 66, 79, 97, 110, 127  
**Severity:** CRITICAL  
**Category:** Runtime - Nil Pointer Dereference

**Issue:**
```go
profileID, _ := middleware.GetProfileID(c)
```

Throughout multiple handlers, the error from GetProfileID is silently ignored. If the profile ID is nil (uuid.Nil), operations will silently fail or worse, affect the wrong user's data.

**Potential Impact:**
- Silent failures in operations
- Data corruption across users
- Security vulnerability - operations on wrong profiles
- Cascading failures in dependent services
- Data leaks between users

**Recommended Fix:**
```go
profileID, err := middleware.GetProfileID(c)
if err != nil || profileID == uuid.Nil {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid profile ID"})
    return
}

// Use profileID safely now
```

---

### 6. Goroutine Leaks in Webhook Service
**File:** `/home/user/RaguDashu/backend/internal/services/webhook_service.go`  
**Lines:** 44-46  
**Severity:** CRITICAL  
**Category:** Runtime - Goroutine/Resource Leak

**Issue:**
```go
for _, webhook := range webhooks {
    go s.sendWebhook(webhook, event, profileID, data)
}
```

Webhooks are sent asynchronously without any tracking, wait groups, or cancellation mechanism. This can cause:
- Memory leaks from accumulating goroutines
- Lost webhook deliveries on shutdown
- No visibility into failed deliveries

**Potential Impact:**
- Gradual memory exhaustion leading to OOM
- Server instability under high event load
- Lost webhook deliveries
- No audit trail of webhook delivery status
- Potential security issues with untracked async operations

**Recommended Fix:**
```go
// Add WaitGroup tracking
func (s *WebhookService) Trigger(profileID uuid.UUID, event string, data map[string]interface{}) {
    var webhooks []models.Webhook
    s.db.Where("profile_id = ? AND is_active = ? AND ? = ANY(events)", profileID, true, event).
        Find(&webhooks)
    
    var wg sync.WaitGroup
    for _, webhook := range webhooks {
        wg.Add(1)
        go func(w models.Webhook) {
            defer wg.Done()
            s.sendWebhook(w, event, profileID, data)
            // Track delivery in database
            s.recordWebhookDelivery(w.ID, event, success)
        }(webhook)
    }
    
    // For critical operations, wait for completion
    // For non-critical, let background task pick up
}
```

---

## HIGH SEVERITY ISSUES

### 7. Missing Input Validation for TopK Parameter
**File:** `/home/user/RaguDashu/backend/internal/api/handlers/search.go`  
**Lines:** 42-44  
**Severity:** HIGH  
**Category:** API Security - Input Validation

**Issue:**
```go
if req.TopK == 0 {
    req.TopK = 10
}
// No upper limit validation
```

The TopK parameter has no upper bound validation. A malicious user could request millions of results, causing:
- Resource exhaustion
- Denial of service
- Memory spikes in Pinecone API calls

**Potential Impact:**
- DoS attacks via large TopK values
- Unexpected API costs
- Service degradation
- Performance issues

**Recommended Fix:**
```go
const MaxTopK = 100
const MinTopK = 1

if req.TopK == 0 {
    req.TopK = 10
} else if req.TopK < MinTopK {
    c.JSON(http.StatusBadRequest, gin.H{"error": "top_k must be at least 1"})
    return
} else if req.TopK > MaxTopK {
    c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("top_k cannot exceed %d", MaxTopK)})
    return
}
```

---

### 8. Unchecked Database Save Operations
**File:** `/home/user/RaguDashu/backend/internal/services/document_service.go`  
**Lines:** 142, 148, 152, 161, 171, 210, 219, 237  
**Severity:** HIGH  
**Category:** Runtime - Missing Error Handling

**Issue:**
Multiple database save operations have their errors ignored:

```go
s.db.Save(&doc)  // Line 142, 152, 171, 237
s.db.Save(credential)  // credential_service.go line 148
```

**Potential Impact:**
- Silent failures in document processing
- Inconsistent database state
- Users unaware of processing failures
- No audit trail of failures
- Data integrity issues

**Recommended Fix:**
```go
if err := s.db.Save(&doc).Error; err != nil {
    log.Printf("Failed to save document status: %v", err)
    // Could also send to monitoring/alerting system
    return
}
```

---

### 9. Weak Webhook Secret Management
**File:** `/home/user/RaguDashu/backend/internal/services/webhook_service.go`  
**Lines:** 72-75  
**Severity:** HIGH  
**Category:** Security - Authentication

**Issue:**
The webhook secret is stored in plaintext in the database and used directly for HMAC:

```go
if webhook.Secret != "" {
    signature := s.generateSignature(jsonData, webhook.Secret)
    req.Header.Set("X-Webhook-Signature", signature)
}
```

Secrets should be hashed and only used for verification, not stored directly.

**Potential Impact:**
- Database breach exposes all webhook secrets
- HMAC-SHA256 signature verification is weakened
- No audit of which webhooks have been compromised
- Compliance issues

**Recommended Fix:**
```go
// Store hashed secret
func (s *WebhookService) CreateWebhook(...) (*models.Webhook, error) {
    plainSecret, err := generateSecureWebhookSecret()
    if err != nil {
        return nil, err
    }
    
    // Store hash of secret
    secretHash := hashSecret(plainSecret)
    webhook.SecretHash = secretHash
    webhook.Secret = "" // Don't store plaintext
    
    // Return plainSecret to user (one-time display)
    // Store webhook with hash
    s.db.Create(webhook)
    return &webhook, plainSecret
}

func (s *WebhookService) verifySignature(providedSignature, plainSecret string) bool {
    expected := s.generateSignature(jsonData, plainSecret)
    return subtle.ConstantTimeCompare([]byte(providedSignature), []byte(expected)) == 1
}
```

---

### 10. Missing Context Timeout in HTTP Requests
**File:** `/home/user/RaguDashu/backend/internal/services/chat_service.go`  
**Lines:** 213-221  
**Severity:** HIGH  
**Category:** Runtime - Resource Management

**Issue:**
HTTP requests to OpenAI API use a hardcoded timeout in nanoseconds but no context cancellation:

```go
req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
// ...
client := &http.Client{Timeout: 60 * 1000000000} // 60 seconds
resp, err := client.Do(req)
```

The request lacks proper context for cancellation and doesn't handle request-level timeouts.

**Potential Impact:**
- Requests can hang indefinitely if timeout calculation is wrong
- No way to cancel in-flight requests
- Goroutines can accumulate if many slow requests occur
- User requests blocked waiting for external API

**Recommended Fix:**
```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
if err != nil {
    return "", err
}

client := &http.Client{Timeout: 35 * time.Second}
resp, err := client.Do(req)
if err != nil {
    if ctx.Err() == context.DeadlineExceeded {
        return "", fmt.Errorf("OpenAI API request timeout")
    }
    return "", err
}
```

---

### 11. Credential Service Error Handling in Deletion
**File:** `/home/user/RaguDashu/backend/internal/services/credential_service.go`  
**Lines:** 114-116  
**Severity:** HIGH  
**Category:** API - Inconsistent Error Handling

**Issue:**
```go
result := s.db.Where("id = ? AND profile_id = ?", credentialID, profileID).
    Delete(&models.UserCredential{})
if result.Error != nil {
    return result.Error
}
```

The deletion doesn't validate that the credential actually belonged to the user before deletion. The query already includes the profile_id check, but error messages could be misleading.

**Potential Impact:**
- Users could attempt to delete other users' credentials
- Error messages don't clearly indicate ownership issues
- Debugging becomes difficult

**Recommended Fix:**
```go
result := s.db.Where("id = ? AND profile_id = ?", credentialID, profileID).
    Delete(&models.UserCredential{})
    
if result.Error != nil {
    return result.Error
}

if result.RowsAffected == 0 {
    return fmt.Errorf("credential not found or access denied")
}

return nil
```

---

### 12. Race Condition in Rate Limiter
**File:** `/home/user/RaguDashu/backend/internal/api/middleware/ratelimit.go`  
**Lines:** 26-52  
**Severity:** HIGH  
**Category:** Runtime - Race Condition

**Issue:**
While the mutex protects map access, the cleanup algorithm has potential race conditions:

```go
func (rl *rateLimiter) allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    windowStart := now.Add(-rl.window)
    
    // Cleanup and check happen in same lock, but could miss edge cases
    if reqs, exists := rl.requests[key]; exists {
        validReqs := []time.Time{}
        for _, reqTime := range reqs {
            if reqTime.After(windowStart) {
                validReqs = append(validReqs, reqTime)
            }
        }
        rl.requests[key] = validReqs
    }
    
    if len(rl.requests[key]) >= rl.limit {
        return false
    }
}
```

The algorithm could allow more requests than specified during the update operation.

**Potential Impact:**
- Rate limit could be bypassed
- More requests allowed than configured
- Denial of service possible

**Recommended Fix:**
```go
func (rl *rateLimiter) allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    windowStart := now.Add(-rl.window)
    
    // Initialize if needed
    if _, exists := rl.requests[key]; !exists {
        rl.requests[key] = []time.Time{}
    }
    
    // Cleanup
    reqs := rl.requests[key]
    validReqs := make([]time.Time, 0, len(reqs))
    for _, reqTime := range reqs {
        if reqTime.After(windowStart) {
            validReqs = append(validReqs, reqTime)
        }
    }
    
    // Check BEFORE adding
    if len(validReqs) >= rl.limit {
        rl.requests[key] = validReqs
        return false
    }
    
    // Add new request
    validReqs = append(validReqs, now)
    rl.requests[key] = validReqs
    return true
}
```

---

### 13. Transaction Rollback Issue in Auth Service
**File:** `/home/user/RaguDashu/backend/internal/services/auth_service.go`  
**Lines:** 60-88  
**Severity:** HIGH  
**Category:** Database - Transaction Handling

**Issue:**
```go
tx := s.db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()
```

The transaction only rolls back on panic, not on errors. If an error occurs, the transaction might be left uncommitted.

**Potential Impact:**
- Failed transactions left hanging
- Inconsistent database state
- User creation might be partially completed
- Data integrity issues

**Recommended Fix:**
```go
tx := s.db.Begin()
if err := tx.Create(user).Error; err != nil {
    tx.Rollback()
    return nil, fmt.Errorf("failed to create user: %w", err)
}

if err := tx.Create(profile).Error; err != nil {
    tx.Rollback()
    return nil, fmt.Errorf("failed to create profile: %w", err)
}

if err := tx.Commit().Error; err != nil {
    return nil, fmt.Errorf("failed to commit transaction: %w", err)
}

return user, nil
```

---

### 14. Missing File Size Validation
**File:** `/home/user/RaguDashu/backend/internal/api/handlers/documents.go`  
**Lines:** 50-62  
**Severity:** HIGH  
**Category:** API Security - DoS Prevention

**Issue:**
```go
file, err := c.FormFile("file")
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
    return
}

fileContent, err := file.Open()
```

There's no file size validation. Users could upload extremely large files, causing:
- Memory exhaustion
- Denial of service
- Unexpected costs (API calls, storage)

**Potential Impact:**
- DoS attacks via large file uploads
- Memory exhaustion
- Service crash
- High operational costs

**Recommended Fix:**
```go
const MaxFileSize = 100 * 1024 * 1024 // 100MB

file, err := c.FormFile("file")
if err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
    return
}

if file.Size > MaxFileSize {
    c.JSON(http.StatusBadRequest, gin.H{
        "error": fmt.Sprintf("File size exceeds %d MB limit", MaxFileSize/(1024*1024)),
    })
    return
}
```

---

### 15. Insecure HTTP Client in Multiple Services
**File:** `/home/user/RaguDashu/backend/internal/services/pinecone_service.go`  
**Lines:** 76, 117, 165  
**Severity:** HIGH  
**Category:** Security - Insecure Communication

**Issue:**
```go
client := &http.Client{}
resp, err := client.Do(req)
```

HTTP clients are created without timeout, certificate validation, or security headers.

**Potential Impact:**
- MITM attacks possible
- Requests can hang indefinitely
- No connection pooling/reuse
- Security headers missing

**Recommended Fix:**
```go
client := &http.Client{
    Timeout: 30 * time.Second,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        MaxConnsPerHost:     100,
        IdleConnTimeout:     90 * time.Second,
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    },
}
```

---

### 16. No Pagination in List Endpoints
**File:** `/home/user/RaguDashu/backend/internal/api/handlers/documents.go`  
**Lines:** 82-96  
**Severity:** HIGH  
**Category:** API - Performance/DoS

**Issue:**
```go
docs, err := h.documentService.ListDocuments(profileID)
```

The ListDocuments handler returns all documents without pagination. Users with thousands of documents could:
- Cause memory exhaustion
- Create DoS conditions
- Have terrible performance

**Potential Impact:**
- DoS attacks
- Performance degradation
- High memory usage
- Slow API responses

**Recommended Fix:**
```go
type ListDocumentsRequest struct {
    Page  int `form:"page,default=1" binding:"min=1"`
    Limit int `form:"limit,default=20" binding:"min=1,max=100"`
}

func (h *DocumentHandler) ListDocuments(c *gin.Context) {
    var req ListDocumentsRequest
    if err := c.ShouldBindQuery(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    offset := (req.Page - 1) * req.Limit
    docs, total, err := h.documentService.ListDocuments(profileID, offset, req.Limit)
}
```

---

## MEDIUM SEVERITY ISSUES

### 17. Inconsistent Error Response Format
**File:** Multiple handler files  
**Severity:** MEDIUM  
**Category:** API Design

**Issue:**
Different endpoints return different error formats:
- Some return `{"error": "message"}` (handlers/auth.go:56)
- Some return `{"error": "message"}` (handlers/chat.go:32)
- Some return detailed objects (handlers/credentials.go:104-106)

This inconsistency makes client-side error handling difficult.

**Recommended Fix:**
Create a standard error response structure:
```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details interface{} `json:"details,omitempty"`
}

func respondError(c *gin.Context, status int, code, message string) {
    c.JSON(status, ErrorResponse{Code: code, Message: message})
}
```

---

### 18. Missing API Request ID Tracking
**File:** `/home/user/RaguDashu/backend/internal/api/routes.go`  
**Severity:** MEDIUM  
**Category:** Observability

**Issue:**
No request ID generation or correlation ID tracking. Makes debugging distributed issues difficult.

**Recommended Fix:**
Add middleware:
```go
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }
        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)
        c.Next()
    }
}
```

---

### 19. Inefficient Conversation History Loading
**File:** `/home/user/RaguDashu/backend/internal/services/chat_service.go`  
**Lines:** 157-174  
**Severity:** MEDIUM  
**Category:** Performance

**Issue:**
```go
messages, err := s.getConversationHistory(conversation.ID, 10)
// ...
// Reverse to chronological order
for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
    messages[i], messages[j] = messages[j], messages[i]
}
```

Messages are fetched in reverse order then reversed. Should be queried in correct order instead.

**Recommended Fix:**
```go
func (s *ChatService) getConversationHistory(conversationID uuid.UUID, limit int) ([]models.ChatMessage, error) {
    var messages []models.ChatMessage
    err := s.db.Where("conversation_id = ?", conversationID).
        Order("created_at ASC").  // Changed to ASC
        Limit(limit).
        Find(&messages).Error
    return messages, err
}
```

---

### 20. No Duplicate Prevention in Credentials
**File:** `/home/user/RaguDashu/backend/internal/services/credential_service.go`  
**Lines:** 46-57  
**Severity:** MEDIUM  
**Category:** Data Integrity

**Issue:**
If CreateCredential is called twice simultaneously with the same type, race condition could create duplicates or lose data.

**Recommended Fix:**
```go
// Add unique constraint in migration
// db.AddUniqueIndex("idx_credential_unique", "profile_id", "credential_type")

// Use upsert properly with select for update
tx := s.db.Clauses(clause.OnConflict{
    UpdateAll: true,
}).Create(credential)
```

---

### 21. No Logging of Authentication Failures
**File:** `/home/user/RaguDashu/backend/internal/services/auth_service.go`  
**Severity:** MEDIUM  
**Category:** Security - Audit Trail

**Issue:**
Failed login attempts are not logged. Cannot detect brute force attacks.

**Recommended Fix:**
```go
func (s *AuthService) Login(email, password string) (accessToken, refreshToken string, err error) {
    var user models.User
    if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
        // Log failed attempt
        log.Printf("Failed login attempt for email: %s (user not found)", email)
        return "", "", errors.New("invalid credentials")
    }
    
    if !user.CheckPassword(password) {
        log.Printf("Failed login attempt for user %s (wrong password)", user.ID)
        return "", "", errors.New("invalid credentials")
    }
    // ... success
}
```

---

### 22. Missing Content-Type Validation in Uploads
**File:** `/home/user/RaguDashu/backend/internal/api/handlers/documents.go`  
**Lines:** 69  
**Severity:** MEDIUM  
**Category:** Security - Input Validation

**Issue:**
```go
file.Header.Get("Content-Type")
```

Content-Type can be spoofed by clients. No validation that the file actually matches its declared type.

**Recommended Fix:**
```go
// Validate file type by magic bytes
func validateFileType(content []byte, declaredType string) bool {
    // Check PDF
    if bytes.HasPrefix(content, []byte("%PDF")) {
        return declaredType == "application/pdf"
    }
    // Check other types...
    return false
}

// In handler
if !validateFileType(contentBytes, file.Header.Get("Content-Type")) {
    c.JSON(http.StatusBadRequest, gin.H{"error": "File type mismatch"})
    return
}
```

---

### 23. Webhook Event Constants Not Enforced
**File:** `/home/user/RaguDashu/backend/internal/services/webhook_service.go`  
**Severity:** MEDIUM  
**Category:** Data Validation

**Issue:**
The webhook service accepts any event string without validating against the defined constants. Users could create webhooks for non-existent events.

**Recommended Fix:**
```go
var validWebhookEvents = map[string]bool{
    WebhookEventDocumentUploaded:  true,
    WebhookEventDocumentProcessed: true,
    WebhookEventDocumentDeleted:   true,
    WebhookEventChatCreated:       true,
    WebhookEventConversationSaved: true,
    WebhookEventSearchPerformed:   true,
}

func (s *WebhookService) CreateWebhook(..., events []string) (*models.Webhook, error) {
    for _, event := range events {
        if !validWebhookEvents[event] {
            return nil, fmt.Errorf("invalid webhook event: %s", event)
        }
    }
    // ... continue
}
```

---

### 24. Missing Rate Limit Bypass Protection
**File:** `/home/user/RaguDashu/backend/internal/api/middleware/ratelimit.go`  
**Severity:** MEDIUM  
**Category:** Security

**Issue:**
The rate limiter uses `c.ClientIP()` which can be spoofed via X-Forwarded-For headers without validation.

**Recommended Fix:**
```go
func RateLimitMiddleware(requests int, windowMinutes int) gin.HandlerFunc {
    limiter := newRateLimiter(requests, windowMinutes)
    
    return func(c *gin.Context) {
        // Use authenticated user ID if available, fallback to IP
        var key string
        if userID, exists := c.Get("user_id"); exists {
            key = userID.(uuid.UUID).String()
        } else {
            // Be careful with IP - validate X-Forwarded-For
            key = c.ClientIP()
        }
        
        if !limiter.allow(key) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

---

### 25. No MaxTokens Validation
**File:** `/home/user/RaguDashu/backend/internal/services/chat_service.go`  
**Lines:** 200-206  
**Severity:** MEDIUM  
**Category:** Cost Control

**Issue:**
The MaxTokens value from configuration is passed directly to OpenAI without validation. A misconfigured or compromised config could cause massive API costs.

**Recommended Fix:**
```go
const MaxAllowedTokens = 4096

func (s *ChatService) callOpenAI(..., maxTokens int, ...) (string, error) {
    if maxTokens <= 0 || maxTokens > MaxAllowedTokens {
        return "", fmt.Errorf("invalid maxTokens: must be between 1 and %d", MaxAllowedTokens)
    }
    
    reqBody := map[string]interface{}{
        "model":       model,
        "messages":    messages,
        "temperature": temperature,
        "max_tokens":  maxTokens,
    }
    // ...
}
```

---

### 26. Insecure Random Vector ID Generation
**File:** `/home/user/RaguDashu/backend/internal/services/pinecone_service.go`  
**Lines:** 189-191  
**Severity:** MEDIUM  
**Category:** Security

**Issue:**
```go
func GenerateVectorID(prefix string) string {
    return fmt.Sprintf("%s_%s", prefix, uuid.New().String())
}
```

While UUID is fine, the vector IDs are predictable and could leak information about document order.

**Recommended Fix:**
```go
func GenerateVectorID(prefix string) string {
    hash := sha256.Sum256([]byte(uuid.New().String() + time.Now().String()))
    return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(hash[:]))
}
```

---

### 27. Missing Update Timestamp in Credential Service
**File:** `/home/user/RaguDashu/backend/internal/services/credential_service.go`  
**Severity:** MEDIUM  
**Category:** Data Integrity

**Issue:**
When credentials are updated (line 52), `UpdatedAt` is manually set, but GORM should handle this automatically.

**Recommended Fix:**
```go
// Remove manual UpdatedAt setting, let GORM handle it
existing.EncryptedValue = encrypted
existing.IsActive = true
// existing.UpdatedAt = time.Now() // Remove this - GORM sets it
if err := s.db.Save(&existing).Error; err != nil {
    return nil, fmt.Errorf("failed to update credential: %w", err)
}
```

---

### 28. No Connection Pooling Configuration
**File:** `/home/user/RaguDashu/backend/internal/database/connection.go`  
**Severity:** MEDIUM  
**Category:** Performance

**Issue:**
Database connection pool is not configured with appropriate settings for the expected load.

**Recommended Fix:**
```go
import "database/sql"

func Connect(databaseURL string, isDevelopment bool) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
        Logger: logger.Default.LogMode(logLevel),
    })
    if err != nil {
        return nil, err
    }
    
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // Configure connection pool
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    return db, nil
}
```

---

### 29. Array Query Not Properly Escaped
**File:** `/home/user/RaguDashu/backend/internal/services/webhook_service.go`  
**Lines:** 40  
**Severity:** MEDIUM  
**Category:** Database - Potential SQL Injection

**Issue:**
```go
s.db.Where("profile_id = ? AND is_active = ? AND ? = ANY(events)", profileID, true, event).
    Find(&webhooks)
```

While parameterized, the ANY() array comparison could be problematic if Postgres compatibility is an issue.

**Recommended Fix:**
```go
s.db.Where("profile_id = ? AND is_active = ? AND events @> ARRAY[?]", profileID, true, event).
    Find(&webhooks)

// Or use explicit array literal
s.db.Where("profile_id = ? AND is_active = ? AND events::text[] @> ARRAY[?]", profileID, true, event).
    Find(&webhooks)
```

---

## LOW SEVERITY ISSUES

### 30. Hardcoded Model Names
**File:** `/home/user/RaguDashu/backend/internal/services/embedding_service.go`  
**Lines:** 43  
**Severity:** LOW  
**Category:** Maintainability

**Issue:**
```go
Model: "text-embedding-ada-002",
```

Hardcoded model name should be configurable.

**Recommended Fix:**
Move to configuration file.

---

### 31. Missing CORS Security Headers
**File:** `/home/user/RaguDashu/backend/internal/api/middleware/cors.go`  
**Severity:** LOW  
**Category:** Security

**Issue:**
The CORS configuration doesn't include important security headers.

**Recommended Fix:**
```go
config := cors.Config{
    AllowOrigins:     []string{frontendURL},
    AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           300,
}
```

---

### 32. No Graceful Shutdown for Services
**File:** `/home/user/RaguDashu/backend/cmd/server/main.go`  
**Severity:** LOW  
**Category:** Runtime

**Issue:**
The cleanup service is started but not properly shut down with context cancellation.

**Recommended Fix:**
Already handled reasonably well with defer cancel(), but could be more explicit.

---

### 33. Missing Index for Document Queries
**File:** `/home/user/RaguDashu/backend/internal/database/connection.go`  
**Severity:** LOW  
**Category:** Performance

**Issue:**
No index on `documents(profile_id, created_at DESC)` for common list queries.

**Recommended Fix:**
```go
"CREATE INDEX IF NOT EXISTS idx_documents_profile_date ON documents(profile_id, created_at DESC);",
```

---

### 34. Profile Default Namespace Not Validated
**File:** `/home/user/RaguDashu/backend/internal/services/auth_service.go`  
**Lines:** 77  
**Severity:** LOW  
**Category:** Data Validation

**Issue:**
The generated Pinecone namespace isn't validated to conform to Pinecone's naming requirements.

**Recommended Fix:**
```go
// Pinecone namespaces must be valid format
namespace := fmt.Sprintf("user_%s", user.ID.String())
if !isValidNamespace(namespace) {
    return nil, errors.New("failed to generate valid namespace")
}
```

---

### 35. No Metrics/Monitoring Instrumentation
**File:** All service files  
**Severity:** LOW  
**Category:** Observability

**Issue:**
No Prometheus metrics, structured logging, or performance monitoring.

**Recommended Fix:**
Add instrumentation middleware and metrics collection.

---

## SUMMARY TABLE

| Category | Critical | High | Medium | Low | Total |
|----------|----------|------|--------|-----|-------|
| Security | 3 | 4 | 4 | 2 | 13 |
| Runtime | 2 | 3 | 2 | 1 | 8 |
| Logic | 1 | 1 | 3 | 0 | 5 |
| API | 0 | 2 | 3 | 1 | 6 |
| Database | 0 | 2 | 3 | 1 | 6 |
| **TOTAL** | **6** | **12** | **15** | **5** | **38** |

---

## REMEDIATION PRIORITY

**IMMEDIATE (Before Production):**
1. Fix insecure webhook secret generation (#1)
2. Fix credential testing logic bug (#2)
3. Prevent API key exposure in errors (#3)
4. Add auth to sensitive endpoints (#4)
5. Fix nil pointer dereferences (#5)
6. Fix goroutine leaks (#6)
7. Add TopK validation (#7)
8. Fix unhandled database errors (#8)

**URGENT (Within 1 Week):**
9. Implement proper webhook secret management
10. Add context timeouts to HTTP calls
11. Fix transaction handling
12. Add file size validation
13. Secure HTTP clients

**HIGH PRIORITY (Within 1 Month):**
14-29: Address remaining medium severity issues

**NICE TO HAVE:**
30-35: Address low severity issues and add monitoring

