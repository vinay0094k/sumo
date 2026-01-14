import { useState } from 'react';
import { login as loginApi, setToken } from '@/lib/auth';
import { useAuthContext } from '@/context/AuthContext';

export function useAuth() {
  const { setUser, logout } = useAuthContext();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const login = async (email: string, password: string) => {
    setLoading(true);
    setError(null);
    try {
      const data = await loginApi(email, password);
      setToken(data.token);
      setUser(data.user);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed');
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { login, logout, loading, error };
}
