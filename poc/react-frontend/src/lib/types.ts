
// Core API Types based on Guide Part 2
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginationParams {
  limit?: number;
  offset?: number;
}

export interface PaginationResponse {
  total: number;
  limit: number;
  offset: number;
  nextOffset?: number;
}

// Verification Types from Guide (types/api.ts)
export interface Verification {
  verificationId: string;
  verificationStatus: 'CORRECT' | 'INCORRECT' | 'PENDING' | 'ERROR' | 'PROCESSING' | 'COMPLETED';
  vendingMachineId?: string;
  verificationType?: 'LAYOUT_VS_CHECKING' | 'PREVIOUS_VS_CURRENT';
  verificationAt: string; // ISO Date String
  overallAccuracy?: number;
  referenceImageUrl?: string;
  checkingImageUrl?: string;
  turn1ProcessedPath?: string;
  turn2ProcessedPath?: string;
  llmAnalysis?: string; // Retained as per previous discussions for fallback
  previousVerificationId?: string; // Reference to previous verification for PREVIOUS_VS_CURRENT type
  createdAt?: string; // ISO Date String
  updatedAt?: string; // ISO Date String
  completedAt?: string;
  rawData?: Record<string, any>;
}


export interface VerificationListResponse {
  results: Verification[];
  pagination: PaginationResponse;
}

export interface VerificationContext {
  verificationType: 'LAYOUT_VS_CHECKING' | 'PREVIOUS_VS_CURRENT';
  referenceImageUrl: string; // S3 URI
  checkingImageUrl: string;  // S3 URI
  vendingMachineId?: string; // Optional, placed inside context
}

export interface CreateVerificationRequest {
  verificationContext: VerificationContext;
}

// S3 Integration Types (from the provided React/Next.js S3 Integration Guide)
export type BucketType = 'reference' | 'checking';

export interface BrowserResponse {
  currentPath: string;
  parentPath?: string;
  items: BrowserItem[];
  totalItems: number;
}

export interface BrowserItem {
  name: string;
  path: string; // Full S3 key/path
  type: 'folder' | 'image' | 'file';
  size?: number;
  lastModified?: string; // ISO timestamp
  contentType?: string; // MIME type
  // Note: The guide does not explicitly put a presigned 'url' field here.
  // It implies getImageViewUrl should be called. We'll adapt components.
}

export interface UploadResponse {
  success: boolean;
  message: string;
  files?: UploadedFile[];
  errors?: string[];
  renderResult?: RenderResult;
}

export interface RenderResult {
  rendered: boolean;
  layoutId?: number;
  layoutPrefix?: string;
  processedKey?: string;
  message?: string;
}

export interface UploadedFile {
  originalName: string;
  key: string; // S3 key where file was stored
  size: number;
  contentType: string;
  bucket: string;
  url?: string; // Optional access URL (e.g., S3 URI or presigned if backend provides it post-upload)
}

export interface ViewResponse {
  presignedUrl: string;
}

// Health Check Types from Guide (types/api.ts)
export interface ServiceInfo {
  status: 'healthy' | 'unhealthy' | 'degraded';
  message?: string;
  details?: Record<string, any>;
}

export interface HealthResponse {
  status: 'healthy' | 'unhealthy' | 'degraded';
  version?: string;
  timestamp?: string; // ISO Date String
  services?: Record<string, ServiceInfo>;
}

export interface VerificationConversationResponse {
  turn1?: string | { content?: string; [key: string]: any };
  turn2?: string | { content?: string; [key: string]: any };
  turn1Content?: { content?: string; [key: string]: any };
  turn2Content?: { content?: string; [key: string]: any };
  [key: string]: any;
}

export interface VerificationStatusResponse {
  verificationId: string;
  status: 'COMPLETED' | 'RUNNING' | 'FAILED';
  currentStatus?: string;
  verificationStatus?: string;
  s3References: {
    turn1Processed?: string;
    turn2Processed?: string;
  };
  summary: {
    message: string;
    verificationAt: string;
    verificationStatus: string;
    overallAccuracy?: number;
    correctPositions?: number;
    discrepantPositions?: number;
  };
  llmResponse?: string;
  verificationSummary?: Record<string, any>;
}

// Frontend-specific types
export interface ApiError {
  message: string;
  statusCode?: number;
  details?: any;
}

export interface UploadHistoryItem {
  id: string; 
  fileName: string;
  s3Url: string; // s3://bucket/key URI or the URL returned by API
  bucket: string;
  key: string; 
  timestamp: string; 
  type: BucketType;
}

// Genkit flow types
export interface PredictiveHealthCheckInput {
  historicalPerformanceData: string;
}

export interface PredictiveHealthCheckOutput {
  prediction: string;
  recommendations: string;
}
