'use client';

import { createContext, useContext, useState, ReactNode } from 'react';
import { Message, Session } from '@/lib/types';

interface ChatContextType {
  messages: Message[];
  setMessages: (messages: Message[]) => void;
  currentSession: Session | null;
  setCurrentSession: (session: Session | null) => void;
  addMessage: (message: Message) => void;
  clearMessages: () => void;
}

const ChatContext = createContext<ChatContextType | undefined>(undefined);

export function ChatProvider({ children }: { children: ReactNode }) {
  const [messages, setMessages] = useState<Message[]>([]);
  const [currentSession, setCurrentSession] = useState<Session | null>(null);

  const addMessage = (message: Message) => {
    setMessages((prev) => [...prev, message]);
  };

  const clearMessages = () => {
    setMessages([]);
  };

  return (
    <ChatContext.Provider
      value={{
        messages,
        setMessages,
        currentSession,
        setCurrentSession,
        addMessage,
        clearMessages,
      }}
    >
      {children}
    </ChatContext.Provider>
  );
}

export function useChatContext() {
  const context = useContext(ChatContext);
  if (!context) {
    throw new Error('useChatContext must be used within ChatProvider');
  }
  return context;
}
