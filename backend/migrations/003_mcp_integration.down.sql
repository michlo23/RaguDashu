-- Rollback Migration 003: MCP Integration

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_mcp_servers_updated_at ON mcp_servers;
DROP FUNCTION IF EXISTS update_mcp_servers_updated_at();

-- Remove column from chat_configurations
ALTER TABLE chat_configurations DROP COLUMN IF EXISTS enabled_mcp_tools;

-- Drop tables (in reverse order of dependencies)
DROP TABLE IF EXISTS mcp_tool_usage;
DROP TABLE IF EXISTS oauth2_states;
DROP TABLE IF EXISTS mcp_servers;
