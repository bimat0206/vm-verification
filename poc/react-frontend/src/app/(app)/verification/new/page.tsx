
"use client";

import React, { useState, useEffect, useCallback } from 'react';
import NextImage from 'next/image';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { ScrollArea } from '@/components/ui/scroll-area';
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import {
  CheckCircle, AlertCircle, Loader2, ArrowRight, ArrowLeft,
  Folder, FileText, Home, ArrowUpToLine, Sparkles, RefreshCw, Eye, Copy,
  Image as ImageIconLucide
} from 'lucide-react';
import apiClient from '@/lib/api-client';
import type { Verification, ApiError, BrowserItem, BrowserResponse, BucketType, CreateVerificationRequest, VerificationContext } from '@/lib/types';

import { cn } from '@/lib/utils';
import { useAppConfig } from '@/config';

interface WizardStepDefinition {
  number: number;
  title: string;
}

const wizardSteps: WizardStepDefinition[] = [
  { number: 1, title: "Type" },
  { number: 2, title: "Reference" }, // Title will be dynamically changed for PREVIOUS_VS_CURRENT
  { number: 3, title: "Checking" },
  { number: 4, title: "Review" },
  { number: 5, title: "Result" },
];

const GradientText: React.FC<{ children: React.ReactNode; className?: string }> = ({ children, className }) => (
  <span className={cn("bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 bg-clip-text text-transparent", className)}>
    {children}
  </span>
);

interface S3ImageSelectorStepContentProps {
  title: string;
  bucketType: BucketType;
  currentSelectionS3Uri: string | null;
  onFileConfirmed: (s3Uri: string) => void;
  stepActive: boolean;
  appConfig: {
    referenceS3BucketName: string;
    checkingS3BucketName: string;
  };
}

const S3ImageSelectorStepContent: React.FC<S3ImageSelectorStepContentProps> = ({ title, bucketType, currentSelectionS3Uri, onFileConfirmed, stepActive, appConfig }) => {
  const [currentS3Path, setCurrentS3Path] = useState<string>(''); 
  const [parentS3Path, setParentS3Path] = useState<string | undefined>(undefined); 
  const [pathInput, setPathInput] = useState<string>('');
  const [s3Items, setS3Items] = useState<BrowserItem[]>([]);
  
  const [selectedItemForPreview, setSelectedItemForPreview] = useState<BrowserItem | null>(null);
  const [previewDisplayUrl, setPreviewDisplayUrl] = useState<string | null>(null); 
  const [isPreviewLoading, setIsPreviewLoading] = useState<boolean>(false);
  
  const [isS3ListingLoading, setIsS3ListingLoading] = useState<boolean>(false);
  const [s3Error, setS3Error] = useState<string | null>(null);
  

  const fetchS3Items = useCallback(async (path: string) => {
    if (!stepActive) return;
    setIsS3ListingLoading(true);
    setS3Error(null);
    setSelectedItemForPreview(null); 
    setPreviewDisplayUrl(null);    
    try {
      const response: BrowserResponse = await apiClient.browseFolder(bucketType, path, true); 
      setS3Items(response.items || []);
      setCurrentS3Path(response.currentPath);
      setParentS3Path(response.parentPath);
      setPathInput(response.currentPath);
    } catch (err) {
      const apiErr = err as ApiError;
      setS3Error(`Failed to browse S3: ${apiErr.message || 'Unknown error'}`);
      console.error("S3 Browsing Error", apiErr.message);
      setS3Items([]); 
    } finally {
      setIsS3ListingLoading(false);
    }
  }, [bucketType, stepActive]);

  useEffect(() => {
    if (stepActive) {
      fetchS3Items(''); 
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stepActive, bucketType]); 

  useEffect(() => {
    if (selectedItemForPreview && (selectedItemForPreview.type === 'image' || (selectedItemForPreview.type === 'file' && selectedItemForPreview.contentType?.startsWith('image/')))) {
      setIsPreviewLoading(true);
      setPreviewDisplayUrl(null); 
      apiClient.getImageUrl(selectedItemForPreview.path, bucketType, true)
        .then(url => {
          if (typeof url === 'string' && url.startsWith('http')) {
            setPreviewDisplayUrl(url);
          } else {
            console.warn("[New Verification Page] S3ImageSelector: Received invalid URL for preview:", url);
            setPreviewDisplayUrl(null);
          }
        })
        .catch(err => {
          console.error("[New Verification Page] S3ImageSelector: Could not load preview:", err.message);
          setPreviewDisplayUrl(null);
        })
        .finally(() => {
          setIsPreviewLoading(false);
        });
    } else {
      setPreviewDisplayUrl(null); 
    }
  }, [selectedItemForPreview, bucketType]);


  const handleItemClick = (item: BrowserItem) => {
    if (item.type === 'folder') { 
      let pathToFetch = item.path;
      if (pathToFetch && !pathToFetch.endsWith('/')) { 
        pathToFetch += '/';
      }
      fetchS3Items(pathToFetch);
      setSelectedItemForPreview(null); 
      setPreviewDisplayUrl(null);
    } else {
      setSelectedItemForPreview(item);
    }
  };

  const handleGoUp = () => {
    if (parentS3Path !== undefined) {
      fetchS3Items(parentS3Path);
    }
  };

  const handleGoToRoot = () => {
    fetchS3Items('');
  };

  const handlePathInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPathInput(e.target.value);
  };

  const handleBrowsePathInput = () => {
    fetchS3Items(pathInput);
  };
  
  const handleConfirmSelection = async () => {
    if (!selectedItemForPreview) {
      console.error("Selection Error", "No file selected for preview.");
      return;
    }
    if (selectedItemForPreview.type === 'folder') {
      console.error("Selection Error", "Please select a file, not a folder.");
      return;
    }
    if (!(selectedItemForPreview.type === 'image' || (selectedItemForPreview.type === 'file' && selectedItemForPreview.contentType?.startsWith('image/')))) {
      console.error("Invalid File Type", "Please select an image file (e.g., PNG, JPG).");
      return;
    }

    const actualBucketName = bucketType === 'reference' ? appConfig.referenceS3BucketName : appConfig.checkingS3BucketName;
    const s3Uri = `s3://${actualBucketName}/${selectedItemForPreview.path}`;
    onFileConfirmed(s3Uri);
  };
  
  const getIconForItemType = (itemType: BrowserItem['type'], itemContentType?: string) => {
    if (itemType === 'folder') return <Folder className="w-5 h-5 mr-3 flex-shrink-0 text-sky-400" />;
    if (itemType === 'image' || itemContentType?.startsWith('image/')) return <ImageIconLucide className="w-5 h-5 mr-3 flex-shrink-0 text-green-400" />;
    return <FileText className="w-5 h-5 mr-3 flex-shrink-0 text-muted-foreground" />; 
  };

  const displayedS3Items = s3Items.filter(item =>
    item.type === 'folder' ||
    item.type === 'image' ||
    (item.type === 'file' && item.contentType?.startsWith('image/'))
  );

  return (
    <div className="bg-background p-6 rounded-2xl border border-border h-[70vh] flex flex-col">
      <h3 className="font-bold text-lg text-primary-foreground mb-4">{title}</h3>
      
      <div className="flex items-center gap-2 mb-4 p-2 bg-card rounded-lg border border-border flex-shrink-0">
        <Button onClick={handleGoUp} variant="outline" size="icon" title="Go Up" disabled={isS3ListingLoading || parentS3Path === undefined} className="p-2 hover:bg-accent/10 rounded-md text-muted-foreground hover:text-primary-foreground"><ArrowUpToLine className="w-5 h-5" /></Button>
        <Button onClick={handleGoToRoot} variant="outline" size="icon" title="Go to Root" disabled={isS3ListingLoading} className="p-2 hover:bg-accent/10 rounded-md text-muted-foreground hover:text-primary-foreground"><Home className="w-5 h-5" /></Button>
        <span className="px-3 py-2 bg-background rounded-md text-sm font-mono text-blue-400 flex-grow overflow-x-auto whitespace-nowrap">s3://{bucketType === 'reference' ? appConfig.referenceS3BucketName : appConfig.checkingS3BucketName}/{currentS3Path}</span>
         <Button onClick={() => fetchS3Items(currentS3Path)} variant="outline" size="icon" title="Refresh" disabled={isS3ListingLoading} className="p-2 hover:bg-accent/10 rounded-md text-muted-foreground hover:text-primary-foreground">
          <RefreshCw className={`w-5 h-5 ${isS3ListingLoading && !displayedS3Items.length ? 'animate-spin' : ''}`} />
        </Button>
      </div>

      <div className="flex-grow flex gap-6 overflow-hidden">
        <div className="w-1/2 flex-shrink-0 border-r border-border pr-2">
          <ScrollArea className="h-full">
            {isS3ListingLoading && !displayedS3Items.length ? (
              <div className="flex items-center justify-center h-full text-muted-foreground"><Loader2 className="w-5 h-5 animate-spin mr-2"/>Loading...</div>
            ) : s3Error ? (
              <div className="text-destructive p-2 text-sm flex items-center"><AlertCircle className="w-4 h-4 mr-2"/>{s3Error}</div>
            ) : displayedS3Items.length === 0 ? (
              <div className="text-muted-foreground text-center p-4 text-sm">This folder is empty or contains no image files.</div>
            ) : (
              displayedS3Items.map(item => (
                <div
                  key={item.path}
                  onClick={() => handleItemClick(item)}
                  className={cn(
                    "flex items-center p-2 rounded-md cursor-pointer transition-colors duration-200",
                    selectedItemForPreview?.path === item.path ? 'bg-purple-600/30' : 'hover:bg-accent/10'
                  )}
                >
                  {getIconForItemType(item.type, item.contentType)}
                  <span className="text-sm font-mono truncate text-primary-foreground">{item.name}</span>
                </div>
              ))
            )}
          </ScrollArea>
        </div>

        <div className="w-1/2 flex flex-col">
          <p className="text-sm text-muted-foreground mb-2 flex-shrink-0">Preview</p>
          <div className="flex-grow bg-card rounded-lg p-4 flex items-center justify-center border border-border relative overflow-hidden">
            {isPreviewLoading ? (
                 <Loader2 className="w-8 h-8 animate-spin text-primary" />
            ) : !selectedItemForPreview ? (
              <p className="text-muted-foreground">Select an image file to preview</p>
            ) : (selectedItemForPreview.type === 'image' || (selectedItemForPreview.type === 'file' && selectedItemForPreview.contentType?.startsWith('image/'))) && previewDisplayUrl ? (
              <NextImage src={previewDisplayUrl} alt={`Preview of ${selectedItemForPreview.name}`} fill style={{objectFit: "contain"}} className="rounded-md" data-ai-hint="s3 file preview"/>
            ) : (
              <p className="text-muted-foreground text-sm p-4 text-center">Cannot preview this file type ({selectedItemForPreview?.contentType || selectedItemForPreview?.type || 'unknown'}). Select an image file.</p>
            )}
          </div>
          {selectedItemForPreview && (selectedItemForPreview.type === 'image' || (selectedItemForPreview.type === 'file' && selectedItemForPreview.contentType?.startsWith('image/'))) && (
             <Button 
                onClick={handleConfirmSelection} 
                className="w-full mt-4 flex-shrink-0 text-primary-foreground font-bold py-2 px-4 rounded-lg bg-gradient-to-r from-green-500 to-teal-500 hover:opacity-90" 
                disabled={isPreviewLoading || isS3ListingLoading || (!previewDisplayUrl && !isPreviewLoading)}
              >
                <CheckCircle className="mr-2"/> Select this Image
             </Button>
          )}
        </div>
      </div>

      <div className="flex-shrink-0 mt-4 pt-4 border-t border-border">
        <p className="text-sm text-muted-foreground">Final Selection:</p>
        {currentSelectionS3Uri ? (
          <div className="flex items-center gap-2 mt-2">
            <NextImage src="https://placehold.co/40x40/EC4899/FFFFFF.png?text=✓" alt="selected thumbnail" width={40} height={40} className="rounded-md" data-ai-hint="selection checkmark" />
            <span className="font-mono text-sm text-green-400 truncate" title={currentSelectionS3Uri}>{currentSelectionS3Uri}</span>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground italic mt-1">No image selected for this step.</p>
        )}
      </div>
    </div>
  );
};

const StepIndicator: React.FC<{ number: number; title: string; active: boolean }> = ({ number, title, active }) => (
    <div className="flex items-center">
        <div className={cn(
            "w-8 h-8 rounded-full flex items-center justify-center font-bold text-sm transition-all duration-300 border-2",
            active ? "bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 text-primary-foreground border-transparent shadow-lg" : "bg-border text-muted-foreground border-border"
        )}>
            {number}
        </div>
        <p className={cn("ml-3 font-medium", active ? 'text-primary-foreground' : 'text-muted-foreground')}>{title}</p>
    </div>
);


const formatPercentageValue = (value: any): string => {
  if (value === undefined || value === null) return 'N/A';
  const num = parseFloat(String(value));
  if (isNaN(num)) return 'N/A';

  // Smart percentage formatting:
  // If the value is <= 1, treat it as a decimal (0.7 = 70%)
  // If the value is > 1, treat it as already a percentage (70 = 70%)
  if (num <= 1) {
    return `${(num * 100).toFixed(1)}%`;
  } else {
    return `${num.toFixed(1)}%`;
  }
};

export default function NewVerificationPage() {
  const { config: appConfig, isLoading: configLoading } = useAppConfig();
  const [currentStep, setCurrentStep] = useState(1);
  const [verificationType, setVerificationType] = useState<'LAYOUT_VS_CHECKING' | 'PREVIOUS_VS_CURRENT'>('LAYOUT_VS_CHECKING');
  const [referenceFileS3Uri, setReferenceFileS3Uri] = useState<string | null>(null);
  const [checkingFileS3Uri, setCheckingFileS3Uri] = useState<string | null>(null);

  const [vendingMachineId, setVendingMachineId] = useState<string>('');

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [submissionError, setSubmissionError] = useState<string | null>(null);
  const [verificationResult, setVerificationResult] = useState<Verification | null>(null);
  const [isPolling, setIsPolling] = useState(false);
  const [pollingStatus, setPollingStatus] = useState<string>('');
  

  const handleNextStep = () => {
    if (currentStep === 2 && !referenceFileS3Uri) { 
      setSubmissionError("Please select a reference image.");
      return;
    }
    if (currentStep === 3 && !checkingFileS3Uri) { 
      setSubmissionError("Please select a checking image.");
      return;
    }
    setSubmissionError(null); 
    setCurrentStep(prev => Math.min(prev + 1, wizardSteps.length));
  };

  const handleBackStep = () => {
    setSubmissionError(null); 
    setCurrentStep(prev => Math.max(prev - 1, 1));
  };

  const handleSubmitVerification = async () => {
    if (!referenceFileS3Uri || !checkingFileS3Uri) {
      setSubmissionError("Both reference and checking images must be selected.");
      return;
    }
    setIsSubmitting(true);
    setSubmissionError(null);
    setVerificationResult(null);

    try {
      const verificationContext: VerificationContext = {
        verificationType: verificationType,
        referenceImageUrl: referenceFileS3Uri,
        checkingImageUrl: checkingFileS3Uri,
      };

      if (vendingMachineId.trim()) {
        verificationContext.vendingMachineId = vendingMachineId.trim();
      }
      
      const payload: CreateVerificationRequest = {
        verificationContext: verificationContext
      };
      
      const result = await apiClient.initiateVerification(payload, true);
      setVerificationResult(result);
      console.log("Verification Initiated", `ID: ${result.verificationId}, Status: ${result.verificationStatus}`);
      setCurrentStep(wizardSteps.length);

      // Start polling for results - always poll since API returns 202 Accepted
      setIsPolling(true);
      setPollingStatus('Verification submitted... Looking for results (dynamic intervals: 30s → 15s → 5s → 5s → 10s)');

      try {
        // For temporary IDs, we need to find the actual verification by matching the request details
        let actualVerificationId = result.verificationId;

        if (result.verificationId.includes('-temp')) {
          console.log('Temporary verification ID detected, will search during polling...');
          setPollingStatus('Verification submitted, monitoring for completion...');

          // We'll search for the actual verification during polling attempts
          // This is more reliable than trying to search immediately
        }

        const finalResult = await apiClient.pollVerificationResults(
          actualVerificationId,
          (updatedVerification, nextInterval) => {
            // Update polling status message with dynamic interval
            const nextIntervalText = nextInterval ? `${nextInterval / 1000}s` : 'final check';
            setPollingStatus(`Status: ${updatedVerification.verificationStatus} - Processing... (next check in ${nextIntervalText})`);

            // Update the verification result with latest data
            setVerificationResult(updatedVerification);
          },
          50, // Increased max attempts for dynamic intervals
          true // debug mode
        );

        // Final update with complete results
        setVerificationResult(finalResult);
        setPollingStatus('Verification completed!');

      } catch (pollingError) {
        console.error("Polling failed:", pollingError);
        const errorMessage = (pollingError as Error).message || '';

        if (errorMessage.includes('timed out') && actualVerificationId.includes('-temp')) {
          setPollingStatus('Verification is taking longer than expected. It may still be processing in the background. Please check the verification results page in a few minutes.');
        } else if (errorMessage.includes('not found') && actualVerificationId.includes('-temp')) {
          setPollingStatus('Verification is still being processed. Please wait a moment and check the verification results page.');
        } else {
          setPollingStatus('Failed to get verification results. Please check the verification details manually.');
        }
      } finally {
        setIsPolling(false);
      }

    } catch (error) {
      const err = error as ApiError;
      setSubmissionError(err.message || "Failed to initiate verification.");
      console.error("Submission Error", err.message);
    } finally {
      setIsSubmitting(false);
    }
  };
  
  const resetWizard = () => {
    setCurrentStep(1);
    setVerificationType('LAYOUT_VS_CHECKING');
    setReferenceFileS3Uri(null);
    setCheckingFileS3Uri(null);
    setVendingMachineId('');
    setIsSubmitting(false);
    setSubmissionError(null);
    setVerificationResult(null);
    setIsPolling(false);
    setPollingStatus('');
  };

  const getStepTitle = (stepNumber: number): string => {
    const stepConfig = wizardSteps.find(s => s.number === stepNumber);
    if (!stepConfig) return "Unknown Step";

    if (stepNumber === 2 && verificationType === 'PREVIOUS_VS_CURRENT') {
      return "Previous Image";
    }
    return stepConfig.title;
  };

  // Show loading state while config is loading
  if (configLoading) {
    return (
      <div className="bg-background text-foreground min-h-screen font-body p-4 md:p-8 flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
          <p className="text-lg text-muted-foreground">Loading configuration...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-background text-foreground min-h-screen font-body p-4 md:p-8 flex items-center justify-center">
      <div className="w-full max-w-7xl">
        <header className="text-center mb-10">
          <h1 className="text-4xl md:text-5xl font-bold">
            <GradientText>New Verification</GradientText>
          </h1>
          <p className="text-muted-foreground mt-2 text-lg">
            Follow the steps below to initiate a new visual verification.
          </p>
        </header>

        <div className="flex justify-between items-center mb-10 px-0 md:px-4">
          {wizardSteps.map((step, index) => (
            <React.Fragment key={step.number}>
              <StepIndicator number={step.number} title={getStepTitle(step.number)} active={currentStep >= step.number} />
              {index < wizardSteps.length - 1 && (
                <div className={cn(
                  "flex-1 h-0.5 mx-2 md:mx-4",
                   currentStep > step.number ? "bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500" : "bg-border"
                )}></div>
              )}
            </React.Fragment>
          ))}
        </div>
        
        <Card className="bg-card rounded-2xl border-border shadow-xl flex flex-col">
          <CardContent className="p-0 flex flex-col"> 
            {currentStep === 1 && (
              <div className="p-6 md:p-8">
                <h2 className="text-xl font-bold mb-4 text-primary-foreground">Step 1: Choose Verification Type</h2>
                <Select value={verificationType} onValueChange={(value: 'LAYOUT_VS_CHECKING' | 'PREVIOUS_VS_CURRENT') => setVerificationType(value)}>
                    <SelectTrigger className="w-full h-12 text-base bg-input border-border focus:ring-primary text-primary-foreground">
                    <SelectValue placeholder="Select verification type" />
                    </SelectTrigger>
                    <SelectContent>
                    <SelectItem value="LAYOUT_VS_CHECKING">Layout vs Checking (Planogram vs Actual)</SelectItem>
                    <SelectItem value="PREVIOUS_VS_CURRENT">Previous vs Current (Snapshot Comparison)</SelectItem>
                    </SelectContent>
                </Select>
              </div>
            )}

            {currentStep === 2 && (
              <S3ImageSelectorStepContent
                title={verificationType === 'PREVIOUS_VS_CURRENT' ? "Step 2: Select Previous Image (from Checking Bucket)" : "Step 2: Select Reference Image"}
                bucketType={verificationType === 'PREVIOUS_VS_CURRENT' ? "checking" : "reference"}
                currentSelectionS3Uri={referenceFileS3Uri}
                onFileConfirmed={setReferenceFileS3Uri}
                stepActive={currentStep === 2}
                appConfig={appConfig}
              />
            )}

            {currentStep === 3 && (
              <S3ImageSelectorStepContent
                title="Step 3: Select Checking Image"
                bucketType="checking"
                currentSelectionS3Uri={checkingFileS3Uri}
                onFileConfirmed={setCheckingFileS3Uri}
                stepActive={currentStep === 3}
                appConfig={appConfig}
              />
            )}

            {currentStep === 4 && ( 
              <div className="p-6 md:p-8">
                <h2 className="text-xl font-bold mb-6 text-primary-foreground">Step 4: Review and Submit</h2>
                <div className="bg-background p-6 rounded-lg border border-border space-y-4 text-sm">
                  <div>
                    <strong className="text-muted-foreground w-32 inline-block">Type:</strong> 
                    <span className="font-mono text-primary-foreground">{verificationType}</span>
                  </div>
                  <div>
                    <strong className="text-muted-foreground w-32 inline-block">
                      {verificationType === 'PREVIOUS_VS_CURRENT' ? "Previous Img:" : "Reference:"}
                    </strong> 
                    <span className="font-mono text-green-400 break-all">{referenceFileS3Uri || 'Not Selected'}</span>
                  </div>
                  <div>
                    <strong className="text-muted-foreground w-32 inline-block">Checking:</strong> 
                    <span className="font-mono text-green-400 break-all">{checkingFileS3Uri || 'Not Selected'}</span>
                  </div>
                   <div className="space-y-1">
                    <Label htmlFor="vendingMachineId" className="text-sm font-medium text-muted-foreground">Vending Machine ID (Optional)</Label>
                    <Input 
                        id="vendingMachineId" 
                        value={vendingMachineId} 
                        onChange={(e) => setVendingMachineId(e.target.value)} 
                        placeholder="e.g., VM001"
                        className="bg-input border-border text-primary-foreground"
                    />
                  </div>
                </div>
                {submissionError && <Alert variant="destructive" className="mt-4"><AlertCircle className="h-4 w-4" /><AlertTitle>Error</AlertTitle><AlertDescription>{submissionError}</AlertDescription></Alert>}
              </div>
            )}

            {currentStep === 5 && verificationResult && (
              <div className="p-6 md:p-8">
                <h2 className="text-xl font-bold mb-6 text-primary-foreground">
                  <Sparkles className="inline mr-2" />
                  Step 5: Verification Result
                </h2>

                {/* Main Result Card */}
                <div className="bg-background p-6 rounded-lg border border-border space-y-6">

                  {/* Basic Information */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <p className="text-muted-foreground text-sm mb-1">Verification ID</p>
                      <p className="font-mono text-blue-400 text-lg">{verificationResult.verificationId}</p>
                    </div>
                    <div>
                      <p className="text-muted-foreground text-sm mb-1">Status</p>
                      <span className={cn(
                        "inline-flex items-center font-bold px-3 py-1 rounded-full text-sm",
                        verificationResult.verificationStatus === 'CORRECT' ? "bg-green-500/20 text-green-300" :
                        verificationResult.verificationStatus === 'INCORRECT' ? "bg-orange-500/20 text-orange-300" :
                        verificationResult.verificationStatus === 'COMPLETED' ? "bg-green-500/20 text-green-300" :
                        verificationResult.verificationStatus === 'PENDING' || verificationResult.verificationStatus === 'PROCESSING' ? "bg-yellow-500/20 text-yellow-300" :
                        verificationResult.verificationStatus === 'FAILED' || verificationResult.verificationStatus === 'ERROR' ? "bg-red-500/20 text-red-300" :
                        "bg-muted text-muted-foreground"
                      )}>
                        {verificationResult.verificationStatus === 'PENDING' || verificationResult.verificationStatus === 'PROCESSING' ? (
                          <Loader2 className="w-4 h-4 animate-spin mr-2" />
                        ) : verificationResult.verificationStatus === 'CORRECT' || verificationResult.verificationStatus === 'COMPLETED' ? (
                          <CheckCircle className="w-4 h-4 mr-2" />
                        ) : verificationResult.verificationStatus === 'INCORRECT' ? (
                          <AlertCircle className="w-4 h-4 mr-2" />
                        ) : verificationResult.verificationStatus === 'FAILED' || verificationResult.verificationStatus === 'ERROR' ? (
                          <AlertCircle className="w-4 h-4 mr-2" />
                        ) : null}
                        {verificationResult.verificationStatus}
                      </span>
                    </div>
                  </div>

                  {/* Polling Status */}
                  {isPolling && (
                    <Alert className="bg-blue-500/10 border-blue-500/20">
                      <Loader2 className="w-4 h-4 animate-spin" />
                      <AlertTitle className="text-blue-400">Processing Verification</AlertTitle>
                      <AlertDescription className="text-blue-300">
                        {pollingStatus}
                      </AlertDescription>
                    </Alert>
                  )}

                  {/* Accuracy and Summary Information */}
                  {(verificationResult.overallAccuracy !== undefined || verificationResult.correctPositions !== undefined) ? (
                    <div className="bg-card p-4 rounded-lg border border-border">
                      <h3 className="text-lg font-semibold text-primary-foreground mb-3 flex items-center">
                        <Eye className="w-5 h-5 mr-2" />
                        Verification Summary
                      </h3>
                      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        {verificationResult.overallAccuracy !== undefined && (
                          <div className="text-center">
                            <p className="text-2xl font-bold text-green-400">
                              {formatPercentageValue(verificationResult.overallAccuracy)}
                            </p>
                            <p className="text-muted-foreground text-sm">Overall Accuracy</p>
                          </div>
                        )}
                        {verificationResult.correctPositions !== undefined && (
                          <div className="text-center">
                            <p className="text-2xl font-bold text-blue-400">
                              {verificationResult.correctPositions}
                            </p>
                            <p className="text-muted-foreground text-sm">Correct Positions</p>
                          </div>
                        )}
                        {verificationResult.discrepantPositions !== undefined && (
                          <div className="text-center">
                            <p className="text-2xl font-bold text-orange-400">
                              {verificationResult.discrepantPositions}
                            </p>
                            <p className="text-muted-foreground text-sm">Discrepancies</p>
                          </div>
                        )}
                      </div>
                    </div>
                  ) : !isPolling && verificationResult.verificationStatus && !['PENDING', 'PROCESSING'].includes(verificationResult.verificationStatus) ? (
                    <div className="bg-yellow-500/10 border border-yellow-500/20 rounded-lg p-4">
                      <h3 className="text-lg font-semibold text-yellow-400 mb-2 flex items-center">
                        <AlertCircle className="w-5 h-5 mr-2" />
                        Summary Data Not Available
                      </h3>
                      <p className="text-yellow-400/80 text-sm">
                        This verification has completed ({verificationResult.verificationStatus}) but summary data
                        (accuracy metrics, position counts) is not available. This may indicate the verification
                        failed during processing or the summary data was not generated.
                      </p>
                    </div>
                  ) : null}

                  {/* LLM Analysis */}
                  <div className="bg-card p-4 rounded-lg border border-border">
                    <div className="flex items-center justify-between mb-3">
                      <h3 className="text-lg font-semibold text-primary-foreground flex items-center">
                        <Sparkles className="w-5 h-5 mr-2" />
                        AI-Generated Analysis
                      </h3>
                      {verificationResult.llmAnalysis && (
                        <div className="flex items-center gap-3">
                          <span className="text-xs text-muted-foreground">
                            {verificationResult.llmAnalysis.length.toLocaleString()} characters
                          </span>
                          <Button
                            onClick={() => {
                              navigator.clipboard.writeText(verificationResult.llmAnalysis || '');
                              // You could add a toast notification here
                              console.log('Analysis copied to clipboard');
                            }}
                            variant="outline"
                            size="sm"
                            className="h-7 px-2 text-xs"
                          >
                            <Copy className="w-3 h-3 mr-1" />
                            Copy
                          </Button>
                        </div>
                      )}
                    </div>

                    {verificationResult.llmAnalysis ? (
                      <div className="space-y-4">
                        {/* Full Analysis Display */}
                        <div className="bg-black/20 rounded-md p-4 border border-border/50">
                          <div className="prose prose-invert prose-sm max-w-none">
                            <pre className="font-mono text-sm whitespace-pre-wrap text-primary-foreground/90 leading-relaxed overflow-x-auto">
                              {verificationResult.llmAnalysis}
                            </pre>
                          </div>
                        </div>

                        {/* Analysis Summary */}
                        <div className="bg-blue-500/10 border border-blue-500/20 rounded-md p-3">
                          <p className="text-blue-400 text-sm flex items-center">
                            <Eye className="w-4 h-4 mr-2" />
                            Complete AI analysis displayed above. This includes detailed verification results,
                            discrepancy analysis, and recommendations.
                          </p>
                        </div>
                      </div>
                    ) : isPolling ? (
                      <div className="bg-black/20 rounded-md p-8 border border-border/50">
                        <div className="flex items-center justify-center text-muted-foreground">
                          <Loader2 className="w-6 h-6 animate-spin mr-3" />
                          <span>AI analysis in progress...</span>
                        </div>
                      </div>
                    ) : (
                      <div className="bg-black/20 rounded-md p-8 border border-border/50">
                        <div className="flex items-center justify-center text-muted-foreground">
                          <AlertCircle className="w-6 h-6 mr-3" />
                          <span>No analysis available yet.</span>
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Debug Information (only in development) */}
                  {process.env.NODE_ENV === 'development' && (
                    <details className="bg-muted/20 p-3 rounded-md">
                      <summary className="cursor-pointer text-muted-foreground text-sm">Debug Information</summary>
                      <div className="mt-2 space-y-3">
                        <div>
                          <h4 className="text-xs font-semibold text-muted-foreground mb-1">Verification Summary Status</h4>
                          <pre className="text-xs text-muted-foreground whitespace-pre-wrap">
                            {JSON.stringify({
                              verificationId: verificationResult.verificationId,
                              status: verificationResult.verificationStatus,
                              overallAccuracy: verificationResult.overallAccuracy,
                              correctPositions: verificationResult.correctPositions,
                              discrepantPositions: verificationResult.discrepantPositions,
                              hasVerificationSummary: !!verificationResult.verificationSummary,
                              summaryDisplayCondition: (verificationResult.overallAccuracy !== undefined || verificationResult.correctPositions !== undefined),
                              hasLlmAnalysis: !!verificationResult.llmAnalysis,
                              llmAnalysisLength: verificationResult.llmAnalysis?.length || 0,
                              isPolling: isPolling,
                              pollingStatus: pollingStatus
                            }, null, 2)}
                          </pre>
                        </div>
                        {verificationResult.verificationSummary && (
                          <div>
                            <h4 className="text-xs font-semibold text-muted-foreground mb-1">Raw Verification Summary</h4>
                            <pre className="text-xs text-muted-foreground whitespace-pre-wrap">
                              {JSON.stringify(verificationResult.verificationSummary, null, 2)}
                            </pre>
                          </div>
                        )}
                      </div>
                    </details>
                  )}
                </div>
              </div>
            )}
            
            <div className={cn(
              "flex", 
              (currentStep === 1 || currentStep === 4 || currentStep === 5) ? "mt-8 p-6 md:p-8 pt-0" : "mt-auto p-6 md:p-8" 
            )}>
                {currentStep > 1 && currentStep < 5 && (
                  <Button variant="outline" onClick={handleBackStep} className="text-muted-foreground font-bold py-3 px-8 rounded-lg bg-card hover:bg-accent/10 h-11 text-base border-border mr-auto" disabled={isSubmitting}>
                    <ArrowLeft className="mr-2"/> Back
                  </Button>
                )}
                 <div className="ml-auto"> 
                  {currentStep < 4 && (
                    <Button 
                      onClick={handleNextStep} 
                      disabled={ (currentStep === 2 && !referenceFileS3Uri) || (currentStep === 3 && !checkingFileS3Uri) } 
                      className="btn-gradient font-bold py-3 px-8 rounded-lg h-11 text-base disabled:opacity-50"
                    >
                      Next Step <ArrowRight className="ml-2"/>
                    </Button>
                  )}
                  {currentStep === 4 && (
                    <Button onClick={handleSubmitVerification} disabled={!referenceFileS3Uri || !checkingFileS3Uri || isSubmitting} className="btn-gradient font-bold py-3 px-8 rounded-lg h-11 text-base">
                        {isSubmitting ? <Loader2 className="mr-2 animate-spin"/> : <CheckCircle className="mr-2"/>}
                        Submit Verification
                    </Button>
                  )}
                  {currentStep === 5 && (
                    <div className="flex gap-3">

                      <Button onClick={resetWizard} className="btn-gradient font-bold py-3 px-8 rounded-lg h-11 text-base">
                          <RefreshCw className="mr-2"/> Start New Verification
                      </Button>
                    </div>
                  )}
                </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}

    
