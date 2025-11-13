-- Migration: 002_enhanced_features
-- Description: Add folders, tags, analytics, audit logs, webhooks, templates, sessions, 2FA
-- Created: 2025-01-13

-- Document folders for organization
CREATE TABLE IF NOT EXISTS document_folders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    parent_id UUID REFERENCES document_folders(id) ON DELETE CASCADE,
    color VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Document tags
CREATE TABLE IF NOT EXISTS document_tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_id, name)
);

-- Usage metrics for analytics
CREATE TABLE IF NOT EXISTS usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    documents_count INTEGER DEFAULT 0,
    search_count INTEGER DEFAULT 0,
    chat_count INTEGER DEFAULT 0,
    tokens_used INTEGER DEFAULT 0,
    storage_bytes BIGINT DEFAULT 0,
    api_calls_count INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(profile_id, date)
);

-- Audit logs for compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    profile_id UUID REFERENCES user_profiles(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_id UUID,
    details JSONB,
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Webhooks
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(1000) NOT NULL,
    events TEXT[],
    secret VARCHAR(500),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Conversation templates
CREATE TABLE IF NOT EXISTS conversation_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID REFERENCES user_profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(50),
    model VARCHAR(100) NOT NULL,
    system_prompt TEXT NOT NULL,
    temperature DECIMAL(2,1) DEFAULT 0.7,
    max_tokens INTEGER DEFAULT 1000,
    top_k INTEGER DEFAULT 5,
    is_public BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- User sessions for security
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(500) UNIQUE NOT NULL,
    ip_address VARCHAR(50),
    user_agent VARCHAR(500),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Two-factor authentication
CREATE TABLE IF NOT EXISTS two_factor_auth (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    secret VARCHAR(500) NOT NULL,
    enabled BOOLEAN DEFAULT false,
    backup_codes TEXT[],
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Add columns to existing documents table
ALTER TABLE documents ADD COLUMN IF NOT EXISTS tags TEXT[];
ALTER TABLE documents ADD COLUMN IF NOT EXISTS folder_id UUID REFERENCES document_folders(id) ON DELETE SET NULL;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS is_favorite BOOLEAN DEFAULT false;
ALTER TABLE documents ADD COLUMN IF NOT EXISTS metadata JSONB;

-- Add columns to existing chat_conversations table
ALTER TABLE chat_conversations ADD COLUMN IF NOT EXISTS folder VARCHAR(255);
ALTER TABLE chat_conversations ADD COLUMN IF NOT EXISTS is_favorite BOOLEAN DEFAULT false;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_document_folders_profile_id ON document_folders(profile_id);
CREATE INDEX IF NOT EXISTS idx_document_folders_parent_id ON document_folders(parent_id);
CREATE INDEX IF NOT EXISTS idx_document_tags_profile_id ON document_tags(profile_id);
CREATE INDEX IF NOT EXISTS idx_usage_metrics_profile_date ON usage_metrics(profile_id, date);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_profile_id ON audit_logs(profile_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_webhooks_profile_id ON webhooks(profile_id);
CREATE INDEX IF NOT EXISTS idx_conversation_templates_profile_id ON conversation_templates(profile_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_two_factor_auth_user_id ON two_factor_auth(user_id);
CREATE INDEX IF NOT EXISTS idx_documents_folder_id ON documents(folder_id);
CREATE INDEX IF NOT EXISTS idx_documents_tags ON documents USING GIN(tags);

-- Apply updated_at triggers
CREATE TRIGGER update_document_folders_updated_at BEFORE UPDATE ON document_folders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_usage_metrics_updated_at BEFORE UPDATE ON usage_metrics
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_conversation_templates_updated_at BEFORE UPDATE ON conversation_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_two_factor_auth_updated_at BEFORE UPDATE ON two_factor_auth
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default conversation templates
INSERT INTO conversation_templates (profile_id, name, description, icon, model, system_prompt, temperature, max_tokens, top_k, is_public)
VALUES
    (NULL, 'Code Assistant', 'Analyzes code documentation and helps with programming questions', '💻', 'gpt-4', 'You are a expert programming assistant with access to code documentation. Help users understand code, debug issues, and suggest improvements. Always cite specific documentation when available.', 0.3, 1500, 7, true),
    (NULL, 'Research Assistant', 'Helps with academic research and paper analysis', '📚', 'gpt-4', 'You are a research assistant helping with academic work. Analyze research papers, summarize findings, and provide insights. Always cite sources and maintain academic rigor.', 0.5, 2000, 10, true),
    (NULL, 'Customer Support', 'Answers customer questions using support documentation', '🎧', 'gpt-3.5-turbo', 'You are a helpful customer support agent. Use the knowledge base to answer questions clearly and professionally. Be empathetic and solution-focused.', 0.7, 1000, 5, true),
    (NULL, 'Meeting Summarizer', 'Summarizes meeting notes and action items', '📝', 'gpt-4', 'You are a meeting assistant. Summarize key points, action items, and decisions from meeting transcripts. Be concise and organized.', 0.5, 1500, 5, true),
    (NULL, 'Creative Writer', 'Helps with creative writing using reference materials', '✍️', 'gpt-4', 'You are a creative writing assistant. Help users write stories, articles, and content. Use reference materials for inspiration while maintaining creativity.', 0.9, 2000, 3, true)
ON CONFLICT DO NOTHING;
