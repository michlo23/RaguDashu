import { api } from './api';

export interface MCPServer {
  id: string;
  profile_id: string;
  name: string;
  description: string;
  server_url: string;
  auth_type: 'none' | 'api_key' | 'oauth2';
  available_tools?: MCPTool[];
  available_resources?: MCPResource[];
  available_prompts?: MCPPrompt[];
  is_active: boolean;
  last_synced_at?: string;
  last_tested_at?: string;
  test_result?: string;
  created_at: string;
  updated_at: string;
}

export interface MCPTool {
  name: string;
  description: string;
  inputSchema: Record<string, any>;
}

export interface MCPResource {
  uri: string;
  name: string;
  description: string;
  mimeType: string;
}

export interface MCPPrompt {
  name: string;
  description: string;
  arguments: Array<{
    name: string;
    description: string;
    required: boolean;
  }>;
}

export interface OAuth2Config {
  client_id: string;
  client_secret: string;
  auth_url: string;
  token_url: string;
  scopes: string[];
}

export interface CreateMCPServerRequest {
  name: string;
  description: string;
  server_url: string;
  auth_type: 'none' | 'api_key' | 'oauth2';
  api_key?: string;
  oauth2_config?: OAuth2Config;
}

export interface UpdateMCPServerRequest {
  name?: string;
  description?: string;
  api_key?: string;
  is_active?: boolean;
}

export interface ToolCallRequest {
  tool_name: string;
  input: Record<string, any>;
}

export interface ToolCallResponse {
  content: any;
  isError: boolean;
  durationMs: number;
}

export interface OAuth2InitiateRequest {
  redirect_url: string;
}

export interface OAuth2InitiateResponse {
  authorization_url: string;
}

export const mcpService = {
  // List all MCP servers
  async listServers(): Promise<MCPServer[]> {
    const response = await api.get('/mcp/servers');
    return response.data;
  },

  // Get MCP server by ID
  async getServer(id: string): Promise<MCPServer> {
    const response = await api.get(`/mcp/servers/${id}`);
    return response.data;
  },

  // Create new MCP server
  async createServer(data: CreateMCPServerRequest): Promise<MCPServer> {
    const response = await api.post('/mcp/servers', data);
    return response.data;
  },

  // Update MCP server
  async updateServer(id: string, data: UpdateMCPServerRequest): Promise<MCPServer> {
    const response = await api.put(`/mcp/servers/${id}`, data);
    return response.data;
  },

  // Delete MCP server
  async deleteServer(id: string): Promise<void> {
    await api.delete(`/mcp/servers/${id}`);
  },

  // Sync capabilities from MCP server
  async syncCapabilities(id: string): Promise<void> {
    await api.post(`/mcp/servers/${id}/sync`);
  },

  // Call MCP tool
  async callTool(id: string, request: ToolCallRequest): Promise<ToolCallResponse> {
    const response = await api.post(`/mcp/servers/${id}/tools/call`, request);
    return response.data;
  },

  // Initiate OAuth 2 flow
  async initiateOAuth2(id: string, request: OAuth2InitiateRequest): Promise<OAuth2InitiateResponse> {
    const response = await api.post(`/mcp/servers/${id}/oauth2/initiate`, request);
    return response.data;
  },

  // Refresh OAuth 2 token
  async refreshToken(id: string): Promise<void> {
    await api.post(`/mcp/servers/${id}/oauth2/refresh`);
  },
};
