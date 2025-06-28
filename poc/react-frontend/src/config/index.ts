
'use client';

import { useEffect, useState } from 'react';
import ConfigLoader from '@/lib/config-loader';

export interface AppConfig {
  apiBaseUrl: string;
  apiKey: string;
  referenceS3BucketName: string;
  checkingS3BucketName: string;
}

// Initialize a singleton instance of the ConfigLoader
const configLoader = new ConfigLoader();

// Create initial config with environment variables
// This will be updated when the server config loads
const initialConfig: AppConfig = {
  apiBaseUrl: process.env.NEXT_PUBLIC_API_BASE_URL || process.env.NEXT_PUBLIC_API_ENDPOINT || 'https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1',
  apiKey: process.env.NEXT_PUBLIC_API_KEY || '',
  referenceS3BucketName: process.env.NEXT_PUBLIC_REFERENCE_BUCKET || process.env.NEXT_PUBLIC_S3_REFERENCE_BUCKET_NAME || 'kootoro-dev-s3-reference-f6d3xl',
  checkingS3BucketName: process.env.NEXT_PUBLIC_CHECKING_BUCKET || process.env.NEXT_PUBLIC_S3_CHECKING_BUCKET_NAME || 'kootoro-dev-s3-checking-f6d3xl',
};

// Async function to get config from the ConfigLoader
export async function getConfig(): Promise<AppConfig> {
  const config = await configLoader.getAll();
  
  return {
    apiBaseUrl: config.API_ENDPOINT || initialConfig.apiBaseUrl,
    apiKey: config.API_KEY || initialConfig.apiKey,
    referenceS3BucketName: config.REFERENCE_BUCKET || initialConfig.referenceS3BucketName,
    checkingS3BucketName: config.CHECKING_BUCKET || initialConfig.checkingS3BucketName,
  };
}

// React hook for components that need config
export function useAppConfig() {
  const [config, setConfig] = useState<AppConfig>(initialConfig);
  const [isLoading, setIsLoading] = useState(true);
  
  useEffect(() => {
    let isMounted = true;
    
    async function loadConfig() {
      try {
        const updatedConfig = await getConfig();
        if (isMounted) {
          setConfig(updatedConfig);
          setIsLoading(false);
        }
      } catch (error) {
        console.error('Failed to load config:', error);
        if (isMounted) {
          setIsLoading(false);
        }
      }
    }
    
    loadConfig();
    
    return () => {
      isMounted = false;
    };
  }, []);
  
  return { config, isLoading };
}

// For compatibility with existing code
const config = initialConfig;

// Log initial configuration for debugging
console.log('Initial API Configuration:', {
  apiBaseUrl: config.apiBaseUrl,
  apiKey: config.apiKey ? '***REDACTED***' : 'NOT_SET',
  referenceS3BucketName: config.referenceS3BucketName,
  checkingS3BucketName: config.checkingS3BucketName,
});

// Load the full configuration asynchronously and log when ready
getConfig().then(fullConfig => {
  console.log('Full API Configuration loaded:', {
    apiBaseUrl: fullConfig.apiBaseUrl,
    apiKey: fullConfig.apiKey ? '***REDACTED***' : 'NOT_SET',
    referenceS3BucketName: fullConfig.referenceS3BucketName,
    checkingS3BucketName: fullConfig.checkingS3BucketName,
    loadedFromServer: configLoader.isLoadedFromServer(),
  });
});

export default config;
