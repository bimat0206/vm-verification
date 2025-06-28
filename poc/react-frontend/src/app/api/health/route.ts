import { NextResponse } from 'next/server';
import { SecretsManagerClient, GetSecretValueCommand } from '@aws-sdk/client-secrets-manager';

// Define the health status interface
interface HealthStatus {
  status: string;
  timestamp: string;
  environment: string;
  secrets: {
    configSecretAvailable: boolean;
    apiKeySecretAvailable: boolean;
  };
  environment_vars: {
    CONFIG_SECRET_SET: boolean;
    API_KEY_SECRET_NAME_SET: boolean;
    AWS_REGION_SET: boolean;
  };
  message: string;
}

// Simple health check endpoint that also verifies secrets
export async function GET() {
  // Initialize with default values including environment_vars
  const healthStatus: HealthStatus = {
    status: 'ok',
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || 'unknown',
    secrets: {
      configSecretAvailable: false,
      apiKeySecretAvailable: false,
    },
    environment_vars: {
      CONFIG_SECRET_SET: false,
      API_KEY_SECRET_NAME_SET: false,
      AWS_REGION_SET: false
    },
    message: 'System is operational'
  };

  // Check if we can access the secrets
  try {
    const configSecretName = process.env.CONFIG_SECRET;
    const apiKeySecretName = process.env.API_KEY_SECRET_NAME;
    const region = process.env.AWS_REGION || 'us-east-1';

    if (configSecretName) {
      try {
        const secretsClient = new SecretsManagerClient({ region });
        const command = new GetSecretValueCommand({ SecretId: configSecretName });
        const response = await secretsClient.send(command);
        healthStatus.secrets.configSecretAvailable = !!response.SecretString;
      } catch (error) {
        console.error('Health check - Error accessing CONFIG_SECRET:', error);
      }
    }

    if (apiKeySecretName) {
      try {
        const secretsClient = new SecretsManagerClient({ region });
        const command = new GetSecretValueCommand({ SecretId: apiKeySecretName });
        const response = await secretsClient.send(command);
        healthStatus.secrets.apiKeySecretAvailable = !!response.SecretString;
      } catch (error) {
        console.error('Health check - Error accessing API_KEY_SECRET_NAME:', error);
      }
    }

    // Update environment variables information
    healthStatus.environment_vars = {
      CONFIG_SECRET_SET: !!process.env.CONFIG_SECRET,
      API_KEY_SECRET_NAME_SET: !!process.env.API_KEY_SECRET_NAME,
      AWS_REGION_SET: !!process.env.AWS_REGION
    };

  } catch (error) {
    console.error('Health check - Error checking secrets:', error);
    healthStatus.status = 'warning';
    healthStatus.message = 'System operational but secret checking failed';
  }

  return NextResponse.json(healthStatus);
}