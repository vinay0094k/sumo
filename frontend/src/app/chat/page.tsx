'use client';

import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { SessionSelector } from '@/components/chat/SessionSelector';
import { ChatBox } from '@/components/chat/ChatBox';

export default function ChatPage() {
  return (
    <ProtectedRoute>
      <div className="flex h-[calc(100vh-8rem)]">
        <div className="w-64">
          <SessionSelector />
        </div>
        <div className="flex-1">
          <ChatBox />
        </div>
      </div>
    </ProtectedRoute>
  );
}
