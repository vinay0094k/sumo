'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { MessageSquare, Upload } from 'lucide-react';

export function Sidebar() {
  const pathname = usePathname();

  const links = [
    { href: '/chat', label: 'Chat', icon: MessageSquare },
    { href: '/ingest', label: 'Ingest Documents', icon: Upload },
  ];

  return (
    <aside className="w-64 border-r bg-gray-50 h-full">
      <nav className="p-4 space-y-2">
        {links.map(({ href, label, icon: Icon }) => (
          <Link
            key={href}
            href={href}
            className={`flex items-center gap-3 px-4 py-2 rounded-lg transition-colors ${
              pathname === href
                ? 'bg-blue-100 text-blue-700'
                : 'hover:bg-gray-100'
            }`}
          >
            <Icon size={20} />
            <span>{label}</span>
          </Link>
        ))}
      </nav>
    </aside>
  );
}
