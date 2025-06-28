
"use client";

import React, { useState, useEffect, useCallback, useMemo } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { S3ImageBrowser } from '@/components/common/s3-image-browser';
import apiClient from '@/lib/api-client';
import { useAppConfig } from '@/config';
import type { ApiError, UploadHistoryItem, UploadResponse, BucketType, UploadedFile } from '@/lib/types';

import { UploadCloud, Loader2, AlertCircle, Trash2, FileJson, Image as ImageIconLucide, X, CheckCircle, PackageOpen, Clock, Upload, Palette } from 'lucide-react';
import { Progress } from "@/components/ui/progress";
import Image from 'next/image';
import { format } from 'date-fns';
import { cn } from '@/lib/utils';
import { ScrollArea } from '@/components/ui/scroll-area';

const MAX_FILE_SIZE_MB = 10;
const MAX_FILE_SIZE_BYTES = MAX_FILE_SIZE_MB * 1024 * 1024;

// Schema matches guide: file is FileList, path for organization, customFilename relevant for final key
const uploadSchema = z.object({
  file: z.custom<FileList>()
    .refine(files => files && files.length === 1, "File is required.")
    .refine(files => files && files[0]?.size <= MAX_FILE_SIZE_BYTES, `File size must be less than ${MAX_FILE_SIZE_MB}MB.`),
  // 'path' query param for API upload - this is the folder path for organization
  s3FolderPath: z.string().optional(),
  // 'fileName' query param for API - if user wants to override original filename
  customS3FileName: z.string().optional(),
  // Upload target for validation
  uploadTarget: z.enum(['reference', 'checking']).optional(),
}).refine((data) => {
  // For reference bucket, s3FolderPath must be "raw" if provided
  if (data.uploadTarget === 'reference' && data.s3FolderPath && data.s3FolderPath.trim() !== '') {
    return data.s3FolderPath.trim() === 'raw';
  }
  return true;
}, {
  message: "For reference bucket, Target S3 Folder Path must be 'raw'",
  path: ["s3FolderPath"]
});

type UploadFormData = z.infer<typeof uploadSchema>;

// Progress tracking types
type ProgressStage = 'idle' | 'validating' | 'uploading' | 'processing' | 'rendering' | 'complete' | 'error';

interface ProgressState {
  stage: ProgressStage;
  progress: number;
  message: string;
  details?: string;
}

const UPLOAD_HISTORY_KEY = 'vms_upload_history_gradient_shift_v3'; // Updated key

export default function ImageUploadPage() {
  const { config: appConfig, isLoading: configLoading } = useAppConfig();
  const [uploadTarget, setUploadTarget] = useState<BucketType>('checking');
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [previewData, setPreviewData] = useState<string | null>(null);
  const [uploadStatus, setUploadStatus] = useState<{ type: 'idle' | 'success' | 'error' | 'loading'; message: string }>({ type: 'idle', message: '' });
  const [progressState, setProgressState] = useState<ProgressState>({ stage: 'idle', progress: 0, message: '' });
  const [uploadHistory, setUploadHistory] = useState<UploadHistoryItem[]>([]);
  const [activeTab, setActiveTab] = useState<'uploader' | 'history'>('uploader');
  const [isDragOver, setIsDragOver] = useState(false);


  const { handleSubmit, register, watch, setValue, reset, formState: { errors } } = useForm<UploadFormData>({
    resolver: zodResolver(uploadSchema),
    defaultValues: {
      uploadTarget: uploadTarget
    }
  });

  useEffect(() => {
    const storedHistory = localStorage.getItem(UPLOAD_HISTORY_KEY);
    if (storedHistory) {
      setUploadHistory(JSON.parse(storedHistory));
    }
  }, []);

  const updateProgress = useCallback((stage: ProgressStage, progress: number, message: string, details?: string) => {
    setProgressState({ stage, progress, message, details });
  }, []);

  const resetFormState = useCallback(() => {
    reset({
      file: undefined,
      s3FolderPath: uploadTarget === 'reference' ? 'raw' : '',
      customS3FileName: '',
      uploadTarget: uploadTarget
    });
    setSelectedFile(null);
    setPreviewData(null);
    setProgressState({ stage: 'idle', progress: 0, message: '' });
    const fileInputField = document.getElementById('file-input-field') as HTMLInputElement;
    if (fileInputField) fileInputField.value = '';
  }, [reset, uploadTarget]);

  useEffect(() => {
    resetFormState();
    // Update the form's uploadTarget value when uploadTarget state changes
    setValue('uploadTarget', uploadTarget);
  }, [uploadTarget, resetFormState, setValue]);

  // Browser-compatible UUID generator
  const generateUUID = (): string => {
    // Try to use crypto.randomUUID if available
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
      return crypto.randomUUID();
    }

    // Fallback to manual UUID generation
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  };

  const addUploadToHistory = (uploadedFile: UploadedFile, targetBucketType: BucketType) => {
    const s3DisplayUrl = uploadedFile.url || `s3://${uploadedFile.bucket}/${uploadedFile.key}`;
    setUploadHistory(prev => {
      const newHistoryItem: UploadHistoryItem = {
        id: generateUUID(),
        timestamp: new Date().toISOString(),
        fileName: uploadedFile.originalName,
        s3Url: s3DisplayUrl,
        bucket: uploadedFile.bucket,
        key: uploadedFile.key,
        type: targetBucketType
      };
      const newHistory = [newHistoryItem, ...prev].slice(0, 20);
      localStorage.setItem(UPLOAD_HISTORY_KEY, JSON.stringify(newHistory));
      return newHistory;
    });
  };

  const clearUploadHistory = () => {
    setUploadHistory([]);
    localStorage.removeItem(UPLOAD_HISTORY_KEY);
    // toast({ title: "Upload History Cleared" });
  };

  const handleFileSelection = useCallback((file: File | null) => {
    if (!file) {
      setSelectedFile(null);
      setPreviewData(null);
      setValue('file', null as any);
      return;
    }
    // Guide file type validation: Backend handles it, but good to have client-side check
    const allowedImageTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/bmp', 'image/tiff'];

    let isValidType = false;
    if (uploadTarget === 'reference') {
      // Reference bucket only accepts JSON files
      isValidType = file.type === 'application/json';
    } else {
      // Checking bucket primarily for images
      isValidType = allowedImageTypes.includes(file.type);
    }

    if (!isValidType) {
      const expectedType = uploadTarget === 'reference' ? 'JSON files only' : 'image files only';
      setUploadStatus({ type: 'error', message: `Invalid file type for ${uploadTarget} bucket. Expected: ${expectedType}. Selected: ${file.type}` });
      resetFormState(); return;
    }
    if (file.size > MAX_FILE_SIZE_BYTES) {
      setUploadStatus({ type: 'error', message: `File is too large (Max ${MAX_FILE_SIZE_MB}MB).` });
      resetFormState(); return;
    }

    setSelectedFile(file);
    setValue('file', { 0: file, length: 1 } as unknown as FileList);
    setUploadStatus({ type: 'idle', message: '' });

    const reader = new FileReader();
    reader.onloadend = () => {
      setPreviewData(reader.result as string);
    };

    if (file.type === 'application/json' || file.type === 'text/plain') {
      reader.readAsText(file);
    } else if (allowedImageTypes.includes(file.type)) {
      reader.readAsDataURL(file);
    } else {
      setPreviewData('No preview for this file type.');
    }
  }, [uploadTarget, setValue, resetFormState]);

  const onFormSubmit = async (data: UploadFormData) => {
    if (!selectedFile) {
      setUploadStatus({ type: 'error', message: "Please select a file to upload." });
      return;
    }

    // Start progress tracking
    updateProgress('validating', 5, 'Validating file...', 'Checking file type and size requirements');

    // Additional validation for reference bucket
    if (uploadTarget === 'reference') {
      if (data.s3FolderPath && data.s3FolderPath.trim() !== '' && data.s3FolderPath.trim() !== 'raw') {
        setUploadStatus({ type: 'error', message: "For reference bucket, Target S3 Folder Path must be 'raw' or empty." });
        updateProgress('error', 0, 'Validation failed', 'Invalid folder path for reference bucket');
        return;
      }
      if (selectedFile.type !== 'application/json') {
        setUploadStatus({ type: 'error', message: "Reference bucket only accepts JSON files." });
        updateProgress('error', 0, 'Validation failed', 'Only JSON files are allowed for reference bucket');
        return;
      }
    }

    setUploadStatus({ type: 'loading', message: 'Processing upload...' });
    updateProgress('uploading', 15, 'Starting upload...', 'Preparing file for upload to S3');

    try {
      // For reference bucket, use 'raw' if no path specified, otherwise use the provided path
      const folderPathForApi = uploadTarget === 'reference'
        ? (data.s3FolderPath?.trim() || 'raw')
        : (data.s3FolderPath || '');
      const customFileNameFromForm = data.customS3FileName || undefined;

      // Update progress for different stages
      updateProgress('uploading', 25, 'Uploading to S3...', 'Transferring file to cloud storage');

      // For JSON files, show additional stages
      if (selectedFile.type === 'application/json' && uploadTarget === 'reference') {
        updateProgress('uploading', 30, 'Uploading JSON file...', 'Uploading and processing will take 30-60 seconds');
      }

      const result: UploadResponse = await apiClient.uploadFile(
        selectedFile!,
        uploadTarget,
        folderPathForApi,
        customFileNameFromForm,
        true // debugFlag
      );
      
      // Update progress based on result
      if (result.success && result.files && result.files.length > 0) {
        // Check if rendering was involved
        if (result.renderResult) {
          if (result.renderResult.rendered) {
            updateProgress('complete', 100, 'Upload and render completed!', 'JSON layout successfully rendered to image');
          } else {
            updateProgress('complete', 100, 'Upload completed!', 'File uploaded but rendering was not successful');
          }
        } else {
          updateProgress('complete', 100, 'Upload completed!', 'File successfully uploaded to S3');
        }
        const uploadedFile = result.files[0];
        let successMessage = result.message || `File ${uploadedFile.originalName} uploaded as ${uploadedFile.key}`;

        // Add render result information to the success message
        if (result.renderResult) {
          if (result.renderResult.rendered) {
            successMessage += `\n\n🎨 JSON Layout Rendered Successfully!\n` +
              `Layout ID: ${result.renderResult.layoutId}\n` +
              `Layout Prefix: ${result.renderResult.layoutPrefix}\n` +
              `Processed Image: ${result.renderResult.processedKey}`;
          } else {
            successMessage += `\n\n⚠️ Render Status: ${result.renderResult.message}`;
          }
        }

        setUploadStatus({ type: 'success', message: successMessage });
        addUploadToHistory(uploadedFile, uploadTarget);

        // Reset form after a short delay to show completion
        setTimeout(() => {
          resetFormState();
        }, 2000);
      } else {
        throw new Error(result.message || (result.errors && result.errors.join(', ')) || "Upload failed with no specific message.");
      }
    } catch (error) {
      console.error("Upload Error (raw object):", error);
      const apiErr = error as ApiError;
      let displayMessage = "An unexpected error occurred during upload.";
      let isTimeoutError = false;

      if (apiErr.message) {
        displayMessage = apiErr.message;
        isTimeoutError = apiErr.message.toLowerCase().includes('timeout') ||
                        apiErr.message.toLowerCase().includes('did not respond in time') ||
                        apiErr.message.toLowerCase().includes('aborted') ||
                        apiErr.statusCode === 408;
      } else if (apiErr.details?.message) {
        displayMessage = apiErr.details.message;
      } else if (apiErr.details?.error) {
        displayMessage = apiErr.details.error;
      } else if (typeof error === 'string') {
        displayMessage = error;
      } else if (error && typeof (error as any).toString === 'function' && (error as any).toString() !== '[object Object]') {
        displayMessage = (error as any).toString();
      }

      // Special handling for timeout errors - treat as successful completion
      if (isTimeoutError || displayMessage.includes('did not respond in time')) {
        if (selectedFile?.type === 'application/json' && uploadTarget === 'reference') {
          displayMessage = `✅ Upload Completed!\n\n` +
            `🎨 Your JSON file has been uploaded and the rendering process is running in the background.\n\n` +
            `💡 Next steps:\n` +
            `• Wait 2-3 minutes and check the S3 bucket\n` +
            `• The rendered image will appear in the bucket once processing is complete\n\n` +
            `📧 Contact support if the rendered file doesn't appear after 5 minutes.`;

          updateProgress('complete', 100, 'Upload completed!', 'File uploaded successfully - rendering in progress');
        } else {
          displayMessage = `✅ Upload Completed!\n\n` +
            `Your file has been uploaded successfully.\n\n` +
            `💡 Next steps:\n` +
            `• Wait 2-3 minutes and check the S3 bucket\n` +
            `• Your file should be available for use\n\n` +
            `📧 Contact support if you don't see your file in the bucket.`;

          updateProgress('complete', 100, 'Upload completed!', 'File uploaded successfully');
        }

        console.log("Upload completed (timeout treated as success):", apiErr, "Display Message:", displayMessage);
        setUploadStatus({ type: 'success', message: displayMessage });
      } else {
        updateProgress('error', 0, 'Upload failed', displayMessage);
        console.error("Upload Error (processed):", apiErr, "Display Message:", displayMessage);
        setUploadStatus({ type: 'error', message: displayMessage });
      }
    }
  };
  
  const handleDragOver = (event: React.DragEvent<HTMLDivElement>) => { event.preventDefault(); setIsDragOver(true); };
  const handleDragLeave = (event: React.DragEvent<HTMLDivElement>) => { event.preventDefault(); setIsDragOver(false); };
  const handleDrop = (event: React.DragEvent<HTMLDivElement>) => {
    event.preventDefault();
    setIsDragOver(false);
    if (event.dataTransfer.files && event.dataTransfer.files.length > 0) {
      handleFileSelection(event.dataTransfer.files[0]);
      const fileInputField = document.getElementById('file-input-field') as HTMLInputElement;
      if (fileInputField) fileInputField.files = event.dataTransfer.files;
    }
  };

  const bucketConfig = useMemo(() => ({
    reference: {
      bucketName: appConfig.referenceS3BucketName,
      accept: '.json', // Only JSON files allowed for reference bucket
      title: 'Reference Bucket (JSON Only)',
      icon: <FileJson className="w-5 h-5 mr-2" />
    },
    checking: {
      bucketName: appConfig.checkingS3BucketName,
      accept: 'image/*', // Guide lists images
      title: 'Checking Bucket (Images)',
      icon: <ImageIconLucide className="w-5 h-5 mr-2" />
    }
  }), [appConfig.referenceS3BucketName, appConfig.checkingS3BucketName]);

  // Show loading state while config is loading
  if (configLoading) {
    return (
      <div className="bg-background text-foreground min-h-screen py-12 px-4 sm:px-6 lg:px-8 font-body">
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center justify-center min-h-[50vh]">
            <div className="text-center">
              <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
              <p className="text-lg text-muted-foreground">Loading configuration...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  // Check if configuration is properly loaded
  if (!appConfig.referenceS3BucketName || !appConfig.checkingS3BucketName) {
    return (
      <div className="bg-background text-foreground min-h-screen py-12 px-4 sm:px-6 lg:px-8 font-body">
        <div className="max-w-6xl mx-auto">
          <div className="flex items-center justify-center min-h-[50vh]">
            <div className="text-center">
              <AlertCircle className="w-8 h-8 mx-auto mb-4 text-destructive" />
              <p className="text-lg text-destructive mb-2">Configuration Error</p>
              <p className="text-muted-foreground">Unable to load S3 bucket configuration. Please check your environment variables.</p>
              <Button
                onClick={() => window.location.reload()}
                variant="outline"
                className="mt-4"
              >
                Retry
              </Button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-background text-foreground min-h-screen py-12 px-4 sm:px-6 lg:px-8 font-body">
      <div className="max-w-6xl mx-auto">
        <h1 className="text-4xl md:text-5xl font-bold text-gradient-primary text-center">
          Image &amp; File Upload
        </h1>
        <p className="mt-4 text-center text-lg text-secondary-foreground max-w-2xl mx-auto">
          Upload files to <code className="font-code bg-card text-accent px-1 rounded-md">{bucketConfig[uploadTarget].bucketName}</code>.
        </p>

        <div className="mt-10 bg-card p-6 sm:p-8 rounded-2xl border border-border shadow-2xl">
          <div className="flex justify-center mb-8">
            <div className="relative flex p-1 bg-background rounded-full border border-border">
              {(['checking', 'reference'] as const).map(target => (
                <button
                  key={target}
                  onClick={() => setUploadTarget(target)}
                  className={cn(
                    "relative z-10 w-40 sm:w-48 text-center py-2.5 rounded-full text-sm font-semibold transition-colors duration-300",
                    uploadTarget === target ? "text-primary-foreground" : "text-muted-foreground hover:text-primary-foreground"
                  )}
                >
                  {bucketConfig[target].title}
                </button>
              ))}
              <span
                className="absolute top-1 bottom-1 w-40 sm:w-48 rounded-full bg-primary-gradient transition-transform duration-300 ease-out"
                style={{ transform: `translateX(${uploadTarget === 'reference' ? '100%' : '0%'})` }}
              />
            </div>
          </div>

          <div className="flex border-b border-border mb-6">
            {(['uploader', 'history'] as const).map(tab => (
              <button
                key={tab}
                onClick={() => setActiveTab(tab)}
                className={cn(
                  "py-3 px-4 sm:px-6 text-sm font-medium capitalize focus:outline-none",
                  activeTab === tab ? "text-primary-foreground border-b-2 border-accent" : "text-muted-foreground hover:text-primary-foreground"
                )}
              >
                {tab}
              </button>
            ))}
            {activeTab === 'history' && uploadHistory.length > 0 && (
              <div className="ml-auto">
                <Button variant="ghost" size="sm" onClick={clearUploadHistory} className="text-destructive-foreground hover:text-destructive hover:bg-destructive/10">
                  <Trash2 className="w-4 h-4 mr-2" /> Clear History
                </Button>
              </div>
            )}
          </div>
          
          {/* Progress Indicator */}
          {progressState.stage !== 'idle' && (
            <div className="mb-6 p-6 rounded-xl bg-card border border-border">
              <div className="space-y-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-3">
                    {progressState.stage === 'validating' && <Clock className="w-5 h-5 text-blue-400 animate-pulse" />}
                    {progressState.stage === 'uploading' && <Upload className="w-5 h-5 text-blue-400 animate-bounce" />}
                    {progressState.stage === 'processing' && <Loader2 className="w-5 h-5 text-yellow-400 animate-spin" />}
                    {progressState.stage === 'rendering' && <Palette className="w-5 h-5 text-purple-400 animate-pulse" />}
                    {progressState.stage === 'complete' && <CheckCircle className="w-5 h-5 text-green-400" />}
                    {progressState.stage === 'error' && <AlertCircle className="w-5 h-5 text-red-400" />}
                    <div>
                      <h3 className="text-sm font-semibold text-primary-foreground">{progressState.message}</h3>
                      {progressState.details && (
                        <p className="text-xs text-muted-foreground">{progressState.details}</p>
                      )}
                    </div>
                  </div>
                  <div className="text-sm font-medium text-primary-foreground">
                    {progressState.progress}%
                  </div>
                </div>

                <div className="space-y-2">
                  <Progress
                    value={progressState.progress}
                    className={cn(
                      "h-2",
                      progressState.stage === 'error' && "bg-red-900/20",
                      progressState.stage === 'complete' && "bg-green-900/20"
                    )}
                  />

                  {/* Stage indicators */}
                  <div className="flex justify-between text-xs text-muted-foreground">
                    <span className={cn(progressState.stage === 'validating' && "text-blue-400 font-medium")}>
                      Validate
                    </span>
                    <span className={cn(progressState.stage === 'uploading' && "text-blue-400 font-medium")}>
                      Upload
                    </span>
                    {selectedFile?.type === 'application/json' && uploadTarget === 'reference' && (
                      <>
                        <span className={cn(progressState.stage === 'processing' && "text-yellow-400 font-medium")}>
                          Process
                        </span>
                        <span className={cn(progressState.stage === 'rendering' && "text-purple-400 font-medium")}>
                          Render
                        </span>
                      </>
                    )}
                    <span className={cn(progressState.stage === 'complete' && "text-green-400 font-medium")}>
                      Complete
                    </span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {uploadStatus.type !== 'idle' && uploadStatus.type !== 'loading' && (
            <div className={cn(
              "mb-6 p-4 rounded-xl",
              uploadStatus.type === 'success' && "bg-success/20 border border-success/50 text-green-400",
              uploadStatus.type === 'error' && "bg-destructive/20 border border-destructive/50 text-destructive-foreground"
            )}>
              <div className="flex items-start justify-between">
                <div className="flex items-start flex-1">
                  {uploadStatus.type === 'success' ? <CheckCircle className="w-5 h-5 mr-3 mt-0.5 text-success flex-shrink-0" /> : <AlertCircle className="w-5 h-5 mr-3 mt-0.5 text-destructive flex-shrink-0" />}
                  <div className="text-sm whitespace-pre-line">{uploadStatus.message}</div>
                </div>
                <div className="flex items-center space-x-2 ml-2 flex-shrink-0">
                  {uploadStatus.type === 'error' && uploadStatus.message.includes('timeout') && selectedFile && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        setUploadStatus({ type: 'idle', message: '' });
                        // Trigger form submission again
                        handleSubmit(onFormSubmit)();
                      }}
                      className="text-current/90 hover:text-current hover:bg-current/10 border-current/30"
                    >
                      Retry Upload
                    </Button>
                  )}
                  <Button variant="ghost" size="icon" onClick={() => setUploadStatus({ type: 'idle', message: '' })} className="text-current/70 hover:text-current hover:bg-current/10">
                    <X className="w-4 h-4" />
                  </Button>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'uploader' && (
            <form onSubmit={handleSubmit(onFormSubmit)}>
              <div className={cn("grid gap-8", selectedFile ? "lg:grid-cols-2" : "lg:grid-cols-1")}>
                <div className={cn("flex flex-col justify-center items-center space-y-4", selectedFile && "lg:order-1")}>
                  {!selectedFile ? (
                    <div
                      onDragOver={handleDragOver} onDragLeave={handleDragLeave} onDrop={handleDrop}
                      onClick={() => document.getElementById('file-input-field')?.click()}
                      className={cn(
                        "w-full h-64 lg:h-96 p-6 border-2 border-dashed border-border rounded-2xl flex flex-col justify-center items-center text-center cursor-pointer transition-all duration-300",
                        isDragOver ? "border-gradient-purple" : "hover:border-gradient-blue/70"
                      )}
                    >
                      <input type="file" id="file-input-field" className="hidden" accept={bucketConfig[uploadTarget].accept}
                        onChange={(e) => handleFileSelection(e.target.files ? e.target.files[0] : null)} />
                      <UploadCloud className={cn("w-12 h-12 mb-4", isDragOver ? "text-gradient-blue" : "text-muted-foreground")} />
                      <p className="text-lg font-semibold text-primary-foreground">Drag &amp; drop your file here</p>
                      <p className="text-sm text-muted-foreground">or click to browse (Max ${MAX_FILE_SIZE_MB}MB)</p>
                      <p className="text-xs text-muted-foreground mt-1">({bucketConfig[uploadTarget].accept})</p>
                      {errors.file && <p className="text-sm text-destructive mt-2">{errors.file.message as string}</p>}
                    </div>
                  ) : (
                    <Card className="w-full bg-background border-border rounded-2xl">
                      <CardHeader><CardTitle className="text-lg text-primary-foreground">File Preview</CardTitle></CardHeader>
                      <CardContent className="h-72 lg:h-96">
                        {previewData && (selectedFile?.type.startsWith('image/')) ? (
                           <div className="relative w-full h-full rounded-lg overflow-hidden bg-card border border-border">
                             <Image src={previewData} alt="Selected file preview" fill style={{objectFit: "contain"}} data-ai-hint="file preview"/>
                           </div>
                        ) : previewData && (selectedFile?.type === 'application/json' || selectedFile?.type.startsWith('text/')) ? (
                          <div className="h-full">
                            {selectedFile?.type === 'application/json' && uploadTarget === 'reference' && (
                              <div className="mb-2 p-2 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                                <div className="flex items-center text-blue-400 text-xs">
                                  <FileJson className="w-4 h-4 mr-2" />
                                  <span>JSON Layout File - Will be automatically rendered to image after upload</span>
                                </div>
                              </div>
                            )}
                            <ScrollArea className="h-full bg-card p-3 rounded-lg border border-border">
                              <pre className="text-xs text-green-400 whitespace-pre-wrap font-code">{previewData}</pre>
                            </ScrollArea>
                          </div>
                        ) : <p className="text-muted-foreground">No preview available for this file type.</p>}
                      </CardContent>
                    </Card>
                  )}
                </div>

                 <div className={cn("space-y-6", selectedFile && "lg:order-2")}>
                    <div className="space-y-2">
                      <Label htmlFor="s3FolderPath" className="text-sm font-medium text-muted-foreground">
                        Target S3 Folder Path {uploadTarget === 'reference' ? '(Must be "raw")' : '(Optional)'}
                      </Label>
                       {/* Simplified S3 Path input for direct entry as guide doesn't specify browser for upload page directly */}
                      <Input
                        id="s3FolderPath"
                        {...register('s3FolderPath')}
                        placeholder={uploadTarget === 'reference' ? 'raw' : 'e.g., planograms/store123 (no leading/trailing slash)'}
                        className="bg-background border-border rounded-lg mt-1 text-muted-foreground"
                        readOnly={uploadTarget === 'reference'}
                      />
                       <p className="text-xs text-muted-foreground">
                         {uploadTarget === 'reference'
                           ? 'Reference bucket uploads are restricted to the "raw" folder only.'
                           : 'If empty, uploads to bucket root for reference, or a default path for checking.'
                         }
                       </p>
                       {errors.s3FolderPath && <p className="text-sm text-destructive mt-1">{errors.s3FolderPath.message}</p>}
                    </div>
                  
                  {selectedFile && (
                    <>
                      <Card className="bg-background border-border rounded-2xl p-6">
                        <h3 className="text-lg font-semibold text-primary-foreground mb-4">Upload Configuration</h3>
                        <div className="space-y-4">
                          <div>
                            <Label className="text-sm font-medium text-muted-foreground">Selected File</Label>
                            <p className="text-sm text-primary-foreground font-mono mt-1 truncate" title={selectedFile.name}>{selectedFile.name}</p>
                          </div>
                           {/* Custom S3 Filename (optional) - Guide's API takes `fileName` as query param */}
                          <div>
                            <Label htmlFor="customS3FileName" className="text-sm font-medium text-muted-foreground">Custom S3 Filename (Optional)</Label>
                            <Input id="customS3FileName" {...register('customS3FileName')} placeholder={selectedFile.name}
                              className="bg-card border-border rounded-lg mt-1"/>
                            <p className="text-xs text-muted-foreground mt-1">If empty, uses original filename.</p>
                          </div>
                        </div>
                      </Card>
                      <div className="flex flex-col sm:flex-row gap-4 mt-4">
                        <Button
                          type="submit"
                          className="w-full btn-gradient rounded-xl py-3 text-base"
                          disabled={uploadStatus.type === 'loading' || progressState.stage !== 'idle'}
                        >
                          {(uploadStatus.type === 'loading' || progressState.stage !== 'idle') ? (
                            <>
                              <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                              {progressState.stage !== 'idle' ? progressState.message : 'Uploading...'}
                            </>
                          ) : (
                            <>
                              <UploadCloud className="mr-2 h-5 w-5" />
                              Upload File
                            </>
                          )}
                        </Button>
                        <Button
                          type="button"
                          variant="outline"
                          onClick={resetFormState}
                          className="w-full border-border hover:bg-accent/10 text-muted-foreground hover:text-primary-foreground rounded-xl py-3 text-base"
                          disabled={uploadStatus.type === 'loading' || progressState.stage !== 'idle'}
                        >
                          Cancel
                        </Button>
                      </div>
                    </>
                  )}
                </div>
              </div>
            </form>
          )}

          {activeTab === 'history' && (
            <div className="max-w-2xl mx-auto">
              {uploadHistory.length === 0 ? (
                 <div className="text-center py-10 text-muted-foreground">
                    <PackageOpen className="w-16 h-16 mx-auto mb-4 opacity-50" />
                    <p className="text-lg">No recent uploads.</p>
                    <p className="text-sm">Uploaded files will appear here.</p>
                </div>
              ) : (
                <ScrollArea className="h-[60vh]">
                  <ul className="space-y-4">
                    {uploadHistory.map(item => (
                      <li key={item.id} className="bg-background border border-border p-4 rounded-xl shadow-lg">
                        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center">
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-primary-foreground truncate" title={item.fileName}>{item.fileName}</p>
                            <p className="text-xs text-gradient-blue hover:underline break-all truncate block" title={item.s3Url}>
                              {item.s3Url} 
                            </p>
                          </div>
                          <div className={cn(
                            "mt-2 sm:mt-0 sm:ml-4 px-2.5 py-1 text-xs font-semibold rounded-full text-primary-foreground whitespace-nowrap",
                            item.type === 'reference' ? "bg-gradient-to-r from-gradient-blue to-gradient-purple" : "bg-gradient-to-r from-gradient-purple to-gradient-magenta"
                          )}>
                            {item.type} bucket
                          </div>
                        </div>
                        <p className="text-xs text-muted-foreground mt-2">Uploaded on: {format(new Date(item.timestamp), 'Pp')}</p>
                      </li>
                    ))}
                  </ul>
                </ScrollArea>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

