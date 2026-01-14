'use client';

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { CheckCircle, Clock, XCircle, Loader2 } from 'lucide-react';

interface IngestionStatusProps {
  status: 'pending' | 'processing' | 'completed' | 'failed';
  message?: string;
}

export function IngestionStatus({ status, message }: IngestionStatusProps) {
  const statusConfig = {
    pending: {
      icon: Clock,
      color: 'text-gray-500',
      label: 'Pending',
      animate: false,
    },
    processing: {
      icon: Loader2,
      color: 'text-blue-500',
      label: 'Processing',
      animate: true,
    },
    completed: {
      icon: CheckCircle,
      color: 'text-green-500',
      label: 'Completed',
      animate: false,
    },
    failed: {
      icon: XCircle,
      color: 'text-red-500',
      label: 'Failed',
      animate: false,
    },
  };

  const config = statusConfig[status];
  const Icon = config.icon;

  return (
    <Card>
      <CardHeader>
        <CardTitle>Ingestion Status</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex items-center gap-3">
          <Icon
            size={24}
            className={`${config.color} ${config.animate ? 'animate-spin' : ''}`}
          />
          <div>
            <p className="font-medium">{config.label}</p>
            {message && <p className="text-sm text-gray-600">{message}</p>}
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
