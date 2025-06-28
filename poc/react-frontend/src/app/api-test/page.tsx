'use client';

import { useState } from 'react';
import apiClient from '@/lib/api-client';
import { Button } from '@/components/ui/button';

export default function ApiTestPage() {
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const testHealthCheck = async () => {
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      console.log('Testing health check...');
      const response = await apiClient.healthCheck(true); // Enable debug mode
      console.log('Health check response:', response);
      setResult(response);
    } catch (err: any) {
      console.error('Health check error:', err);
      setError(err.message || 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  const testVerification = async () => {
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      console.log('Testing verification initiation...');
      const response = await apiClient.initiateVerification(
        'LAYOUT_VS_CHECKING',
        's3://test-bucket/ref.jpg',
        's3://test-bucket/check.jpg',
        false,
        true // Enable debug mode
      );
      console.log('Verification response:', response);
      setResult(response);
    } catch (err: any) {
      console.error('Verification error:', err);
      setError(err.message || 'Unknown error');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">API Test Page</h1>
      
      <div className="space-y-4 mb-6">
        <Button onClick={testHealthCheck} disabled={loading}>
          {loading ? 'Testing...' : 'Test Health Check'}
        </Button>
        
        <Button onClick={testVerification} disabled={loading}>
          {loading ? 'Testing...' : 'Test Verification'}
        </Button>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
          <strong>Error:</strong> {error}
        </div>
      )}

      {result && (
        <div className="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">
          <strong>Result:</strong>
          <pre className="mt-2 text-sm overflow-auto">
            {JSON.stringify(result, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
