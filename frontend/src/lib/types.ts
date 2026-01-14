export interface User {
  id: string;
  email: string;
  name: string;
}

export interface Message {
  id: string;
  content: string;
  role: 'user' | 'assistant';
  timestamp: string;
}

export interface Session {
  id: string;
  name: string;
  userId: string;
  createdAt: string;
}

export interface Document {
  id: string;
  name: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  uploadedAt: string;
}

export interface ChatRequest {
  sessionId: string;
  message: string;
}

export interface ChatResponse {
  reply: string;
  sessionId: string;
}

export interface IngestRequest {
  documentName: string;
  text: string;
}

export interface IngestResponse {
  message: string;
  chunks: number;
}
