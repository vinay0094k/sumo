'use client';

import { useState, useEffect } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Document } from '@/lib/types';
import { FileText, CheckCircle, Clock, XCircle, Loader2 } from 'lucide-react';

export function DocumentList() {
  const [documents, setDocuments] = useState<Document[]>([]);

  useEffect(() => {
    // TODO: Fetch documents from API
    // Placeholder data
    setDocuments([]);
  }, []);

  const getStatusIcon = (status: Document['status']) => {
    switch (status) {
      case 'completed':
        return <CheckCircle size={16} className="text-green-500" />;
      case 'processing':
        return <Loader2 size={16} className="text-blue-500 animate-spin" />;
      case 'failed':
        return <XCircle size={16} className="text-red-500" />;
      default:
        return <Clock size={16} className="text-gray-500" />;
    }
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Uploaded Documents</CardTitle>
      </CardHeader>
      <CardContent>
        <ScrollArea className="h-[400px]">
          {documents.length === 0 ? (
            <div className="text-center text-gray-500 py-8">
              No documents uploaded yet
            </div>
          ) : (
            <div className="space-y-2">
              {documents.map((doc) => (
                <div
                  key={doc.id}
                  className="flex items-center gap-3 p-3 border rounded-lg hover:bg-gray-50"
                >
                  <FileText size={20} className="text-gray-400" />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium truncate">{doc.name}</p>
                    <p className="text-xs text-gray-500">
                      {new Date(doc.uploadedAt).toLocaleString()}
                    </p>
                  </div>
                  {getStatusIcon(doc.status)}
                </div>
              ))}
            </div>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
