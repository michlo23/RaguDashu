-- Migration 003: MCP (Model Context Protocol) Integration
-- Adds support for connecting to external MCP servers with OAuth 2 authorization

-- MCP Servers table
CREATE TABLE IF NOT EXISTS mcp_servers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    server_url VARCHAR(500) NOT NULL,
    auth_type VARCHAR(50) NOT NULL CHECK (auth_type IN ('none', 'api_key', 'oauth2')),

    -- API Key authentication (encrypted)
    encrypted_api_key TEXT,

    -- OAuth 2 configuration (JSONB)
    oauth2_config JSONB,

    -- OAuth 2 tokens (encrypted)
    encrypted_access_token TEXT,
    encrypted_refresh_token TEXT,
    token_expires_at TIMESTAMP,

    -- MCP capabilities (cached from server)
    available_tools JSONB,
    available_resources JSONB,
    available_prompts JSONB,
    last_synced_at TIMESTAMP,

    -- Status tracking
    is_active BOOLEAN DEFAULT true,
    last_tested_at TIMESTAMP,
    test_result VARCHAR(255),
    last_used_at TIMESTAMP,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT mcp_servers_profile_name_unique UNIQUE (profile_id, name)
);

-- Indexes for MCP servers
CREATE INDEX IF NOT EXISTS idx_mcp_servers_profile_id ON mcp_servers(profile_id);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_is_active ON mcp_servers(is_active);
CREATE INDEX IF NOT EXISTS idx_mcp_servers_auth_type ON mcp_servers(auth_type);

-- OAuth 2 states table (temporary storage for OAuth flow)
CREATE TABLE IF NOT EXISTS oauth2_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    state VARCHAR(64) NOT NULL UNIQUE,
    profile_id UUID NOT NULL REFERENCES user_profiles(id) ON DELETE CASCADE,
    mcp_server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    code_verifier VARCHAR(128) NOT NULL,
    redirect_url VARCHAR(500) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for OAuth 2 states
CREATE INDEX IF NOT EXISTS idx_oauth2_states_state ON oauth2_states(state);
CREATE INDEX IF NOT EXISTS idx_oauth2_states_expires_at ON oauth2_states(expires_at);

-- MCP tool usage tracking
CREATE TABLE IF NOT EXISTS mcp_tool_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES chat_conversations(id) ON DELETE CASCADE,
    message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
    mcp_server_id UUID NOT NULL REFERENCES mcp_servers(id) ON DELETE CASCADE,
    tool_name VARCHAR(255) NOT NULL,
    tool_input JSONB,
    tool_output JSONB,
    error_message TEXT,
    duration_ms INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for MCP tool usage
CREATE INDEX IF NOT EXISTS idx_mcp_tool_usage_conversation_id ON mcp_tool_usage(conversation_id);
CREATE INDEX IF NOT EXISTS idx_mcp_tool_usage_message_id ON mcp_tool_usage(message_id);
CREATE INDEX IF NOT EXISTS idx_mcp_tool_usage_mcp_server_id ON mcp_tool_usage(mcp_server_id);
CREATE INDEX IF NOT EXISTS idx_mcp_tool_usage_created_at ON mcp_tool_usage(created_at);

-- Add enabled_mcp_tools column to chat_configurations
ALTER TABLE chat_configurations ADD COLUMN IF NOT EXISTS enabled_mcp_tools JSONB DEFAULT '[]'::jsonb;

-- Create a trigger to update updated_at timestamp on mcp_servers
CREATE OR REPLACE FUNCTION update_mcp_servers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_mcp_servers_updated_at
    BEFORE UPDATE ON mcp_servers
    FOR EACH ROW
    EXECUTE FUNCTION update_mcp_servers_updated_at();

-- Comments for documentation
COMMENT ON TABLE mcp_servers IS 'Configuration for MCP (Model Context Protocol) servers that provide external tools and resources';
COMMENT ON COLUMN mcp_servers.auth_type IS 'Authentication method: none, api_key, or oauth2';
COMMENT ON COLUMN mcp_servers.oauth2_config IS 'OAuth 2 configuration including client_id, client_secret, auth_url, token_url, and scopes';
COMMENT ON COLUMN mcp_servers.available_tools IS 'Cached list of tools available from the MCP server';
COMMENT ON COLUMN mcp_servers.available_resources IS 'Cached list of resources available from the MCP server';
COMMENT ON COLUMN mcp_servers.available_prompts IS 'Cached list of prompts available from the MCP server';

COMMENT ON TABLE oauth2_states IS 'Temporary storage for OAuth 2 authorization flow state (expires after 10 minutes)';
COMMENT ON COLUMN oauth2_states.code_verifier IS 'PKCE code verifier for secure OAuth 2 flow';

COMMENT ON TABLE mcp_tool_usage IS 'Tracking of MCP tool calls made during chat conversations';
COMMENT ON COLUMN mcp_tool_usage.tool_input IS 'Input parameters passed to the tool';
COMMENT ON COLUMN mcp_tool_usage.tool_output IS 'Response from the tool';
COMMENT ON COLUMN mcp_tool_usage.duration_ms IS 'Time taken for tool call in milliseconds';

COMMENT ON COLUMN chat_configurations.enabled_mcp_tools IS 'Array of MCP server IDs enabled for this chat configuration';
