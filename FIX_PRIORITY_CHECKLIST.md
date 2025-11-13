# RAG Dashboard Backend - Fix Priority Checklist

## PHASE 1: CRITICAL SECURITY FIXES (Must complete before any deployment)
**Estimated Time: 4-6 hours**
**Target: Complete by end of this week**

### Security & Data Integrity
- [ ] **webhook_service.go:94-95** - Replace timestamp secret with `crypto/rand`
  - Files to modify: `webhook_service.go`
  - Complexity: LOW
  - Time: 30 mins
  - Test: Create webhook and verify secret format

- [ ] **credential_service.go:128** - Fix TestCredential empty string bug
  - Files to modify: `credential_service.go`
  - Complexity: LOW
  - Time: 15 mins
  - Test: Test credential validation endpoint

- [ ] **document_service.go:168 (and similar)** - Sanitize error messages to remove API keys
  - Files to modify: `document_service.go`, `chat_service.go`, `embedding_service.go`
  - Complexity: MEDIUM
  - Time: 1 hour
  - Test: Upload document with invalid API key, check error message

- [ ] **routes.go** - Add authentication requirement to `/api/health` and `/api/status`
  - Files to modify: `routes.go`, `system.go`
  - Complexity: LOW
  - Time: 20 mins
  - Test: Verify endpoints require auth token

### Runtime Stability
- [ ] **chat.go, search.go, credentials.go, documents.go, webhooks.go, analytics.go**
  - Add nil/error checks for `middleware.GetProfileID()` calls
  - Files to modify: All handler files
  - Complexity: LOW
  - Time: 1.5 hours
  - Test: Call endpoints without auth, verify proper error response

- [ ] **webhook_service.go:44-46** - Add WaitGroup to prevent goroutine leaks
  - Files to modify: `webhook_service.go`
  - Complexity: MEDIUM
  - Time: 1 hour
  - Test: Trigger many webhooks, monitor goroutine count

---

## PHASE 2: HIGH PRIORITY ISSUES (Complete within 1 week)
**Estimated Time: 8-12 hours**

### Input Validation & DoS Prevention
- [ ] **search.go:42-44** - Add TopK bounds validation (1-100)
  - Files to modify: `search.go`
  - Complexity: LOW
  - Time: 20 mins

- [ ] **documents.go:50-62** - Add file size validation (max 100MB)
  - Files to modify: `documents.go`
  - Complexity: LOW
  - Time: 30 mins

- [ ] **chat_service.go:200-206** - Add MaxTokens validation
  - Files to modify: `chat_service.go`
  - Complexity: LOW
  - Time: 20 mins

### Error Handling & Data Integrity
- [ ] **document_service.go:142, 148, 152, etc.** - Check all `db.Save()` errors
  - Files to modify: `document_service.go`, `credential_service.go`, `webhook_service.go`
  - Complexity: MEDIUM
  - Time: 1.5 hours
  - Test: Simulate database errors, verify proper handling

- [ ] **auth_service.go:60-88** - Fix transaction handling
  - Files to modify: `auth_service.go`
  - Complexity: MEDIUM
  - Time: 1 hour
  - Test: Create user, simulate errors, verify rollback

### HTTP/Network Security
- [ ] **chat_service.go:213-221** - Add proper context timeouts
  - Files to modify: `chat_service.go`, `embedding_service.go`, `pinecone_service.go`
  - Complexity: MEDIUM
  - Time: 1.5 hours
  - Test: Timeout a slow request, verify proper error

- [ ] **pinecone_service.go:76, 117, 165** - Secure HTTP clients
  - Files to modify: `pinecone_service.go`
  - Complexity: MEDIUM
  - Time: 1 hour
  - Test: Verify timeouts and TLS settings

- [ ] **webhook_service.go:72-75** - Hash webhook secrets instead of plaintext
  - Files to modify: `webhook_service.go`, database migration
  - Complexity: HIGH
  - Time: 2 hours
  - Test: Create webhook, verify secret is hashed in DB

### API Design Issues
- [ ] **documents.go:82-96** - Add pagination to list endpoints
  - Files to modify: `documents.go`, `document_service.go`
  - Complexity: MEDIUM
  - Time: 1.5 hours
  - Test: List with page/limit parameters

- [ ] **ratelimit.go:26-52** - Fix rate limiter algorithm
  - Files to modify: `ratelimit.go`
  - Complexity: MEDIUM
  - Time: 1.5 hours
  - Test: Verify rate limit is properly enforced

---

## PHASE 3: MEDIUM PRIORITY ISSUES (Complete within 1 month)
**Estimated Time: 16-24 hours**

### Security & Audit Trail
- [ ] **auth_service.go** - Add logging of authentication failures
  - Files to modify: `auth_service.go`
  - Complexity: LOW
  - Time: 1 hour
  - Test: Failed login attempts should be logged

- [ ] **ratelimit.go** - Prevent IP spoofing via X-Forwarded-For
  - Files to modify: `ratelimit.go`
  - Complexity: MEDIUM
  - Time: 1 hour

- [ ] **documents.go:69** - Add Content-Type validation via magic bytes
  - Files to modify: `documents.go`
  - Complexity: MEDIUM
  - Time: 1.5 hours

### Data & API Consistency
- [ ] **Multiple handlers** - Implement consistent error response format
  - Files to modify: All handler files
  - Complexity: MEDIUM
  - Time: 2 hours
  - Test: Verify all error responses follow same format

- [ ] **routes.go** - Add request ID tracking middleware
  - Files to modify: `routes.go`, new middleware file
  - Complexity: MEDIUM
  - Time: 1.5 hours
  - Test: Verify X-Request-ID header in responses

- [ ] **chat_service.go:157-174** - Optimize conversation history loading
  - Files to modify: `chat_service.go`
  - Complexity: LOW
  - Time: 30 mins

- [ ] **webhook_service.go** - Enforce webhook event constants
  - Files to modify: `webhook_service.go`
  - Complexity: LOW
  - Time: 30 mins

### Performance & Resource Management
- [ ] **database/connection.go** - Configure connection pooling
  - Files to modify: `connection.go`
  - Complexity: LOW
  - Time: 30 mins
  - Test: Load test verifies proper pooling

- [ ] **credential_service.go:46-57** - Add unique constraint for credentials
  - Files to modify: Database migration, `credential_service.go`
  - Complexity: MEDIUM
  - Time: 1 hour

- [ ] **credential_service.go:52** - Remove manual UpdatedAt setting
  - Files to modify: `credential_service.go`
  - Complexity: LOW
  - Time: 15 mins

### Code Quality
- [ ] **pinecone_service.go:189-191** - Improve vector ID randomness
  - Files to modify: `pinecone_service.go`
  - Complexity: LOW
  - Time: 30 mins

---

## PHASE 4: LOW PRIORITY ISSUES (Nice to have)
**Estimated Time: 4-8 hours**

- [ ] **embedding_service.go:43** - Make model name configurable
- [ ] **cors.go** - Add security headers (MaxAge, etc.)
- [ ] **All services** - Add Prometheus metrics instrumentation
- [ ] **database/connection.go** - Add compound indexes for common queries
- [ ] **auth_service.go:77** - Validate Pinecone namespace format

---

## Testing Strategy

### Unit Tests Required
- [ ] Webhook secret generation
- [ ] Credential testing
- [ ] Error sanitization
- [ ] Input validation (TopK, file size, tokens)
- [ ] Rate limiter algorithm
- [ ] Transaction rollback

### Integration Tests Required
- [ ] End-to-end webhook delivery
- [ ] Auth flow with proper error handling
- [ ] Document upload with size validation
- [ ] Chat with context timeouts
- [ ] Search with pagination

### Load Tests Required
- [ ] Rate limiter under concurrent load
- [ ] Goroutine leak detection over time
- [ ] Connection pool under load
- [ ] Large document upload impact

---

## Deployment Checklist

### Before Phase 1 Deployment
- [ ] All critical issues fixed
- [ ] Unit tests passing
- [ ] Security review of fixes
- [ ] No new goroutine leaks detected

### Before Phase 2 Deployment
- [ ] All high priority issues fixed
- [ ] Integration tests passing
- [ ] Load tests passing
- [ ] Database migration tested

### Before Phase 3 Deployment
- [ ] All medium priority issues fixed
- [ ] Full regression testing
- [ ] Performance benchmarks
- [ ] Security audit pass

### Before Production
- [ ] All phases complete
- [ ] Penetration testing
- [ ] GDPR/HIPAA/SOC2 compliance check
- [ ] Documentation updated
- [ ] Incident response plan prepared

---

## File-by-File Fix Guide

### webhook_service.go (3 issues)
1. Line 94-95: Replace timestamp with crypto/rand secret
2. Line 72-75: Hash secrets instead of plaintext storage
3. Line 44-46: Add WaitGroup for goroutine tracking

### credential_service.go (2 issues)
1. Line 128: Fix empty string bug in TestCredential
2. Line 46-57: Add unique constraint for race condition prevention

### chat_service.go (3 issues)
1. Line 213-221: Add context timeouts to HTTP requests
2. Line 157-174: Optimize message history loading
3. Line 200-206: Add MaxTokens validation

### document_service.go (2 issues)
1. Line 168: Sanitize error messages
2. Multiple lines: Check all db.Save() errors

### auth_service.go (2 issues)
1. Line 60-88: Fix transaction handling
2. Add login failure logging

### pinecone_service.go (2 issues)
1. Lines 76, 117, 165: Secure HTTP clients
2. Line 189-191: Improve vector ID randomness

### documents.go (handlers) (2 issues)
1. Line 50-62: Add file size validation
2. Line 82-96: Add pagination

### search.go (1 issue)
1. Line 42-44: Add TopK validation

### routes.go (1 issue)
1. Add auth requirement to health/status endpoints

### Multiple handler files (1 critical issue)
- Add nil checks for GetProfileID calls

### ratelimit.go (1 issue)
1. Fix rate limiter algorithm

---

## Success Metrics

After completing all phases:

- Zero critical security vulnerabilities
- 100% of database operations checked for errors
- No goroutine leaks over 24-hour runtime
- All HTTP requests have proper timeouts
- Proper authentication on all protected endpoints
- Consistent error response formats
- Audit logging for security events
- Proper pagination on all list endpoints
- Input validation on all user inputs
- Connection pooling properly configured

---

## Questions to Answer During Fixes

1. Should health/status endpoints be public or require auth?
2. What should max file size be? (currently 100MB suggested)
3. What should max TopK be? (currently 100 suggested)
4. Should webhook delivery be synchronous or async?
5. How long should HTTP timeouts be? (currently 30-60 seconds suggested)
6. Should we implement webhook delivery retry logic?
7. Should we log API keys in audit logs at all?

---

## References

- [OWASP Top 10](https://owasp.org/Top10/)
- [Go Security Best Practices](https://golang.org/doc/security)
- [GDPR API Security](https://gdpr-info.eu/)
- [Go HTTP Client Best Practices](https://golang.org/pkg/net/http/)
- [Webhook Security](https://docs.github.com/en/developers/webhooks-and-events/webhooks/securing-your-webhooks)

