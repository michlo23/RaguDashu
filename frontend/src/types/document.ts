export interface Document {
  id: string;
  filename: string;
  file_type: string;
  file_size: number;
  chunk_count: number;
  upload_status: 'processing' | 'completed' | 'failed';
  error_message?: string;
  created_at: string;
}

export interface SearchResult {
  id: string;
  score: number;
  text: string;
  source: string;
  metadata: Record<string, any>;
}
