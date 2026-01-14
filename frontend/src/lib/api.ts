import axios from 'axios';
import { API_URL, CHAT_ENDPOINT, INGEST_ENDPOINT } from '@/env';
import { getToken } from './auth';
import type { ChatRequest, ChatResponse, IngestRequest, IngestResponse, Session } from './types';

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const chatApi = {
  sendMessage: (data: ChatRequest) => api.post<ChatResponse>(CHAT_ENDPOINT, data),
  getSessions: () => api.get<Session[]>('/sessions'),
  createSession: (name: string) => api.post<Session>('/sessions', { name }),
};

export const ingestApi = {
  uploadDocument: (data: IngestRequest) => api.post<IngestResponse>(INGEST_ENDPOINT, data),
};

export default api;
