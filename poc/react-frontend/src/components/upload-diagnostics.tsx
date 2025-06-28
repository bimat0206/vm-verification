'use client';

import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';

export function UploadDiagnostics() {
  const [diagnostics, setDiagnostics] = useState<string>('');
  const [isRunning, setIsRunning] = useState(false);

  const runDiagnostics = async () => {
    setIsRunning(true);
    setDiagnostics('Running diagnostics...\n');

    try {
      // Test 1: Check environment variables
      setDiagnostics(prev => prev + '\n1. Environment Variables:\n');
      setDiagnostics(prev => prev + `   API_BASE_URL: ${process.env.NEXT_PUBLIC_API_BASE_URL || 'NOT SET'}\n`);
      setDiagnostics(prev => prev + `   API_KEY: ${process.env.NEXT_PUBLIC_API_KEY ? 'SET' : 'NOT SET'}\n`);
      setDiagnostics(prev => prev + `   AWS_REGION: ${process.env.NEXT_PUBLIC_AWS_REGION || 'NOT SET'}\n`);

      // Test 2: Check API client configuration
      setDiagnostics(prev => prev + '\n2. API Client Configuration:\n');
      try {
        const { getConfig } = await import('@/config');
        const config = await getConfig();
        setDiagnostics(prev => prev + `   Config loaded: YES\n`);
        setDiagnostics(prev => prev + `   API Base URL: ${config.apiBaseUrl}\n`);
        setDiagnostics(prev => prev + `   API Key: ${config.apiKey ? 'SET' : 'NOT SET'}\n`);
      } catch (configError: any) {
        setDiagnostics(prev => prev + `   Config error: ${configError.message}\n`);
      }

      // Test 3: Test API connectivity
      setDiagnostics(prev => prev + '\n3. API Connectivity:\n');
      try {
        const { default: apiClient } = await import('@/lib/api-client');
        const healthResponse = await apiClient.healthCheck(true);
        setDiagnostics(prev => prev + `   Health check: SUCCESS\n`);
        setDiagnostics(prev => prev + `   Status: ${healthResponse.status}\n`);
      } catch (apiError: any) {
        setDiagnostics(prev => prev + `   Health check: FAILED\n`);
        setDiagnostics(prev => prev + `   Error: ${apiError.message}\n`);
      }

      // Test 4: Test file upload preparation
      setDiagnostics(prev => prev + '\n4. Upload Preparation:\n');
      try {
        // Create a small test file
        const testFile = new File(['test content'], 'test.txt', { type: 'text/plain' });
        setDiagnostics(prev => prev + `   Test file created: ${testFile.name} (${testFile.size} bytes)\n`);
        
        // Test FormData creation
        const formData = new FormData();
        formData.append('file', testFile);
        setDiagnostics(prev => prev + `   FormData created: SUCCESS\n`);
        
        // Test URL construction
        const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL || '';
        const uploadUrl = new URL(`${baseUrl}/api/images/upload`);
        uploadUrl.searchParams.append('bucketType', 'checking');
        uploadUrl.searchParams.append('fileName', 'test.txt');
        setDiagnostics(prev => prev + `   Upload URL: ${uploadUrl.toString()}\n`);
        
      } catch (prepError: any) {
        setDiagnostics(prev => prev + `   Upload preparation: FAILED\n`);
        setDiagnostics(prev => prev + `   Error: ${prepError.message}\n`);
      }

      // Test 5: Browser capabilities
      setDiagnostics(prev => prev + '\n5. Browser Capabilities:\n');
      setDiagnostics(prev => prev + `   Fetch API: ${typeof fetch !== 'undefined' ? 'AVAILABLE' : 'NOT AVAILABLE'}\n`);
      setDiagnostics(prev => prev + `   FormData: ${typeof FormData !== 'undefined' ? 'AVAILABLE' : 'NOT AVAILABLE'}\n`);
      setDiagnostics(prev => prev + `   File API: ${typeof File !== 'undefined' ? 'AVAILABLE' : 'NOT AVAILABLE'}\n`);
      setDiagnostics(prev => prev + `   User Agent: ${navigator.userAgent}\n`);

      setDiagnostics(prev => prev + '\n✅ Diagnostics completed!\n');

    } catch (error: any) {
      setDiagnostics(prev => prev + `\n❌ Diagnostics failed: ${error.message}\n`);
      console.error('Diagnostics error:', error);
    } finally {
      setIsRunning(false);
    }
  };

  return (
    <Card className="w-full max-w-4xl mx-auto">
      <CardHeader>
        <CardTitle>Upload Diagnostics</CardTitle>
      </CardHeader>
      <CardContent className="space-y-4">
        <Button 
          onClick={runDiagnostics} 
          disabled={isRunning}
          className="w-full"
        >
          {isRunning ? 'Running Diagnostics...' : 'Run Upload Diagnostics'}
        </Button>
        
        {diagnostics && (
          <div className="bg-black/30 p-4 rounded-md border border-border/50">
            <pre className="text-sm font-mono text-muted-foreground whitespace-pre-wrap">
              {diagnostics}
            </pre>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
