# Testing Guide

This document describes how to test the RAG Dashboard application.

## Running Tests

### Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test ./internal/services/
go test ./internal/api/handlers/

# Run specific test
go test -run TestEncryptionService_Encrypt_Decrypt ./internal/services/
```

### Test Coverage

Current test coverage:
- ✅ `encryption_service.go` - Encryption/decryption, key validation
- ✅ `text.go` - Text chunking, token estimation, truncation
- ✅ `auth.go` (handlers) - Registration, login, token validation

### Manual API Testing

**Health Check**

```bash
curl http://localhost:8080/api/health
```

**Register User**

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'
```

**Login**

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Get Current User**

```bash
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Add Credentials**

```bash
curl -X POST http://localhost:8080/api/credentials \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "credential_type": "openai_api_key",
    "value": "sk-your-openai-key"
  }'
```

**List Credentials**

```bash
curl http://localhost:8080/api/credentials \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Search**

```bash
curl -X POST http://localhost:8080/api/search \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "What is RAG?",
    "index_id": "your-index-id",
    "top_k": 5
  }'
```

**Send Chat Message**

```bash
curl -X POST http://localhost:8080/api/chat/send \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config_id": "your-config-id",
    "message": "Hello, how are you?"
  }'
```

## Integration Testing

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# Wait for services to be healthy
sleep 10

# Run integration tests
./scripts/integration-tests.sh
```

### End-to-End Test Flow

1. **Setup**: Create test user
2. **Auth**: Login and get token
3. **Credentials**: Add OpenAI and Pinecone keys
4. **Documents**: Upload a test document
5. **Search**: Perform semantic search
6. **Chat Config**: Create chat configuration
7. **Chat**: Send message and verify response
8. **Cleanup**: Delete test data

## Load Testing

### Using hey

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Test health endpoint (1000 requests, 10 concurrent)
hey -n 1000 -c 10 http://localhost:8080/api/health

# Test login endpoint
hey -n 100 -c 5 -m POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  http://localhost:8080/api/auth/login
```

### Using Apache Bench

```bash
# Install ab
sudo apt-get install apache2-utils  # Ubuntu/Debian
brew install httpd  # macOS

# Test health endpoint
ab -n 1000 -c 10 http://localhost:8080/api/health
```

## Security Testing

### Test Rate Limiting

```bash
# Should succeed for first 100 requests
for i in {1..100}; do
  curl http://localhost:8080/api/health
done

# Should return 429 Too Many Requests
curl http://localhost:8080/api/health
```

### Test Authentication

```bash
# Should return 401 Unauthorized
curl http://localhost:8080/api/auth/me

# Should return 401 with invalid token
curl http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer invalid-token"
```

### Test Multi-Tenancy Isolation

1. Create two users
2. User A uploads document
3. User B tries to access User A's document
4. Verify User B gets 404 or 403

## Frontend Testing

### Manual Testing Checklist

- [ ] Registration works
- [ ] Login works
- [ ] Token refresh works on 401
- [ ] Dashboard loads correctly
- [ ] Navigation works
- [ ] Logout works
- [ ] Protected routes redirect to login
- [ ] Forms validate properly
- [ ] Error messages display correctly

### Browser Testing

Test on:
- [ ] Chrome/Chromium
- [ ] Firefox
- [ ] Safari
- [ ] Edge

### Responsive Testing

Test on:
- [ ] Desktop (1920x1080)
- [ ] Tablet (768x1024)
- [ ] Mobile (375x667)

## Database Testing

### Test Migrations

```bash
# Run migrations
./backend/scripts/run-migrations.sh

# Verify tables exist
psql $DATABASE_URL -c "\dt"

# Check indexes
psql $DATABASE_URL -c "\di"

# Test rollback
psql $DATABASE_URL -f backend/migrations/001_initial_schema.down.sql

# Re-run migrations
./backend/scripts/run-migrations.sh
```

### Test Data Isolation

```sql
-- Create test users
INSERT INTO users (email, password_hash) VALUES
  ('user1@test.com', 'hash1'),
  ('user2@test.com', 'hash2');

-- Verify users can't see each other's data
SELECT * FROM documents WHERE profile_id = 'user1-profile-id';
SELECT * FROM documents WHERE profile_id = 'user2-profile-id';
```

## Performance Testing

### Benchmark Results

Expected performance (local development):

- Health check: < 1ms
- Login: < 50ms
- Search: < 500ms (depends on Pinecone)
- Chat: < 2s (depends on OpenAI)
- Document upload: < 100ms (chunking only, embedding async)

### Monitoring

```bash
# Watch logs
docker-compose logs -f backend

# Monitor database connections
psql $DATABASE_URL -c "SELECT * FROM pg_stat_activity;"

# Check memory usage
docker stats
```

## Continuous Integration

### GitHub Actions (Example)

```yaml
name: Tests

on: [push, pull_request]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Run tests
        run: |
          cd backend
          go test -v ./...
```

## Test Data

### Sample Documents

Create test documents in `backend/testdata/`:
- `sample.txt` - Plain text file
- `sample_large.txt` - Large file (>10KB)

### Sample API Keys (for testing)

Use test/mock keys that don't incur charges:
- OpenAI: Use mock service or test mode
- Pinecone: Use test index with small dimensions

## Troubleshooting Tests

### Tests fail with database errors

```bash
# Ensure test database exists
createdb ragdb_test

# Use separate DATABASE_URL for tests
export DATABASE_URL="postgresql://localhost/ragdb_test?sslmode=disable"
```

### Tests fail with import errors

```bash
go mod tidy
go mod download
go clean -modcache
```

### Integration tests timeout

- Increase timeout in test configuration
- Check if services are running
- Verify network connectivity

---

**Need help? Check [DEPLOYMENT.md](DEPLOYMENT.md) or open an issue.**
