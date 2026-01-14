'use client';

import Link from 'next/link';
import { useAuthContext } from '@/context/AuthContext';
import { Button } from '@/components/ui/button';

export function Header() {
  const { user, logout } = useAuthContext();

  return (
    <header className="border-b bg-white">
      <div className="container mx-auto px-4 py-3 flex items-center justify-between">
        <Link href="/" className="text-xl font-bold">
          Sumo AI
        </Link>
        
        {user && (
          <div className="flex items-center gap-4">
            <span className="text-sm text-gray-600">{user.email}</span>
            <Button onClick={logout} variant="outline" size="sm">
              Logout
            </Button>
          </div>
        )}
      </div>
    </header>
  );
}
