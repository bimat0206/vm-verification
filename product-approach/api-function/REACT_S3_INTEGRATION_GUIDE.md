# React/Next.js S3 Integration Guide
## Vending Machine Verification System - Image Browser & Management

This guide provides comprehensive instructions for integrating the S3 image management API with React/Next.js frontend applications based on the actual API implementations.

## 🏗️ API Architecture Overview

### Available Endpoints

Based on the current API implementation, there are three main endpoints for S3 operations:

1. **Browser API** (`/api/images/browser`) - Browse S3 bucket contents
2. **Upload API** (`/api/images/upload`) - Upload files to S3 buckets
3. **View API** (`/api/images/{key}/view`) - Get presigned URLs for image viewing

### Supported Bucket Types
- `reference` - Reference images bucket
- `checking` - Checking images bucket

## 📋 API Specifications

### 1. Browser API - S3 Bucket Navigation

**Endpoint**: `GET /api/images/browser/{path}?bucketType={type}`

**Purpose**: Browse folder structure and list files in S3 buckets

**Request Parameters**:
- `bucketType` (query): `reference` | `checking` (default: `reference`)
- `path` (path parameter): Optional folder path to browse

**Response Structure**:
```typescript
interface BrowserResponse {
  currentPath: string;        // Current folder path
  parentPath?: string;        // Parent folder path (if not root)
  items: BrowserItem[];       // Array of files and folders
  totalItems: number;         // Total count of items
}

interface BrowserItem {
  name: string;              // Display name
  path: string;              // Full S3 key/path
  type: "folder" | "image" | "file";  // Item type
  size?: number;             // File size in bytes (files only)
  lastModified?: string;     // ISO timestamp (files only)
  contentType?: string;      // MIME type (files only)
}
```

**Design Considerations**:
- Supports hierarchical folder navigation with breadcrumbs
- Distinguishes between images and other file types
- Provides file metadata for display
- Handles URL encoding for special characters in paths

### 2. Upload API - File Upload to S3

**Endpoint**: `POST /api/images/upload?bucketType={type}&fileName={name}&path={folder}`

**Purpose**: Upload files to S3 buckets with organization

**Request Parameters**:
- `bucketType` (query): `reference` | `checking` (default: `reference`)
- `fileName` (query): Name of the file being uploaded
- `path` (query): Optional folder path for organization

**Request Body**: Raw file content (binary or text)

**Response Structure**:
```typescript
interface UploadResponse {
  success: boolean;
  message: string;
  files?: UploadedFile[];     // Array of uploaded files
  errors?: string[];          // Array of error messages
}

interface UploadedFile {
  originalName: string;       // Original filename
  key: string;               // S3 key where file was stored
  size: number;              // File size in bytes
  contentType: string;       // Detected MIME type
  bucket: string;            // Target bucket name
  url?: string;              // Optional access URL
}
```

**File Type Support**:
- **Images**: `.jpg`, `.jpeg`, `.png`, `.gif`, `.bmp`, `.webp`, `.tiff`, `.tif`
- **Documents**: `.pdf`, `.txt`, `.json`, `.csv`, `.xml`
- **Size Limit**: 10MB maximum per file

**Design Considerations**:
- Automatic content type detection
- File type validation on server side
- Organized storage with custom folder paths
- Error handling for oversized or invalid files

### 3. View API - Image Access URLs

**Endpoint**: `GET /api/images/{key}/view?bucketType={type}`

**Purpose**: Generate presigned URLs for secure image access

**Request Parameters**:
- `key` (path parameter): S3 object key (URL encoded)
- `bucketType` (query): `reference` | `checking` (default: `reference`)

**Response Structure**:
```typescript
interface ViewResponse {
  presignedUrl: string;       // Temporary access URL (1 hour expiration)
}
```

**Design Considerations**:
- Presigned URLs expire after 1 hour for security
- Handles URL encoding/decoding for special characters
- Validates object existence before generating URL
- Supports both bucket types

## � React Integration Patterns

### Environment Setup

```typescript
// .env.local (Next.js)
NEXT_PUBLIC_API_BASE_URL=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api
NEXT_PUBLIC_API_KEY=WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH

// .env (Create React App)
REACT_APP_API_BASE_URL=https://hpux2uegnd.execute-api.us-east-1.amazonaws.com/v1/api
REACT_APP_API_KEY=WgGMX8xBxV9Ci3HtHJt6e7WF6VcIPojiahSXHUjH
```

### TypeScript Interfaces

```typescript
// types/s3.ts
export type BucketType = 'reference' | 'checking';

export interface BrowserResponse {
  currentPath: string;
  parentPath?: string;
  items: BrowserItem[];
  totalItems: number;
}

export interface BrowserItem {
  name: string;
  path: string;
  type: 'folder' | 'image' | 'file';
  size?: number;
  lastModified?: string;
  contentType?: string;
}

export interface UploadResponse {
  success: boolean;
  message: string;
  files?: UploadedFile[];
  errors?: string[];
}

export interface UploadedFile {
  originalName: string;
  key: string;
  size: number;
  contentType: string;
  bucket: string;
  url?: string;
}

export interface ViewResponse {
  presignedUrl: string;
}
```

### API Service Implementation

```typescript
// services/s3Service.ts
class S3Service {
  private baseURL: string;
  private apiKey: string;

  constructor() {
    this.baseURL = process.env.NEXT_PUBLIC_API_BASE_URL || process.env.REACT_APP_API_BASE_URL || '';
    this.apiKey = process.env.NEXT_PUBLIC_API_KEY || process.env.REACT_APP_API_KEY || '';
  }

  private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const url = `${this.baseURL}${endpoint}`;
    const response = await fetch(url, {
      ...options,
      headers: {
        'X-Api-Key': this.apiKey,
        ...options.headers,
      },
    });

    if (!response.ok) {
      throw new Error(`API Error: ${response.status} ${response.statusText}`);
    }

    return response.json();
  }

  // Browse S3 bucket contents
  async browseFolder(bucketType: BucketType, path: string = ''): Promise<BrowserResponse> {
    const endpoint = path
      ? `/images/browser/${encodeURIComponent(path)}?bucketType=${bucketType}`
      : `/images/browser?bucketType=${bucketType}`;

    return this.request<BrowserResponse>(endpoint, { method: 'GET' });
  }

  // Get presigned URL for image viewing
  async getImageViewUrl(imageKey: string, bucketType: BucketType): Promise<string> {
    const response = await this.request<ViewResponse>(
      `/images/${encodeURIComponent(imageKey)}/view?bucketType=${bucketType}`,
      { method: 'GET' }
    );
    return response.presignedUrl;
  }

  // Upload file to S3 bucket
  async uploadFile(
    file: File,
    bucketType: BucketType,
    path: string = ''
  ): Promise<UploadResponse> {
    const endpoint = `/images/upload?bucketType=${bucketType}&fileName=${encodeURIComponent(file.name)}&path=${encodeURIComponent(path)}`;

    return this.request<UploadResponse>(endpoint, {
      method: 'POST',
      headers: {
        'Content-Type': file.type || 'application/octet-stream',
      },
      body: file,
    });
  }
}

export const s3Service = new S3Service();
```

## 🎨 React Component Design Patterns

### 1. S3 Browser Hook

```typescript
// hooks/useS3Browser.ts
import { useState, useEffect, useCallback } from 'react';
import { s3Service } from '../services/s3Service';
import { BrowserResponse, BucketType } from '../types/s3';

export const useS3Browser = (bucketType: BucketType, initialPath: string = '') => {
  const [data, setData] = useState<BrowserResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [currentPath, setCurrentPath] = useState(initialPath);

  const loadFolder = useCallback(async (path: string = '') => {
    setLoading(true);
    setError(null);

    try {
      const response = await s3Service.browseFolder(bucketType, path);
      setData(response);
      setCurrentPath(path);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load folder');
    } finally {
      setLoading(false);
    }
  }, [bucketType]);

  useEffect(() => {
    loadFolder(currentPath);
  }, [loadFolder, currentPath]);

  const navigateToPath = useCallback((path: string) => {
    setCurrentPath(path);
  }, []);

  const navigateUp = useCallback(() => {
    if (data?.parentPath !== undefined) {
      setCurrentPath(data.parentPath);
    }
  }, [data?.parentPath]);

  const refresh = useCallback(() => {
    loadFolder(currentPath);
  }, [loadFolder, currentPath]);

  return {
    data,
    loading,
    error,
    currentPath,
    navigateToPath,
    navigateUp,
    refresh,
  };
};
```

### 2. S3 Browser Component

```tsx
// components/S3Browser.tsx
import React from 'react';
import { useS3Browser } from '../hooks/useS3Browser';
import { BucketType, BrowserItem } from '../types/s3';

interface S3BrowserProps {
  bucketType: BucketType;
  onImageSelect?: (item: BrowserItem) => void;
  className?: string;
}

export const S3Browser: React.FC<S3BrowserProps> = ({
  bucketType,
  onImageSelect,
  className = ''
}) => {
  const { data, loading, error, currentPath, navigateToPath, navigateUp, refresh } =
    useS3Browser(bucketType);

  const handleItemClick = (item: BrowserItem) => {
    if (item.type === 'folder') {
      navigateToPath(item.path);
    } else if (item.type === 'image' && onImageSelect) {
      onImageSelect(item);
    }
  };

  if (loading) {
    return (
      <div className={`flex items-center justify-center p-8 ${className}`}>
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
        <span className="ml-2 text-gray-600">Loading...</span>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`bg-red-50 border border-red-200 rounded-md p-4 ${className}`}>
        <p className="text-red-800">{error}</p>
        <button
          onClick={refresh}
          className="mt-2 px-3 py-1 bg-red-600 text-white rounded hover:bg-red-700"
        >
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className={`bg-white rounded-lg shadow-md p-6 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-800">
          {bucketType.charAt(0).toUpperCase() + bucketType.slice(1)} Bucket
        </h3>
        <button
          onClick={refresh}
          className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
        >
          Refresh
        </button>
      </div>

      {/* Breadcrumb */}
      <div className="flex items-center mb-4 text-sm text-gray-600">
        <span className="font-medium">{bucketType}/</span>
        {currentPath && <span>{currentPath}</span>}
      </div>

      {/* Navigation */}
      {data?.parentPath !== undefined && (
        <button
          onClick={navigateUp}
          className="flex items-center mb-4 text-blue-600 hover:text-blue-800 transition-colors"
        >
          <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back to parent folder
        </button>
      )}

      {/* Items Grid */}
      {data && data.items.length > 0 ? (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {data.items.map((item) => (
            <div
              key={item.path}
              onClick={() => handleItemClick(item)}
              className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 cursor-pointer transition-colors"
            >
              <div className="flex flex-col items-center">
                {/* Icon */}
                <div className="mb-2">
                  {item.type === 'folder' ? (
                    <svg className="w-8 h-8 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
                    </svg>
                  ) : (
                    <svg className="w-8 h-8 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
                    </svg>
                  )}
                </div>

                {/* Name */}
                <p className="text-sm font-medium text-gray-800 text-center truncate w-full">
                  {item.name}
                </p>

                {/* Metadata */}
                {item.type !== 'folder' && (
                  <div className="mt-1 text-xs text-gray-500 text-center">
                    {item.size && <div>{(item.size / 1024).toFixed(1)} KB</div>}
                    {item.lastModified && (
                      <div>{new Date(item.lastModified).toLocaleDateString()}</div>
                    )}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="text-center py-8 text-gray-500">
          <svg className="w-12 h-12 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
          </svg>
          <p>This folder is empty</p>
        </div>
      )}
    </div>
  );
};
```

### 3. Image Preview Component

```tsx
// components/ImagePreview.tsx
import React, { useState, useEffect } from 'react';
import { s3Service } from '../services/s3Service';
import { BrowserItem, BucketType } from '../types/s3';

interface ImagePreviewProps {
  item: BrowserItem;
  bucketType: BucketType;
  onClose: () => void;
  isOpen: boolean;
}

export const ImagePreview: React.FC<ImagePreviewProps> = ({
  item,
  bucketType,
  onClose,
  isOpen
}) => {
  const [imageUrl, setImageUrl] = useState<string>('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (isOpen && item.type === 'image') {
      loadImage();
    }
  }, [isOpen, item]);

  const loadImage = async () => {
    setLoading(true);
    setError(null);

    try {
      const url = await s3Service.getImageViewUrl(item.path, bucketType);
      setImageUrl(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load image');
    } finally {
      setLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-75 flex items-center justify-center z-50">
      <div className="relative max-w-4xl max-h-full p-4">
        {/* Close Button */}
        <button
          onClick={onClose}
          className="absolute top-4 right-4 text-white hover:text-gray-300 z-10"
        >
          <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>

        {/* Image Content */}
        <div className="bg-white rounded-lg p-6">
          <h3 className="text-lg font-semibold mb-4">{item.name}</h3>

          {loading && (
            <div className="flex items-center justify-center p-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
              <span className="ml-2">Loading image...</span>
            </div>
          )}

          {error && (
            <div className="bg-red-50 border border-red-200 rounded-md p-4">
              <p className="text-red-800">{error}</p>
            </div>
          )}

          {imageUrl && !loading && !error && (
            <div className="text-center">
              <img
                src={imageUrl}
                alt={item.name}
                className="max-w-full max-h-96 object-contain rounded"
              />

              {/* Metadata */}
              <div className="mt-4 text-sm text-gray-600 space-y-1">
                {item.size && <p>Size: {(item.size / 1024).toFixed(1)} KB</p>}
                {item.lastModified && (
                  <p>Modified: {new Date(item.lastModified).toLocaleString()}</p>
                )}
                {item.contentType && <p>Type: {item.contentType}</p>}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
```

### 4. File Upload Component

```tsx
// components/FileUpload.tsx
import React, { useState, useCallback } from 'react';
import { s3Service } from '../services/s3Service';
import { BucketType } from '../types/s3';

interface FileUploadProps {
  bucketType: BucketType;
  uploadPath?: string;
  onUploadComplete?: () => void;
  className?: string;
}

export const FileUpload: React.FC<FileUploadProps> = ({
  bucketType,
  uploadPath = '',
  onUploadComplete,
  className = ''
}) => {
  const [uploading, setUploading] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [uploadStatus, setUploadStatus] = useState<string>('');

  const handleFileUpload = async (files: FileList) => {
    if (files.length === 0) return;

    setUploading(true);
    setUploadStatus('');

    try {
      const uploadPromises = Array.from(files).map(file =>
        s3Service.uploadFile(file, bucketType, uploadPath)
      );

      const results = await Promise.all(uploadPromises);
      const successCount = results.filter(r => r.success).length;

      setUploadStatus(`Successfully uploaded ${successCount} of ${files.length} files`);

      if (onUploadComplete) {
        onUploadComplete();
      }
    } catch (error) {
      setUploadStatus(`Upload failed: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setUploading(false);
    }
  };

  const handleDrop = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);

    if (e.dataTransfer.files) {
      handleFileUpload(e.dataTransfer.files);
    }
  }, []);

  const handleDragOver = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: React.DragEvent) => {
    e.preventDefault();
    setDragOver(false);
  }, []);

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files) {
      handleFileUpload(e.target.files);
    }
  };

  return (
    <div className={`bg-white rounded-lg shadow-md p-6 ${className}`}>
      <h3 className="text-lg font-semibold mb-4">Upload Files</h3>

      <div
        onDrop={handleDrop}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
          dragOver
            ? 'border-blue-500 bg-blue-50'
            : 'border-gray-300 hover:border-gray-400'
        }`}
      >
        {uploading ? (
          <div className="flex items-center justify-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
            <span className="ml-2">Uploading...</span>
          </div>
        ) : (
          <>
            <svg className="w-12 h-12 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
            </svg>
            <p className="text-gray-600 mb-2">Drag and drop files here, or</p>
            <label className="inline-block px-4 py-2 bg-blue-500 text-white rounded cursor-pointer hover:bg-blue-600">
              Browse Files
              <input
                type="file"
                multiple
                onChange={handleFileSelect}
                className="hidden"
                accept="image/*,.pdf,.txt,.json,.csv,.xml"
              />
            </label>
            <p className="text-xs text-gray-500 mt-2">
              Supported: Images, PDF, TXT, JSON, CSV, XML (Max 10MB each)
            </p>
          </>
        )}
      </div>

      {uploadStatus && (
        <div className={`mt-4 p-3 rounded ${
          uploadStatus.includes('Successfully')
            ? 'bg-green-50 text-green-800'
            : 'bg-red-50 text-red-800'
        }`}>
          {uploadStatus}
        </div>
      )}
    </div>
  );
};
```

## 🔧 Usage Examples

### Complete S3 Browser Application

```tsx
// App.tsx
import React, { useState } from 'react';
import { S3Browser } from './components/S3Browser';
import { ImagePreview } from './components/ImagePreview';
import { FileUpload } from './components/FileUpload';
import { BucketType, BrowserItem } from './types/s3';

export const S3BrowserApp: React.FC = () => {
  const [selectedBucket, setSelectedBucket] = useState<BucketType>('reference');
  const [selectedImage, setSelectedImage] = useState<BrowserItem | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleImageSelect = (item: BrowserItem) => {
    setSelectedImage(item);
  };

  const handleUploadComplete = () => {
    // Refresh the browser to show new files
    setRefreshKey(prev => prev + 1);
  };

  return (
    <div className="min-h-screen bg-gray-100 p-4">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">S3 Image Browser</h1>

        {/* Bucket Selector */}
        <div className="mb-6">
          <div className="flex space-x-4">
            <button
              onClick={() => setSelectedBucket('reference')}
              className={`px-4 py-2 rounded ${
                selectedBucket === 'reference'
                  ? 'bg-blue-500 text-white'
                  : 'bg-white text-gray-700 border'
              }`}
            >
              Reference Images
            </button>
            <button
              onClick={() => setSelectedBucket('checking')}
              className={`px-4 py-2 rounded ${
                selectedBucket === 'checking'
                  ? 'bg-blue-500 text-white'
                  : 'bg-white text-gray-700 border'
              }`}
            >
              Checking Images
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* File Upload */}
          <div className="lg:col-span-1">
            <FileUpload
              bucketType={selectedBucket}
              onUploadComplete={handleUploadComplete}
            />
          </div>

          {/* S3 Browser */}
          <div className="lg:col-span-2">
            <S3Browser
              key={`${selectedBucket}-${refreshKey}`}
              bucketType={selectedBucket}
              onImageSelect={handleImageSelect}
            />
          </div>
        </div>

        {/* Image Preview Modal */}
        {selectedImage && (
          <ImagePreview
            item={selectedImage}
            bucketType={selectedBucket}
            isOpen={!!selectedImage}
            onClose={() => setSelectedImage(null)}
          />
        )}
      </div>
    </div>
  );
};
```

### Next.js Page Example

```tsx
// pages/s3-browser.tsx (Next.js)
import { NextPage } from 'next';
import { S3BrowserApp } from '../components/S3BrowserApp';

const S3BrowserPage: NextPage = () => {
  return <S3BrowserApp />;
};

export default S3BrowserPage;
```

## 📚 Best Practices

### 1. Error Handling
- Always implement proper error boundaries
- Provide user-friendly error messages
- Include retry mechanisms for failed operations
- Log errors for debugging purposes

### 2. Performance Optimization
- Implement lazy loading for large folders
- Use React.memo for expensive components
- Debounce search and filter operations
- Cache API responses when appropriate

### 3. User Experience
- Show loading states for all async operations
- Provide visual feedback for drag and drop
- Implement keyboard navigation
- Support mobile touch gestures

### 4. Security Considerations
- Validate file types on both client and server
- Implement file size limits
- Use presigned URLs for secure access
- Never expose API keys in client-side code

### 5. Accessibility
- Include proper ARIA labels
- Ensure keyboard navigation works
- Provide alt text for images
- Support screen readers

## 🎨 Tailwind CSS Classes Reference

### Layout Classes
```css
/* Containers */
.bg-white .rounded-lg .shadow-md .p-6
.max-w-6xl .mx-auto
.grid .grid-cols-1 .lg:grid-cols-3 .gap-6

/* Flexbox */
.flex .items-center .justify-between
.flex-col .space-x-4 .space-y-1

/* Responsive */
.md:grid-cols-3 .lg:grid-cols-4
.sm:text-sm .md:text-base
```

### Interactive Elements
```css
/* Buttons */
.px-4 .py-2 .bg-blue-500 .text-white .rounded
.hover:bg-blue-600 .transition-colors
.disabled:opacity-50 .disabled:cursor-not-allowed

/* Form Elements */
.border .border-gray-300 .rounded-md .p-2
.focus:ring-2 .focus:ring-blue-500 .focus:border-blue-500

/* States */
.hover:bg-gray-50 .cursor-pointer
.animate-spin .animate-pulse
```

### Status Colors
```css
/* Success */
.bg-green-50 .text-green-800 .border-green-200

/* Error */
.bg-red-50 .text-red-800 .border-red-200

/* Warning */
.bg-yellow-50 .text-yellow-800 .border-yellow-200

/* Info */
.bg-blue-50 .text-blue-800 .border-blue-200
```

## 🚨 Common Issues & Solutions

### CORS Errors
- Ensure API Gateway has proper CORS configuration
- Check that all required headers are included in requests
- Verify the API key is being sent correctly

### File Upload Failures
- Validate file size before upload (10MB limit)
- Check file type against allowed extensions
- Handle network timeouts with retry logic

### Image Loading Issues
- Verify presigned URLs are not expired (1 hour limit)
- Check that the image key exists in the bucket
- Handle URL encoding for special characters

### Performance Problems
- Implement pagination for large folders
- Use React.memo to prevent unnecessary re-renders
- Consider virtual scrolling for very large lists

This guide provides a complete foundation for integrating S3 bucket browsing and image management into React/Next.js applications using the vending machine verification system APIs.

interface S3BrowserProps {
  bucketType: BucketType;
  onImageSelect?: (item: S3Item) => void;
}

export const S3Browser: React.FC<S3BrowserProps> = ({ bucketType, onImageSelect }) => {
  const [items, setItems] = useState<S3Item[]>([]);
  const [currentPath, setCurrentPath] = useState<string>('');
  const [parentPath, setParentPath] = useState<string | undefined>();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const loadFolder = async (path: string = '') => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await s3Service.browseFolder(bucketType, path);
      setItems(response.items);
      setCurrentPath(response.currentPath);
      setParentPath(response.parentPath);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load folder');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFolder();
  }, [bucketType]);

  const handleItemClick = (item: S3Item) => {
    if (item.type === 'folder') {
      loadFolder(item.key);
    } else if (item.type === 'file' && onImageSelect) {
      onImageSelect(item);
    }
  };

  const navigateUp = () => {
    if (parentPath !== undefined) {
      loadFolder(parentPath);
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-md p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold text-gray-800">
          {bucketType.charAt(0).toUpperCase() + bucketType.slice(1)} Bucket
        </h3>
        <button
          onClick={() => loadFolder(currentPath)}
          className="px-3 py-1 bg-blue-500 text-white rounded hover:bg-blue-600 transition-colors"
          disabled={loading}
        >
          Refresh
        </button>
      </div>

      {/* Breadcrumb */}
      <div className="flex items-center mb-4 text-sm text-gray-600">
        <span className="font-medium">{bucketType}/</span>
        {currentPath && <span>{currentPath}</span>}
      </div>

      {/* Navigation */}
      {parentPath !== undefined && (
        <button
          onClick={navigateUp}
          className="flex items-center mb-4 text-blue-600 hover:text-blue-800 transition-colors"
        >
          <svg className="w-4 h-4 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
          Back to parent folder
        </button>
      )}

      {/* Loading State */}
      {loading && (
        <div className="flex items-center justify-center py-8">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500"></div>
          <span className="ml-2 text-gray-600">Loading...</span>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4 mb-4">
          <p className="text-red-800">{error}</p>
        </div>
      )}

      {/* Items Grid */}
      {!loading && !error && (
        <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
          {items.map((item) => (
            <div
              key={item.key}
              onClick={() => handleItemClick(item)}
              className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50 cursor-pointer transition-colors"
            >
              <div className="flex flex-col items-center">
                {/* Icon */}
                <div className="mb-2">
                  {item.type === 'folder' ? (
                    <svg className="w-8 h-8 text-blue-500" fill="currentColor" viewBox="0 0 20 20">
                      <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
                    </svg>
                  ) : (
                    <svg className="w-8 h-8 text-green-500" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4zm12 12H4l4-8 3 6 2-4 3 6z" clipRule="evenodd" />
                    </svg>
                  )}
                </div>
                
                {/* Name */}
                <p className="text-sm font-medium text-gray-800 text-center truncate w-full">
                  {item.name}
                </p>
                
                {/* Metadata */}
                {item.type === 'file' && (
                  <div className="mt-1 text-xs text-gray-500 text-center">
                    {item.size && <div>{(item.size / 1024).toFixed(1)} KB</div>}
                    {item.lastModified && (
                      <div>{new Date(item.lastModified).toLocaleDateString()}</div>
                    )}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Empty State */}
      {!loading && !error && items.length === 0 && (
        <div className="text-center py-8 text-gray-500">
          <svg className="w-12 h-12 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 21h10a2 2 0 002-2V9.414a1 1 0 00-.293-.707l-5.414-5.414A1 1 0 0012.586 3H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
          </svg>
          <p>This folder is empty</p>
        </div>
      )}
    </div>
  );
};
```
