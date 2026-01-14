'use client';

import { ProtectedRoute } from '@/components/auth/ProtectedRoute';
import { DocumentUpload } from '@/components/ingest/DocumentUpload';
import { DocumentList } from '@/components/ingest/DocumentList';

export default function IngestPage() {
  return (
    <ProtectedRoute>
      <div className="container mx-auto px-4 py-8">
        <h1 className="text-3xl font-bold mb-8">Document Ingestion</h1>
        
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="space-y-6">
            <DocumentUpload />
          </div>
          
          <div>
            <DocumentList />
          </div>
        </div>
      </div>
    </ProtectedRoute>
  );
}
