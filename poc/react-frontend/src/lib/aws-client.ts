'use client';

import { S3Client } from '@aws-sdk/client-s3';
import { SecretsManagerClient, GetSecretValueCommand } from '@aws-sdk/client-secrets-manager';
import { DynamoDBClient } from '@aws-sdk/client-dynamodb';
import { DynamoDBDocumentClient } from '@aws-sdk/lib-dynamodb';

export class AWSClient {
  private s3Client: S3Client | null = null;
  private secretsManagerClient: SecretsManagerClient | null = null;
  private dynamoDBClient: DynamoDBClient | null = null;
  private dynamoDBDocClient: DynamoDBDocumentClient | null = null;

  constructor() {
    try {
      // Initialize AWS clients with default configuration
      // In a browser environment, credentials should be provided via AWS Cognito or similar
      const region = process.env.NEXT_PUBLIC_AWS_REGION || 'us-east-1';
      
      // Note: In a browser environment, you typically need to use AWS Cognito
      // or temporary credentials. Direct AWS SDK usage in browsers requires
      // proper authentication setup.
      
      this.s3Client = new S3Client({ 
        region,
        // Add credentials configuration as needed for your setup
      });
      
      this.secretsManagerClient = new SecretsManagerClient({ 
        region,
        // Add credentials configuration as needed for your setup
      });
      
      this.dynamoDBClient = new DynamoDBClient({ 
        region,
        // Add credentials configuration as needed for your setup
      });
      
      this.dynamoDBDocClient = DynamoDBDocumentClient.from(this.dynamoDBClient);
      
    } catch (error) {
      console.error('Failed to initialize AWS clients:', error);
      throw error;
    }
  }

  async getSecret(secretName: string): Promise<any> {
    if (!this.secretsManagerClient) {
      throw new Error('SecretsManager client not initialized');
    }

    try {
      const command = new GetSecretValueCommand({ SecretId: secretName });
      const response = await this.secretsManagerClient.send(command);
      return response;
    } catch (error) {
      console.error(`Failed to get secret ${secretName}:`, error);
      throw error;
    }
  }

  getS3Client(): S3Client | null {
    return this.s3Client;
  }

  getSecretsManagerClient(): SecretsManagerClient | null {
    return this.secretsManagerClient;
  }

  getDynamoDBClient(): DynamoDBClient | null {
    return this.dynamoDBClient;
  }

  getDynamoDBDocumentClient(): DynamoDBDocumentClient | null {
    return this.dynamoDBDocClient;
  }
}

export default AWSClient;
