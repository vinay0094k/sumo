'use client';

import { useState } from 'react';
import { useSession } from '@/hooks/useSession';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Plus } from 'lucide-react';

export function SessionSelector() {
  const { sessions, currentSession, selectSession, createSession, loading } = useSession();
  const [newSessionName, setNewSessionName] = useState('');
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const handleCreateSession = async () => {
    if (!newSessionName.trim()) return;
    await createSession(newSessionName);
    setNewSessionName('');
    setIsDialogOpen(false);
  };

  return (
    <div className="h-full border-r bg-gray-50 flex flex-col">
      <div className="p-4 border-b">
        <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
          <DialogTrigger asChild>
            <Button className="w-full" size="sm">
              <Plus size={16} className="mr-2" />
              New Session
            </Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create New Session</DialogTitle>
            </DialogHeader>
            <div className="space-y-4">
              <Input
                placeholder="Session name"
                value={newSessionName}
                onChange={(e) => setNewSessionName(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleCreateSession()}
              />
              <Button onClick={handleCreateSession} className="w-full" disabled={loading}>
                Create
              </Button>
            </div>
          </DialogContent>
        </Dialog>
      </div>

      <ScrollArea className="flex-1">
        <div className="p-2 space-y-1">
          {sessions.map((session) => (
            <button
              key={session.id}
              onClick={() => selectSession(session)}
              className={`w-full text-left px-3 py-2 rounded-lg transition-colors ${
                currentSession?.id === session.id
                  ? 'bg-blue-100 text-blue-700'
                  : 'hover:bg-gray-100'
              }`}
            >
              <div className="font-medium truncate">{session.name}</div>
              <div className="text-xs text-gray-500">
                {new Date(session.createdAt).toLocaleDateString()}
              </div>
            </button>
          ))}
        </div>
      </ScrollArea>
    </div>
  );
}
