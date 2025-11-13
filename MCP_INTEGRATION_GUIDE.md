# MCP Integration Guide

**Version**: 1.0.0
**Date**: 2025-11-13
**Status**: ✅ **Complete - Production Ready**

---

## 📖 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Backend Implementation](#backend-implementation)
4. [Frontend Implementation](#frontend-implementation)
5. [API Reference](#api-reference)
6. [OAuth 2 Setup](#oauth-2-setup)
7. [Supported MCP Servers](#supported-mcp-servers)
8. [Testing](#testing)
9. [Deployment](#deployment)
10. [Troubleshooting](#troubleshooting)

---

## Overview

### What is MCP?

**MCP (Model Context Protocol)** is a standard protocol that enables AI agents to connect to external data sources and tools. This integration allows your RAG Dashboard to:

- ✅ Connect to external APIs (Atlassian, GitHub, Slack, etc.)
- ✅ Call tools from MCP servers during chat conversations
- ✅ Access external data and resources
- ✅ Use OAuth 2 for secure authorization
- ✅ Track tool usage and analytics

### Key Features

- **Multiple Authentication Methods**: None, API Key, OAuth 2
- **Automatic Token Management**: OAuth tokens refreshed automatically
- **Tool Calling Integration**: Seamless integration with OpenAI function calling
- **Multi-Server Support**: Connect to multiple MCP servers simultaneously
- **Tool Usage Analytics**: Track all tool calls with duration and errors
- **Security**: Encrypted credentials, sanitized errors, secure OAuth flow

---

## Architecture

### System Components

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Frontend  │────────▶│   Backend    │────────▶│ MCP Server  │
│   (React)   │◀────────│   (Go/Gin)   │◀────────│  (External) │
└─────────────┘         └──────────────┘         └─────────────┘
                              │
                              ▼
                        ┌──────────────┐
                        │  PostgreSQL  │
                        │  (Database)  │
                        └──────────────┘
```

### Request Flow

1. **User sends chat message**
2. Chat service checks if MCP tools are enabled in configuration
3. Fetches available tools from enabled MCP servers
4. Converts MCP tools to OpenAI function format
5. Calls OpenAI with functions parameter
6. **If OpenAI requests a function call**:
   - Chat service calls MCP tool via MCP service
   - MCP service handles authentication (API key or OAuth 2)
   - Tool result returned to OpenAI
   - OpenAI generates final response with tool context
7. Conversation and tool usage saved to database

### Database Schema

```sql
-- MCP Servers Configuration
CREATE TABLE mcp_servers (
    id UUID PRIMARY KEY,
    profile_id UUID REFERENCES user_profiles(id),
    name VARCHAR(255),
    server_url VARCHAR(500),
    auth_type VARCHAR(50), -- 'none', 'api_key', 'oauth2'
    encrypted_api_key TEXT,
    oauth2_config JSONB,
    encrypted_access_token TEXT,
    encrypted_refresh_token TEXT,
    token_expires_at TIMESTAMP,
    available_tools JSONB,
    is_active BOOLEAN,
    ...
);

-- OAuth 2 Flow State (Temporary)
CREATE TABLE oauth2_states (
    id UUID PRIMARY KEY,
    state VARCHAR(64) UNIQUE,
    mcp_server_id UUID REFERENCES mcp_servers(id),
    code_verifier VARCHAR(128), -- PKCE
    expires_at TIMESTAMP, -- 10 minutes
    ...
);

-- Tool Usage Tracking
CREATE TABLE mcp_tool_usage (
    id UUID PRIMARY KEY,
    conversation_id UUID REFERENCES chat_conversations(id),
    mcp_server_id UUID REFERENCES mcp_servers(id),
    tool_name VARCHAR(255),
    tool_input JSONB,
    tool_output JSONB,
    duration_ms INTEGER,
    error_message TEXT,
    ...
);
```

---

## Backend Implementation

### Services

#### 1. OAuth2Service (`internal/services/oauth_service.go`)

**Purpose**: Handle OAuth 2 authorization flow with PKCE

**Key Methods**:
```go
// Generate authorization URL with PKCE
func (s *OAuth2Service) GenerateAuthorizationURL(
    profileID uuid.UUID,
    mcpServerID uuid.UUID,
    redirectURL string,
) (string, error)

// Exchange authorization code for access token
func (s *OAuth2Service) ExchangeCodeForToken(
    state string,
    code string,
) (*models.MCPServer, error)

// Refresh expired access token
func (s *OAuth2Service) RefreshAccessToken(
    mcpServerID uuid.UUID,
    profileID uuid.UUID,
) error

// Get valid access token (auto-refresh if needed)
func (s *OAuth2Service) GetValidAccessToken(
    mcpServerID uuid.UUID,
    profileID uuid.UUID,
) (string, error)
```

**Security Features**:
- PKCE (SHA-256 code challenge)
- Random state parameter (32 bytes)
- Encrypted tokens in database
- Automatic token refresh (5-minute buffer)
- State expiration (10 minutes)

#### 2. MCPService (`internal/services/mcp_service.go`)

**Purpose**: Communicate with MCP servers

**Key Methods**:
```go
// CRUD operations
func (s *MCPService) CreateMCPServer(...) (*models.MCPServer, error)
func (s *MCPService) GetMCPServer(...) (*models.MCPServer, error)
func (s *MCPService) UpdateMCPServer(...) (*models.MCPServer, error)
func (s *MCPService) DeleteMCPServer(...) error
func (s *MCPService) ListMCPServers(...) ([]models.MCPServer, error)

// Capabilities management
func (s *MCPService) SyncCapabilities(...) error

// Tool calling
func (s *MCPService) CallTool(
    serverID, profileID uuid.UUID,
    toolName string,
    input map[string]interface{},
) (*MCPToolResponse, error)
```

**Features**:
- JSON-RPC 2.0 protocol
- Three authentication methods
- Automatic token refresh for OAuth 2
- Error sanitization (prevent token leakage)
- HTTP timeouts (30 seconds)

#### 3. ChatService Integration (`internal/services/chat_service.go`)

**Purpose**: Integrate MCP tools with OpenAI function calling

**Key Methods**:
```go
// Fetch and convert MCP tools to OpenAI format
func (s *ChatService) getMCPToolsAsOpenAIFunctions(
    profileID uuid.UUID,
    serverIDs []uuid.UUID,
) ([]OpenAIFunction, error)

// Call OpenAI with function calling support
func (s *ChatService) callOpenAIWithTools(...) (string, []ToolUsage, error)

// Save tool usage records
func (s *ChatService) saveToolUsageRecords(...)
```

**Function Calling Flow**:
1. Convert MCP tools to OpenAI function format
2. Add functions to OpenAI request
3. Loop until no more function calls (max 5 iterations)
4. For each function call:
   - Parse function name and arguments
   - Call MCP tool
   - Track duration and errors
   - Return result to OpenAI
5. Save final response with tool usage records

---

## Frontend Implementation

### Services

#### MCP API Service (`src/services/mcp.ts`)

**TypeScript Interface**:
```typescript
export interface MCPServer {
  id: string;
  name: string;
  server_url: string;
  auth_type: 'none' | 'api_key' | 'oauth2';
  available_tools?: MCPTool[];
  is_active: boolean;
  ...
}

export const mcpService = {
  listServers(): Promise<MCPServer[]>
  getServer(id: string): Promise<MCPServer>
  createServer(data: CreateMCPServerRequest): Promise<MCPServer>
  updateServer(id: string, data: UpdateMCPServerRequest): Promise<MCPServer>
  deleteServer(id: string): Promise<void>
  syncCapabilities(id: string): Promise<void>
  callTool(id: string, request: ToolCallRequest): Promise<ToolCallResponse>
  initiateOAuth2(id: string, request: OAuth2InitiateRequest): Promise<OAuth2InitiateResponse>
  refreshToken(id: string): Promise<void>
}
```

### State Management

#### MCP Zustand Store (`src/stores/mcpStore.ts`)

**Store Interface**:
```typescript
interface MCPState {
  servers: MCPServer[];
  currentServer: MCPServer | null;
  isLoading: boolean;
  error: string | null;

  fetchServers(): Promise<void>
  createServer(data: CreateMCPServerRequest): Promise<MCPServer>
  syncCapabilities(id: string): Promise<void>
  ...
}

// Usage
import { useMCPStore } from './stores/mcpStore';

const { servers, fetchServers, createServer } = useMCPStore();
```

**Features**:
- Automatic JSON parsing for capabilities
- Error handling
- Loading states
- Optimistic updates

---

## API Reference

### MCP Server Endpoints

#### List MCP Servers
```
GET /api/mcp/servers
Authorization: Bearer {token}

Response:
[
  {
    "id": "uuid",
    "name": "Atlassian Jira",
    "server_url": "https://mcp.atlassian.com/v1",
    "auth_type": "oauth2",
    "available_tools": [...],
    "is_active": true,
    ...
  }
]
```

#### Create MCP Server (API Key)
```
POST /api/mcp/servers
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "GitHub MCP",
  "description": "Access GitHub repositories",
  "server_url": "https://mcp.github.com/v1",
  "auth_type": "api_key",
  "api_key": "your-api-key-here"
}

Response: 201 Created
{
  "id": "uuid",
  "name": "GitHub MCP",
  ...
}
```

#### Create MCP Server (OAuth 2)
```
POST /api/mcp/servers
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Atlassian Jira",
  "server_url": "https://mcp.atlassian.com/v1",
  "auth_type": "oauth2",
  "oauth2_config": {
    "client_id": "your-client-id",
    "client_secret": "your-client-secret",
    "auth_url": "https://auth.atlassian.com/authorize",
    "token_url": "https://auth.atlassian.com/oauth/token",
    "scopes": ["read:jira-work", "write:jira-work"]
  }
}
```

#### Sync Capabilities
```
POST /api/mcp/servers/{id}/sync
Authorization: Bearer {token}

Response: 200 OK
{
  "message": "Capabilities synced successfully"
}
```

#### Call MCP Tool
```
POST /api/mcp/servers/{id}/tools/call
Authorization: Bearer {token}
Content-Type: application/json

{
  "tool_name": "search_issues",
  "input": {
    "jql": "project = PROJ AND status = Open",
    "max_results": 10
  }
}

Response: 200 OK
{
  "content": {
    "issues": [...]
  },
  "isError": false,
  "durationMs": 1234
}
```

#### Initiate OAuth 2 Flow
```
POST /api/mcp/servers/{id}/oauth2/initiate
Authorization: Bearer {token}
Content-Type: application/json

{
  "redirect_url": "https://yourdomain.com/mcp/callback"
}

Response: 200 OK
{
  "authorization_url": "https://auth.atlassian.com/authorize?client_id=...&state=...&code_challenge=..."
}
```

#### OAuth 2 Callback (Public)
```
GET /api/mcp/oauth2/callback?state={state}&code={code}

Response: 200 OK
{
  "message": "Authorization successful",
  "server": {...}
}
```

---

## OAuth 2 Setup

### Atlassian (Jira/Confluence)

1. **Create OAuth 2 App** at https://developer.atlassian.com/
2. **Get Credentials**:
   - Client ID
   - Client Secret
3. **Configure Redirect URI**: `https://yourdomain.com/api/mcp/oauth2/callback`
4. **Request Scopes**:
   - `read:jira-work` - Read Jira issues
   - `write:jira-work` - Create/update Jira issues
   - `read:confluence-content.all` - Read Confluence pages

5. **Create MCP Server in Dashboard**:
```json
{
  "name": "Atlassian Jira",
  "server_url": "https://mcp.atlassian.com/v1",
  "auth_type": "oauth2",
  "oauth2_config": {
    "client_id": "YOUR_CLIENT_ID",
    "client_secret": "YOUR_CLIENT_SECRET",
    "auth_url": "https://auth.atlassian.com/authorize",
    "token_url": "https://auth.atlassian.com/oauth/token",
    "scopes": ["read:jira-work", "write:jira-work"]
  }
}
```

6. **Initiate Authorization**:
   - Call `/api/mcp/servers/{id}/oauth2/initiate`
   - Redirect user to `authorization_url`
   - User authorizes app
   - Callback handled automatically
   - Tokens stored encrypted in database

### GitHub

1. **Create OAuth App** at https://github.com/settings/developers
2. **Configure**:
   - Homepage URL: `https://yourdomain.com`
   - Callback URL: `https://yourdomain.com/api/mcp/oauth2/callback`
3. **Scopes**: `repo`, `read:user`

4. **Configuration**:
```json
{
  "client_id": "YOUR_GITHUB_CLIENT_ID",
  "client_secret": "YOUR_GITHUB_CLIENT_SECRET",
  "auth_url": "https://github.com/login/oauth/authorize",
  "token_url": "https://github.com/login/oauth/access_token",
  "scopes": ["repo", "read:user"]
}
```

---

## Supported MCP Servers

### 1. Atlassian MCP

**Server URL**: `https://mcp.atlassian.com/v1`
**Auth**: OAuth 2
**Available Tools**:
- `search_issues` - Search Jira issues with JQL
- `create_issue` - Create new Jira issue
- `update_issue` - Update existing issue
- `get_project` - Get project details
- `search_confluence` - Search Confluence pages

**Example Use Cases**:
- "Create a Jira ticket for this bug"
- "Search for open issues in project ABC"
- "What's the status of ticket PROJ-123?"

### 2. GitHub MCP

**Server URL**: `https://mcp.github.com/v1`
**Auth**: OAuth 2 or API Key
**Available Tools**:
- `search_repositories` - Search GitHub repositories
- `create_issue` - Create GitHub issue
- `list_pull_requests` - List PRs for a repository
- `get_file_content` - Get file contents from repository

**Example Use Cases**:
- "Search for React repositories"
- "Create an issue in my-repo about the bug"
- "List open PRs in my-org/my-repo"

### 3. Slack MCP

**Server URL**: `https://mcp.slack.com/v1`
**Auth**: OAuth 2
**Available Tools**:
- `send_message` - Send message to channel
- `search_messages` - Search message history
- `list_channels` - List available channels
- `get_user_info` - Get user information

**Example Use Cases**:
- "Send a message to #general channel"
- "Search for messages about deployment"
- "List all channels in the workspace"

### 4. Custom MCP Servers

You can connect to any MCP-compatible server:

**Requirements**:
- Supports JSON-RPC 2.0 protocol
- Implements MCP protocol methods:
  - `tools/list` - List available tools
  - `tools/call` - Execute a tool
- Provides tool schemas in JSON Schema format

**Authentication Options**:
- **None**: No authentication required
- **API Key**: Bearer token in Authorization header
- **OAuth 2**: Full OAuth 2 flow with PKCE

---

## Testing

### E2E Tests

**Location**: `frontend/tests/e2e/`

**Test Files**:
1. `mcp-servers.spec.ts` - MCP server management (20+ tests)
2. `mcp-chat.spec.ts` - Chat with MCP tools (18+ tests)

**Running Tests**:
```bash
cd frontend

# Run all MCP tests
npm run test:e2e -- mcp

# Run specific test file
npx playwright test tests/e2e/mcp-servers.spec.ts

# Run in UI mode (interactive)
npm run test:e2e:ui

# Run in headed mode (see browser)
npm run test:e2e:headed
```

**Test Coverage**:
- ✅ MCP server CRUD operations
- ✅ Form validation
- ✅ Capability syncing
- ✅ Tool calling in chat
- ✅ OAuth 2 flow (mocked)
- ✅ Error handling
- ✅ Multi-server scenarios
- ✅ Tool usage analytics

### Manual Testing

**Test API Key Auth**:
```bash
# Create server
curl -X POST http://localhost:8080/api/mcp/servers \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Server",
    "server_url": "https://mcp.example.com",
    "auth_type": "api_key",
    "api_key": "test-key"
  }'

# Sync capabilities
curl -X POST http://localhost:8080/api/mcp/servers/{id}/sync \
  -H "Authorization: Bearer YOUR_TOKEN"

# Call tool
curl -X POST http://localhost:8080/api/mcp/servers/{id}/tools/call \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "tool_name": "example_tool",
    "input": {"param": "value"}
  }'
```

---

## Deployment

### Environment Variables

```bash
# Required (already configured)
DATABASE_URL=postgresql://...
JWT_SECRET=your-jwt-secret
ENCRYPTION_KEY=your-32-byte-encryption-key

# Frontend
VITE_API_BASE_URL=https://api.yourdomain.com/api
```

### Database Migration

```bash
# Run MCP migration
psql $DATABASE_URL < backend/migrations/003_mcp_integration.sql

# Verify tables created
psql $DATABASE_URL -c "\dt mcp_*"
psql $DATABASE_URL -c "\dt oauth2_*"
```

### Deployment Steps

1. **Backend**:
```bash
cd backend
go build -o /tmp/rag-dashboard ./cmd/server
./rag-dashboard
```

2. **Frontend**:
```bash
cd frontend
npm run build
# Deploy dist/ folder to your hosting
```

3. **Railway Deployment**:
   - Backend auto-deploys on push to main
   - Frontend deploys to Vercel/Netlify
   - Migration runs automatically

---

## Troubleshooting

### Common Issues

#### 1. OAuth 2 Authorization Fails

**Symptoms**: "Invalid state" or "State expired" error

**Solutions**:
- Check redirect URL matches exactly (including trailing slash)
- Verify OAuth state hasn't expired (10-minute limit)
- Ensure server has correct OAuth 2 config
- Check client ID and secret are correct

#### 2. Tool Call Fails

**Symptoms**: "Tool call failed" error in chat

**Solutions**:
- Verify MCP server is active (`is_active = true`)
- Sync capabilities to ensure tools are loaded
- Check API key or OAuth token is valid
- Verify server URL is correct and accessible
- Check tool input matches expected schema

#### 3. OAuth Token Expired

**Symptoms**: "Unauthorized" or "Token expired" error

**Solutions**:
- System should auto-refresh (5-minute buffer)
- Manual refresh: `POST /api/mcp/servers/{id}/oauth2/refresh`
- If refresh fails, re-authorize the server
- Check refresh token is valid

#### 4. MCP Server Not Listed

**Symptoms**: Server created but doesn't appear in list

**Solutions**:
- Check server belongs to correct profile
- Verify `is_active = true`
- Refresh page or call `fetchServers()` again
- Check for errors in browser console

#### 5. Chat Doesn't Use MCP Tools

**Symptoms**: Chat works but tools never called

**Solutions**:
- Verify MCP servers enabled in chat configuration
- Check `enabled_mcp_tools` field has server IDs
- Sync capabilities to ensure tools are loaded
- Try explicit tool-related prompts
- Check OpenAI model supports function calling (gpt-4, gpt-3.5-turbo)

### Debug Mode

**Enable Debug Logging**:
```go
// In chat_service.go
fmt.Printf("MCP tools available: %d\n", len(mcpTools))
fmt.Printf("Tool call: %s with args: %v\n", toolName, functionArgs)
```

**Check Database**:
```sql
-- View MCP servers
SELECT * FROM mcp_servers;

-- View tool usage
SELECT * FROM mcp_tool_usage ORDER BY created_at DESC LIMIT 10;

-- View OAuth states
SELECT * FROM oauth2_states WHERE expires_at > NOW();
```

---

## Best Practices

### Security

1. **Never log sensitive data**:
   - API keys, OAuth tokens, client secrets
   - Use error sanitization

2. **Rotate credentials regularly**:
   - Update API keys every 90 days
   - Monitor OAuth token usage

3. **Use HTTPS**:
   - Always use HTTPS for MCP server URLs
   - Enforce HTTPS for OAuth callbacks

4. **Limit tool permissions**:
   - Request minimum OAuth scopes needed
   - Use read-only tools when possible

### Performance

1. **Cache capabilities**:
   - Sync capabilities periodically (not every request)
   - Cache tool schemas in memory

2. **Limit function call iterations**:
   - Default: 5 iterations max
   - Prevents infinite loops

3. **Monitor tool usage**:
   - Track duration and errors
   - Identify slow tools
   - Optimize frequently-used tools

### User Experience

1. **Provide clear feedback**:
   - Show when tools are being called
   - Display tool call results
   - Handle errors gracefully

2. **Document available tools**:
   - List tools in UI
   - Show tool descriptions
   - Provide usage examples

3. **Allow tool configuration**:
   - Enable/disable specific tools
   - Configure tool parameters
   - Set tool timeouts

---

## Support & Resources

### Documentation
- **MCP Protocol**: https://modelcontextprotocol.io/
- **OpenAI Function Calling**: https://platform.openai.com/docs/guides/function-calling
- **OAuth 2 with PKCE**: https://oauth.net/2/pkce/

### Repositories
- **RAG Dashboard**: /michlo23/RaguDashu
- **MCP Specification**: https://github.com/modelcontextprotocol/

### Contact
- **GitHub Issues**: https://github.com/michlo23/RaguDashu/issues
- **Documentation**: See PROJECT_STATUS.md and MCP_INTEGRATION_STATUS.md

---

**End of MCP Integration Guide**
