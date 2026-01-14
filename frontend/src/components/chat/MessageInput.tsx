'use client';

import { useState, FormEvent } from 'react';
import { useChat } from '@/hooks/useChat';
import { Button } from '@/components/ui/button';
import { Textarea } from '@/components/ui/textarea';
import { Send } from 'lucide-react';

export function MessageInput() {
  const [message, setMessage] = useState('');
  const { sendMessage, loading } = useChat();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (!message.trim() || loading) return;

    await sendMessage(message);
    setMessage('');
  };

  return (
    <form onSubmit={handleSubmit} className="flex gap-2 p-4 border-t bg-white">
      <Textarea
        value={message}
        onChange={(e) => setMessage(e.target.value)}
        placeholder="Type your message..."
        className="resize-none"
        rows={2}
        disabled={loading}
        onKeyDown={(e) => {
          if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSubmit(e);
          }
        }}
      />
      <Button type="submit" disabled={loading || !message.trim()} size="icon">
        <Send size={20} />
      </Button>
    </form>
  );
}
