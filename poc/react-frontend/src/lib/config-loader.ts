'use client';

export interface ConfigData {
  API_ENDPOINT: string;
  REGION: string;
  CHECKING_BUCKET: string;
  DYNAMODB_CONVERSATION_TABLE: string;
  DYNAMODB_VERIFICATION_TABLE: string;
  REFERENCE_BUCKET: string;
  AWS_DEFAULT_REGION: string;
  API_KEY: string;
  // Legacy support
  DYNAMODB_TABLE: string;
  S3_BUCKET: string;
}

export class ConfigLoader {
  private config: Partial<ConfigData> = {};
  private loadedFromSecret = false;
  private isLoading = false;
  private loadPromise: Promise<void> | null = null;

  constructor() {
    // Initialize with environment variables first
    this.loadFromEnvVars();
    // Then attempt to load from server API
    this.loadPromise = this.loadFromServerApi();
  }

  /**
   * Ensures configuration is loaded before accessing it
   */
  private async ensureLoaded(): Promise<void> {
    if (this.loadPromise) {
      await this.loadPromise;
    }
  }

  /**
   * Loads configuration from the server-side API endpoint
   */
  private async loadFromServerApi(): Promise<void> {
    if (this.isLoading) return;

    this.isLoading = true;
    try {
      console.log('Loading configuration from server API');

      // Only try to load from server API if we're in the browser
      if (typeof window === 'undefined') {
        console.log('Server-side rendering detected, skipping server API call');
        throw new Error('Cannot load from server API during SSR');
      }

      // Ensure we have a proper base URL for the fetch request
      const baseUrl = window.location.origin;
      const configUrl = `${baseUrl}/api/config`;

      console.log(`Fetching config from: ${configUrl}`);
      const response = await fetch(configUrl);

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(`Failed to load config: ${response.status} ${errorText}`);
      }

      const configData = await response.json();
      
      if (configData && typeof configData === 'object') {
        // Map the server config to our internal config
        this.config = {
          ...this.config, // Keep existing values
          API_ENDPOINT: configData.API_ENDPOINT || this.config.API_ENDPOINT || '',
          REGION: configData.REGION || this.config.REGION || '',
          CHECKING_BUCKET: configData.CHECKING_BUCKET || this.config.CHECKING_BUCKET || '',
          DYNAMODB_CONVERSATION_TABLE: configData.DYNAMODB_CONVERSATION_TABLE || this.config.DYNAMODB_CONVERSATION_TABLE || '',
          DYNAMODB_VERIFICATION_TABLE: configData.DYNAMODB_VERIFICATION_TABLE || this.config.DYNAMODB_VERIFICATION_TABLE || '',
          REFERENCE_BUCKET: configData.REFERENCE_BUCKET || this.config.REFERENCE_BUCKET || '',
          AWS_DEFAULT_REGION: configData.AWS_DEFAULT_REGION || this.config.AWS_DEFAULT_REGION || '',
          API_KEY: configData.API_KEY || this.config.API_KEY || '',
          // Legacy support
          DYNAMODB_TABLE: configData.DYNAMODB_TABLE || this.config.DYNAMODB_TABLE || '',
          S3_BUCKET: configData.S3_BUCKET || this.config.S3_BUCKET || ''
        };
        
        this.loadedFromSecret = true;
        console.log('Successfully loaded configuration from server API');
      }
    } catch (error) {
      console.error('Error loading from server API:', error);
      console.log('Using fallback configuration from environment variables');
    } finally {
      this.isLoading = false;
      this.loadPromise = null;
    }
  }

  /**
   * Loads configuration from environment variables as fallback
   */
  private loadFromEnvVars(): void {
    console.log('Loading configuration from environment variables');
    this.config = {
      API_ENDPOINT: process.env.NEXT_PUBLIC_API_ENDPOINT || '',
      REGION: process.env.NEXT_PUBLIC_REGION || '',
      CHECKING_BUCKET: process.env.NEXT_PUBLIC_CHECKING_BUCKET || '',
      DYNAMODB_CONVERSATION_TABLE: process.env.NEXT_PUBLIC_DYNAMODB_CONVERSATION_TABLE || '',
      DYNAMODB_VERIFICATION_TABLE: process.env.NEXT_PUBLIC_DYNAMODB_VERIFICATION_TABLE || '',
      REFERENCE_BUCKET: process.env.NEXT_PUBLIC_REFERENCE_BUCKET || '',
      AWS_DEFAULT_REGION: process.env.NEXT_PUBLIC_AWS_DEFAULT_REGION || '',
      API_KEY: process.env.NEXT_PUBLIC_API_KEY || '',
      // Legacy support
      DYNAMODB_TABLE: process.env.NEXT_PUBLIC_DYNAMODB_TABLE || '',
      S3_BUCKET: process.env.NEXT_PUBLIC_S3_BUCKET || ''
    };
  }

  /**
   * Gets a specific configuration value
   */
  async get<K extends keyof ConfigData>(key: K, defaultValue?: string): Promise<string> {
    await this.ensureLoaded();
    return this.config[key] as string || defaultValue || '';
  }

  /**
   * Gets a specific configuration value synchronously (may not have loaded from server yet)
   */
  getSync<K extends keyof ConfigData>(key: K, defaultValue?: string): string {
    return this.config[key] as string || defaultValue || '';
  }

  /**
   * Gets all configuration values
   */
  async getAll(): Promise<Partial<ConfigData>> {
    await this.ensureLoaded();
    return { ...this.config };
  }

  /**
   * Gets all configuration values synchronously (may not have loaded from server yet)
   */
  getAllSync(): Partial<ConfigData> {
    return { ...this.config };
  }

  /**
   * Checks if configuration was loaded from server secrets
   */
  async isLoadedFromServer(): Promise<boolean> {
    await this.ensureLoaded();
    return this.loadedFromSecret;
  }
}

export default ConfigLoader;
