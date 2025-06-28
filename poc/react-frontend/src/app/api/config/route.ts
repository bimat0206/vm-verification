import { NextResponse } from 'next/server';
import { SecretsManagerClient, GetSecretValueCommand } from '@aws-sdk/client-secrets-manager';

// This function runs on the server side, not in the browser
export async function GET() {
  try {
    // For development, return configuration from environment variables directly
    if (process.env.NODE_ENV === 'development') {
      const clientConfig = {
        API_ENDPOINT: process.env.NEXT_PUBLIC_API_BASE_URL || process.env.NEXT_PUBLIC_API_ENDPOINT || '',
        REGION: process.env.NEXT_PUBLIC_AWS_REGION || process.env.NEXT_PUBLIC_REGION || 'us-east-1',
        CHECKING_BUCKET: process.env.NEXT_PUBLIC_CHECKING_BUCKET || process.env.NEXT_PUBLIC_S3_CHECKING_BUCKET_NAME || '',
        REFERENCE_BUCKET: process.env.NEXT_PUBLIC_REFERENCE_BUCKET || process.env.NEXT_PUBLIC_S3_REFERENCE_BUCKET_NAME || '',
        API_KEY: process.env.NEXT_PUBLIC_API_KEY || ''
      };

      console.log('Development mode: returning config from environment variables');
      return NextResponse.json(clientConfig);
    }

    // For production, try to get from AWS Secrets Manager
    const configSecretName = process.env.CONFIG_SECRET || process.env.NEXT_PUBLIC_CONFIG_SECRET;
    const apiKeySecretName = process.env.API_KEY_SECRET_NAME || process.env.NEXT_PUBLIC_API_KEY_SECRET_NAME;

    if (!configSecretName) {
      console.error('CONFIG_SECRET environment variable not found');
      return NextResponse.json(
        { error: 'Configuration not available' },
        { status: 500 }
      );
    }

    // Initialize AWS client - this uses AWS SDK on the server with ECS task role credentials
    const region = process.env.AWS_REGION || 'us-east-1';
    const secretsClient = new SecretsManagerClient({ region });

    // Get configuration secret
    const configCommand = new GetSecretValueCommand({ SecretId: configSecretName });
    const configResponse = await secretsClient.send(configCommand);

    if (!configResponse.SecretString) {
      throw new Error('Secret string is empty');
    }

    // Parse configuration secret
    const configData = JSON.parse(configResponse.SecretString);

    // If API_KEY_SECRET_NAME is provided in the environment, fetch it separately
    let apiKey = null;
    if (apiKeySecretName) {
      try {
        const apiKeyCommand = new GetSecretValueCommand({ SecretId: apiKeySecretName });
        const apiKeyResponse = await secretsClient.send(apiKeyCommand);
        
        if (apiKeyResponse.SecretString) {
          // API key might be plain string or JSON
          try {
            const parsed = JSON.parse(apiKeyResponse.SecretString);
            // Extract the actual API key string from the JSON object
            apiKey = parsed.api_key || parsed.API_KEY || parsed.apiKey || parsed.key;
            // If none of the expected keys exist, use the whole object as fallback
            if (!apiKey && typeof parsed === 'string') {
              apiKey = parsed;
            }
          } catch {
            // If not valid JSON, use the string directly
            apiKey = apiKeyResponse.SecretString;
          }
        }
      } catch (error) {
        console.error(`Error fetching API key secret: ${error}`);
        // Continue without API key - we still want to return the config
      }
    }

    // Prepare safe client-side configuration (don't expose sensitive data)
    const clientConfig = {
      API_ENDPOINT: configData.API_ENDPOINT || '',
      REGION: configData.REGION || '',
      CHECKING_BUCKET: configData.CHECKING_BUCKET || '',
      REFERENCE_BUCKET: configData.REFERENCE_BUCKET || '',
      // Include API key if available
      ...(apiKey ? { API_KEY: apiKey } : {})
    };

    return NextResponse.json(clientConfig);
  } catch (error) {
    console.error('Error fetching configuration:', error);
    return NextResponse.json(
      { error: 'Failed to load configuration' },
      { status: 500 }
    );
  }
}