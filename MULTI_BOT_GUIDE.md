# Multi-Bot RAG System Guide

**Last Updated**: 2025-11-13
**Feature Status**: ✅ **Fully Implemented**

---

## Overview

The RAG Dashboard supports **multiple RAG bots per user**, each with independent:
- ✅ **Knowledge sources** (different document sets)
- ✅ **MCP tool access** (different external integrations)
- ✅ **AI behavior** (different personalities, models, temperatures)
- ✅ **Complete data isolation** (multi-tenant architecture)

---

## 🔒 Data Separation & Security

### Multi-Tenant Architecture

**Every user's data is completely isolated** using profile-based separation:

```sql
-- All data tables include profile_id for isolation
user_profiles (id, email, ...)
  ├── user_credentials (profile_id, ...)
  ├── documents (profile_id, ...)
  ├── pinecone_indexes (profile_id, ...)
  ├── chat_configurations (profile_id, ...)  ← RAG Bots
  ├── chat_conversations (profile_id, ...)
  ├── chat_messages (via conversation)
  ├── mcp_servers (profile_id, ...)
  ├── mcp_tool_usage (via conversation)
  └── webhooks (profile_id, ...)
```

### Security Enforcement

All API endpoints enforce profile-based isolation:

```go
// Every protected endpoint uses this pattern
profileID, ok := middleware.MustGetProfileID(c)
if !ok {
    return // 401 Unauthorized
}

// All database queries filter by profile_id
db.Where("profile_id = ?", profileID).Find(&records)
```

**Result**: Users can ONLY access their own data. No cross-user data leakage possible.

---

## 🤖 Multiple RAG Bots (Chat Configurations)

### Concept

A **Chat Configuration** is a "RAG Bot" - a configured AI assistant with:
- Specific knowledge base (Pinecone index)
- Specific tool access (MCP servers)
- Specific personality (system prompt)
- Specific AI settings (model, temperature)

### Database Model

```go
type ChatConfiguration struct {
    ID              uuid.UUID      // Unique bot ID
    ProfileID       uuid.UUID      // Owner (data isolation)
    Name            string         // "Customer Support Bot", "Code Assistant"

    // Knowledge Source
    IndexID         *uuid.UUID     // Which documents to use (NULL = no RAG)
    TopK            int            // How many context chunks (default: 5)

    // MCP Tools (External Integrations)
    EnabledMCPTools datatypes.JSON // Array of MCP server UUIDs

    // AI Behavior
    Model           string         // "gpt-4", "gpt-3.5-turbo", "gpt-4-turbo"
    SystemPrompt    string         // Bot personality and instructions
    Temperature     float32        // 0.0-1.0 (creativity level)
    MaxTokens       int            // Response length limit

    IsDefault       bool           // Quick access bot
    CreatedAt       time.Time
    UpdatedAt       time.Time
}
```

---

## 📋 Usage Examples

### Example 1: Customer Support Bot

**Scenario**: Bot that answers customer questions using product documentation and creates Jira tickets.

```bash
POST /api/chat/configurations
Authorization: Bearer {token}

{
  "name": "Customer Support Bot",
  "index_id": "uuid-of-product-docs-index",
  "enabled_mcp_tools": ["uuid-of-jira-mcp-server"],
  "system_prompt": "You are a helpful customer support assistant. Use the product documentation to answer questions accurately. When a customer reports a bug, create a Jira ticket.",
  "model": "gpt-4",
  "temperature": 0.3,
  "max_tokens": 1000,
  "top_k": 5
}
```

**Capabilities**:
- ✅ Searches product documentation (RAG)
- ✅ Creates Jira tickets (MCP tool)
- ✅ Consistent, accurate responses (low temperature)

---

### Example 2: Engineering Assistant

**Scenario**: Bot that helps with code using engineering docs and accesses GitHub.

```bash
POST /api/chat/configurations

{
  "name": "Engineering Assistant",
  "index_id": "uuid-of-code-docs-index",
  "enabled_mcp_tools": [
    "uuid-of-github-mcp-server",
    "uuid-of-slack-mcp-server"
  ],
  "system_prompt": "You are an expert software engineer. Help with code reviews, debugging, and technical questions. Use the codebase documentation. You can create GitHub issues and send Slack notifications.",
  "model": "gpt-4-turbo",
  "temperature": 0.5,
  "max_tokens": 2000,
  "top_k": 10
}
```

**Capabilities**:
- ✅ Searches code documentation (RAG with more context: top_k=10)
- ✅ Creates GitHub issues (MCP tool)
- ✅ Sends Slack notifications (MCP tool)
- ✅ Moderate creativity (temperature 0.5)

---

### Example 3: Marketing Content Writer

**Scenario**: Bot that generates marketing copy using brand guidelines.

```bash
POST /api/chat/configurations

{
  "name": "Marketing Content Writer",
  "index_id": "uuid-of-brand-docs-index",
  "enabled_mcp_tools": [],
  "system_prompt": "You are a creative marketing copywriter. Write engaging, on-brand content. Follow the brand guidelines in the knowledge base. Be creative and persuasive.",
  "model": "gpt-4",
  "temperature": 0.9,
  "max_tokens": 1500,
  "top_k": 3
}
```

**Capabilities**:
- ✅ Uses brand guidelines (RAG)
- ✅ No external tools (focused on writing)
- ✅ High creativity (temperature 0.9)

---

### Example 4: Data Analyst Bot

**Scenario**: Bot with NO documents (no RAG) but many tool integrations.

```bash
POST /api/chat/configurations

{
  "name": "Data Analyst",
  "index_id": null,
  "enabled_mcp_tools": [
    "uuid-of-database-mcp-server",
    "uuid-of-analytics-mcp-server",
    "uuid-of-spreadsheet-mcp-server"
  ],
  "system_prompt": "You are a data analyst. Query databases, analyze data, and create reports. Use the available tools to access data sources.",
  "model": "gpt-4",
  "temperature": 0.2,
  "max_tokens": 2000,
  "top_k": 5
}
```

**Capabilities**:
- ✅ No RAG (no document knowledge base)
- ✅ Multiple data tool integrations
- ✅ Very consistent (low temperature for data accuracy)

---

## 🔄 Complete Workflow

### 1. Setup Phase

```bash
# Step 1: Upload documents for different knowledge bases
# (Documents automatically go to user's Pinecone indexes)

# Upload customer support docs
POST /api/documents/upload
Files: customer_docs.pdf, faq.pdf, product_guide.pdf

# Upload engineering docs
POST /api/documents/upload
Files: api_docs.pdf, architecture.pdf, coding_standards.pdf

# Upload brand guidelines
POST /api/documents/upload
Files: brand_guide.pdf, tone_of_voice.pdf, examples.pdf

# Step 2: Configure MCP servers
POST /api/mcp/servers
{
  "name": "Jira Integration",
  "server_url": "https://mcp.atlassian.com",
  "auth_type": "oauth2",
  ...
}

POST /api/mcp/servers
{
  "name": "GitHub Integration",
  "server_url": "https://mcp.github.com",
  "auth_type": "api_key",
  ...
}

# Step 3: Create multiple RAG bots
POST /api/chat/configurations
{ "name": "Customer Support Bot", ... }

POST /api/chat/configurations
{ "name": "Engineering Assistant", ... }

POST /api/chat/configurations
{ "name": "Marketing Writer", ... }
```

### 2. Usage Phase

```bash
# List all bots
GET /api/chat/configurations
Response:
[
  { "id": "bot1-uuid", "name": "Customer Support Bot", ... },
  { "id": "bot2-uuid", "name": "Engineering Assistant", ... },
  { "id": "bot3-uuid", "name": "Marketing Writer", ... }
]

# Chat with Customer Support Bot
POST /api/chat/send
{
  "config_id": "bot1-uuid",
  "message": "How do I reset my password?"
}
→ Bot searches customer docs and responds

# Chat with Engineering Assistant
POST /api/chat/send
{
  "config_id": "bot2-uuid",
  "message": "Create a GitHub issue for the login bug"
}
→ Bot creates GitHub issue using MCP tool

# Chat with Marketing Writer
POST /api/chat/send
{
  "config_id": "bot3-uuid",
  "message": "Write a product launch announcement"
}
→ Bot uses brand guidelines to write creative copy
```

### 3. Conversation Management

```bash
# Each bot maintains separate conversations
GET /api/chat/conversations
Response:
[
  { "id": "conv1", "config_id": "bot1-uuid", "title": "Password Reset Help", ... },
  { "id": "conv2", "config_id": "bot2-uuid", "title": "Bug Fixes", ... },
  { "id": "conv3", "config_id": "bot3-uuid", "title": "Launch Copy", ... }
]

# Continue conversation with specific bot
POST /api/chat/send
{
  "conversation_id": "conv1-uuid",
  "config_id": "bot1-uuid",
  "message": "Thanks! What about 2FA setup?"
}
```

---

## 🎯 Key Features

### Per-Bot Configuration

| Feature | Customer Support | Engineering | Marketing |
|---------|-----------------|-------------|-----------|
| **Knowledge Base** | Product docs | Code docs | Brand guidelines |
| **MCP Servers** | Jira | GitHub, Slack | None |
| **Model** | GPT-4 | GPT-4 Turbo | GPT-4 |
| **Temperature** | 0.3 (precise) | 0.5 (balanced) | 0.9 (creative) |
| **Max Tokens** | 1000 | 2000 | 1500 |
| **TopK (RAG)** | 5 chunks | 10 chunks | 3 chunks |

### Data Isolation

```
User A's Data (profile_id = uuid-A)
├── Bot 1: Support (config_id = bot1-A)
│   ├── Knowledge: Product Docs Index A
│   ├── Tools: Jira Server A
│   └── Conversations: conv1-A, conv2-A
└── Bot 2: Engineering (config_id = bot2-A)
    ├── Knowledge: Code Docs Index A
    ├── Tools: GitHub Server A
    └── Conversations: conv3-A, conv4-A

User B's Data (profile_id = uuid-B)  ← COMPLETELY SEPARATE
├── Bot 1: Support (config_id = bot1-B)
│   ├── Knowledge: Product Docs Index B
│   ├── Tools: Jira Server B
│   └── Conversations: conv1-B, conv2-B
└── Bot 2: Custom (config_id = bot2-B)
    ├── Knowledge: Custom Docs Index B
    ├── Tools: Slack Server B
    └── Conversations: conv3-B
```

**No cross-contamination**: User A cannot access User B's bots, documents, conversations, or MCP servers.

---

## 🔧 API Reference

### Create Bot

```
POST /api/chat/configurations
Authorization: Bearer {token}

Request Body:
{
  "name": "string",              // Required: Bot name
  "index_id": "uuid | null",     // Optional: Pinecone index (null = no RAG)
  "enabled_mcp_tools": ["uuid"], // Optional: Array of MCP server IDs
  "system_prompt": "string",     // Required: Bot instructions
  "model": "string",             // Required: "gpt-4", "gpt-3.5-turbo", etc.
  "temperature": 0.7,            // Optional: 0.0-1.0 (default: 0.7)
  "max_tokens": 1000,            // Optional: (default: 1000)
  "top_k": 5,                    // Optional: RAG chunks (default: 5)
  "is_default": false            // Optional: (default: false)
}

Response: 201 Created
{
  "id": "uuid",
  "profile_id": "uuid",
  "name": "Customer Support Bot",
  ...
}
```

### List Bots

```
GET /api/chat/configurations
Authorization: Bearer {token}

Response: 200 OK
[
  {
    "id": "uuid",
    "name": "Customer Support Bot",
    "model": "gpt-4",
    ...
  },
  ...
]
```

### Update Bot

```
PUT /api/chat/configurations/{id}
Authorization: Bearer {token}

Request Body: (all fields optional)
{
  "name": "Updated Name",
  "enabled_mcp_tools": ["new-uuid"],
  "temperature": 0.5,
  ...
}
```

### Delete Bot

```
DELETE /api/chat/configurations/{id}
Authorization: Bearer {token}

Response: 200 OK
```

### Chat with Bot

```
POST /api/chat/send
Authorization: Bearer {token}

Request Body:
{
  "config_id": "uuid",           // Required: Which bot to use
  "conversation_id": "uuid",     // Optional: Continue conversation
  "message": "string"            // Required: User message
}

Response: 200 OK
{
  "conversation_id": "uuid",
  "message": "Bot response...",
  "context": [...]               // RAG context chunks
}
```

---

## 💡 Use Cases

### Enterprise Support Center

**Multiple departments, each with their own bot**:
- Sales Bot (sales docs + CRM tools)
- Technical Support Bot (tech docs + Jira tools)
- Billing Bot (billing docs + payment system tools)
- HR Bot (HR policies + calendar tools)

### Development Team

**Different aspects of software development**:
- Code Review Bot (codebase + GitHub)
- Documentation Bot (docs + wiki tools)
- DevOps Bot (infrastructure docs + deployment tools)
- Bug Triage Bot (bug reports + Jira)

### Content Creation

**Different content types**:
- Blog Writer Bot (blog guidelines + high creativity)
- Email Writer Bot (email templates + moderate creativity)
- Social Media Bot (social guidelines + very high creativity)
- Technical Writer Bot (technical docs + low creativity)

### Personal Assistant Collection

**Individual user with multiple specialized bots**:
- Research Bot (research papers + web search tools)
- Coding Assistant (programming docs + GitHub)
- Writing Helper (writing guides + grammar tools)
- Task Manager (no docs + calendar/task tools)

---

## 🚀 Best Practices

### 1. Bot Naming

✅ **Good**: Descriptive, purpose-clear names
- "Customer Support - Product X"
- "Engineering - Backend Services"
- "Marketing - Blog Content"

❌ **Bad**: Generic, unclear names
- "Bot 1"
- "Assistant"
- "Test"

### 2. Knowledge Base Selection

✅ **Do**:
- Use focused, relevant document sets per bot
- Keep knowledge bases updated
- Use appropriate TopK values (5-10 for most cases)

❌ **Don't**:
- Mix unrelated documents in one index
- Use all documents for every bot
- Set TopK too high (wastes context, adds noise)

### 3. Temperature Settings

| Use Case | Temperature | Reasoning |
|----------|-------------|-----------|
| Customer Support | 0.2 - 0.4 | Need consistent, accurate answers |
| Technical Docs | 0.3 - 0.5 | Balance accuracy and readability |
| Creative Writing | 0.7 - 0.9 | Encourage variety and creativity |
| Data Analysis | 0.1 - 0.3 | Maximize precision |
| Brainstorming | 0.8 - 1.0 | Maximum creativity |

### 4. MCP Tool Selection

✅ **Do**:
- Only enable tools the bot actually needs
- Test tools before enabling in production bots
- Monitor tool usage for errors

❌ **Don't**:
- Enable all tools for every bot
- Mix unrelated tools (e.g., HR + GitHub)
- Ignore tool call errors

### 5. System Prompts

✅ **Good System Prompt**:
```
You are a customer support assistant for ProductX.

Your responsibilities:
- Answer product questions using the documentation
- Create Jira tickets for bugs or feature requests
- Escalate to human support for account issues

Guidelines:
- Be friendly and professional
- Always search the docs before answering
- If unsure, say so and offer to escalate
- Use tools only when explicitly needed
```

❌ **Bad System Prompt**:
```
You are an AI assistant. Help the user.
```

---

## 📊 Monitoring & Analytics

### Per-Bot Metrics

```bash
# Get analytics for specific bot
GET /api/analytics/dashboard?config_id={bot-uuid}

Response:
{
  "total_conversations": 156,
  "total_messages": 1243,
  "avg_response_time": "2.3s",
  "tool_usage": {
    "create_jira_issue": 23,
    "search_github": 45
  },
  "rag_metrics": {
    "avg_chunks_retrieved": 5.2,
    "avg_relevance_score": 0.85
  }
}
```

### Tool Usage by Bot

```sql
-- Tool usage analytics per configuration
SELECT
  c.name AS bot_name,
  m.tool_name,
  COUNT(*) AS usage_count,
  AVG(m.duration_ms) AS avg_duration
FROM mcp_tool_usage m
JOIN chat_conversations conv ON m.conversation_id = conv.id
JOIN chat_configurations c ON conv.config_id = c.id
WHERE c.profile_id = ?
GROUP BY c.name, m.tool_name
ORDER BY usage_count DESC;
```

---

## 🔐 Security Considerations

### 1. Profile Isolation

✅ **Enforced at every level**:
- Database queries filter by `profile_id`
- API middleware validates profile ownership
- Foreign key constraints prevent cross-user access

### 2. MCP Server Access

✅ **Per-user MCP servers**:
- Each user has their own MCP server configurations
- OAuth tokens are user-specific
- API keys encrypted per-user

### 3. Document Access

✅ **Pinecone namespace isolation**:
- Each user's documents in separate namespace: `user_{profile_id}`
- Vector search scoped to user's namespace
- No cross-user document leakage

---

## ❓ FAQ

**Q: Can I share a bot with another user?**
A: Not currently. Each bot belongs to one user. Future feature could add bot templates or exports.

**Q: How many bots can I create?**
A: No hard limit. Create as many as needed. Best practice: 5-10 focused bots rather than 50 generic ones.

**Q: Can one bot access another bot's conversations?**
A: No. Each bot has independent conversations. However, all conversations belong to the same user and can be viewed.

**Q: Can I use the same MCP server in multiple bots?**
A: Yes! One MCP server (e.g., Jira) can be enabled in multiple bots. Each bot will have access to the same tools.

**Q: Can I switch a conversation from one bot to another?**
A: Not directly. Conversations are tied to the bot (config_id) they started with. You'd need to start a new conversation with the other bot.

**Q: What happens if I delete a bot?**
A: Conversations remain (they're not deleted), but you can't send new messages using that config. Best practice: mark inactive rather than delete.

**Q: Can bots share documents?**
A: Yes, indirectly. Documents are uploaded to your Pinecone indexes. Multiple bots can use the same `index_id` to access the same document set.

---

## 📝 Summary

**Multi-Bot System**: ✅ Fully Implemented

**Key Capabilities**:
- ✅ **Unlimited bots** per user
- ✅ **Complete data isolation** (multi-tenant)
- ✅ **Per-bot knowledge bases** (different documents)
- ✅ **Per-bot tool access** (different MCP servers)
- ✅ **Per-bot AI behavior** (different personalities)
- ✅ **Independent conversations** per bot
- ✅ **Full API support** (CRUD operations)

**Security**:
- ✅ Profile-based isolation enforced everywhere
- ✅ No cross-user data access possible
- ✅ Encrypted credentials per user
- ✅ Namespace isolation for documents

**Production Ready**: Yes, fully functional via API. Optional: Build UI for better bot management.

---

**Next Steps**: See `MCP_INTEGRATION_GUIDE.md` for MCP server setup and `FEATURES.md` for complete feature list.
