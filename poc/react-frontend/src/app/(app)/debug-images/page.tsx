'use client';

import { useState } from 'react';
import NextImage from 'next/image';
import apiClient from '@/lib/api-client';

export default function DebugImagesPage() {
  const [imageKey, setImageKey] = useState('AATM 3.jpg');
  const [bucketType, setBucketType] = useState<'reference' | 'checking'>('checking');
  const [result, setResult] = useState<string>('');
  const [loading, setLoading] = useState(false);

  const testImageUrl = async () => {
    setLoading(true);
    setResult('');
    
    try {
      console.log(`Testing image URL for key: "${imageKey}" in bucket: ${bucketType}`);
      const url = await apiClient.getImageUrl(imageKey, bucketType, true); // Enable debug
      setResult(`Success! URL: ${url}`);
      console.log('Generated URL:', url);
      
      // Test if the URL is accessible
      try {
        const response = await fetch(url, { method: 'HEAD' });
        if (response.ok) {
          setResult(prev => prev + '\n\n✅ URL is accessible (HEAD request successful)');
        } else {
          setResult(prev => prev + `\n\n❌ URL returned status: ${response.status} ${response.statusText}`);
        }
      } catch (fetchError) {
        setResult(prev => prev + `\n\n❌ URL fetch failed: ${fetchError}`);
      }
      
    } catch (error: any) {
      setResult(`Error: ${error.message}`);
      console.error('Error getting image URL:', error);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mx-auto p-6">
      <h1 className="text-2xl font-bold mb-6">Debug Image URLs</h1>
      
      <div className="space-y-4 max-w-md">
        <div>
          <label className="block text-sm font-medium mb-2">Image Key:</label>
          <input
            type="text"
            value={imageKey}
            onChange={(e) => setImageKey(e.target.value)}
            className="w-full p-2 border border-gray-300 rounded"
            placeholder="e.g., AATM 3.jpg"
          />
        </div>
        
        <div>
          <label className="block text-sm font-medium mb-2">Bucket Type:</label>
          <select
            value={bucketType}
            onChange={(e) => setBucketType(e.target.value as 'reference' | 'checking')}
            className="w-full p-2 border border-gray-300 rounded"
          >
            <option value="checking">Checking</option>
            <option value="reference">Reference</option>
          </select>
        </div>
        
        <button
          onClick={testImageUrl}
          disabled={loading}
          className="w-full bg-blue-500 text-white p-2 rounded hover:bg-blue-600 disabled:opacity-50"
        >
          {loading ? 'Testing...' : 'Test Image URL'}
        </button>
      </div>
      
      {result && (
        <div className="mt-6">
          <h2 className="text-lg font-semibold mb-2">Result:</h2>
          <pre className="bg-gray-100 p-4 rounded whitespace-pre-wrap text-sm">
            {result}
          </pre>

          {result.includes('Success!') && (
            <div className="mt-4">
              <h3 className="text-md font-semibold mb-2">Image Display Test:</h3>
              <div className="space-y-4">
                <div>
                  <h4 className="text-sm font-medium mb-1">Using regular img tag:</h4>
                  <img
                    src={result.split('URL: ')[1]?.split('\n')[0]}
                    alt="Test with img tag"
                    className="max-w-xs border border-gray-300 rounded"
                    onError={(e) => {
                      console.error('img tag failed to load:', e);
                      (e.target as HTMLImageElement).style.border = '2px solid red';
                    }}
                    onLoad={() => console.log('img tag loaded successfully')}
                  />
                </div>
                <div>
                  <h4 className="text-sm font-medium mb-1">Using Next.js Image component:</h4>
                  <div className="relative w-64 h-48 border border-gray-300 rounded">
                    <NextImage
                      src={result.split('URL: ')[1]?.split('\n')[0] || ''}
                      alt="Test with Next.js Image"
                      fill
                      className="object-contain"
                      unoptimized={true}
                      onError={(e) => {
                        console.error('Next.js Image failed to load:', e);
                      }}
                      onLoad={() => console.log('Next.js Image loaded successfully')}
                    />
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}
      
      <div className="mt-8">
        <h2 className="text-lg font-semibold mb-2">Instructions:</h2>
        <ol className="list-decimal list-inside space-y-1 text-sm">
          <li>Enter an image key (filename) that exists in the S3 bucket</li>
          <li>Select the bucket type (checking or reference)</li>
          <li>Click "Test Image URL" to generate and test the presigned URL</li>
          <li>Check the browser console for detailed debug logs</li>
          <li>The result will show if the URL was generated successfully and if it's accessible</li>
        </ol>
      </div>
    </div>
  );
}
