
'use client';

import type { AppConfig } from '@/config';
import type {
  ApiError,
  Verification,
  CreateVerificationRequest,
  BucketType,
  BrowserResponse,
  UploadResponse,
  ViewResponse,
  HealthResponse,
  VerificationListResponse,
  VerificationConversationResponse,
  VerificationStatusResponse
} from './types';
import appConfigInstance, { getConfig } from '@/config';

class ApiClient {
  private baseUrl: string;
  private apiKey: string;
  private configLoadPromise: Promise<AppConfig> | null = null;
  private configLoaded = false;

  constructor(initialConfig: AppConfig) {
    // Start with initial config values
    this.baseUrl = initialConfig.apiBaseUrl.endsWith('/') 
      ? initialConfig.apiBaseUrl.slice(0, -1) 
      : initialConfig.apiBaseUrl;
    this.apiKey = initialConfig.apiKey;
    
    // Log initial config status
    if (!this.baseUrl) {
      console.error("API_ENDPOINT not found in initial configuration. Attempting to load from server.");
    }
    if (!this.apiKey || this.apiKey === 'your-default-api-key') {
      console.warn("API key not found or using default in initial config. Attempting to load from server.");
    }
    
    // Load the full config asynchronously
    this.loadFullConfig();
  }
  
  /**
   * Loads the full configuration from server
   */
  private async loadFullConfig(): Promise<void> {
    try {
      // Load full config from the server
      const fullConfig = await getConfig();
      
      // Update client with full config values
      this.baseUrl = fullConfig.apiBaseUrl.endsWith('/') 
        ? fullConfig.apiBaseUrl.slice(0, -1) 
        : fullConfig.apiBaseUrl;
      this.apiKey = fullConfig.apiKey;
      
      this.configLoaded = true;
      
      // Log updated config status
      if (!this.baseUrl) {
        console.error("API_ENDPOINT still not found after loading full configuration.");
      }
      if (!this.apiKey) {
        console.warn("API key still not found after loading full configuration.");
      }
    } catch (error) {
      console.error("Failed to load full configuration:", error);
    }
  }
  
  /**
   * Ensures config is loaded before making requests
   */
  private async ensureConfigLoaded(): Promise<void> {
    if (!this.configLoaded) {
      if (!this.configLoadPromise) {
        this.configLoadPromise = getConfig();
      }
      
      try {
        const fullConfig = await this.configLoadPromise;
        
        // Update with loaded values
        this.baseUrl = fullConfig.apiBaseUrl.endsWith('/') 
          ? fullConfig.apiBaseUrl.slice(0, -1) 
          : fullConfig.apiBaseUrl;
        this.apiKey = fullConfig.apiKey;
        
        this.configLoaded = true;
      } catch (error) {
        console.error("Error loading config in ensureConfigLoaded:", error);
        // Continue with initial values
      } finally {
        this.configLoadPromise = null;
      }
    }
  }

  private async makeRequest<T>(
    method: string,
    endpoint: string, // This is path relative to /api, e.g., /images/browser or /health
    params?: Record<string, any>,
    data?: any, // Can be JSON object, File, or FormData
    contentTypeOverride?: string, // Explicit content type override
    debug: boolean = false
  ): Promise<T> {
    // Ensure config is loaded before making the request
    await this.ensureConfigLoaded();
    
    const headers: HeadersInit = {
      'X-Api-Key': this.apiKey,
      'Accept': 'application/json'
    };

    const apiSpecificPath = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;
    const fullUrlString = `${this.baseUrl}/api${apiSpecificPath}`;

    let requestUrl: URL;
    try {
      requestUrl = new URL(fullUrlString);
      if (params) {
        Object.keys(params).forEach(key => {
          if (params[key] !== undefined && params[key] !== null) {
            requestUrl.searchParams.append(key, String(params[key]));
          }
        });
      }
    } catch (e: any) {
      console.error(`[API Client DEBUG] Invalid URL constructed: ${fullUrlString}`, e);
      throw { message: `Invalid API URL: ${fullUrlString}. ${e.message}`, statusCode: 0, details: { originalError: e.toString() } } as ApiError;
    }

    const options: RequestInit = { method, headers };

    if (data) {
      if (data instanceof FormData) {
        options.body = data;
        // DO NOT set Content-Type for FormData; browser does it with boundary.
      } else if (data instanceof File) {
        options.body = data;
        headers['Content-Type'] = contentTypeOverride || data.type || 'application/octet-stream';
      } else {
        options.body = JSON.stringify(data);
        headers['Content-Type'] = contentTypeOverride || 'application/json';
      }
    } else if (contentTypeOverride) {
      headers['Content-Type'] = contentTypeOverride;
    }


    if (debug) {
      let bodySummary = data;
      if (data instanceof File) bodySummary = { name: data.name, size: data.size, type: data.type };
      else if (data instanceof FormData) {
        bodySummary = {};
        // @ts-ignore
        for (let [key, value] of data.entries()) {
            // @ts-ignore
            bodySummary[key] = value instanceof File ? { name: value.name, size: value.size, type: value.type } : value;
        }
      }
      console.log(`[API Client DEBUG] Request (makeRequest): ${method} ${requestUrl.toString()}`, { headers: {...headers}, bodySummary });
    }

    try {
      const response = await fetch(requestUrl.toString(), options);
      const responseText = await response.clone().text();

      if (debug) {
        console.log(`[API Client DEBUG] Response Status: ${response.status} ${response.statusText} for ${requestUrl.toString()}`);
        console.log(`[API Client DEBUG] RAW Response Body Text:`, responseText);
      }

      if (!response.ok) {
        let errorBody;
        try {
          errorBody = JSON.parse(responseText);
        } catch (e) {
          errorBody = { message: responseText || response.statusText, rawResponse: responseText };
        }

        if (debug) {
          console.error(`[API Client DEBUG] API Error Body:`, {
            status: response.status,
            statusText: response.statusText,
            url: requestUrl.toString(),
            errorBody: errorBody,
            rawResponse: responseText
          });
        }

        // Handle empty error bodies more gracefully
        let errorMessage = errorBody.message || errorBody.error;
        if (!errorMessage || errorMessage.trim() === '') {
          if (responseText && responseText.trim() !== '') {
            errorMessage = responseText;
          } else {
            // Provide more specific error messages based on status code
            switch (response.status) {
              case 404:
                errorMessage = `Resource not found`;
                break;
              case 403:
                errorMessage = `Access forbidden - check API key`;
                break;
              case 401:
                errorMessage = `Unauthorized - invalid API key`;
                break;
              case 500:
                errorMessage = `Internal server error`;
                break;
              case 502:
                errorMessage = `Bad gateway - API server may be down`;
                break;
              case 503:
                errorMessage = `Service unavailable - API server temporarily down`;
                break;
              case 504:
                errorMessage = `Gateway timeout - API server took too long to respond`;
                break;
              default:
                errorMessage = `API Error ${response.status}: ${response.statusText}`;
            }
          }
        }

        throw {
          message: errorMessage,
          statusCode: response.status,
          details: {
            ...errorBody,
            url: requestUrl.toString(),
            rawResponse: responseText
          }
        } as ApiError;
      }

      if (response.status === 204) {
        if (debug) console.log(`[API Client DEBUG] Response Status 204, returning empty object.`);
        return {} as T;
      }

      const responseContentType = response.headers.get("content-type");
      if (responseContentType?.includes("application/json")) {
        try {
            const jsonData = JSON.parse(responseText);
            if (debug) console.log(`[API Client DEBUG] Parsed JSON Response:`, jsonData);
            return jsonData as T;
        } catch (parseError: any) {
            if (debug) console.error(`[API Client DEBUG] JSON Parse Error: `, parseError, `for response text: ${responseText}`);
            throw { message: `Failed to parse JSON response: ${parseError.message}`, statusCode: response.status, details: { responseBody: responseText } } as ApiError;
        }
      }
      if (debug) console.log(`[API Client DEBUG] Response not JSON, returning raw text as 'unexpectedResponse'. Content-Type: ${responseContentType}`);
      return { unexpectedResponse: responseText } as unknown as T;
    } catch (error: any) {
      const errorMsg = error.message || 'Unknown fetch error';
      const isFetchTypeError = error.name === 'TypeError' || errorMsg.toLowerCase().includes('failed to fetch');

      if (debug) {
        console.error(`[API Client DEBUG] Fetch/Processing Error for ${requestUrl.toString()}:`, {
          error: error,
          message: errorMsg,
          name: error.name,
          stack: error.stack,
          url: requestUrl.toString()
        });
      }

      if (isFetchTypeError) {
        const detailedMsg = `Network request to API failed: "${errorMsg}".
        This can be caused by:
        1. BROWSER EXTENSIONS (like ad blockers/privacy tools) interfering. Try disabling them.
        2. NETWORK ISSUES (CORS, DNS, firewall). Check your connection and if the API is reachable.
        3. API SERVER being down or misconfigured (e.g. incorrect URL path, method, or body format expected by API Gateway/Lambda).
        URL: ${requestUrl.toString()}`;
        if (debug) console.warn(`[API Client DEBUG] A 'TypeError' or "Failed to fetch" occurred. This is OFTEN due to browser extensions, CORS, network issues, or fundamental API request misconfiguration. Check API server status & URL. URL: ${requestUrl.toString()}`);
         throw {
            message: detailedMsg,
            statusCode: 0,
            details: { originalError: error.toString(), name: error.name, url: requestUrl.toString() }
        } as ApiError;
      }

      if (error.statusCode !== undefined) throw error;

      throw {
        message: `Unexpected error during API request: ${errorMsg}`,
        statusCode: 0,
        details: { originalError: error.toString(), name: error.name, url: requestUrl.toString() }
      } as ApiError;
    }
  }

  async healthCheck(debug: boolean = false): Promise<HealthResponse> {
    return this.makeRequest<HealthResponse>('GET', '/health', undefined, undefined, undefined, debug);
  }

  async initiateVerification(data: CreateVerificationRequest, debug: boolean = false): Promise<Verification> {
    const response = await fetch(`${this.baseUrl}/api/verifications`, {
      method: 'POST',
      headers: {
        'X-Api-Key': this.apiKey,
        'Content-Type': 'application/json',
        'Accept': 'application/json'
      },
      body: JSON.stringify(data)
    });

    if (debug) {
      console.log(`[API Client DEBUG] Verification creation response: ${response.status} ${response.statusText}`);
    }

    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(`Failed to create verification: ${response.status} ${errorText}`);
    }

    // Handle 202 Accepted with empty body - API accepts request but processes asynchronously
    if (response.status === 202) {
      // Generate a temporary verification ID based on timestamp (format: verif-YYYYMMDDHHMMSS-temp)
      const now = new Date();
      const timestamp = now.toISOString().replace(/[-:T]/g, '').replace(/\..+/, '').substring(0, 14);
      const tempId = `verif-${timestamp}-temp`;

      if (debug) {
        console.log(`[API Client DEBUG] 202 Accepted - creating temporary verification object with ID: ${tempId}`);
      }

      // Return a temporary verification object that will be updated by polling
      return {
        verificationId: tempId,
        verificationStatus: 'PENDING',
        verificationType: data.verificationContext.verificationType,
        vendingMachineId: data.verificationContext.vendingMachineId || '',
        referenceImageUrl: data.verificationContext.referenceImageUrl,
        checkingImageUrl: data.verificationContext.checkingImageUrl,
        verificationAt: new Date().toISOString(),
        rawData: { tempId, originalRequest: data },
      } as Verification;
    }

    // Handle normal response with verification data
    const responseText = await response.text();
    if (responseText.trim() === '') {
      throw new Error('Empty response from verification API');
    }

    const apiResponse = JSON.parse(responseText);
    return {
      verificationId: apiResponse.verificationId,
      verificationStatus: apiResponse.verificationStatus,
      vendingMachineId: apiResponse.vendingMachineId,
      verificationType: apiResponse.verificationType || apiResponse.verificationContext?.verificationType,
      verificationAt: apiResponse.verificationAt,
      referenceImageUrl: apiResponse.referenceImageUrl,
      checkingImageUrl: apiResponse.checkingImageUrl,
      createdAt: apiResponse.createdAt,
      updatedAt: apiResponse.updatedAt,
      overallAccuracy: apiResponse.overallAccuracy,
      turn1ProcessedPath: apiResponse.turn1ProcessedPath,
      turn2ProcessedPath: apiResponse.turn2ProcessedPath,
      llmAnalysis: apiResponse.llmAnalysis,
      previousVerificationId: apiResponse.previousVerificationId || apiResponse.previous_verification_id,
      rawData: apiResponse,
    } as Verification;
  }

  async listVerifications(params?: {
    verificationStatus?: Verification['verificationStatus'];
    vendingMachineId?: string;
    verificationId?: string;
    verificationType?: string;
    limit?: number;
    offset?: number;
    sortBy?: string;
    dateRangeStart?: string;
    dateRangeEnd?: string;
  }, debug: boolean = false): Promise<VerificationListResponse> {
    const response = await this.makeRequest<VerificationListResponse>('GET', '/verifications', params, undefined, undefined, debug);

    if (response && response.results) {
      response.results = response.results.map(apiItem => ({
        ...apiItem,
        verificationType: (apiItem as any).verificationType || (apiItem as any).verificationContext?.verificationType,
        overallAccuracy: (apiItem as any).overallAccuracy || (apiItem as any).overall_accuracy,
        previousVerificationId: (apiItem as any).previousVerificationId || (apiItem as any).previous_verification_id,
        rawData: apiItem,
      }));
    }
    return response;
  }

  async getVerificationDetails(verificationId: string, debug: boolean = false): Promise<Verification> {
    // Use the list API to get verification details from the verification-results table
    const listResponse = await this.listVerifications({
      verificationId: verificationId,
      limit: 1
    }, debug);

    if (!listResponse.results || listResponse.results.length === 0) {
      throw new Error(`Verification ${verificationId} not found`);
    }

    const mainDetails = listResponse.results[0];

    if (debug) {
      console.log(`[API Client DEBUG] Verification details found:`, {
        id: mainDetails.verificationId,
        status: mainDetails.verificationStatus,
        type: mainDetails.verificationType,
        hasRawData: !!mainDetails.rawData,
        hasLlmAnalysis: !!mainDetails.llmAnalysis
      });
    }

    // Extract LLM analysis from the verification data
    let llmAnalysis = '';

    // First try the top-level llmAnalysis field
    if (mainDetails.llmAnalysis) {
      llmAnalysis = mainDetails.llmAnalysis;
    } else if (mainDetails.rawData) {
      // Try to extract from rawData using various field names
      const rawData = mainDetails.rawData;
      const analysisFields = [
        'llmAnalysis',
        'llm_analysis',
        'analysis',
        'verificationAnalysis',
        'verification_analysis',
        'aiAnalysis',
        'ai_analysis',
        'turn2Content',
        'turn2_content',
        'content'
      ];

      for (const field of analysisFields) {
        if (rawData[field]) {
          if (typeof rawData[field] === 'string' && rawData[field].trim().length > 0) {
            llmAnalysis = rawData[field] as string;
            break;
          } else if (typeof rawData[field] === 'object' && rawData[field].content) {
            llmAnalysis = rawData[field].content as string;
            break;
          }
        }
      }
    }

    // Calculate overall accuracy from verification summary
    let overallAccuracy = mainDetails.overallAccuracy;
    const rawData = (mainDetails.rawData || mainDetails) as any;
    if (!overallAccuracy && rawData.verificationSummary) {
      const summary = rawData.verificationSummary;
      if (summary.correct_positions && summary.total_positions_checked) {
        overallAccuracy = summary.correct_positions / summary.total_positions_checked;
      }
    }

    return {
      verificationId: mainDetails.verificationId,
      verificationAt: mainDetails.verificationAt,
      verificationStatus: mainDetails.verificationStatus,
      verificationType: mainDetails.verificationType,
      vendingMachineId: mainDetails.vendingMachineId,
      referenceImageUrl: mainDetails.referenceImageUrl,
      checkingImageUrl: mainDetails.checkingImageUrl,
      layoutId: rawData.layoutId,
      layoutPrefix: rawData.layoutPrefix,
      overallAccuracy: overallAccuracy,
      correctPositions: rawData.verificationSummary?.correct_positions,
      discrepantPositions: rawData.verificationSummary?.discrepant_positions,
      llmAnalysis: llmAnalysis,
      previousVerificationId: mainDetails.previousVerificationId || rawData.previous_verification_id,
      verificationSummary: rawData.verificationSummary,
      rawData: rawData,
    } as Verification;
  }

  async getVerificationStatus(verificationId: string, debug: boolean = false): Promise<VerificationStatusResponse> {
    return this.makeRequest<VerificationStatusResponse>('GET', '/verifications/status', { verificationId }, undefined, undefined, debug);
  }



  async pollVerificationResults(
    verificationId: string,
    onUpdate?: (verification: Verification, nextInterval?: number) => void,
    maxAttempts: number = 80, // Increased to handle longer processing times (up to ~15 minutes)
    debug: boolean = false
  ): Promise<Verification> {
    let attempts = 0;

    // Dynamic polling intervals: Start fast, then slow down for longer processing
    const getPollingInterval = (attemptNumber: number): number => {
      if (attemptNumber < 5) return 10000;  // First 5 attempts: 10 seconds
      if (attemptNumber < 15) return 15000; // Next 10 attempts: 15 seconds
      if (attemptNumber < 30) return 20000; // Next 15 attempts: 20 seconds
      return 30000; // Remaining attempts: 30 seconds
    };

    while (attempts < maxAttempts) {
      try {
        let verification: Verification;

        // If we have a temporary ID, try to find the actual verification
        if (verificationId.includes('-temp')) {
          if (debug) {
            console.log(`[API Client DEBUG] Attempting to find actual verification for temp ID: ${verificationId}`);
          }

          // Try to get recent verifications and find a match
          try {
            const recentVerifications = await this.listVerifications({ limit: 10 }, debug);

            if (debug) {
              console.log(`[API Client DEBUG] Found ${recentVerifications.results?.length || 0} recent verifications`);
            }

            // Look for verifications created within the last 10 minutes
            const cutoffTime = Date.now() - 600000; // 10 minutes
            let foundVerification: Verification | null = null;

            if (recentVerifications.results && recentVerifications.results.length > 0) {
              // Try to find a verification that matches our timeframe
              for (const candidate of recentVerifications.results) {
                const verificationTime = new Date(candidate.verificationAt).getTime();
                if (verificationTime > cutoffTime) {
                  if (debug) {
                    console.log(`[API Client DEBUG] Checking candidate verification: ${candidate.verificationId} (${new Date(candidate.verificationAt).toISOString()})`);
                  }

                  try {
                    foundVerification = await this.getVerificationDetails(candidate.verificationId, debug);
                    if (debug) {
                      console.log(`[API Client DEBUG] Successfully found verification: ${candidate.verificationId}`);
                    }
                    break;
                  } catch (detailError) {
                    if (debug) {
                      console.log(`[API Client DEBUG] Could not get details for ${candidate.verificationId}, trying next...`);
                    }
                    continue;
                  }
                }
              }
            }

            if (foundVerification) {
              verification = foundVerification;
            } else {
              // If we can't find a verification yet, continue polling (don't throw error)
              if (debug) {
                console.log(`[API Client DEBUG] No matching verification found yet, will continue polling (attempt ${attempts + 1}/${maxAttempts})`);
              }
              throw new Error('Verification still being processed');
            }
          } catch (searchError) {
            if (debug) {
              console.log(`[API Client DEBUG] Error searching for verification: ${(searchError as Error).message}`);
            }
            // Don't throw immediately, let the polling continue
            throw new Error('Verification still being processed');
          }
        } else {
          verification = await this.getVerificationDetails(verificationId, debug);
        }

        const currentInterval = getPollingInterval(attempts);
        const nextInterval = attempts < maxAttempts - 1 ? getPollingInterval(attempts + 1) : 0;

        if (debug) {
          console.log(`[API Client DEBUG] Poll attempt ${attempts + 1}/${maxAttempts}:`, {
            verificationId: verification.verificationId,
            status: verification.verificationStatus,
            hasLlmAnalysis: !!verification.llmAnalysis,
            overallAccuracy: verification.overallAccuracy,
            currentInterval: `${currentInterval / 1000}s`,
            nextInterval: nextInterval > 0 ? `${nextInterval / 1000}s` : 'none'
          });
        }

        // Call the update callback if provided with progress information
        if (onUpdate) {
          // Add progress percentage and estimated time remaining
          const progressPercent = Math.min(95, (attempts / maxAttempts) * 100);
          const estimatedTimeRemaining = nextInterval > 0 ? Math.round((maxAttempts - attempts) * (currentInterval / 1000)) : 0;

          onUpdate(verification, nextInterval, {
            attempt: attempts + 1,
            maxAttempts,
            progressPercent,
            estimatedTimeRemaining
          });
        }

        // Check if verification is complete based on status and presence of LLM analysis
        const isComplete = verification.verificationStatus === 'COMPLETED' ||
                          verification.verificationStatus === 'CORRECT' ||
                          verification.verificationStatus === 'INCORRECT' ||
                          verification.verificationStatus === 'ERROR' ||
                          verification.verificationStatus === 'FAILED' ||
                          (verification.llmAnalysis && verification.llmAnalysis.trim().length > 0);

        if (isComplete) {
          return verification;
        }

        // Wait before next poll using dynamic interval
        if (attempts < maxAttempts - 1) {
          await new Promise(resolve => setTimeout(resolve, currentInterval));
        }

      } catch (error) {
        if (debug) {
          console.error(`[API Client DEBUG] Poll attempt ${attempts + 1} failed:`, error);
        }

        const errorMessage = (error as Error).message || '';
        const statusCode = (error as ApiError).statusCode;

        // Determine if we should continue polling
        const shouldContinuePolling =
          // For temp IDs, always continue polling unless it's a fatal error
          (verificationId.includes('-temp') && (
            statusCode === 404 ||
            errorMessage.includes('not found') ||
            errorMessage.includes('still being processed') ||
            errorMessage.includes('No matching verification')
          )) ||
          // For regular IDs, only continue on 404
          (!verificationId.includes('-temp') && statusCode === 404);

        // Fatal errors that should stop polling immediately
        const isFatalError = statusCode === 403 || statusCode === 401 ||
                           (statusCode !== undefined && statusCode >= 500 && statusCode !== 503 && statusCode !== 504);

        if (isFatalError) {
          if (debug) {
            console.error(`[API Client DEBUG] Fatal error encountered, stopping polling: ${errorMessage}`);
          }
          throw error;
        }

        if (!shouldContinuePolling) {
          // For non-temp IDs with non-404 errors, throw immediately
          if (debug) {
            console.error(`[API Client DEBUG] Non-recoverable error for verification ${verificationId}: ${errorMessage}`);
          }
          throw error;
        }

        if (debug) {
          console.log(`[API Client DEBUG] Continuing to poll (attempt ${attempts + 1}/${maxAttempts}): ${errorMessage}`);
        }

        // Wait the interval before next attempt
        if (attempts < maxAttempts - 1) {
          const currentInterval = getPollingInterval(attempts);
          await new Promise(resolve => setTimeout(resolve, currentInterval));
        }
      }

      attempts++;
    }

    // Calculate total time spent polling
    const totalTimeMinutes = Math.round((maxAttempts * 18000) / 60000); // Approximate average interval

    // Provide more specific timeout error messages
    if (verificationId.includes('-temp')) {
      throw new Error(`Verification polling timed out after ${maxAttempts} attempts (~${totalTimeMinutes} minutes).

The verification may still be processing in the background. This can happen when:
• The verification involves complex image analysis
• The backend is processing a high volume of requests
• There are temporary performance issues

Please try:
1. Check the verification results page in a few minutes
2. Refresh the page to see if results are available
3. Create a new verification if the issue persists

If this problem continues, please contact support.`);
    } else {
      throw new Error(`Verification polling timed out after ${maxAttempts} attempts (~${totalTimeMinutes} minutes) for verification ${verificationId}.

The verification may still be processing. Please check the verification results page or try refreshing the page.`);
    }
  }

  async getVerificationConversation(verificationId: string, debug: boolean = false): Promise<VerificationConversationResponse> {
    return this.makeRequest<VerificationConversationResponse>('GET', `/verifications/${verificationId}/conversation`, undefined, undefined, undefined, debug);
  }

  async getPreviousVerificationAnalysis(previousVerificationId: string, debug: boolean = false): Promise<string | null> {
    try {
      if (debug) {
        console.log(`[API Client DEBUG] Fetching previous verification analysis for: ${previousVerificationId}`);
      }

      // First, try to get analysis content from the conversation endpoint (which fetches from S3)
      try {
        if (debug) {
          console.log(`[API Client DEBUG] Attempting to fetch conversation data for ${previousVerificationId}`);
        }

        const conversationData = await this.getVerificationConversation(previousVerificationId, debug);

        if (conversationData?.turn2Content?.content) {
          if (debug) {
            console.log(`[API Client DEBUG] Found turn2 content from conversation endpoint (${conversationData.turn2Content.content.length} characters)`);
          }
          return conversationData.turn2Content.content;
        }

        if (conversationData?.turn1Content?.content) {
          if (debug) {
            console.log(`[API Client DEBUG] Found turn1 content from conversation endpoint (${conversationData.turn1Content.content.length} characters)`);
          }
          return conversationData.turn1Content.content;
        }

        if (debug) {
          console.log(`[API Client DEBUG] No content found in conversation endpoint, falling back to verification record search`);
        }
      } catch (conversationError: any) {
        if (debug) {
          console.log(`[API Client DEBUG] Conversation endpoint failed, falling back to verification record search:`, conversationError.message);
        }
        // Continue to fallback method below
      }

      // Fallback: Get the previous verification from the verification-results table via list API
      let listResponse;
      try {
        listResponse = await this.listVerifications({
          verificationId: previousVerificationId,
          limit: 1
        }, debug);
      } catch (error: any) {
        if (debug) {
          console.log(`[API Client DEBUG] Error fetching verification list for ${previousVerificationId}:`, error);
        }

        // Handle specific error cases
        if (error.statusCode === 404 || error.message?.includes('not found')) {
          if (debug) {
            console.log(`[API Client DEBUG] Previous verification ${previousVerificationId} not found (404)`);
          }
          return null;
        }

        // Re-throw other errors
        throw error;
      }

      if (!listResponse.results || listResponse.results.length === 0) {
        if (debug) {
          console.log(`[API Client DEBUG] Previous verification ${previousVerificationId} not found in verification list`);
        }
        return null;
      }

      const previousVerification = listResponse.results[0];

      if (debug) {
        console.log(`[API Client DEBUG] Previous verification found:`, {
          id: previousVerification.verificationId,
          type: previousVerification.verificationType,
          hasRawData: !!previousVerification.rawData,
          hasLlmAnalysis: !!previousVerification.llmAnalysis,
          hasTurn2ProcessedPath: !!(previousVerification.turn2ProcessedPath || previousVerification.rawData?.turn2ProcessedPath),
          llmAnalysisLength: previousVerification.llmAnalysis ? previousVerification.llmAnalysis.length : 0
        });
      }

      // Try to extract analysis content from the verification data
      let analysisContent: string | null = null;

      // Check for turn2ProcessedPath or turn1ProcessedPath in the verification record
      const turn2Path = previousVerification.turn2ProcessedPath || previousVerification.rawData?.turn2ProcessedPath;
      const turn1Path = previousVerification.turn1ProcessedPath || previousVerification.rawData?.turn1ProcessedPath;

      if (turn2Path || turn1Path) {
        if (debug) {
          console.log(`[API Client DEBUG] Found processed paths in verification record:`, {
            turn2Path,
            turn1Path
          });
          console.log(`[API Client DEBUG] Note: S3 content should have been fetched by conversation endpoint. This suggests the conversation endpoint may not be working properly.`);
        }
      }

      // Check for analysis content in various possible fields
      if (previousVerification.rawData) {
        const rawData = previousVerification.rawData;

        // Try different possible field names for analysis content
        const analysisFields = [
          'llmAnalysis',
          'llm_analysis',
          'analysis',
          'verificationAnalysis',
          'verification_analysis',
          'aiAnalysis',
          'ai_analysis',
          'turn2Content',
          'turn2_content',
          'content',
          'description',
          'summary',
          'analysisResult',
          'analysis_result',
          'verificationResult',
          'verification_result'
        ];

        for (const field of analysisFields) {
          if (rawData[field]) {
            if (typeof rawData[field] === 'string' && rawData[field].trim().length > 0) {
              analysisContent = rawData[field] as string;
              if (debug) {
                console.log(`[API Client DEBUG] Found analysis content in field '${field}' (${analysisContent.length} characters)`);
              }
              break;
            } else if (typeof rawData[field] === 'object' && rawData[field].content) {
              analysisContent = rawData[field].content as string;
              if (debug) {
                console.log(`[API Client DEBUG] Found analysis content in field '${field}.content' (${analysisContent.length} characters)`);
              }
              break;
            }
          }
        }
      }

      // Fallback to top-level llmAnalysis field
      if (!analysisContent && previousVerification.llmAnalysis) {
        analysisContent = previousVerification.llmAnalysis;
        if (debug) {
          console.log(`[API Client DEBUG] Using top-level llmAnalysis field (${analysisContent.length} characters)`);
        }
      }

      if (!analysisContent) {
        if (debug) {
          console.log(`[API Client DEBUG] No analysis content found for previous verification ${previousVerificationId}. This verification may not have completed LLM analysis or was created before analysis features were implemented.`);
        }
        return "No analysis content is available for this previous verification. This may occur if:\n\n1. The previous verification was created before LLM analysis features were implemented\n2. The previous verification failed during the analysis phase\n3. The analysis content was not properly stored\n4. The S3 content could not be retrieved\n\nVerification ID: " + previousVerificationId;
      }

      return analysisContent;
    } catch (error: any) {
      if (debug) {
        console.error(`[API Client DEBUG] Error fetching previous verification analysis for ${previousVerificationId}:`, error);
      }

      // Handle specific error cases more gracefully
      if (error.statusCode === 404 || error.message?.includes('not found')) {
        if (debug) {
          console.log(`[API Client DEBUG] Previous verification ${previousVerificationId} not found, returning null`);
        }
        return null;
      }

      // For other errors, provide a more user-friendly message but don't throw
      const errorMessage = `Unable to load previous verification analysis: ${error.message || 'Unknown error'}`;
      if (debug) {
        console.warn(`[API Client DEBUG] Returning error message instead of throwing: ${errorMessage}`);
      }

      return `Error loading previous verification analysis: ${error.message || 'Unknown error'}\n\nVerification ID: ${previousVerificationId}\n\nThis may be a temporary issue. Please try refreshing the page.`;
    }
  }

  async lookupVerification(params: {
    vendingMachineId?: string;
    startDate?: string;
    endDate?: string;
    limit?: number;
  }, debug: boolean = false): Promise<Verification[]> {
    const apiResponse = await this.makeRequest<any[]>('GET', '/verifications/lookup', params, undefined, undefined, debug);
    return apiResponse.map(item => ({
        ...item,
        verificationType: item.verificationType || item.verificationContext?.verificationType,
        overallAccuracy: item.overallAccuracy || item.overall_accuracy,
        rawData: item,
    })) as Verification[];
  }


  async browseFolder(bucketType: BucketType, path: string = '', debug: boolean = false): Promise<BrowserResponse> {
    const endpoint = path
      ? `/images/browser/${encodeURIComponent(path)}`
      : `/images/browser`;
    return this.makeRequest<BrowserResponse>('GET', endpoint, { bucketType }, undefined, undefined, debug);
  }

  async getImageUrl(imageKey: string, bucketType: BucketType, debug: boolean = false): Promise<string> {
    try {
      const response = await this.makeRequest<ViewResponse>(
        'GET',
        `/images/${encodeURIComponent(imageKey)}/view`,
        { bucketType },
        undefined,
        'application/json', // Changed from 'text/plain' to 'application/json'
        debug
      );
      if (typeof response.presignedUrl !== 'string' || !response.presignedUrl.startsWith('http')) {
        if(debug) console.error(`[API Client DEBUG] getImageUrl for ${imageKey} did not return a valid presigned URL string. Received:`, response);
        // Fallback to a placeholder or throw more specific error for UI
        return "https://placehold.co/600x400.png?text=Invalid+URL+Received";
      }

      // Validate the URL structure before returning
      try {
        const urlTest = new URL(response.presignedUrl);
        if (debug) {
          console.log(`[API Client DEBUG] Presigned URL validation successful:`, {
            host: urlTest.host,
            pathname: urlTest.pathname,
            searchParams: urlTest.searchParams.toString().substring(0, 100) + '...',
            fullUrl: response.presignedUrl.substring(0, 200) + '...'
          });
        }
      } catch (urlError) {
        if (debug) console.error(`[API Client DEBUG] Invalid presigned URL structure:`, urlError);
        return "https://placehold.co/600x400.png?text=Invalid+URL+Structure";
      }

      if (debug) {
        console.log(`[API Client DEBUG] Final presigned URL being returned:`, response.presignedUrl);
      }

      return response.presignedUrl;
    } catch (error: any) {
      const errorMessage = error?.message || 'Unknown error';
      const statusCode = error?.statusCode;

      if (debug) {
        console.error(`[API Client DEBUG] getImageUrl failed for ${imageKey} (bucket: ${bucketType}):`, {
          error: errorMessage,
          statusCode: statusCode,
          details: error?.details
        });
      }

      // Provide more specific error messages based on status code
      if (statusCode === 404) {
        throw new Error(`The requested image does not exist or is not accessible`);
      } else if (statusCode === 403) {
        throw new Error(`Access denied to the requested image`);
      } else if (statusCode === 500 || statusCode === 502 || statusCode === 503) {
        throw new Error(`Server error when accessing image: ${errorMessage}`);
      } else {
        throw new Error(`Failed to get image URL: ${errorMessage}`);
      }
    }
  }

  /**
   * Clean multipart form data boundaries from file content
   * This prevents corrupted JSON files from being uploaded to S3
   */
  private cleanJsonContent(content: string): string {
    // Remove multipart form data boundaries and headers
    let cleaned = content;

    // Remove boundary headers at the start
    cleaned = cleaned.replace(/^------WebKitFormBoundary[a-zA-Z0-9]+\r?\n/, '');
    cleaned = cleaned.replace(/^Content-Disposition: form-data; name="file"; filename="[^"]+"\r?\n/, '');
    cleaned = cleaned.replace(/^Content-Type: application\/json\r?\n/, '');
    cleaned = cleaned.replace(/^\r?\n/, ''); // Remove empty line after headers

    // Remove boundary footers at the end
    cleaned = cleaned.replace(/\r?\n------WebKitFormBoundary[a-zA-Z0-9]+--\r?\n?$/, '');

    // Trim any remaining whitespace
    return cleaned.trim();
  }

  /**
   * Validate that the content is valid JSON
   */
  private validateJsonContent(content: string): { valid: boolean; error?: string } {
    try {
      JSON.parse(content);
      return { valid: true };
    } catch (error: any) {
      return { valid: false, error: error.message };
    }
  }

  /**
   * Process and validate JSON file before upload
   */
  private async processJsonFile(file: File, debug: boolean = false): Promise<File> {
    if (file.type !== 'application/json' && !file.name.toLowerCase().endsWith('.json')) {
      // Not a JSON file, return as-is
      return file;
    }

    try {
      // Read file content
      const content = await file.text();

      if (debug) {
        console.log(`[API Client DEBUG] Original JSON file size: ${content.length} bytes`);
      }

      // Check if content contains multipart boundaries
      const hasMultipartBoundaries = content.includes('WebKitFormBoundary');

      if (hasMultipartBoundaries) {
        if (debug) {
          console.warn('[API Client DEBUG] Detected multipart boundaries in JSON file, cleaning...');
        }

        // Clean the content
        const cleanedContent = this.cleanJsonContent(content);

        if (debug) {
          console.log(`[API Client DEBUG] Cleaned JSON file size: ${cleanedContent.length} bytes`);
        }

        // Validate the cleaned JSON
        const validation = this.validateJsonContent(cleanedContent);
        if (!validation.valid) {
          throw new Error(`Invalid JSON after cleaning: ${validation.error}`);
        }

        // Create a new file with cleaned content
        const cleanedFile = new File([cleanedContent], file.name, {
          type: 'application/json',
          lastModified: file.lastModified
        });

        if (debug) {
          console.log('[API Client DEBUG] Successfully cleaned and validated JSON file');
        }

        return cleanedFile;
      } else {
        // Validate existing JSON content
        const validation = this.validateJsonContent(content);
        if (!validation.valid) {
          throw new Error(`Invalid JSON file: ${validation.error}`);
        }

        if (debug) {
          console.log('[API Client DEBUG] JSON file is already clean and valid');
        }

        return file;
      }
    } catch (error: any) {
      if (debug) {
        console.error('[API Client DEBUG] Error processing JSON file:', error.message);
      }
      throw new Error(`Failed to process JSON file: ${error.message}`);
    }
  }

  async uploadFile(
    file: File,
    bucketType: BucketType,
    folderPathParam?: string, // S3 folder path, optional
    customS3FileNameParam?: string, // S3 object name, optional (overrides file.name)
    debugFlag?: boolean // Debug flag, optional
  ): Promise<UploadResponse> {

    const actualDebug = typeof debugFlag === 'boolean' ? debugFlag : false;
    const actualFolderPath = typeof folderPathParam === 'string' ? folderPathParam : '';

    let s3ObjectNameForQuery: string;
    if (typeof customS3FileNameParam === 'string' && customS3FileNameParam.trim() !== '') {
      s3ObjectNameForQuery = customS3FileNameParam.trim();
    } else {
      s3ObjectNameForQuery = file.name;
    }

    if (!s3ObjectNameForQuery || s3ObjectNameForQuery.trim() === '') {
        s3ObjectNameForQuery = "uploaded_file_unknown.bin"; // Fallback filename if all else fails
        if(actualDebug) {
            console.warn("[API Client DEBUG] S3 object name was empty or invalid, using fallback:", s3ObjectNameForQuery);
        }
    }

    // Process and validate JSON files before upload
    let processedFile: File;
    try {
      processedFile = await this.processJsonFile(file, actualDebug);
    } catch (error: any) {
      if (actualDebug) {
        console.error('[API Client DEBUG] JSON processing failed:', error.message);
      }
      throw {
        message: `File validation failed: ${error.message}`,
        statusCode: 400,
        details: { originalError: error.message }
      } as ApiError;
    }

    if (actualDebug) {
      console.log(`[API Client DEBUG] uploadFile (FormData approach) called with:`, {
        originalFileName: file.name,
        processedFileName: processedFile.name,
        originalFileSize: file.size,
        processedFileSize: processedFile.size,
        fileType: file.type,
        bucketType,
        passedFolderPath: folderPathParam,
        actualFolderPath,
        passedCustomS3FileName: customS3FileNameParam,
        resolvedS3ObjectNameForQuery: s3ObjectNameForQuery,
        passedDebugFlag: debugFlag,
        actualDebug,
        fileWasProcessed: processedFile !== file
      });
    }

    const uploadUrl = new URL(`${this.baseUrl}/api/images/upload`); // Correct API Gateway path
    uploadUrl.searchParams.append('bucketType', bucketType);
    uploadUrl.searchParams.append('fileName', s3ObjectNameForQuery);
    if (actualFolderPath) {
      uploadUrl.searchParams.append('path', actualFolderPath);
    }

    // Create FormData for multipart/form-data upload (required by Lambda function)
    const formData = new FormData();
    formData.append('file', processedFile, s3ObjectNameForQuery);

    const headers: HeadersInit = {
      'X-Api-Key': this.apiKey,
      'Accept': 'application/json'
      // Note: Don't set Content-Type header when using FormData - browser will set it automatically with boundary
    };

    // Create AbortController for timeout handling (25 seconds to avoid API Gateway timeout)
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 25000); // 25 second timeout

    const options: RequestInit = {
      method: 'POST',
      headers,
      body: formData,
      signal: controller.signal,
    };

    if (actualDebug) {
      console.log(`[API Client DEBUG] Request (FormData uploadFile): POST ${uploadUrl.toString()}`, {
        headers: {...headers},
        bodySummary: {
          originalName: file.name,
          processedName: processedFile.name,
          originalSize: file.size,
          processedSize: processedFile.size,
          type: processedFile.type,
          formDataKeys: Array.from(formData.keys()),
          fileWasProcessed: processedFile !== file
        }
      });
    }

    try {
      const response = await fetch(uploadUrl.toString(), options);
      clearTimeout(timeoutId); // Clear timeout on successful response
      const responseText = await response.clone().text();

      if (actualDebug) {
        console.log(`[API Client DEBUG] Response Status (FormData uploadFile): ${response.status} ${response.statusText} for ${uploadUrl.toString()}`);
        console.log(`[API Client DEBUG] RAW Response Body Text (FormData uploadFile):`, responseText);
      }

      if (!response.ok) {
        let errorBody;
        try { errorBody = JSON.parse(responseText); } catch (e) { errorBody = { message: responseText || response.statusText }; }
        if (actualDebug) console.error(`[API Client DEBUG] API Error Body (direct uploadFile):`, errorBody);
        const errorMessage = errorBody.message || errorBody.error || `API Error ${response.status}`;
        throw { message: errorMessage, statusCode: response.status, details: errorBody } as ApiError;
      }

      if (response.status === 204) {
        if (actualDebug) console.log(`[API Client DEBUG] Response Status 204 (FormData uploadFile), returning potentially empty but successful.`);
        // For 204, the Go backend might not return a body, but it's a success.
        // We can construct a minimal success response if needed by the frontend.
        return { success: true, message: "Upload successful (204 No Content)", files: [{ originalName: s3ObjectNameForQuery, key: (actualFolderPath ? actualFolderPath + "/" : "") + s3ObjectNameForQuery, size: processedFile.size, contentType: processedFile.type || 'application/octet-stream', bucket: bucketType === 'reference' ? appConfigInstance.referenceS3BucketName : appConfigInstance.checkingS3BucketName }] } as UploadResponse;
      }

      const responseContentType = response.headers.get("content-type");
      if (responseContentType?.includes("application/json")) {
        try {
            const jsonData = JSON.parse(responseText);
            if (actualDebug) console.log(`[API Client DEBUG] Parsed JSON Response (direct uploadFile):`, jsonData);
            return jsonData as UploadResponse;
        } catch (parseError: any) {
            if (actualDebug) console.error(`[API Client DEBUG] JSON Parse Error (direct uploadFile): `, parseError, `for response text: ${responseText}`);
            throw { message: `Failed to parse JSON response: ${parseError.message}`, statusCode: response.status, details: { responseBody: responseText } } as ApiError;
        }
      }
      if (actualDebug) console.log(`[API Client DEBUG] Response not JSON (FormData uploadFile), but considered success. Content-Type: ${responseContentType}`);
      return { success: true, message: "Upload successful, but response was not standard JSON.", files: [{ originalName: s3ObjectNameForQuery, key: (actualFolderPath ? actualFolderPath + "/" : "") + s3ObjectNameForQuery, size: processedFile.size, contentType: processedFile.type || 'application/octet-stream', bucket: bucketType === 'reference' ? appConfigInstance.referenceS3BucketName : appConfigInstance.checkingS3BucketName }] } as UploadResponse;

    } catch (error: any) {
      clearTimeout(timeoutId); // Clear timeout on error
      const errorMsg = error.message || 'Unknown fetch error';
      const isAbortError = error.name === 'AbortError';
      const isFetchTypeError = error.name === 'TypeError' || errorMsg.toLowerCase().includes('failed to fetch');

      if (actualDebug) {
        console.error(`[API Client DEBUG] Fetch/Processing Error (direct uploadFile) for ${uploadUrl.toString()}:`, error);
      }

      if (isAbortError) {
        const timeoutMsg = `Upload request timed out after 60 seconds. The file upload and rendering process is taking longer than expected. This can happen with large JSON files or complex layouts that require more processing time.`;
        if (actualDebug) console.warn(`[API Client DEBUG] Upload request was aborted due to timeout.`);
        throw {
          message: timeoutMsg,
          statusCode: 408, // Request Timeout
          details: { originalError: error.toString(), name: error.name, url: uploadUrl.toString(), timeout: true }
        } as ApiError;
      }

      if (isFetchTypeError) {
        const detailedMsg = `Network request to API failed: "${errorMsg}".
        This can be caused by:
        1. BROWSER EXTENSIONS (like ad blockers/privacy tools) interfering. Try disabling them.
        2. NETWORK ISSUES (CORS, DNS, firewall). Check your connection and if the API is reachable.
        3. API SERVER being down or misconfigured (e.g. incorrect URL path, method, or body format expected by API Gateway/Lambda).
        URL: ${uploadUrl.toString()}`;
        if (actualDebug) console.warn(`[API Client DEBUG] A 'TypeError' or "Failed to fetch" occurred during direct uploadFile. URL: ${uploadUrl.toString()}`);
         throw {
            message: detailedMsg,
            statusCode: 0,
            details: { originalError: error.toString(), name: error.name, url: uploadUrl.toString() }
        } as ApiError;
      }

      if (error.statusCode !== undefined) throw error;

      throw {
        message: `Unexpected error during API request (FormData uploadFile): ${errorMsg}`,
        statusCode: 0,
        details: { originalError: error.toString(), name: error.name, url: uploadUrl.toString() }
      } as ApiError;
    }
  }
}

const apiClient = new ApiClient(appConfigInstance);
export default apiClient;
    