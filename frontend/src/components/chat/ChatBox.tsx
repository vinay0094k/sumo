'use client';

import { ChatHistory } from './ChatHistory';
import { MessageInput } from './MessageInput';

export function ChatBox() {
  return (
    <div className="flex flex-col h-full">
      <ChatHistory />
      <MessageInput />
    </div>
  );
}
