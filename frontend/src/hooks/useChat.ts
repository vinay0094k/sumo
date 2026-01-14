import { useState } from 'react';
import { chatApi } from '@/lib/api';
import { useChatContext } from '@/context/ChatContext';
import { Message } from '@/lib/types';

export function useChat() {
  const { messages, addMessage, currentSession } = useChatContext();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const sendMessage = async (content: string) => {
    if (!currentSession) {
      setError('No session selected');
      return;
    }

    const userMessage: Message = {
      id: Date.now().toString(),
      content,
      role: 'user',
      timestamp: new Date().toISOString(),
    };

    addMessage(userMessage);
    setLoading(true);
    setError(null);

    try {
      const { data } = await chatApi.sendMessage({
        sessionId: currentSession.id,
        message: content,
      });

      const aiMessage: Message = {
        id: (Date.now() + 1).toString(),
        content: data.reply,
        role: 'assistant',
        timestamp: new Date().toISOString(),
      };

      addMessage(aiMessage);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to send message');
    } finally {
      setLoading(false);
    }
  };

  return { messages, sendMessage, loading, error };
}
