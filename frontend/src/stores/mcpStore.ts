import { create } from 'zustand';
import { mcpService, MCPServer, CreateMCPServerRequest, UpdateMCPServerRequest } from '../services/mcp';

interface MCPState {
  servers: MCPServer[];
  currentServer: MCPServer | null;
  isLoading: boolean;
  error: string | null;

  // Actions
  fetchServers: () => Promise<void>;
  fetchServer: (id: string) => Promise<void>;
  createServer: (data: CreateMCPServerRequest) => Promise<MCPServer>;
  updateServer: (id: string, data: UpdateMCPServerRequest) => Promise<MCPServer>;
  deleteServer: (id: string) => Promise<void>;
  syncCapabilities: (id: string) => Promise<void>;
  refreshToken: (id: string) => Promise<void>;
  clearError: () => void;
  setCurrentServer: (server: MCPServer | null) => void;
}

export const useMCPStore = create<MCPState>((set, get) => ({
  servers: [],
  currentServer: null,
  isLoading: false,
  error: null,

  fetchServers: async () => {
    set({ isLoading: true, error: null });
    try {
      const servers = await mcpService.listServers();
      // Parse JSON fields
      const parsedServers = servers.map(server => ({
        ...server,
        available_tools: server.available_tools ? (typeof server.available_tools === 'string' ? JSON.parse(server.available_tools) : server.available_tools) : [],
        available_resources: server.available_resources ? (typeof server.available_resources === 'string' ? JSON.parse(server.available_resources) : server.available_resources) : [],
        available_prompts: server.available_prompts ? (typeof server.available_prompts === 'string' ? JSON.parse(server.available_prompts) : server.available_prompts) : [],
      }));
      set({ servers: parsedServers, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Failed to fetch MCP servers',
        isLoading: false
      });
      throw error;
    }
  },

  fetchServer: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      const server = await mcpService.getServer(id);
      // Parse JSON fields
      const parsedServer = {
        ...server,
        available_tools: server.available_tools ? (typeof server.available_tools === 'string' ? JSON.parse(server.available_tools) : server.available_tools) : [],
        available_resources: server.available_resources ? (typeof server.available_resources === 'string' ? JSON.parse(server.available_resources) : server.available_resources) : [],
        available_prompts: server.available_prompts ? (typeof server.available_prompts === 'string' ? JSON.parse(server.available_prompts) : server.available_prompts) : [],
      };
      set({ currentServer: parsedServer, isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Failed to fetch MCP server',
        isLoading: false
      });
      throw error;
    }
  },

  createServer: async (data: CreateMCPServerRequest) => {
    set({ isLoading: true, error: null });
    try {
      const server = await mcpService.createServer(data);
      // Refresh server list
      await get().fetchServers();
      set({ isLoading: false });
      return server;
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Failed to create MCP server',
        isLoading: false
      });
      throw error;
    }
  },

  updateServer: async (id: string, data: UpdateMCPServerRequest) => {
    set({ isLoading: true, error: null });
    try {
      const server = await mcpService.updateServer(id, data);
      // Refresh server list
      await get().fetchServers();
      set({ isLoading: false });
      return server;
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Failed to update MCP server',
        isLoading: false
      });
      throw error;
    }
  },

  deleteServer: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      await mcpService.deleteServer(id);
      // Remove from local state
      set(state => ({
        servers: state.servers.filter(s => s.id !== id),
        currentServer: state.currentServer?.id === id ? null : state.currentServer,
        isLoading: false,
      }));
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Failed to delete MCP server',
        isLoading: false
      });
      throw error;
    }
  },

  syncCapabilities: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      await mcpService.syncCapabilities(id);
      // Refresh server to get updated capabilities
      await get().fetchServer(id);
      // Also refresh server list
      await get().fetchServers();
      set({ isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Failed to sync capabilities',
        isLoading: false
      });
      throw error;
    }
  },

  refreshToken: async (id: string) => {
    set({ isLoading: true, error: null });
    try {
      await mcpService.refreshToken(id);
      // Refresh server to get updated token info
      await get().fetchServer(id);
      set({ isLoading: false });
    } catch (error: any) {
      set({
        error: error.response?.data?.error || 'Failed to refresh token',
        isLoading: false
      });
      throw error;
    }
  },

  clearError: () => set({ error: null }),

  setCurrentServer: (server: MCPServer | null) => set({ currentServer: server }),
}));
