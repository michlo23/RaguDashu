import api from './api';
import { ChatRequest, ChatResponse, ChatConversation, ChatConfiguration } from '../types/chat';

export const chatService = {
  async sendMessage(data: ChatRequest): Promise<ChatResponse> {
    const response = await api.post<ChatResponse>('/chat/send', data);
    return response.data;
  },

  async getConversation(id: string): Promise<ChatConversation> {
    const response = await api.get<ChatConversation>(`/chat/conversations/${id}`);
    return response.data;
  },

  async listConversations(): Promise<ChatConversation[]> {
    const response = await api.get<ChatConversation[]>('/chat/conversations');
    return response.data;
  },

  async saveConversation(id: string): Promise<void> {
    await api.post(`/chat/conversations/${id}/save`);
  },

  async deleteConversation(id: string): Promise<void> {
    await api.delete(`/chat/conversations/${id}`);
  },

  async createConfiguration(data: Partial<ChatConfiguration>): Promise<ChatConfiguration> {
    const response = await api.post<ChatConfiguration>('/chat/configurations', data);
    return response.data;
  },

  async listConfigurations(): Promise<ChatConfiguration[]> {
    const response = await api.get<ChatConfiguration[]>('/chat/configurations');
    return response.data;
  },
};
