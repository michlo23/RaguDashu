export interface ChatConfiguration {
  id: string;
  name: string;
  model: string;
  index_id?: string;
  system_prompt: string;
  temperature: number;
  max_tokens: number;
  top_k: number;
  is_default: boolean;
}

export interface ChatConversation {
  id: string;
  config_id?: string;
  title: string;
  is_saved: boolean;
  message_count: number;
  last_message_at?: string;
  created_at: string;
}

export interface ChatMessage {
  id: string;
  conversation_id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  retrieved_context?: any;
  created_at: string;
}

export interface ChatRequest {
  conversation_id?: string;
  config_id: string;
  message: string;
}

export interface ChatResponse {
  conversation_id: string;
  message: string;
  context?: Array<{
    text: string;
    score: number;
    source: string;
  }>;
}
