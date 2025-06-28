# React/Next.js TypeScript Integration Guide
## Kootoro Vending Machine Verification System

## 📋 Table of Contents
- [API Overview](#api-overview)
- [Architecture & Design Patterns](#architecture--design-patterns)
- [TypeScript Types & Interfaces](#typescript-types--interfaces)
- [Environment Setup](#environment-setup)
- [API Client Design](#api-client-design)
- [Service Layer Architecture](#service-layer-architecture)
- [Custom Hooks Design](#custom-hooks-design)
- [Component Design Patterns](#component-design-patterns)
- [Error Handling Strategy](#error-handling-strategy)
- [State Management Patterns](#state-management-patterns)
- [Performance Optimization](#performance-optimization)
- [Testing Strategy](#testing-strategy)

## 🔍 API Overview

### Base Configuration
- **API Gateway ID**: `hpux2uegnd`
- **Base URL**: `https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1`
- **Stage**: `v1`
- **Region**: `us-east-1`
- **API Key Required**: Yes
- **CORS Enabled**: Yes

### Available Endpoints
```
GET    /api/health                                    # Health check
GET    /api/verifications                            # List verifications
POST   /api/verifications                            # Create verification
GET    /api/verifications/{id}                       # Get verification details
GET    /api/verifications/{id}/conversation          # Get conversation
GET    /api/verifications/lookup                     # Lookup verification
POST   /api/images/upload                            # Upload image
GET    /api/images/{key}/view                        # Get image view URL
GET    /api/images/browser                           # Browse images
GET    /api/images/browser/{path+}                   # Browse specific path
POST   /workflow                                     # Workflow endpoint
```

## 🏗️ Architecture & Design Patterns

### Layered Architecture
```
┌─────────────────────────────────────┐
│           UI Components             │ ← Presentation Layer
├─────────────────────────────────────┤
│         Custom Hooks               │ ← Business Logic Layer
├─────────────────────────────────────┤
│        Service Layer              │ ← Data Access Layer
├─────────────────────────────────────┤
│        API Client                 │ ← Network Layer
└─────────────────────────────────────┘
```

### Design Principles
- **Separation of Concerns**: Each layer has a single responsibility
- **Dependency Injection**: Services are injected into hooks and components
- **Error Boundaries**: Graceful error handling at component level
- **Type Safety**: Full TypeScript coverage with strict mode
- **Immutable State**: Use of immutable patterns for state management
- **Composition over Inheritance**: Favor hooks and composition patterns

## 🔷 TypeScript Types & Interfaces

### Core API Types (types/api.ts)
```typescript
// Base API Response
export interface ApiResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

// Pagination
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

// Verification Types
export interface Verification {
  verificationId: string;
  verificationStatus: 'CORRECT' | 'INCORRECT' | 'PENDING' | 'ERROR';
  vendingMachineId: string;
  verificationAt: string;
  overallAccuracy?: number;
  referenceImageUrl?: string;
  checkingImageUrl?: string;
}

export interface VerificationListResponse {
  results: Verification[];
  pagination: PaginationResponse;
}

export interface CreateVerificationRequest {
  vendingMachineId: string;
  referenceImageUrl: string;
  checkingImageUrl: string;
}

// Image Types
export interface ImageUploadRequest {
  fileContent: string; // base64
}

export interface ImageUploadResponse {
  message: string;
  s3Key: string;
  bucket: string;
  fileSize: number;
  contentType: string;
}

export interface ImageViewResponse {
  presignedUrl: string;
}

// Health Check Types
export interface ServiceInfo {
  status: 'healthy' | 'unhealthy' | 'degraded';
  message?: string;
  details?: Record<string, string>;
}

export interface HealthResponse {
  status: 'healthy' | 'unhealthy' | 'degraded';
  version: string;
  timestamp: string;
  services: Record<string, ServiceInfo>;
}
```

## 🔧 Environment Setup

### Next.js (.env.local)
```bash
# API Configuration
NEXT_PUBLIC_API_BASE_URL=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api
NEXT_PUBLIC_API_KEY=WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH

# Optional: Environment identifier
NEXT_PUBLIC_ENV=development
```

### Create React App (.env)
```bash
# API Configuration
REACT_APP_API_BASE_URL=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api
REACT_APP_API_KEY=WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH

# Optional: Environment identifier
REACT_APP_ENV=development
```

## 🌐 API Client Design

### Design Pattern: Axios with Interceptors
```typescript
// lib/apiClient.ts
import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import { ApiResponse } from '../types/api';

class ApiClient {
  private client: AxiosInstance;

  constructor() {
    this.client = axios.create({
      baseURL: process.env.NEXT_PUBLIC_API_BASE_URL || process.env.REACT_APP_API_BASE_URL,
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
        'X-Api-Key': process.env.NEXT_PUBLIC_API_KEY || process.env.REACT_APP_API_KEY,
      },
    });

    this.setupInterceptors();
  }

  private setupInterceptors(): void {
    // Request interceptor for logging and auth
    this.client.interceptors.request.use(
      (config) => {
        console.log(`🚀 ${config.method?.toUpperCase()} ${config.url}`);
        return config;
      },
      (error) => Promise.reject(this.handleError(error))
    );

    // Response interceptor for error handling
    this.client.interceptors.response.use(
      (response) => {
        console.log(`✅ ${response.status} ${response.config.url}`);
        return response;
      },
      (error) => Promise.reject(this.handleError(error))
    );
  }

  private handleError(error: any): ApiError {
    const status = error.response?.status || 500;
    const message = error.response?.data?.message || error.message || 'Unknown error';

    return new ApiError(message, status, error.response?.data);
  }

  // Generic request methods
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.get<T>(url, config);
    return response.data;
  }

  async post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.post<T>(url, data, config);
    return response.data;
  }

  async put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.put<T>(url, data, config);
    return response.data;
  }

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await this.client.delete<T>(url, config);
    return response.data;
  }
}

// Custom Error Class
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public data?: any
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export const apiClient = new ApiClient();
```

## 🔐 Authentication

The API uses API Key authentication via the `X-Api-Key` header. The key is automatically included in all requests through the API client configuration.

### Rate Limiting
- **Burst Limit**: 400 requests
- **Rate Limit**: 200 requests/second
- **Throttling**: Enabled at API Gateway level

## 🏢 Service Layer Architecture

### Design Pattern: Repository Pattern with TypeScript

### Abstract Base Service
```typescript
// services/BaseService.ts
import { apiClient } from '../lib/apiClient';

export abstract class BaseService {
  protected async handleRequest<T>(request: () => Promise<T>): Promise<T> {
    try {
      return await request();
    } catch (error) {
      console.error(`${this.constructor.name} Error:`, error);
      throw error;
    }
  }
}
```

### Health Service (services/HealthService.ts)
```typescript
import { BaseService } from './BaseService';
import { HealthResponse } from '../types/api';
import { apiClient } from '../lib/apiClient';

export class HealthService extends BaseService {
  async checkHealth(): Promise<HealthResponse> {
    return this.handleRequest(() =>
      apiClient.get<HealthResponse>('/health')
    );
  }
}

export const healthService = new HealthService();
```

### Verification Service (services/VerificationService.ts)
```typescript
import { BaseService } from './BaseService';
import {
  Verification,
  VerificationListResponse,
  CreateVerificationRequest,
  PaginationParams
} from '../types/api';
import { apiClient } from '../lib/apiClient';

interface VerificationQueryParams extends PaginationParams {
  verificationStatus?: string;
  vendingMachineId?: string;
  startDate?: string;
  endDate?: string;
}

export class VerificationService extends BaseService {
  async listVerifications(params: VerificationQueryParams = {}): Promise<VerificationListResponse> {
    return this.handleRequest(() =>
      apiClient.get<VerificationListResponse>('/verifications', { params })
    );
  }

  async getVerification(verificationId: string): Promise<Verification> {
    return this.handleRequest(() =>
      apiClient.get<Verification>(`/verifications/${verificationId}`)
    );
  }

  async createVerification(data: CreateVerificationRequest): Promise<Verification> {
    return this.handleRequest(() =>
      apiClient.post<Verification>('/verifications', data)
    );
  }

  async getConversation(verificationId: string): Promise<{ turn1: string; turn2: string }> {
    return this.handleRequest(() =>
      apiClient.get<{ turn1: string; turn2: string }>(`/verifications/${verificationId}/conversation`)
    );
  }

  async lookupVerification(params: Record<string, any>): Promise<Verification[]> {
    return this.handleRequest(() =>
      apiClient.get<Verification[]>('/verifications/lookup', { params })
    );
  }
}

export const verificationService = new VerificationService();
```

### Image Service (services/ImageService.ts)
```typescript
import { BaseService } from './BaseService';
import { ImageUploadRequest, ImageUploadResponse, ImageViewResponse } from '../types/api';
import { apiClient } from '../lib/apiClient';

type BucketType = 'reference' | 'checking';

interface ImageBrowserResponse {
  items: Array<{
    key: string;
    size: number;
    lastModified: string;
    isDirectory: boolean;
  }>;
  path: string;
  bucketType: BucketType;
}

export class ImageService extends BaseService {
  private readonly MAX_FILE_SIZE = 10 * 1024 * 1024; // 10MB
  private readonly ALLOWED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];

  async uploadImage(
    file: File,
    bucketType: BucketType = 'reference',
    path: string = 'uploads'
  ): Promise<ImageUploadResponse> {
    this.validateFile(file);

    const base64Content = await this.fileToBase64(file);
    const params = { bucketType, fileName: file.name, path };

    return this.handleRequest(() =>
      apiClient.post<ImageUploadResponse>('/images/upload',
        { fileContent: base64Content } as ImageUploadRequest,
        { params }
      )
    );
  }

  async getImageViewUrl(imageKey: string, bucketType: BucketType = 'reference'): Promise<ImageViewResponse> {
    return this.handleRequest(() =>
      apiClient.get<ImageViewResponse>(`/images/${encodeURIComponent(imageKey)}/view`, {
        params: { bucketType }
      })
    );
  }

  async browseImages(bucketType: BucketType = 'reference', path: string = ''): Promise<ImageBrowserResponse> {
    const endpoint = path ? `/images/browser/${path}` : '/images/browser';

    return this.handleRequest(() =>
      apiClient.get<ImageBrowserResponse>(endpoint, {
        params: { bucketType }
      })
    );
  }

  private fileToBase64(file: File): Promise<string> {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.readAsDataURL(file);
      reader.onload = () => {
        const result = reader.result as string;
        const base64 = result.split(',')[1];
        resolve(base64);
      };
      reader.onerror = error => reject(error);
    });
  }

  private validateFile(file: File): void {
    if (file.size > this.MAX_FILE_SIZE) {
      throw new Error('File size must be less than 10MB');
    }

    if (!this.ALLOWED_TYPES.includes(file.type)) {
      throw new Error('File type not supported. Use JPEG, PNG, GIF, or WebP');
    }
  }
}

export const imageService = new ImageService();
```

## 🎣 Custom Hooks Design

### Design Pattern: Compound Hook Pattern
```typescript
// hooks/useAsyncState.ts
import { useState, useCallback } from 'react';

interface AsyncState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

export function useAsyncState<T>() {
  const [state, setState] = useState<AsyncState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(async (asyncFunction: () => Promise<T>) => {
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const data = await asyncFunction();
      setState({ data, loading: false, error: null });
      return data;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error';
      setState(prev => ({ ...prev, loading: false, error: errorMessage }));
      throw error;
    }
  }, []);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return { ...state, execute, reset };
}
```

### Verification Hooks (hooks/useVerifications.ts)
```typescript
import { useCallback, useEffect } from 'react';
import { verificationService } from '../services/VerificationService';
import { Verification, VerificationListResponse } from '../types/api';
import { useAsyncState } from './useAsyncState';

interface UseVerificationsParams {
  verificationStatus?: string;
  limit?: number;
  autoFetch?: boolean;
}

export function useVerifications(params: UseVerificationsParams = {}) {
  const { data, loading, error, execute, reset } = useAsyncState<VerificationListResponse>();

  const fetchVerifications = useCallback(async (overrideParams = {}) => {
    return execute(() => verificationService.listVerifications({
      limit: 20,
      ...params,
      ...overrideParams,
    }));
  }, [execute, params]);

  const loadMore = useCallback(() => {
    if (data?.pagination.nextOffset) {
      return fetchVerifications({ offset: data.pagination.nextOffset });
    }
  }, [data?.pagination.nextOffset, fetchVerifications]);

  useEffect(() => {
    if (params.autoFetch !== false) {
      fetchVerifications();
    }
  }, [fetchVerifications, params.autoFetch]);

  return {
    verifications: data?.results || [],
    pagination: data?.pagination,
    loading,
    error,
    fetchVerifications,
    loadMore,
    reset,
  };
}
```

### Image Upload Hook (hooks/useImageUpload.ts)
```typescript
import { useState, useCallback } from 'react';
import { imageService } from '../services/ImageService';
import { ImageUploadResponse } from '../types/api';

interface UploadProgress {
  progress: number;
  uploading: boolean;
}

export function useImageUpload() {
  const [uploadState, setUploadState] = useState<UploadProgress>({
    progress: 0,
    uploading: false,
  });
  const [error, setError] = useState<string | null>(null);

  const uploadFile = useCallback(async (
    file: File,
    bucketType: 'reference' | 'checking' = 'reference',
    path: string = 'uploads'
  ): Promise<ImageUploadResponse> => {
    setUploadState({ progress: 0, uploading: true });
    setError(null);

    try {
      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setUploadState(prev => ({
          ...prev,
          progress: Math.min(prev.progress + 10, 90)
        }));
      }, 200);

      const result = await imageService.uploadImage(file, bucketType, path);

      clearInterval(progressInterval);
      setUploadState({ progress: 100, uploading: false });

      // Reset progress after delay
      setTimeout(() => {
        setUploadState({ progress: 0, uploading: false });
      }, 1000);

      return result;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Upload failed';
      setError(errorMessage);
      setUploadState({ progress: 0, uploading: false });
      throw err;
    }
  }, []);

  const reset = useCallback(() => {
    setUploadState({ progress: 0, uploading: false });
    setError(null);
  }, []);

  return {
    ...uploadState,
    error,
    uploadFile,
    reset,
  };
}
```

## 🧩 Component Design Patterns

### Design Pattern: Compound Components with TypeScript
```typescript
// components/VerificationList/VerificationList.tsx
import React, { createContext, useContext } from 'react';
import { Verification } from '../../types/api';

interface VerificationListContextValue {
  verifications: Verification[];
  onSelect?: (verification: Verification) => void;
  selectedId?: string;
}

const VerificationListContext = createContext<VerificationListContextValue | null>(null);

interface VerificationListProps {
  children: React.ReactNode;
  verifications: Verification[];
  onSelect?: (verification: Verification) => void;
  selectedId?: string;
}

export function VerificationList({ children, ...contextValue }: VerificationListProps) {
  return (
    <VerificationListContext.Provider value={contextValue}>
      <div className="verification-list">
        {children}
      </div>
    </VerificationListContext.Provider>
  );
}

// Compound components
VerificationList.Grid = function VerificationGrid({ children }: { children: React.ReactNode }) {
  return <div className="verification-grid">{children}</div>;
};

VerificationList.Item = function VerificationItem({
  verification
}: {
  verification: Verification
}) {
  const context = useContext(VerificationListContext);
  if (!context) throw new Error('VerificationItem must be used within VerificationList');

  const { onSelect, selectedId } = context;
  const isSelected = selectedId === verification.verificationId;

  return (
    <div
      className={`verification-item ${isSelected ? 'selected' : ''}`}
      onClick={() => onSelect?.(verification)}
    >
      <div className="verification-status">{verification.verificationStatus}</div>
      <div className="verification-id">{verification.verificationId}</div>
      <div className="verification-machine">{verification.vendingMachineId}</div>
    </div>
  );
};

// Usage:
// <VerificationList verifications={data} onSelect={handleSelect}>
//   <VerificationList.Grid>
//     {verifications.map(v => (
//       <VerificationList.Item key={v.verificationId} verification={v} />
//     ))}
//   </VerificationList.Grid>
// </VerificationList>
```

## 🚨 Error Handling Strategy

### Design Pattern: Error Boundaries with Context
```typescript
// components/ErrorBoundary/ErrorBoundary.tsx
import React, { Component, ErrorInfo, ReactNode } from 'react';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface State {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo);
    this.props.onError?.(error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback || (
        <div className="error-boundary">
          <h2>Something went wrong</h2>
          <details>
            <summary>Error details</summary>
            <pre>{this.state.error?.stack}</pre>
          </details>
        </div>
      );
    }

    return this.props.children;
  }
}
```

### Global Error Context
```typescript
// contexts/ErrorContext.tsx
import React, { createContext, useContext, useState, ReactNode } from 'react';

interface ErrorContextValue {
  errors: string[];
  addError: (error: string) => void;
  removeError: (index: number) => void;
  clearErrors: () => void;
}

const ErrorContext = createContext<ErrorContextValue | null>(null);

export function ErrorProvider({ children }: { children: ReactNode }) {
  const [errors, setErrors] = useState<string[]>([]);

  const addError = (error: string) => {
    setErrors(prev => [...prev, error]);
    // Auto-remove after 5 seconds
    setTimeout(() => {
      setErrors(prev => prev.slice(1));
    }, 5000);
  };

  const removeError = (index: number) => {
    setErrors(prev => prev.filter((_, i) => i !== index));
  };

  const clearErrors = () => setErrors([]);

  return (
    <ErrorContext.Provider value={{ errors, addError, removeError, clearErrors }}>
      {children}
    </ErrorContext.Provider>
  );
}

export function useError() {
  const context = useContext(ErrorContext);
  if (!context) throw new Error('useError must be used within ErrorProvider');
  return context;
}
```

## 🏪 State Management Patterns

### Design Pattern: Zustand with TypeScript
```typescript
// stores/verificationStore.ts
import { create } from 'zustand';
import { devtools } from 'zustand/middleware';
import { Verification } from '../types/api';

interface VerificationState {
  verifications: Verification[];
  selectedVerification: Verification | null;
  filters: {
    status?: string;
    machineId?: string;
  };

  // Actions
  setVerifications: (verifications: Verification[]) => void;
  addVerification: (verification: Verification) => void;
  selectVerification: (verification: Verification | null) => void;
  setFilters: (filters: Partial<VerificationState['filters']>) => void;
  clearFilters: () => void;
}

export const useVerificationStore = create<VerificationState>()(
  devtools(
    (set, get) => ({
      verifications: [],
      selectedVerification: null,
      filters: {},

      setVerifications: (verifications) =>
        set({ verifications }, false, 'setVerifications'),

      addVerification: (verification) =>
        set(
          (state) => ({
            verifications: [...state.verifications, verification]
          }),
          false,
          'addVerification'
        ),

      selectVerification: (verification) =>
        set({ selectedVerification: verification }, false, 'selectVerification'),

      setFilters: (newFilters) =>
        set(
          (state) => ({
            filters: { ...state.filters, ...newFilters }
          }),
          false,
          'setFilters'
        ),

      clearFilters: () =>
        set({ filters: {} }, false, 'clearFilters'),
    }),
    { name: 'verification-store' }
  )
);
```

## ⚡ Performance Optimization

### Design Pattern: React Query Integration
```typescript
// hooks/useVerificationsQuery.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { verificationService } from '../services/VerificationService';
import { Verification, CreateVerificationRequest } from '../types/api';

export function useVerificationsQuery(params = {}) {
  return useQuery({
    queryKey: ['verifications', params],
    queryFn: () => verificationService.listVerifications(params),
    staleTime: 5 * 60 * 1000, // 5 minutes
    cacheTime: 10 * 60 * 1000, // 10 minutes
  });
}

export function useCreateVerificationMutation() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateVerificationRequest) =>
      verificationService.createVerification(data),
    onSuccess: (newVerification) => {
      // Invalidate and refetch verifications
      queryClient.invalidateQueries({ queryKey: ['verifications'] });

      // Optimistically update the cache
      queryClient.setQueryData(['verifications'], (old: any) => ({
        ...old,
        results: [newVerification, ...(old?.results || [])],
      }));
    },
  });
}
```

### Memoization Patterns
```typescript
// components/VerificationCard/VerificationCard.tsx
import React, { memo } from 'react';
import { Verification } from '../../types/api';

interface VerificationCardProps {
  verification: Verification;
  onSelect?: (verification: Verification) => void;
  isSelected?: boolean;
}

export const VerificationCard = memo<VerificationCardProps>(({
  verification,
  onSelect,
  isSelected = false,
}) => {
  const handleClick = React.useCallback(() => {
    onSelect?.(verification);
  }, [onSelect, verification]);

  const statusColor = React.useMemo(() => {
    switch (verification.verificationStatus) {
      case 'CORRECT': return 'green';
      case 'INCORRECT': return 'red';
      case 'PENDING': return 'yellow';
      default: return 'gray';
    }
  }, [verification.verificationStatus]);

  return (
    <div
      className={`verification-card ${isSelected ? 'selected' : ''}`}
      onClick={handleClick}
      style={{ borderColor: statusColor }}
    >
      <div className="verification-id">{verification.verificationId}</div>
      <div className="verification-status">{verification.verificationStatus}</div>
      <div className="verification-machine">{verification.vendingMachineId}</div>
    </div>
  );
});

VerificationCard.displayName = 'VerificationCard';
```

## 🧪 Testing Strategy

### Service Layer Testing
```typescript
// __tests__/services/VerificationService.test.ts
import { verificationService } from '../../services/VerificationService';
import { apiClient } from '../../lib/apiClient';

// Mock the API client
jest.mock('../../lib/apiClient');
const mockedApiClient = apiClient as jest.Mocked<typeof apiClient>;

describe('VerificationService', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('listVerifications', () => {
    it('should fetch verifications with correct parameters', async () => {
      const mockResponse = {
        results: [{ verificationId: '123', verificationStatus: 'CORRECT' }],
        pagination: { total: 1, limit: 20, offset: 0 }
      };

      mockedApiClient.get.mockResolvedValue(mockResponse);

      const result = await verificationService.listVerifications({ limit: 10 });

      expect(mockedApiClient.get).toHaveBeenCalledWith('/verifications', {
        params: { limit: 10 }
      });
      expect(result).toEqual(mockResponse);
    });

    it('should handle API errors gracefully', async () => {
      const error = new Error('API Error');
      mockedApiClient.get.mockRejectedValue(error);

      await expect(verificationService.listVerifications()).rejects.toThrow('API Error');
    });
  });
});
```

### Hook Testing
```typescript
// __tests__/hooks/useVerifications.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { useVerifications } from '../../hooks/useVerifications';
import { verificationService } from '../../services/VerificationService';

jest.mock('../../services/VerificationService');
const mockedVerificationService = verificationService as jest.Mocked<typeof verificationService>;

describe('useVerifications', () => {
  it('should fetch verifications on mount', async () => {
    const mockData = {
      results: [{ verificationId: '123', verificationStatus: 'CORRECT' }],
      pagination: { total: 1, limit: 20, offset: 0 }
    };

    mockedVerificationService.listVerifications.mockResolvedValue(mockData);

    const { result } = renderHook(() => useVerifications());

    expect(result.current.loading).toBe(true);

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.verifications).toEqual(mockData.results);
    expect(result.current.pagination).toEqual(mockData.pagination);
  });
});
```

### Component Testing with React Testing Library
```typescript
// __tests__/components/VerificationCard.test.tsx
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { VerificationCard } from '../../components/VerificationCard/VerificationCard';
import { Verification } from '../../types/api';

const mockVerification: Verification = {
  verificationId: '123',
  verificationStatus: 'CORRECT',
  vendingMachineId: 'VM001',
  verificationAt: '2024-01-01T00:00:00Z',
};

describe('VerificationCard', () => {
  it('should render verification details', () => {
    render(<VerificationCard verification={mockVerification} />);

    expect(screen.getByText('123')).toBeInTheDocument();
    expect(screen.getByText('CORRECT')).toBeInTheDocument();
    expect(screen.getByText('VM001')).toBeInTheDocument();
  });

  it('should call onSelect when clicked', () => {
    const onSelect = jest.fn();
    render(<VerificationCard verification={mockVerification} onSelect={onSelect} />);

    fireEvent.click(screen.getByRole('button'));

    expect(onSelect).toHaveBeenCalledWith(mockVerification);
  });

  it('should apply selected styling when isSelected is true', () => {
    const { container } = render(
      <VerificationCard verification={mockVerification} isSelected={true} />
    );

    expect(container.firstChild).toHaveClass('selected');
  });
});
```

### Integration Testing
```typescript
// __tests__/integration/VerificationFlow.test.tsx
import React from 'react';
import { render, screen, waitFor, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { VerificationListPage } from '../../pages/VerificationListPage';

// Mock API responses
const mockVerifications = {
  results: [
    { verificationId: '1', verificationStatus: 'CORRECT', vendingMachineId: 'VM001' },
    { verificationId: '2', verificationStatus: 'PENDING', vendingMachineId: 'VM002' },
  ],
  pagination: { total: 2, limit: 20, offset: 0 }
};

describe('Verification Flow Integration', () => {
  let queryClient: QueryClient;

  beforeEach(() => {
    queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } }
    });
  });

  it('should display verifications and handle selection', async () => {
    // Mock API call
    global.fetch = jest.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve(mockVerifications),
    });

    render(
      <QueryClientProvider client={queryClient}>
        <VerificationListPage />
      </QueryClientProvider>
    );

    // Wait for data to load
    await waitFor(() => {
      expect(screen.getByText('VM001')).toBeInTheDocument();
    });

    // Test selection
    fireEvent.click(screen.getByText('VM001'));

    expect(screen.getByText('VM001')).toHaveClass('selected');
  });
});
```

---

## 📚 Summary

This guide provides a comprehensive TypeScript-first approach to integrating with the Kootoro Vending Machine Verification API, focusing on:

- **Clean Architecture**: Layered design with clear separation of concerns
- **Type Safety**: Full TypeScript coverage with strict typing
- **Design Patterns**: Repository pattern, compound components, custom hooks
- **Error Handling**: Comprehensive error boundaries and global error management
- **Performance**: Memoization, React Query, and optimization strategies
- **Testing**: Unit, integration, and component testing strategies

The architecture promotes maintainability, scalability, and developer experience while ensuring robust error handling and optimal performance.
```

## 🎣 React Hooks

### useVerifications Hook (hooks/useVerifications.js)
```javascript
import { useState, useEffect, useCallback } from 'react';
import { verificationService } from '../services/verificationService';

export const useVerifications = (initialParams = {}) => {
  const [verifications, setVerifications] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [pagination, setPagination] = useState({
    total: 0,
    limit: 20,
    offset: 0,
    nextOffset: null,
  });

  const fetchVerifications = useCallback(async (params = {}) => {
    setLoading(true);
    setError(null);

    try {
      const response = await verificationService.listVerifications({
        ...initialParams,
        ...params,
      });

      setVerifications(response.results || []);
      setPagination(response.pagination || {});
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [initialParams]);

  const loadMore = useCallback(() => {
    if (pagination.nextOffset) {
      fetchVerifications({ offset: pagination.nextOffset });
    }
  }, [pagination.nextOffset, fetchVerifications]);

  const refresh = useCallback(() => {
    fetchVerifications({ offset: 0 });
  }, [fetchVerifications]);

  useEffect(() => {
    fetchVerifications();
  }, [fetchVerifications]);

  return {
    verifications,
    loading,
    error,
    pagination,
    loadMore,
    refresh,
    fetchVerifications,
  };
};
```

### useImageUpload Hook (hooks/useImageUpload.js)
```javascript
import { useState, useCallback } from 'react';
import { imageService } from '../services/imageService';

export const useImageUpload = () => {
  const [uploading, setUploading] = useState(false);
  const [progress, setProgress] = useState(0);
  const [error, setError] = useState(null);

  const uploadFile = useCallback(async (file, bucketType = 'reference', path = 'uploads') => {
    setUploading(true);
    setError(null);
    setProgress(0);

    try {
      // Validate file first
      imageService.validateFile(file);

      // Simulate progress for better UX
      const progressInterval = setInterval(() => {
        setProgress(prev => Math.min(prev + 10, 90));
      }, 200);

      const result = await imageService.uploadImage(file, bucketType, path);

      clearInterval(progressInterval);
      setProgress(100);

      return result;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setUploading(false);
      setTimeout(() => setProgress(0), 1000);
    }
  }, []);

  const reset = useCallback(() => {
    setError(null);
    setProgress(0);
  }, []);

  return {
    uploadFile,
    uploading,
    progress,
    error,
    reset,
  };
};
```

### useHealthCheck Hook (hooks/useHealthCheck.js)
```javascript
import { useState, useEffect, useCallback } from 'react';
import { healthService } from '../services/healthService';

export const useHealthCheck = (interval = 30000) => {
  const [health, setHealth] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const checkHealth = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const healthData = await healthService.checkHealth();
      setHealth(healthData);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    checkHealth();

    if (interval > 0) {
      const intervalId = setInterval(checkHealth, interval);
      return () => clearInterval(intervalId);
    }
  }, [checkHealth, interval]);

  return {
    health,
    loading,
    error,
    checkHealth,
  };
};
```
