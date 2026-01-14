import { useState, useEffect } from 'react';
import { chatApi } from '@/lib/api';
import { useChatContext } from '@/context/ChatContext';
import { Session } from '@/lib/types';

export function useSession() {
  const { currentSession, setCurrentSession, clearMessages } = useChatContext();
  const [sessions, setSessions] = useState<Session[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchSessions = async () => {
    setLoading(true);
    setError(null);
    try {
      const { data } = await chatApi.getSessions();
      setSessions(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch sessions');
    } finally {
      setLoading(false);
    }
  };

  const createSession = async (name: string) => {
    setLoading(true);
    setError(null);
    try {
      const { data } = await chatApi.createSession(name);
      setSessions((prev) => [...prev, data]);
      setCurrentSession(data);
      clearMessages();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create session');
    } finally {
      setLoading(false);
    }
  };

  const selectSession = (session: Session) => {
    setCurrentSession(session);
    clearMessages();
  };

  useEffect(() => {
    fetchSessions();
  }, []);

  return {
    sessions,
    currentSession,
    selectSession,
    createSession,
    loading,
    error,
  };
}
