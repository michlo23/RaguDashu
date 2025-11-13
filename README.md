# RAG Dashboard with AI Chatbot

A comprehensive multi-tenant RAG (Retrieval-Augmented Generation) Dashboard that enables users to upload documents, connect Slack workspaces, perform semantic search, and chat with an AI assistant that retrieves context from their own data.

## 🎯 Features

- **Multi-Tenant Architecture**: Each user has isolated data with their own Pinecone namespace
- **Document Management**: Upload PDF and TXT files that are automatically chunked and embedded
- **Semantic Search**: Search across all your uploaded documents using vector similarity
- **AI Chat with RAG**: Chat with GPT-4/GPT-3.5 that retrieves relevant context from your documents
- **Secure Credential Management**: AES-256 encrypted storage of API keys
- **Slack Integration**: Index Slack threads with trust scoring
- **Conversation Management**: Save important chats or let them auto-expire after 60 days
- **Configurable AI Behavior**: Customize system prompts, temperature, model, and context retrieval per conversation

## 🏗️ Tech Stack

### Backend
- **Go 1.23+** with Gin web framework
- **PostgreSQL 15+** for metadata storage
- **GORM** for database operations
- **JWT** for authentication
- **AES-256-GCM** for credential encryption

### Frontend
- **React 18** with TypeScript
- **Vite** for fast development
- **TailwindCSS** for styling
- **Zustand** for state management
- **React Router v6** for navigation

### Infrastructure
- **Pinecone** vector database for embeddings
- **OpenAI API** for embeddings (text-embedding-ada-002) and chat completions

## 📋 Prerequisites

1. **Go 1.23 or higher**
2. **Node.js 18 or higher**
3. **PostgreSQL 15 or higher**
4. **OpenAI API Key** (for embeddings and chat)
5. **Pinecone Account** with API key and index created

## 🚀 Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/michlo23/RaguDashu.git
cd RaguDashu
```

### 2. Setup Database

```bash
# Create PostgreSQL database
createdb ragdb

# Or with psql
psql -U postgres
CREATE DATABASE ragdb;
\q
```

### 3. Backend Setup

```bash
cd backend

# Copy environment file
cp .env.example .env

# Edit .env and configure:
# - DATABASE_URL (your PostgreSQL connection string)
# - JWT_SECRET (min 32 characters)
# - ENCRYPTION_KEY (exactly 32 characters for AES-256)
vim .env

# Install dependencies
go mod download

# Run database migrations (automatic on startup)
# Run the server
go run cmd/server/main.go
```

The backend will start on `http://localhost:8080`

### 4. Frontend Setup

```bash
cd ../frontend

# Install dependencies
npm install

# Copy environment file
cp .env.example .env

# Edit .env if needed (default API URL is correct)
vim .env

# Start development server
npm run dev
```

The frontend will start on `http://localhost:5173`

## 📝 Environment Variables

### Backend (.env)

```bash
# Server
PORT=8080
ENV=development
FRONTEND_URL=http://localhost:5173

# Database
DATABASE_URL=postgresql://raguser:ragpass@localhost:5432/ragdb?sslmode=disable

# JWT (CHANGE IN PRODUCTION!)
JWT_SECRET=your-super-secret-jwt-key-min-32-chars-long-change-this-in-production
JWT_ACCESS_EXPIRE_MINUTES=60
JWT_REFRESH_EXPIRE_DAYS=30

# Encryption (MUST BE EXACTLY 32 BYTES!)
ENCRYPTION_KEY=your-32-byte-encryption-key-change-in-production-!!!!!

# Cleanup
CLEANUP_INTERVAL_HOURS=24
CHAT_RETENTION_DAYS=60

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW_MINUTES=15
```

### Frontend (.env)

```bash
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_NAME=RAG Dashboard
```

## 🔐 Security Features

### Multi-Tenancy Isolation

1. **Database Level**: Each user has a unique profile with isolated data
2. **Pinecone Level**: Each user has a unique namespace (format: `user_{user_id}`)
3. **Application Level**: All queries filter by `profile_id`

### Credential Encryption

All user API keys are encrypted at rest using AES-256-GCM:
- OpenAI API keys
- Pinecone API keys
- Slack tokens

### Authentication

- JWT tokens with configurable expiry
- Refresh token rotation
- Secure password hashing with bcrypt
- Rate limiting on all endpoints

## 📚 API Documentation

### Authentication

```bash
POST /api/auth/register
POST /api/auth/login
POST /api/auth/refresh
GET  /api/auth/me
POST /api/auth/logout
```

### Credentials

```bash
POST   /api/credentials              # Create credential
GET    /api/credentials              # List credentials
DELETE /api/credentials/:id          # Delete credential
POST   /api/credentials/:id/test     # Test credential
```

### Documents

```bash
POST   /api/documents/upload         # Upload document
GET    /api/documents                # List documents
GET    /api/documents/:id            # Get document
DELETE /api/documents/:id            # Delete document
```

### Search

```bash
POST   /api/search                   # Semantic search
```

### Chat

```bash
POST   /api/chat/send                # Send message
POST   /api/chat/configurations      # Create config
GET    /api/chat/configurations      # List configs
GET    /api/chat/conversations       # List conversations
GET    /api/chat/conversations/:id   # Get conversation
POST   /api/chat/conversations/:id/save    # Save conversation
DELETE /api/chat/conversations/:id          # Delete conversation
```

## 🎨 Usage Guide

### 1. First Time Setup

1. **Register an account** at http://localhost:5173/register
2. **Add your API credentials**:
   - Go to Credentials page
   - Add your OpenAI API key
   - Add your Pinecone API key
   - Optionally add Slack token

### 2. Upload Documents

1. Go to Documents page
2. Click "Upload Document"
3. Select a PDF or TXT file
4. Wait for processing (chunking + embedding + upload to Pinecone)

### 3. Configure Chat

1. Go to Chat Configurations
2. Create a new configuration:
   - Choose a name (e.g., "Product Documentation Assistant")
   - Select OpenAI model (GPT-4, GPT-4-Turbo, GPT-3.5-Turbo)
   - Select a Pinecone index (or none for pure chat)
   - Customize system prompt
   - Set temperature, max tokens, and top-k context chunks

### 4. Start Chatting

1. Go to Chat page
2. Select a configuration
3. Start asking questions!
4. The AI will retrieve relevant context from your documents
5. Save important conversations or let them auto-expire

## 🔄 How RAG Works

1. **User sends a message** to the chat
2. **Query embedding** is generated using OpenAI
3. **Semantic search** finds top-K most relevant chunks from Pinecone
4. **Context is injected** into the system prompt
5. **OpenAI generates response** using the retrieved context
6. **Sources are cited** so you know where information came from

## 📊 Database Schema

The application uses the following main tables:

- `users` - User accounts
- `user_profiles` - Multi-tenant profiles
- `user_credentials` - Encrypted API keys
- `pinecone_indexes` - User-specific indexes
- `documents` - Uploaded files metadata
- `document_chunks` - Text chunks with embeddings
- `chat_configurations` - Reusable chat setups
- `chat_conversations` - Chat sessions
- `chat_messages` - Individual messages
- `slack_threads` - Indexed Slack data

## 🧹 Automatic Cleanup

The system includes a cleanup service that runs every 24 hours to:
- Delete unsaved conversations older than 60 days
- Free up database space
- Maintain system performance

## 🛠️ Development

### Backend Development

```bash
cd backend

# Run with auto-reload (install air first)
go install github.com/cosmtrek/air@latest
air

# Run tests
go test ./...

# Build for production
go build -o server cmd/server/main.go
```

### Frontend Development

```bash
cd frontend

# Run dev server
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

## 🔧 Troubleshooting

### Database Connection Errors

- Verify PostgreSQL is running: `pg_isready`
- Check DATABASE_URL in `.env`
- Ensure database exists: `psql -l`

### OpenAI API Errors

- Verify API key is correct
- Check you have credits: https://platform.openai.com/usage
- Ensure you're not rate limited

### Pinecone Errors

- Verify API key and environment
- Check index exists in Pinecone console
- Ensure index dimension matches (1536 for text-embedding-ada-002)
- Format index host correctly: `{index-name}.svc.pinecone.io`

### Frontend Build Errors

```bash
# Clear node_modules and reinstall
rm -rf node_modules package-lock.json
npm install
```

## 📝 TODO / Future Enhancements

- [ ] PDF text extraction (currently only supports TXT files)
- [ ] Batch document upload
- [ ] Real-time Slack sync
- [ ] Export conversations to PDF/Markdown
- [ ] Admin dashboard
- [ ] Usage analytics and cost tracking
- [ ] Multiple profiles per user
- [ ] Shared conversations
- [ ] Advanced search filters
- [ ] Document annotations

## 🔒 Security Considerations for Production

1. **Change default secrets** in `.env`
2. **Use HTTPS** for all connections
3. **Enable PostgreSQL SSL** mode
4. **Set up proper CORS** policies
5. **Implement rate limiting** (already included)
6. **Use environment variables** for secrets (never commit)
7. **Regular security audits**
8. **Backup database** regularly

## 📄 License

This project is private and proprietary.

## 🙋 Support

For issues and questions:
- Check existing issues on GitHub
- Create a new issue with detailed reproduction steps
- Include relevant logs and error messages

## 🚀 Deployment

### Backend Deployment

```bash
# Build binary
cd backend
go build -o server cmd/server/main.go

# Run with production config
ENV=production ./server
```

### Frontend Deployment

```bash
# Build for production
cd frontend
npm run build

# Deploy dist/ folder to your hosting provider
# (Vercel, Netlify, AWS S3 + CloudFront, etc.)
```

### Docker Deployment (Optional)

```bash
# Backend Dockerfile
FROM golang:1.23-alpine
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go
CMD ["./server"]

# Frontend Dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
RUN npm install && npm run build
# Serve with nginx or node server
```

## 💡 Tips

- **Start small**: Upload a few documents first to test
- **Optimize chunks**: 500-1000 words works well for most content
- **Tune top-k**: 3-5 chunks usually provides good context
- **Experiment with temperature**: Lower (0.3-0.5) for factual, higher (0.7-0.9) for creative
- **Use specific system prompts**: Guide the AI on how to use retrieved context
- **Save important chats**: They'll auto-delete after 60 days otherwise

---

**Built with ❤️ using Go, React, and AI**
