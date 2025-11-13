-- Rollback migration: 002_enhanced_features

-- Drop triggers
DROP TRIGGER IF EXISTS update_document_folders_updated_at ON document_folders;
DROP TRIGGER IF EXISTS update_usage_metrics_updated_at ON usage_metrics;
DROP TRIGGER IF EXISTS update_webhooks_updated_at ON webhooks;
DROP TRIGGER IF EXISTS update_conversation_templates_updated_at ON conversation_templates;
DROP TRIGGER IF EXISTS update_two_factor_auth_updated_at ON two_factor_auth;

-- Remove columns from existing tables
ALTER TABLE documents DROP COLUMN IF EXISTS tags;
ALTER TABLE documents DROP COLUMN IF EXISTS folder_id;
ALTER TABLE documents DROP COLUMN IF EXISTS is_favorite;
ALTER TABLE documents DROP COLUMN IF EXISTS metadata;

ALTER TABLE chat_conversations DROP COLUMN IF EXISTS folder;
ALTER TABLE chat_conversations DROP COLUMN IF EXISTS is_favorite;

-- Drop tables in reverse order
DROP TABLE IF EXISTS two_factor_auth CASCADE;
DROP TABLE IF EXISTS user_sessions CASCADE;
DROP TABLE IF EXISTS conversation_templates CASCADE;
DROP TABLE IF EXISTS webhooks CASCADE;
DROP TABLE IF EXISTS audit_logs CASCADE;
DROP TABLE IF EXISTS usage_metrics CASCADE;
DROP TABLE IF EXISTS document_tags CASCADE;
DROP TABLE IF EXISTS document_folders CASCADE;
