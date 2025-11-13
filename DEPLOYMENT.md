# Deployment Guide

This guide covers multiple deployment options for the RAG Dashboard.

## Table of Contents

1. [Railway Deployment](#railway-deployment)
2. [Docker Deployment](#docker-deployment)
3. [Manual Deployment](#manual-deployment)
4. [Environment Variables](#environment-variables)
5. [Database Setup](#database-setup)
6. [Testing](#testing)

---

## Railway Deployment

Railway is the easiest way to deploy this application. It provides managed PostgreSQL and automatic deployments.

### Prerequisites

- Railway account (https://railway.app)
- GitHub account (for automatic deployments)

### Step 1: Create Railway Project

```bash
# Install Railway CLI
npm i -g @railway/cli

# Login to Railway
railway login

# Initialize project
railway init
```

### Step 2: Add PostgreSQL Database

1. Go to your Railway project dashboard
2. Click "New" → "Database" → "Add PostgreSQL"
3. Railway will automatically provision a PostgreSQL database
4. Note the connection string from the "Connect" tab

### Step 3: Deploy Backend

1. In Railway dashboard, click "New" → "GitHub Repo"
2. Select your repository
3. Configure the service:
   - **Root Directory**: `/backend`
   - **Build Command**: (auto-detected from Dockerfile)
   - **Start Command**: `./server`

4. Add environment variables:

```bash
# Required
JWT_SECRET=your-super-secret-jwt-key-min-32-chars-long-change-this
ENCRYPTION_KEY=your-32-byte-encryption-key-here
DATABASE_URL=${{Postgres.DATABASE_URL}}  # Auto-linked from PostgreSQL service

# Optional (with defaults)
PORT=8080
ENV=production
FRONTEND_URL=https://your-frontend-url.railway.app
JWT_ACCESS_EXPIRE_MINUTES=60
JWT_REFRESH_EXPIRE_DAYS=30
CLEANUP_INTERVAL_HOURS=24
CHAT_RETENTION_DAYS=60
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_MINUTES=15
```

### Step 4: Run Database Migrations

**Option A: Using Railway CLI**

```bash
# Connect to your Railway project
railway link

# Run migrations
railway run bash -c "cd backend && ./scripts/run-migrations.sh"
```

**Option B: Using psql directly**

```bash
# Get DATABASE_URL from Railway dashboard
export DATABASE_URL="postgresql://..."

# Run migrations
cd backend
./scripts/run-migrations.sh
```

**Option C: Let the app auto-migrate (easier but less control)**

The app will auto-migrate on startup using GORM. This is already configured.

### Step 5: Deploy Frontend

1. Click "New" → "GitHub Repo" again
2. Select the same repository
3. Configure the service:
   - **Root Directory**: `/frontend`
   - **Build Command**: `npm install && npm run build`
   - **Start Command**: (uses nginx from Dockerfile)

4. Add environment variables:

```bash
VITE_API_BASE_URL=https://your-backend-url.railway.app/api
VITE_APP_NAME=RAG Dashboard
```

### Step 6: Configure Custom Domains (Optional)

1. In Railway dashboard, go to each service
2. Click "Settings" → "Domains"
3. Add your custom domain or use Railway-provided URL

### Railway Auto-Deploy

Once connected to GitHub, Railway will automatically:
- Deploy on every push to main branch
- Run builds in isolated containers
- Handle zero-downtime deployments
- Provide automatic HTTPS

---

## Docker Deployment

### Local Development with Docker Compose

```bash
# Create .env file in root
cp backend/.env.example .env

# Edit .env with your secrets
vim .env

# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down
```

Services will be available at:
- Backend API: http://localhost:8080
- Frontend: http://localhost:5173
- PostgreSQL: localhost:5432

### Production Docker Deployment

**1. Build Images**

```bash
# Backend
cd backend
docker build -t rag-backend:latest .

# Frontend
cd ../frontend
docker build -t rag-frontend:latest .
```

**2. Run with External PostgreSQL**

```bash
# Backend
docker run -d \
  --name rag-backend \
  -p 8080:8080 \
  -e DATABASE_URL="postgresql://user:pass@host:5432/ragdb" \
  -e JWT_SECRET="your-secret" \
  -e ENCRYPTION_KEY="your-32-byte-key" \
  -e FRONTEND_URL="https://your-frontend.com" \
  rag-backend:latest

# Frontend
docker run -d \
  --name rag-frontend \
  -p 80:80 \
  rag-frontend:latest
```

**3. Docker Compose Production**

```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  backend:
    image: rag-backend:latest
    environment:
      DATABASE_URL: ${DATABASE_URL}
      JWT_SECRET: ${JWT_SECRET}
      ENCRYPTION_KEY: ${ENCRYPTION_KEY}
      FRONTEND_URL: ${FRONTEND_URL}
    ports:
      - "8080:8080"
    restart: always

  frontend:
    image: rag-frontend:latest
    ports:
      - "80:80"
    restart: always
```

```bash
docker-compose -f docker-compose.prod.yml up -d
```

---

## Manual Deployment

### Backend Deployment

**1. Build the Backend**

```bash
cd backend
go build -o server cmd/server/main.go
```

**2. Setup Environment**

```bash
# Create .env file
cp .env.example .env

# Edit with production values
vim .env
```

**3. Run Migrations**

```bash
./scripts/run-migrations.sh
```

**4. Start the Server**

```bash
# Using systemd (recommended)
sudo systemctl start rag-backend

# Or using screen/tmux
screen -S rag-backend
./server

# Or with supervisor
supervisorctl start rag-backend
```

### Frontend Deployment

**1. Build the Frontend**

```bash
cd frontend
npm install
npm run build
```

**2. Deploy with Nginx**

```nginx
# /etc/nginx/sites-available/rag-dashboard
server {
    listen 80;
    server_name your-domain.com;

    root /var/www/rag-dashboard/dist;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    # Proxy API requests to backend
    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/rag-dashboard /etc/nginx/sites-enabled/

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx

# Setup SSL with Let's Encrypt
sudo certbot --nginx -d your-domain.com
```

---

## Environment Variables

### Backend Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8080` | Server port |
| `ENV` | No | `development` | Environment (development/production) |
| `FRONTEND_URL` | Yes | - | Frontend URL for CORS |
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `JWT_SECRET` | Yes | - | JWT signing secret (min 32 chars) |
| `JWT_ACCESS_EXPIRE_MINUTES` | No | `60` | Access token expiry |
| `JWT_REFRESH_EXPIRE_DAYS` | No | `30` | Refresh token expiry |
| `ENCRYPTION_KEY` | Yes | - | AES-256 key (exactly 32 bytes) |
| `CLEANUP_INTERVAL_HOURS` | No | `24` | Cleanup service interval |
| `CHAT_RETENTION_DAYS` | No | `60` | Unsaved chat retention |
| `RATE_LIMIT_REQUESTS` | No | `100` | Rate limit requests |
| `RATE_LIMIT_WINDOW_MINUTES` | No | `15` | Rate limit window |

### Frontend Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VITE_API_BASE_URL` | Yes | - | Backend API URL |
| `VITE_APP_NAME` | No | `RAG Dashboard` | Application name |

### Generating Secrets

```bash
# Generate JWT_SECRET (32+ characters)
openssl rand -base64 32

# Generate ENCRYPTION_KEY (exactly 32 bytes)
openssl rand -base64 32 | cut -c1-32
```

---

## Database Setup

### Create Database

```bash
# Local PostgreSQL
createdb ragdb

# Or with psql
psql -U postgres
CREATE DATABASE ragdb;
\q
```

### Run Migrations

**Option 1: Using migration script**

```bash
cd backend
./scripts/run-migrations.sh
```

**Option 2: Using psql directly**

```bash
psql -U raguser -d ragdb -f migrations/001_initial_schema.sql
```

**Option 3: Auto-migrate (development only)**

The application will auto-migrate on startup using GORM.

### Rollback Migrations

```bash
psql -U raguser -d ragdb -f migrations/001_initial_schema.down.sql
```

### Backup Database

```bash
# Backup
pg_dump -U raguser ragdb > backup.sql

# Restore
psql -U raguser ragdb < backup.sql
```

---

## Testing

### Run Backend Tests

```bash
cd backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with verbose output
go test -v ./...

# Run specific package
go test ./internal/services/

# Run specific test
go test -run TestEncryptionService_Encrypt_Decrypt ./internal/services/
```

### Test Results

The test suite includes:
- ✅ Encryption/Decryption tests
- ✅ Text chunking tests
- ✅ Token estimation tests
- ✅ Auth handler integration tests
- ✅ Registration and login flows

### Load Testing

```bash
# Install hey
go install github.com/rakyll/hey@latest

# Test health endpoint
hey -n 1000 -c 10 http://localhost:8080/api/health

# Test auth endpoint
hey -n 100 -c 5 -m POST \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  http://localhost:8080/api/auth/login
```

---

## Production Checklist

Before deploying to production:

- [ ] Change all default secrets (JWT_SECRET, ENCRYPTION_KEY)
- [ ] Set ENV=production
- [ ] Enable HTTPS/SSL
- [ ] Configure proper CORS (FRONTEND_URL)
- [ ] Set up database backups
- [ ] Configure monitoring (logs, metrics)
- [ ] Set up error tracking (Sentry, etc.)
- [ ] Test all API endpoints
- [ ] Verify rate limiting works
- [ ] Test authentication flows
- [ ] Verify multi-tenancy isolation
- [ ] Test document upload and embedding
- [ ] Test chat functionality
- [ ] Set up CI/CD pipeline
- [ ] Configure log rotation
- [ ] Review security headers
- [ ] Test database migrations
- [ ] Verify cleanup service runs

---

## Troubleshooting

### Backend won't start

```bash
# Check logs
docker-compose logs backend

# Verify database connection
psql $DATABASE_URL

# Check environment variables
env | grep DATABASE_URL
```

### Frontend can't connect to backend

1. Check VITE_API_BASE_URL in frontend .env
2. Verify backend is running: `curl http://localhost:8080/api/health`
3. Check CORS configuration in backend
4. Verify firewall rules

### Database migration fails

```bash
# Check if database exists
psql -l

# Verify connection
psql $DATABASE_URL -c "SELECT version();"

# Check migration file syntax
cat migrations/001_initial_schema.sql
```

### Tests failing

```bash
# Clean test cache
go clean -testcache

# Run tests with race detection
go test -race ./...

# Check for import errors
go mod tidy
go mod verify
```

---

## Support

For deployment issues:
- Check logs: `docker-compose logs -f`
- Verify environment variables
- Review [README.md](README.md) for setup instructions
- Open an issue on GitHub

---

**Happy Deploying! 🚀**
