
"use client";

import React, { Suspense, useState, useRef, useCallback, useEffect } from 'react';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Calendar } from '@/components/ui/calendar';
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion'; // Added Accordion imports
import apiClient from '@/lib/api-client';
import type { Verification, ApiError, BucketType } from '@/lib/types';

import { format, subDays, isValid, parseISO } from 'date-fns';
import {
  CalendarIcon, Search, Filter, Loader2, AlertCircle,
  ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight,
  PanelLeft, PanelRight, X, Maximize, ZoomIn, ZoomOut, Move, Eye, FileText,
  CheckCircle, AlertTriangle, Clock, XCircle as XCircleIcon, Sparkles
} from 'lucide-react';
import { Badge } from '@/components/ui/badge';
import { cn } from '@/lib/utils';
import { useSearchParams, useRouter, usePathname } from 'next/navigation';
import Image from 'next/image';
// appConfig is not directly used here anymore for bucket name comparison in image URL fetching
// import appConfig from '@/config'; 

const filterSchema = z.object({
  verificationStatus: z.string().optional(),
  machineId: z.string().optional(),
  verificationType: z.string().optional(),
  dateRangePreset: z.string().optional(),
  customDateFrom: z.date().optional(),
  customDateTo: z.date().optional(),
  sortBy: z.string().optional(),
  quickLookupId: z.string().optional(),
});

type FilterFormData = z.infer<typeof filterSchema>;

const sortOptions: Record<string, string> = {
  "newest": "verificationAt:desc",
  "oldest": "verificationAt:asc",
  "accuracy_desc": "overallAccuracy:desc",
  "accuracy_asc": "overallAccuracy:asc"
};

const parseS3Uri = (s3Uri: string): { bucketName: string; key: string } | null => {
  if (!s3Uri || !s3Uri.startsWith('s3://')) return null;
  const pathWithoutScheme = s3Uri.substring(5);
  const firstSlashIndex = pathWithoutScheme.indexOf('/');
  if (firstSlashIndex === -1) return null;

  const bucketName = pathWithoutScheme.substring(0, firstSlashIndex);
  const key = pathWithoutScheme.substring(firstSlashIndex + 1);
  return { bucketName, key };
};


const ImageContainerWithLoading: React.FC<{ src?: string; alt: string; className?: string; priority?: boolean }> = ({ src, alt, className, priority }) => {
    const [isLoading, setIsLoading] = useState(true);
    const effectiveSrc = src || "https://placehold.co/600x400.png?text=Image+Not+Available";

    useEffect(() => {
        setIsLoading(true); 
    }, [src]);

    return (
        <div className={cn("relative w-full h-full bg-card rounded-lg overflow-hidden", className)}>
            {isLoading && (!src || src.startsWith('s3://') || src === "https://placehold.co/600x400.png?text=Loading...") && (
                <div className="absolute inset-0 flex items-center justify-center bg-card/50">
                    <Loader2 className="w-8 h-8 animate-spin text-primary" />
                </div>
            )}
            {/* Use regular img tag for S3 presigned URLs to avoid Next.js processing issues */}
            {effectiveSrc.includes('amazonaws.com') || effectiveSrc.includes('s3.') ? (
                <img
                    src={effectiveSrc}
                    alt={alt}
                    className={cn("w-full h-full object-contain transition-opacity duration-300", isLoading ? 'opacity-0' : 'opacity-100')}
                    onLoad={() => setIsLoading(false)}
                    onError={() => { setIsLoading(false); console.error(`Failed to load image: ${alt} from ${effectiveSrc}`); }}
                    data-ai-hint="verification image"
                />
            ) : (
                <Image
                    src={effectiveSrc}
                    alt={alt}
                    fill
                    style={{objectFit: "contain"}}
                    onLoad={() => setIsLoading(false)}
                    onError={() => { setIsLoading(false); console.error(`Failed to load image: ${alt} from ${effectiveSrc}`); }}
                    className={cn("transition-opacity duration-300", isLoading ? 'opacity-0' : 'opacity-100')}
                    priority={priority}
                    unoptimized={true}
                    data-ai-hint="verification image"
                />
            )}
        </div>
    );
};

const FullScreenImageViewer: React.FC<{ referenceSrc?: string; checkingSrc?: string; onClose: () => void }> = ({ referenceSrc, checkingSrc, onClose }) => {
    const [transform, setTransform] = useState({ scale: 1, x: 0, y: 0 });
    const viewerRef = useRef<HTMLDivElement>(null);
    const isPanning = useRef(false);
    const startPanPos = useRef({ x: 0, y: 0 });

    const handleWheel = useCallback((e: WheelEvent) => {
        e.preventDefault();
        const scaleFactor = 0.1;
        setTransform(prev => {
            const newScale = prev.scale - (e.deltaY > 0 ? scaleFactor : -scaleFactor);
            return { ...prev, scale: Math.max(0.2, Math.min(newScale, 5)) };
        });
    }, []);

    const handleMouseDown = (e: React.MouseEvent<HTMLDivElement>) => {
        e.preventDefault();
        isPanning.current = true;
        startPanPos.current = { x: e.clientX - transform.x, y: e.clientY - transform.y };
        if (viewerRef.current) viewerRef.current.style.cursor = 'grabbing';
    };

    const handleMouseMove = useCallback((e: MouseEvent) => {
        if (!isPanning.current || !viewerRef.current) return;
        e.preventDefault();
        setTransform(prev => ({
            ...prev,
            x: e.clientX - startPanPos.current.x,
            y: e.clientY - startPanPos.current.y
        }));
    }, []); 

    const handleMouseUp = useCallback(() => {
        isPanning.current = false;
        if (viewerRef.current) viewerRef.current.style.cursor = 'grab';
    }, []);

    useEffect(() => {
        const currentViewerRef = viewerRef.current;
        if (currentViewerRef) {
            currentViewerRef.addEventListener('wheel', handleWheel, { passive: false });
            document.addEventListener('mousemove', handleMouseMove);
            document.addEventListener('mouseup', handleMouseUp);

            const handleEsc = (event: KeyboardEvent) => {
                if (event.key === 'Escape') onClose();
            };
            document.addEventListener('keydown', handleEsc);

            return () => {
                currentViewerRef.removeEventListener('wheel', handleWheel);
                document.removeEventListener('mousemove', handleMouseMove);
                document.removeEventListener('mouseup', handleMouseUp);
                document.removeEventListener('keydown', handleEsc);
            };
        }
    }, [handleWheel, handleMouseMove, handleMouseUp, onClose]);

    return (
        <div
            className="fixed inset-0 bg-black/90 backdrop-blur-md z-[100] flex flex-col p-4"
            ref={viewerRef}
        >
            <div className="flex justify-between items-center mb-4 flex-shrink-0">
                <div className="flex items-center space-x-4 text-muted-foreground text-sm">
                    <span className="flex items-center"><ZoomIn className="w-4 h-4 mr-1"/>/<ZoomOut className="w-4 h-4 mr-1"/> Mouse Wheel: Zoom</span>
                    <span className="flex items-center"><Move className="w-4 h-4 mr-1"/> Click & Drag: Pan</span>
                    <span>ESC: Close</span>
                </div>
                <Button onClick={onClose} variant="ghost" size="icon" className="text-foreground hover:bg-white/10">
                    <X className="h-6 w-6" />
                </Button>
            </div>
            <div
                className="flex-grow flex gap-4 overflow-hidden cursor-grab"
                onMouseDown={handleMouseDown}
            >
                <div
                    className="flex gap-4 w-full h-full transition-transform duration-100 ease-out"
                    style={{ transform: `translate(${transform.x}px, ${transform.y}px) scale(${transform.scale})`}}
                >
                    <div className="w-1/2 h-full rounded-lg border border-border overflow-hidden"><ImageContainerWithLoading src={referenceSrc} alt="Reference Full Screen" priority/></div>
                    <div className="w-1/2 h-full rounded-lg border border-border overflow-hidden"><ImageContainerWithLoading src={checkingSrc} alt="Checking Full Screen" priority/></div>
                </div>
            </div>
        </div>
    );
};

const StatusPill: React.FC<{ status?: Verification['verificationStatus'] }> = ({ status }) => {
  if (!status) return <Badge variant="outline" className={cn("text-xs font-medium px-2.5 py-1", "bg-muted text-muted-foreground border-muted-foreground/30")}><AlertCircle className="w-3.5 h-3.5 mr-1.5" />UNKNOWN</Badge>;

  let MappedIcon = AlertCircle;
  let colorClasses = "bg-muted text-muted-foreground border-muted-foreground/30";

  switch (status) {
    case 'CORRECT':
      MappedIcon = CheckCircle;
      colorClasses = "bg-green-500/20 text-green-400 border-green-500/50";
      break;
    case 'INCORRECT':
      MappedIcon = XCircleIcon;
      colorClasses = "bg-red-500/20 text-red-400 border-red-500/50";
      break;
    case 'PENDING':
      MappedIcon = Clock;
      colorClasses = "bg-yellow-500/20 text-yellow-400 border-yellow-500/50";
      break;
    case 'ERROR':
      MappedIcon = AlertTriangle;
      colorClasses = "bg-orange-500/20 text-orange-400 border-orange-500/50";
      break;
    case 'PROCESSING': 
      MappedIcon = Loader2; 
      colorClasses = "bg-blue-500/20 text-blue-400 border-blue-500/50"; 
      return (
        <Badge variant="outline" className={cn("text-xs font-medium px-2.5 py-1", colorClasses)}>
            <MappedIcon className="w-3.5 h-3.5 mr-1.5 animate-spin" />
            {status}
        </Badge>
      );
    case 'COMPLETED': 
      MappedIcon = CheckCircle; 
      colorClasses = "bg-purple-500/20 text-purple-400 border-purple-500/50";
      break;
  }
  return (
    <Badge variant="outline" className={cn("text-xs font-medium px-2.5 py-1", colorClasses)}>
      <MappedIcon className="w-3.5 h-3.5 mr-1.5" />
      {status}
    </Badge>
  );
};

function VerificationResultsPageContent() {
  
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();

  const [verifications, setVerifications] = useState<Verification[]>([]);
  const [totalVerifications, setTotalVerifications] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [mounted, setMounted] = useState(false);

  const [isSidebarVisible, setIsSidebarVisible] = useState(true);
  const [selectedVerification, setSelectedVerification] = useState<Verification | null>(null);
  const [isFullScreenViewerOpen, setIsFullScreenViewerOpen] = useState(false);

  const [resultsPerPage, setResultsPerPage] = useState(
    parseInt(searchParams.get('limit') || '10', 10)
  );
  const currentPage = parseInt(searchParams.get('page') || '1', 10);

  const [resolvedImageUrls, setResolvedImageUrls] = useState<{ reference?: string; checking?: string }>({});
  const [detailedLlmAnalysis, setDetailedLlmAnalysis] = useState<string | null>(null);
  const [llmAnalysisLoading, setLlmAnalysisLoading] = useState(false);
  const [llmAnalysisError, setLlmAnalysisError] = useState<string | null>(null);
  const [llmAnalysisCache, setLlmAnalysisCache] = useState<Map<string, string | null>>(new Map());

  // Previous analysis state for Previous vs Current verification type
  const [previousAnalysis, setPreviousAnalysis] = useState<string | null>(null);
  const [previousAnalysisLoading, setPreviousAnalysisLoading] = useState(false);
  const [previousAnalysisError, setPreviousAnalysisError] = useState<string | null>(null);
  const [previousAnalysisCache, setPreviousAnalysisCache] = useState<Map<string, string | null>>(new Map());
  const [previousVerificationId, setPreviousVerificationId] = useState<string | null>(null);

  // Analysis comparison modal state
  const [isAnalysisComparisonOpen, setIsAnalysisComparisonOpen] = useState(false);


  const form = useForm<FilterFormData>({
    resolver: zodResolver(filterSchema),
  });

  const { watch, setValue, reset, control } = form;
  const dateRangePreset = watch('dateRangePreset');

  const formatDateSafe = (dateInput: string | Date | undefined, formatPattern: string = 'PPpp'): string => {
    if (!dateInput) return 'N/A';
    try {
      const dateObj = typeof dateInput === 'string' ? parseISO(dateInput) : dateInput;
      if (isValid(dateObj)) {
        return format(dateObj, formatPattern);
      }
      return 'Invalid Date';
    } catch (e) {
      console.error("Date formatting error for:", dateInput, e);
      return 'Invalid Date';
    }
  };

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



  const formatVerificationType = (verificationType: string | undefined): string => {
    console.log(`[DEBUG] Formatting verification type: "${verificationType}"`);
    if (!verificationType) return 'N/A';
    switch (verificationType) {
      case 'LAYOUT_VS_CHECKING':
        return 'Layout vs Checking';
      case 'PREVIOUS_VS_CURRENT':
        return 'Previous vs Current';
      default:
        console.log(`[DEBUG] Unknown verification type, returning as-is: "${verificationType}"`);
        return verificationType;
    }
  };

  const fetchVerifications = useCallback(async (filters: Partial<FilterFormData>, page: number, limit: number) => {
    setLoading(true);
    setError(null);
    try {
      let dateRangeStartStr, dateRangeEndStr;
      if (filters.dateRangePreset === 'LAST_24_HOURS') {
        dateRangeStartStr = subDays(new Date(), 1).toISOString();
      } else if (filters.dateRangePreset === 'LAST_7_DAYS') {
        dateRangeStartStr = subDays(new Date(), 7).toISOString();
      } else if (filters.dateRangePreset === 'CUSTOM' && filters.customDateFrom && filters.customDateTo) {
        if (isValid(filters.customDateFrom)) dateRangeStartStr = format(filters.customDateFrom, "yyyy-MM-dd'T'HH:mm:ss.SSSxxx");
        if (isValid(filters.customDateTo)) dateRangeEndStr = format(filters.customDateTo, "yyyy-MM-dd'T'HH:mm:ss.SSSxxx");
      }

      const apiSortBy = filters.sortBy ? sortOptions[filters.sortBy as keyof typeof sortOptions] : sortOptions.newest;

      const apiCallParams: any = {
        verificationStatus: filters.verificationStatus === 'ALL' || !filters.verificationStatus ? undefined : filters.verificationStatus,
        vendingMachineId: filters.machineId || undefined,
        verificationId: filters.quickLookupId || undefined,
        // Note: verificationType filtering is handled client-side since API doesn't support it yet
        dateRangeStart: dateRangeStartStr,
        dateRangeEnd: dateRangeEndStr,
        sortBy: apiSortBy,
        limit: limit,
        offset: (page - 1) * limit,
      };

      const urlUpdateParams = new URLSearchParams();
      if (filters.verificationStatus && filters.verificationStatus !== 'ALL') urlUpdateParams.set('verificationStatus', filters.verificationStatus);
      if (filters.machineId) urlUpdateParams.set('machineId', filters.machineId);
      if (filters.verificationType && filters.verificationType !== 'ALL') urlUpdateParams.set('verificationType', filters.verificationType);
      if (filters.quickLookupId) urlUpdateParams.set('quickLookupId', filters.quickLookupId);
      if (filters.dateRangePreset && filters.dateRangePreset !== 'ALL_TIME') urlUpdateParams.set('dateRangePreset', filters.dateRangePreset);
      if(filters.customDateFrom && filters.dateRangePreset === 'CUSTOM' && isValid(filters.customDateFrom)) urlUpdateParams.set('customDateFrom', format(filters.customDateFrom, "yyyy-MM-dd'T'HH:mm:ss.SSSxxx"));
      if(filters.customDateTo && filters.dateRangePreset === 'CUSTOM' && isValid(filters.customDateTo)) urlUpdateParams.set('customDateTo', format(filters.customDateTo, "yyyy-MM-dd'T'HH:mm:ss.SSSxxx"));
      if (filters.sortBy && filters.sortBy !== 'newest') urlUpdateParams.set('sortBy', filters.sortBy);

      urlUpdateParams.set('page', String(page));
      urlUpdateParams.set('limit', String(limit));

      router.replace(`${pathname}?${urlUpdateParams.toString()}`, { scroll: false });

      const response = await apiClient.listVerifications(apiCallParams);

      console.log('[DEBUG] API Response from listVerifications:', response);
      console.log('[DEBUG] First verification from API:', response.results[0]);
      console.log('[DEBUG] Pagination info:', response.pagination);

      // Fetch detailed data for each verification to get accuracy information
      const verificationsWithDetails = await Promise.all(
        response.results.map(async (verification) => {
          try {
            const details = await apiClient.getVerificationDetails(verification.verificationId);
            return {
              ...verification,
              overallAccuracy: details.overallAccuracy || details.rawData?.overall_accuracy || details.rawData?.overallAccuracy,
              rawData: details.rawData || verification.rawData
            };
          } catch (error) {
            console.warn(`Failed to fetch details for verification ${verification.verificationId}:`, error);
            return verification;
          }
        })
      );

      // Apply client-side filtering by verification type since API doesn't support it yet
      let filteredVerifications = verificationsWithDetails;
      if (filters.verificationType && filters.verificationType !== 'ALL') {
        filteredVerifications = verificationsWithDetails.filter(verification => {
          const verType = verification.verificationType;
          return verType === filters.verificationType;
        });
      }

      setVerifications(filteredVerifications);
      // Use the total from API response pagination
      setTotalVerifications(response.pagination?.total || filteredVerifications.length);
    } catch (err) {
      const apiErr = err as ApiError;
      setError(apiErr.message || 'Failed to fetch verifications.');
      // toast({ variant: "destructive", title: "Error", description: apiErr.message });
      console.error("Error fetching verifications", apiErr.message);
    } finally {
      setLoading(false);
    }
  }, [router, pathname]);

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    if (!mounted) return;

    const params = new URLSearchParams(searchParams.toString());
    const currentVerificationStatus = params.get('verificationStatus') || 'ALL';
    const currentMachineId = params.get('machineId') || undefined;
    const currentVerificationType = params.get('verificationType') || undefined;
    const currentDateRangePreset = params.get('dateRangePreset') || 'ALL_TIME';
    const currentSortBy = params.get('sortBy') || 'newest';
    const currentCustomDateFromStr = params.get('customDateFrom');
    const currentCustomDateToStr = params.get('customDateTo');
    const quickLookupIdFromUrl = params.get('quickLookupId');

    const filtersFromUrl: FilterFormData = {
      verificationStatus: currentVerificationStatus,
      machineId: currentMachineId,
      verificationType: currentVerificationType,
      dateRangePreset: currentDateRangePreset,
      sortBy: currentSortBy,
      quickLookupId: quickLookupIdFromUrl || '',
      customDateFrom: undefined,
      customDateTo: undefined,
    };

    if (currentCustomDateFromStr) {
        const parsedDate = parseISO(currentCustomDateFromStr);
        if(isValid(parsedDate)) filtersFromUrl.customDateFrom = parsedDate;
    }
    if (currentCustomDateToStr) {
        const parsedDate = parseISO(currentCustomDateToStr);
        if(isValid(parsedDate)) filtersFromUrl.customDateTo = parsedDate;
    }

    reset(filtersFromUrl); 
    fetchVerifications(filtersFromUrl, currentPage, resultsPerPage);
  }, [mounted, searchParams, currentPage, resultsPerPage, fetchVerifications, reset]);


  useEffect(() => {
    const currentVerId = selectedVerification?.verificationId;
    const currentRefS3Uri = selectedVerification?.referenceImageUrl;
    const currentCheckS3Uri = selectedVerification?.checkingImageUrl;

    if (currentVerId) {
      const fetchDetailsAndAnalysis = async () => {
        setResolvedImageUrls({
          reference: "https://placehold.co/600x400.png?text=Loading...",
          checking: "https://placehold.co/600x400.png?text=Loading...",
        });
        setLlmAnalysisLoading(true);
        setLlmAnalysisError(null);
        setDetailedLlmAnalysis(null);

        let refDisplayUrl: string | undefined = undefined;
        let checkDisplayUrl: string | undefined = undefined;

        // Determine the verification type to know which bucket to use for reference image
        const verificationType = selectedVerification.verificationType;
        const referenceBucketType: BucketType = verificationType === 'PREVIOUS_VS_CURRENT' ? 'checking' : 'reference';

        console.log(`[Verifications Page] Verification Type: ${verificationType}, Reference Bucket: ${referenceBucketType}`);

        try {
          if (currentRefS3Uri && currentRefS3Uri.startsWith('s3://')) {
            const parsed = parseS3Uri(currentRefS3Uri);
            if (parsed) {
              const presignedResponseUrl = await apiClient.getImageUrl(parsed.key, referenceBucketType, true);
              if (presignedResponseUrl && typeof presignedResponseUrl === 'string' && presignedResponseUrl.startsWith('http')) {
                refDisplayUrl = presignedResponseUrl;
              } else {
                console.warn(`[Verifications Page] getImageUrl for reference did not return a valid HTTP(S) URL. Received:`, presignedResponseUrl);
              }
            } else {
               console.warn("[Verifications Page] Failed to parse reference S3 URI:", currentRefS3Uri);
            }
          } else if (currentRefS3Uri && currentRefS3Uri.startsWith('http')) {
            refDisplayUrl = currentRefS3Uri;
          } else if (currentRefS3Uri) {
             console.warn("[Verifications Page] Reference image URI is not a valid S3 or HTTP(S) URL:", currentRefS3Uri);
          }
        } catch (e: any) {
          const errorMessage = e?.message || 'Unknown error';
          console.error(`[Verifications Page] Error fetching reference presigned URL for S3 URI "${currentRefS3Uri}":`, errorMessage, e.details || {});

          // Set a fallback URL for missing images instead of leaving it undefined
          if (errorMessage.includes('does not exist') || errorMessage.includes('not accessible')) {
            refDisplayUrl = "https://placehold.co/600x400.png?text=Reference+Image+Not+Found";
          } else {
            refDisplayUrl = "https://placehold.co/600x400.png?text=Error+Loading+Reference+Image";
          }
        }
  
        try {
          if (currentCheckS3Uri && currentCheckS3Uri.startsWith('s3://')) {
            const parsed = parseS3Uri(currentCheckS3Uri);
            if (parsed) {
              const presignedResponseUrl = await apiClient.getImageUrl(parsed.key, 'checking', true);
               if (presignedResponseUrl && typeof presignedResponseUrl === 'string' && presignedResponseUrl.startsWith('http')) {
                  checkDisplayUrl = presignedResponseUrl;
              } else {
                  console.warn(`[Verifications Page] getImageUrl for checking did not return a valid HTTP(S) URL. Received:`, presignedResponseUrl);
              }
            } else {
              console.warn("[Verifications Page] Failed to parse checking S3 URI:", currentCheckS3Uri);
            }
          } else if (currentCheckS3Uri && currentCheckS3Uri.startsWith('http')) {
            checkDisplayUrl = currentCheckS3Uri;
          } else if (currentCheckS3Uri) {
             console.warn("[Verifications Page] Checking image URI is not a valid S3 or HTTP(S) URL:", currentCheckS3Uri);
          }
        } catch (e: any) {
          const errorMessage = e?.message || 'Unknown error';
          console.error(`[Verifications Page] Error fetching checking presigned URL for S3 URI "${currentCheckS3Uri}":`, errorMessage, e.details || {});

          // Set a fallback URL for missing images instead of leaving it undefined
          if (errorMessage.includes('does not exist') || errorMessage.includes('not accessible')) {
            checkDisplayUrl = "https://placehold.co/600x400.png?text=Checking+Image+Not+Found";
          } else {
            checkDisplayUrl = "https://placehold.co/600x400.png?text=Error+Loading+Checking+Image";
          }
        }
        
        setResolvedImageUrls({ reference: refDisplayUrl, checking: checkDisplayUrl });
  
        if (llmAnalysisCache.has(currentVerId)) {
            setDetailedLlmAnalysis(llmAnalysisCache.get(currentVerId) || "No analysis available (cached).");
            setLlmAnalysisLoading(false);
            return;
        }

        let analysisContent: string | null = null;
        try {
            console.log(`[DEBUG] Attempting to fetch conversation data for verification ID: ${currentVerId}`);
            const conversationData = await apiClient.getVerificationConversation(currentVerId, true); // Enable debug
            
            if (conversationData?.turn2) {
                if (typeof conversationData.turn2 === 'string') analysisContent = conversationData.turn2;
                else if (typeof conversationData.turn2 === 'object' && conversationData.turn2.content) analysisContent = conversationData.turn2.content;
            } else if (conversationData?.turn2Content && typeof conversationData.turn2Content === 'object' && conversationData.turn2Content.content) {
                analysisContent = conversationData.turn2Content.content;
            }
            
            if (!analysisContent && conversationData?.turn1) { 
                 if (typeof conversationData.turn1 === 'string') analysisContent = conversationData.turn1;
                else if (typeof conversationData.turn1 === 'object' && conversationData.turn1.content) analysisContent = conversationData.turn1.content;
            }
        } catch (convErr: any) {
            console.warn("Error fetching LLM conversation:", convErr.message);
            if (convErr.message?.includes('not found') || convErr.message?.includes('404')) {
                setLlmAnalysisError(`No conversation data found for this verification. This may be normal if the verification was created before LLM analysis features were implemented or if the analysis step failed. `);
            } else {
                setLlmAnalysisError(`Conversation API Error: ${convErr.message}. `);
            }
        }

        if (!analysisContent && selectedVerification.llmAnalysis) {
            analysisContent = selectedVerification.llmAnalysis;
            if(llmAnalysisError) setLlmAnalysisError(prev => (prev || "") + "Using fallback llmAnalysis field. ");
            else setLlmAnalysisError("Used fallback llmAnalysis field. ");
        }

        if (!analysisContent && selectedVerification.rawData) {
            const fallbackFields = ['analysis', 'llm_analysis', 'verificationAnalysis', 'aiAnalysis', 'description'];
            for (const field of fallbackFields) {
                if (selectedVerification.rawData && typeof selectedVerification.rawData[field] === 'string' && selectedVerification.rawData[field]) {
                    analysisContent = selectedVerification.rawData[field] as string;
                    if(llmAnalysisError) setLlmAnalysisError(prev => (prev || "") + `Using fallback from rawData field: ${field}.`);
                    else setLlmAnalysisError(`Used fallback from rawData field: ${field}.`);
                    break;
                }
            }
        }
        
        const finalAnalysis = analysisContent || "No detailed analysis available.";
        setDetailedLlmAnalysis(finalAnalysis);
        setLlmAnalysisCache(prevCache => new Map(prevCache).set(currentVerId, finalAnalysis));
        setLlmAnalysisLoading(false);

        // Fetch previous analysis if this is a "Previous vs Current" verification type
        // verificationType is already declared above, so we can reuse it
        if (verificationType === 'PREVIOUS_VS_CURRENT') {
          const prevVerificationId = selectedVerification.previousVerificationId ||
                                       selectedVerification.rawData?.previousVerificationId ||
                                       selectedVerification.rawData?.previous_verification_id;

          // Store the previous verification ID for display
          setPreviousVerificationId(prevVerificationId || null);

          console.log('[DEBUG] Previous vs Current verification detected:', {
            selectedVerificationId: selectedVerification.verificationId,
            previousVerificationId: prevVerificationId,
            hasRawData: !!selectedVerification.rawData,
            hasLlmAnalysis: !!selectedVerification.llmAnalysis,
            llmAnalysisLength: selectedVerification.llmAnalysis ? selectedVerification.llmAnalysis.length : 0,
            rawDataKeys: selectedVerification.rawData ? Object.keys(selectedVerification.rawData) : [],
            currentVerificationFields: Object.keys(selectedVerification),
            fullSelectedVerification: selectedVerification
          });

          if (prevVerificationId) {
            setPreviousAnalysisLoading(true);
            setPreviousAnalysisError(null);

            // Check cache first
            if (previousAnalysisCache.has(prevVerificationId)) {
              setPreviousAnalysis(previousAnalysisCache.get(prevVerificationId) || "No previous analysis available (cached).");
              setPreviousAnalysisLoading(false);
            } else {
              try {
                const prevAnalysisContent = await apiClient.getPreviousVerificationAnalysis(prevVerificationId, true);
                if (prevAnalysisContent) {
                  setPreviousAnalysis(prevAnalysisContent);
                  setPreviousAnalysisCache(prevCache => new Map(prevCache).set(prevVerificationId, prevAnalysisContent));
                } else {
                  setPreviousAnalysis("No analysis content available for the previous verification. The previous verification may not have completed processing or may not have generated analysis content.");
                  setPreviousAnalysisCache(prevCache => new Map(prevCache).set(prevVerificationId, null));
                }
              } catch (prevErr: any) {
                console.warn("Error fetching previous verification analysis:", prevErr);

                // More robust error handling
                const errorMessage = prevErr?.message || 'Unknown error';
                const statusCode = prevErr?.statusCode;

                if (statusCode === 404 || errorMessage.includes('not found') || errorMessage.includes('404')) {
                  setPreviousAnalysisError(`Previous verification (${prevVerificationId}) was not found in the database. This may indicate a data integrity issue.`);
                  setPreviousAnalysis(null);
                } else if (statusCode === 403 || statusCode === 401) {
                  setPreviousAnalysisError(`Access denied when fetching previous verification analysis. Please check your permissions.`);
                  setPreviousAnalysis(null);
                } else if (statusCode === 500 || statusCode === 502 || statusCode === 503) {
                  setPreviousAnalysisError(`Server error when fetching previous verification analysis. Please try again later.`);
                  setPreviousAnalysis(null);
                } else if (errorMessage.includes('timeout') || statusCode === 504) {
                  setPreviousAnalysisError(`Request timed out when fetching previous verification analysis. Please try again.`);
                  setPreviousAnalysis(null);
                } else {
                  // For other errors, show a generic message but don't fail completely
                  setPreviousAnalysisError(`Unable to load previous verification analysis: ${errorMessage}. You can still view the current analysis.`);
                  setPreviousAnalysis(null);
                }
              } finally {
                setPreviousAnalysisLoading(false);
              }
            }
          } else {
            setPreviousAnalysisError("No previous verification ID found for this Previous vs Current verification.");
            setPreviousAnalysis(null);
            setPreviousAnalysisLoading(false);
          }
        } else {
          // Reset previous analysis state for non-Previous vs Current verifications
          setPreviousAnalysis(null);
          setPreviousAnalysisError(null);
          setPreviousAnalysisLoading(false);
          setPreviousVerificationId(null);
        }
      };

      fetchDetailsAndAnalysis();
    } else {
      setResolvedImageUrls({});
      setDetailedLlmAnalysis(null);
      setLlmAnalysisLoading(false);
      setLlmAnalysisError(null);
      // Reset previous analysis state
      setPreviousAnalysis(null);
      setPreviousAnalysisLoading(false);
      setPreviousAnalysisError(null);
      setPreviousVerificationId(null);
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedVerification]);


  const onSubmitFilters = (data: FilterFormData) => {
    fetchVerifications(data, 1, resultsPerPage);
  };

  const handleResetFilters = () => {
    const defaultVals: FilterFormData = {
      verificationStatus: 'ALL', machineId: '', verificationType: 'ALL',
      dateRangePreset: 'ALL_TIME', sortBy: 'newest', quickLookupId: '',
      customDateFrom: undefined, customDateTo: undefined,
    };
    reset(defaultVals);
    fetchVerifications(defaultVals, 1, resultsPerPage);
  };

  const totalPages = Math.ceil(totalVerifications / resultsPerPage);

  const handlePageChange = (newPage: number) => {
    if (newPage >= 1 && newPage <= totalPages) {
      const currentFilters = form.getValues();
      fetchVerifications(currentFilters, newPage, resultsPerPage);
    }
  };

  const handleResultsPerPageChange = (value: string) => {
    const newLimit = parseInt(value, 10);
    setResultsPerPage(newLimit);
    const currentFilters = form.getValues();
    fetchVerifications(currentFilters, 1, newLimit);
  };

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
        if (e.key === 'Escape') {
            if(isAnalysisComparisonOpen) setIsAnalysisComparisonOpen(false);
            else if(isFullScreenViewerOpen) setIsFullScreenViewerOpen(false);
            else if(selectedVerification) setSelectedVerification(null);
        }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isFullScreenViewerOpen, selectedVerification, isAnalysisComparisonOpen]);

  const GradientTitle = ({ children }: { children: React.ReactNode }) => (
    <h1 className="text-4xl md:text-5xl font-bold mb-3 text-center text-gradient-primary">
      {children}
    </h1>
  );

  if (!mounted) {
    return <VerificationResultsLoading />;
  }

  return (
    <div className="min-h-screen bg-background text-foreground">
      <main className="container mx-auto px-4 py-8">
        <header className="mb-10 text-center">
          <GradientTitle>Verification Results</GradientTitle>
          <p className="text-lg text-muted-foreground">Review, filter, and analyze all verification outcomes.</p>
        </header>

        <div className="flex flex-row gap-6 items-start">
          <div className={cn(
            "sticky top-[calc(var(--top-nav-bar-height,4rem)+1.5rem)] transition-all duration-300 ease-in-out",
            isSidebarVisible ? "w-80 opacity-100" : "w-0 opacity-0 -mr-6 overflow-hidden"
          )}>
            <div className="h-full bg-card/60 backdrop-blur-lg rounded-xl border border-border p-5 space-y-6 shadow-xl">
              <div>
                <h3 className="text-lg font-semibold text-primary-foreground mb-3">Quick Lookup</h3>
                <form onSubmit={form.handleSubmit(onSubmitFilters)} className="space-y-3">
                    <Input
                        {...form.register('quickLookupId')}
                        placeholder="Enter Verification ID"
                        className="bg-input border-border focus:ring-primary"
                    />
                    <Button type="submit" className="w-full btn-gradient">
                        <Search className="mr-2 h-4 w-4" /> Lookup ID
                    </Button>
                </form>
              </div>

              <hr className="border-border/50" />

              <div>
                <h3 className="text-lg font-semibold text-primary-foreground mb-4 flex items-center">
                    <Filter className="w-5 h-5 mr-2 text-primary" /> Filters & Sorting
                </h3>
                <form onSubmit={form.handleSubmit(onSubmitFilters)} className="space-y-4 text-sm">
                  <div>
                    <Label htmlFor="verificationStatus" className="text-muted-foreground">Status</Label>
                    <Controller name="verificationStatus" control={control} render={({ field }) => (
                      <Select onValueChange={field.onChange} value={field.value || 'ALL'}>
                        <SelectTrigger id="verificationStatus" className="bg-input border-border focus:ring-primary"><SelectValue placeholder="Select status" /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="ALL">All</SelectItem>
                          <SelectItem value="CORRECT">Correct</SelectItem>
                          <SelectItem value="INCORRECT">Incorrect</SelectItem>
                          <SelectItem value="PENDING">Pending</SelectItem>
                          <SelectItem value="ERROR">Error</SelectItem>
                          <SelectItem value="PROCESSING">Processing</SelectItem>
                          <SelectItem value="COMPLETED">Completed</SelectItem>
                        </SelectContent>
                      </Select>
                    )} />
                  </div>
                  <div><Label htmlFor="machineId" className="text-muted-foreground">Vending Machine ID</Label><Input id="machineId" {...form.register('machineId')} placeholder="Enter Machine ID" className="bg-input border-border focus:ring-primary"/></div>
                  <div>
                    <Label htmlFor="verificationType" className="text-muted-foreground">Verification Type</Label>
                    <Controller name="verificationType" control={control} render={({ field }) => (
                      <Select onValueChange={field.onChange} value={field.value || 'ALL'}>
                        <SelectTrigger id="verificationType" className="bg-input border-border focus:ring-primary">
                          <SelectValue placeholder="Select verification type" />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="ALL">All Types</SelectItem>
                          <SelectItem value="LAYOUT_VS_CHECKING">Layout vs Checking</SelectItem>
                          <SelectItem value="PREVIOUS_VS_CURRENT">Previous vs Current</SelectItem>
                        </SelectContent>
                      </Select>
                    )} />
                  </div>
                  <div>
                    <Label htmlFor="dateRangePreset" className="text-muted-foreground">Date Range</Label>
                    <Controller name="dateRangePreset" control={control} render={({ field }) => (
                      <Select onValueChange={(value) => { field.onChange(value); if (value !== 'CUSTOM') {setValue('customDateFrom', undefined); setValue('customDateTo', undefined);} }} value={field.value || 'ALL_TIME'}>
                        <SelectTrigger id="dateRangePreset" className="bg-input border-border focus:ring-primary"><SelectValue placeholder="Select date range" /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="ALL_TIME">All Time</SelectItem>
                          <SelectItem value="LAST_24_HOURS">Last 24 Hours</SelectItem>
                          <SelectItem value="LAST_7_DAYS">Last 7 Days</SelectItem>
                          <SelectItem value="CUSTOM">Custom Range</SelectItem>
                        </SelectContent>
                      </Select>
                    )} />
                  </div>
                  {dateRangePreset === 'CUSTOM' && (
                    <>
                      <div>
                        <Label htmlFor="customDateFrom" className="text-muted-foreground">From</Label>
                        <Controller name="customDateFrom" control={control} render={({ field }) => (
                          <Popover><PopoverTrigger asChild>
                              <Button id="customDateFrom" variant="outline" className={cn("w-full justify-start text-left font-normal bg-input border-border hover:bg-accent/10", !field.value && "text-muted-foreground")}>
                                <CalendarIcon className="mr-2 h-4 w-4" />{field.value && isValid(field.value) ? format(field.value, 'PPP') : <span>Pick a date</span>}
                              </Button>
                          </PopoverTrigger><PopoverContent className="w-auto p-0"><Calendar mode="single" selected={field.value} onSelect={field.onChange} initialFocus /></PopoverContent></Popover>
                        )} />
                      </div>
                      <div>
                        <Label htmlFor="customDateTo" className="text-muted-foreground">To</Label>
                        <Controller name="customDateTo" control={control} render={({ field }) => (
                          <Popover><PopoverTrigger asChild>
                            <Button id="customDateTo" variant="outline" className={cn("w-full justify-start text-left font-normal bg-input border-border hover:bg-accent/10", !field.value && "text-muted-foreground")}>
                              <CalendarIcon className="mr-2 h-4 w-4" />{field.value && isValid(field.value) ? format(field.value, 'PPP') : <span>Pick a date</span>}
                            </Button>
                          </PopoverTrigger><PopoverContent className="w-auto p-0"><Calendar mode="single" selected={field.value} onSelect={field.onChange} initialFocus /></PopoverContent></Popover>
                        )} />
                      </div>
                    </>
                  )}
                  <div>
                    <Label htmlFor="sortBy" className="text-muted-foreground">Sort By</Label>
                    <Controller name="sortBy" control={control} render={({ field }) => (
                      <Select onValueChange={field.onChange} value={field.value || 'newest'}>
                        <SelectTrigger id="sortBy" className="bg-input border-border focus:ring-primary"><SelectValue placeholder="Sort by" /></SelectTrigger>
                        <SelectContent>
                          <SelectItem value="newest">Newest First (Verification At)</SelectItem>
                          <SelectItem value="oldest">Oldest First (Verification At)</SelectItem>
                          <SelectItem value="accuracy_desc">Accuracy (High-Low)</SelectItem>
                          <SelectItem value="accuracy_asc">Accuracy (Low-High)</SelectItem>
                        </SelectContent>
                      </Select>
                    )} />
                  </div>
                  <div>
                    <Label htmlFor="resultsPerPage" className="text-muted-foreground">Results per page</Label>
                    <Select onValueChange={handleResultsPerPageChange} value={String(resultsPerPage)}>
                        <SelectTrigger id="resultsPerPage" className="bg-input border-border focus:ring-primary"><SelectValue placeholder="Select count" /></SelectTrigger>
                        <SelectContent>
                            <SelectItem value="5">5</SelectItem>
                            <SelectItem value="10">10</SelectItem>
                            <SelectItem value="15">15</SelectItem>
                            <SelectItem value="20">20</SelectItem>
                            <SelectItem value="50">50</SelectItem>
                        </SelectContent>
                    </Select>
                  </div>
                  <Button type="submit" className="w-full btn-gradient" disabled={loading}>
                    {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Search className="mr-2 h-4 w-4" />}
                    Apply Filters
                  </Button>
                  <Button type="button" variant="outline" className="w-full" onClick={handleResetFilters} disabled={loading}>
                    Reset All Filters
                  </Button>
                </form>
              </div>
            </div>
          </div>

          <div className={cn("flex-1 transition-all duration-300 ease-in-out", isSidebarVisible ? "max-w-[calc(100%-20rem-1.5rem)]" : "max-w-full")}>
            <div className="mb-4">
              <Button onClick={() => setIsSidebarVisible(!isSidebarVisible)} variant="outline" className="bg-card hover:bg-accent/10 border-border" title={isSidebarVisible ? "Collapse Filters" : "Expand Filters"}>
                {isSidebarVisible ? <PanelLeft className="mr-2 h-4 w-4" /> : <PanelRight className="mr-2 h-4 w-4" />}
                {isSidebarVisible ? "Hide Filters" : "Show Filters"}
              </Button>
            </div>

            <div className="mb-4 p-4 bg-card rounded-lg border border-border shadow-md flex items-center justify-between text-sm">
              <span className="text-muted-foreground">
                Showing {loading || verifications.length === 0 ? 0 : (currentPage - 1) * resultsPerPage + 1}-
                {Math.min((currentPage - 1) * resultsPerPage + verifications.length, totalVerifications)} of {totalVerifications} total results
              </span>
              {totalPages > 1 && (
                <div className="flex items-center space-x-1">
                  <Button variant="outline" size="icon" onClick={() => handlePageChange(1)} disabled={currentPage === 1 || loading} className="h-8 w-8"><ChevronsLeft className="h-4 w-4" /></Button>
                  <Button variant="outline" size="icon" onClick={() => handlePageChange(currentPage - 1)} disabled={currentPage === 1 || loading} className="h-8 w-8"><ChevronLeft className="h-4 w-4" /></Button>
                  <Input
                    type="number"
                    min="1"
                    max={totalPages}
                    value={currentPage}
                    onChange={(e) => { const page = parseInt(e.target.value); if (page >=1 && page <= totalPages) handlePageChange(page);}}
                    className="w-16 h-8 text-center bg-input border-border"
                  />
                  <span className="px-2 text-muted-foreground">of {totalPages}</span>
                  <Button variant="outline" size="icon" onClick={() => handlePageChange(currentPage + 1)} disabled={currentPage === totalPages || loading} className="h-8 w-8"><ChevronRight className="h-4 w-4" /></Button>
                  <Button variant="outline" size="icon" onClick={() => handlePageChange(totalPages)} disabled={currentPage === totalPages || loading} className="h-8 w-8"><ChevronsRight className="h-4 w-4" /></Button>
                </div>
              )}
            </div>

            {loading && <div className="flex justify-center items-center py-20 bg-card rounded-lg border border-border"><Loader2 className="w-12 h-12 animate-spin text-primary" /> <span className="ml-4 text-lg text-muted-foreground">Loading verifications...</span></div>}
            {error && <div className="text-destructive p-6 bg-destructive/10 rounded-lg border border-destructive flex items-center justify-center"><AlertCircle className="w-8 h-8 mr-3"/> <span className="text-lg">{error}</span></div>}

            {!loading && !error && verifications && verifications.length === 0 && (
              <div className="text-muted-foreground text-center py-20 bg-card rounded-lg border border-border">
                <Search className="w-16 h-16 mx-auto mb-4 text-muted-foreground/50" />
                <p className="text-xl">No verifications found.</p>
                <p>Try adjusting your filters or performing a new lookup.</p>
              </div>
            )}

            {!loading && !error && verifications && verifications.length > 0 && (
              <div className="bg-card rounded-lg border border-border shadow-md overflow-x-auto">
                <table className="min-w-full divide-y divide-border">
                  <thead className="bg-card/50">
                    <tr>
                      <th className="px-5 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Verification ID</th>
                      <th className="px-5 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</th>
                      <th className="px-5 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Verification Type</th>
                      <th className="px-5 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Verification Date</th>
                      <th className="px-5 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Accuracy</th>
                      <th className="px-5 py-3 text-center text-xs font-medium text-muted-foreground uppercase tracking-wider">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {verifications.map(verification => (
                      <tr key={verification.verificationId} className="hover:bg-accent/5 transition-colors">
                        <td className="px-5 py-3 whitespace-nowrap text-sm font-mono text-primary hover:underline cursor-pointer" onClick={() => setSelectedVerification(verification)}>{verification.verificationId}</td>
                        <td className="px-5 py-3 whitespace-nowrap"><StatusPill status={verification.verificationStatus} /></td>
                        <td className="px-5 py-3 whitespace-nowrap text-sm text-foreground">
                           {formatVerificationType(verification.verificationType)}
                        </td>
                        <td className="px-5 py-3 whitespace-nowrap text-sm text-muted-foreground">
                           {formatDateSafe(verification.rawData?.verification_at ?? verification.rawData?.verificationAt ?? verification.rawData?.verified_at ?? verification.rawData?.verifiedAt ?? verification.rawData?.timestamp ?? verification.verificationAt)}
                        </td>
                        <td className="px-5 py-3 whitespace-nowrap text-sm text-foreground">
                            {formatPercentageValue(verification.rawData?.overall_accuracy ?? verification.rawData?.overallAccuracy ?? verification.rawData?.accuracy ?? verification.overallAccuracy)}
                        </td>
                        <td className="px-5 py-3 whitespace-nowrap text-center">
                          <Button variant="ghost" size="sm" onClick={() => setSelectedVerification(verification)} className="text-primary hover:text-primary/80">
                            <Eye className="w-4 h-4 mr-1.5" /> View Details
                          </Button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </main>

      {selectedVerification && (
        <div
            className="fixed inset-0 bg-black/80 backdrop-blur-sm z-[60] flex items-center justify-center p-4"
            onClick={() => setSelectedVerification(null)}
        >
            <div
                className="bg-card w-[95vw] h-[95vh] rounded-xl border border-border shadow-2xl flex flex-col"
                onClick={e => e.stopPropagation()}
            >
                <div className="flex items-center justify-between p-2 border-b border-border flex-shrink-0">
                    <h2 className="text-lg font-semibold text-gradient-primary">Verification Details</h2>
                    <Button onClick={() => setSelectedVerification(null)} variant="ghost" size="icon" className="text-muted-foreground hover:text-foreground">
                        <X className="h-4 w-4" />
                    </Button>
                </div>
                <div className="p-2 overflow-hidden flex-1">
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-2 h-full">
                        {/* Left Column: Image Comparison, Summary, Raw Data (Accordion) */}
                        <div className="flex flex-col space-y-3 h-full">
                            {/* Image Comparison Section */}
                            <div className="space-y-2">
                                <h3 className="text-base font-semibold text-primary-foreground text-center">Image Comparison</h3>
                                <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
                                    <div>
                                        <p className="text-center text-muted-foreground text-xs mb-1">Reference Image</p>
                                        <div className="aspect-[4/3] rounded-lg border border-border overflow-hidden">
                                            <ImageContainerWithLoading src={resolvedImageUrls.reference} alt="Reference Layout" />
                                        </div>
                                    </div>
                                    <div>
                                        <p className="text-center text-muted-foreground text-xs mb-1">Checking Image</p>
                                        <div className="aspect-[4/3] rounded-lg border border-border overflow-hidden">
                                            <ImageContainerWithLoading src={resolvedImageUrls.checking} alt="Checking Image" />
                                        </div>
                                    </div>
                                </div>
                                <Button onClick={() => setIsFullScreenViewerOpen(true)} className="w-full btn-gradient text-xs py-1" disabled={!resolvedImageUrls.reference || !resolvedImageUrls.checking || resolvedImageUrls.reference.includes('placehold.co') || resolvedImageUrls.checking.includes('placehold.co')}>
                                    <Maximize className="mr-1 h-3 w-3" /> View Full Screen Comparison
                                </Button>
                            </div>

                            {/* Summary Section */}
                            <Card className="bg-background/50 border-border">
                                <CardHeader className="pb-2 pt-3">
                                    <CardTitle className="text-base text-primary-foreground">Summary</CardTitle>
                                </CardHeader>
                                <CardContent className="space-y-2 text-sm pt-0 pb-3">
                                    <div className="space-y-2">
                                        <div className="flex justify-between items-center">
                                            <span className="text-muted-foreground">ID:</span>
                                            <span className="font-mono text-primary text-xs">{selectedVerification.verificationId}</span>
                                        </div>
                                        {selectedVerification.verificationType === 'PREVIOUS_VS_CURRENT' && previousVerificationId && (
                                            <div className="flex justify-between">
                                                <span className="text-muted-foreground">Previous ID:</span>
                                                <span className="font-mono text-muted-foreground text-xs">{previousVerificationId}</span>
                                            </div>
                                        )}
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">Verified At:</span>
                                            <span className="text-xs">{formatDateSafe(selectedVerification.rawData?.verification_at ?? selectedVerification.rawData?.verificationAt ?? selectedVerification.rawData?.verified_at ?? selectedVerification.rawData?.verifiedAt ?? selectedVerification.rawData?.timestamp ?? selectedVerification.verificationAt)}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">Verification Type:</span>
                                            <span className="text-xs">{formatVerificationType(selectedVerification.verificationType)}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">Overall Accuracy:</span>
                                            <span className="font-medium">{formatPercentageValue(selectedVerification.rawData?.overall_accuracy ?? selectedVerification.rawData?.overallAccuracy ?? selectedVerification.rawData?.accuracy ?? selectedVerification.overallAccuracy)}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">Overall Confidence:</span>
                                            <span>{formatPercentageValue(selectedVerification.rawData?.overall_confidence ?? selectedVerification.rawData?.confidence)}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">Discrepant Positions:</span>
                                            <span>{selectedVerification.rawData?.discrepant_positions ?? selectedVerification.rawData?.discrepancies ?? 'N/A'}</span>
                                        </div>
                                        <div className="flex justify-between">
                                            <span className="text-muted-foreground">Correct Positions:</span>
                                            <span>{selectedVerification.rawData?.correct_positions ?? 'N/A'}</span>
                                        </div>
                                        <div className="flex justify-between items-center">
                                            <span className="text-muted-foreground">Outcome:</span>
                                            <StatusPill status={selectedVerification.verificationStatus} />
                                        </div>
                                    </div>
                                </CardContent>
                            </Card>

                            {/* Raw Data Accordion */}
                            {selectedVerification.rawData && (
                                <Card className="bg-background/50 border-border">
                                    <Accordion type="single" collapsible className="w-full">
                                        <AccordionItem value="raw-data-item" className="border-b-0">
                                            <AccordionTrigger className="px-3 py-2 hover:no-underline">
                                                <CardTitle className="text-base text-primary-foreground">Raw Data</CardTitle>
                                            </AccordionTrigger>
                                            <AccordionContent className="px-3 pb-2">
                                                <div className="text-xs whitespace-pre-wrap font-code overflow-auto bg-black/30 p-3 rounded-md border border-border/50 text-muted-foreground max-h-[200px]">
                                                    {JSON.stringify(selectedVerification.rawData, null, 2)}
                                                </div>
                                            </AccordionContent>
                                        </AccordionItem>
                                    </Accordion>
                                </Card>
                            )}
                        </div>

                        {/* Right Column: AI-Generated Analysis */}
                        <div className="flex flex-col space-y-2 h-full min-h-0">
                          {(llmAnalysisLoading || llmAnalysisError || detailedLlmAnalysis) && (
                            <Card className="bg-background/50 border-border flex flex-col flex-grow min-h-0 overflow-hidden">
                              <CardHeader className="pb-1 pt-2 flex-shrink-0">
                                <CardTitle className="text-lg text-primary-foreground flex items-center">
                                  <Sparkles className="w-5 h-5 mr-2 text-accent"/>
                                  AI-Generated Analysis
                                </CardTitle>
                              </CardHeader>
                              <CardContent className="flex-grow p-2 min-h-0 overflow-hidden">
                                {llmAnalysisLoading ? (
                                  <div className="flex items-center justify-center text-muted-foreground py-8 h-full">
                                    <Loader2 className="w-5 h-5 animate-spin mr-2" /> Loading analysis...
                                  </div>
                                ) : llmAnalysisError ? (
                                  <div className="text-destructive p-3 bg-destructive/10 rounded-md border border-destructive/50 text-sm h-full overflow-y-auto">
                                    <AlertCircle className="w-4 h-4 inline mr-2" /> Error: {llmAnalysisError}
                                  </div>
                                ) : (
                                  <div className="text-sm font-mono text-muted-foreground bg-black/30 p-3 rounded-md whitespace-pre-wrap border border-border/50 h-full overflow-y-auto">
                                    {detailedLlmAnalysis || "No analysis available."}
                                  </div>
                                )}
                              </CardContent>
                            </Card>
                          )}

                          {/* Side-by-Side Analysis Comparison Button - Only for Previous vs Current */}
                          {selectedVerification.verificationType === 'PREVIOUS_VS_CURRENT' && (
                            <Button
                              onClick={() => setIsAnalysisComparisonOpen(true)}
                              className="w-full btn-gradient text-sm py-2 flex-shrink-0"
                              disabled={!detailedLlmAnalysis || previousAnalysisLoading || !!previousAnalysisError}
                            >
                              <FileText className="mr-2 h-4 w-4" />
                              View Analysis Comparison (Previous vs Current)
                            </Button>
                          )}
                        </div>
                    </div>
                </div>
            </div>
        </div>
      )}

      {isFullScreenViewerOpen && selectedVerification && (
        <FullScreenImageViewer
          referenceSrc={resolvedImageUrls.reference}
          checkingSrc={resolvedImageUrls.checking}
          onClose={() => setIsFullScreenViewerOpen(false)}
        />
      )}

      {/* Analysis Comparison Modal */}
      {isAnalysisComparisonOpen && selectedVerification && (
        <div className="fixed inset-0 bg-black/90 backdrop-blur-md z-[100] flex flex-col p-4">
          <div className="flex justify-between items-center mb-4 flex-shrink-0">
            <div className="flex items-center space-x-4 text-muted-foreground text-sm">
              <span className="text-lg font-semibold text-primary-foreground">Analysis Comparison: Previous vs Current</span>
              <span className="text-sm text-muted-foreground">Verification ID: {selectedVerification.verificationId}</span>
            </div>
            <Button onClick={() => setIsAnalysisComparisonOpen(false)} variant="ghost" size="icon" className="text-foreground hover:bg-white/10">
              <X className="h-6 w-6" />
            </Button>
          </div>

          <div className="flex-grow flex gap-4 overflow-hidden">
            <div className="flex gap-4 w-full h-full">
              {/* Previous Analysis */}
              <div className="w-1/2 h-full rounded-lg border border-border overflow-hidden bg-card/50 flex flex-col">
                <div className="p-4 border-b border-border bg-card/80 flex-shrink-0">
                  <h3 className="text-lg font-semibold text-primary-foreground flex items-center">
                    <FileText className="w-5 h-5 mr-2 text-muted-foreground" />
                    Previous Analysis
                  </h3>
                  {previousVerificationId && (
                    <p className="text-sm text-muted-foreground mt-1 font-mono">
                      ID: {previousVerificationId}
                    </p>
                  )}
                </div>
                <div className="flex-1 min-h-0 p-3">
                  {previousAnalysisLoading ? (
                    <div className="flex items-center justify-center text-muted-foreground py-8 h-full">
                      <Loader2 className="w-6 h-6 animate-spin mr-2" /> Loading previous analysis...
                    </div>
                  ) : previousAnalysisError ? (
                    <div className="text-destructive p-3 bg-destructive/10 rounded-md border border-destructive/50 text-sm">
                      <AlertCircle className="w-4 h-4 inline mr-2" /> Error: {previousAnalysisError}
                    </div>
                  ) : (
                    <div className="h-full overflow-y-auto">
                      <div className="text-sm font-mono text-muted-foreground bg-black/30 p-4 rounded-md whitespace-pre-wrap border border-border/50 min-h-full break-words leading-relaxed">
                        {previousAnalysis || "No previous analysis available."}
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* AI-Generated Analysis */}
              <div className="w-1/2 h-full rounded-lg border border-border overflow-hidden bg-card/50 flex flex-col">
                <div className="p-4 border-b border-border bg-card/80 flex-shrink-0">
                  <h3 className="text-lg font-semibold text-primary-foreground flex items-center">
                    <FileText className="w-5 h-5 mr-2 text-accent" />
                    AI-Generated Analysis
                  </h3>
                </div>
                <div className="flex-1 min-h-0 p-3">
                  {llmAnalysisLoading ? (
                    <div className="flex items-center justify-center text-muted-foreground py-8 h-full">
                      <Loader2 className="w-6 h-6 animate-spin mr-2" /> Loading AI analysis...
                    </div>
                  ) : llmAnalysisError ? (
                    <div className="text-destructive p-3 bg-destructive/10 rounded-md border border-destructive/50 text-sm">
                      <AlertCircle className="w-4 h-4 inline mr-2" /> Error: {llmAnalysisError}
                    </div>
                  ) : (
                    <div className="h-full overflow-y-auto">
                      <div className="text-sm font-mono text-muted-foreground bg-black/30 p-4 rounded-md whitespace-pre-wrap border border-border/50 min-h-full break-words leading-relaxed">
                        {detailedLlmAnalysis || "No AI analysis available."}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

const VerificationResultsLoading = () => (
  <div className="min-h-screen bg-background text-foreground">
    <main className="container mx-auto px-4 py-8">
      <header className="mb-10 text-center">
        <h1 className="text-4xl md:text-5xl font-bold mb-3 text-center text-gradient-primary">
          Verification Results
        </h1>
        <p className="text-lg text-muted-foreground">Review, filter, and analyze all verification outcomes.</p>
      </header>
      <div className="flex justify-center items-center py-20">
        <Loader2 className="w-12 h-12 animate-spin text-primary" />
        <span className="ml-4 text-lg text-muted-foreground">Loading verification results...</span>
      </div>
    </main>
  </div>
);

export default function VerificationResultsPage() {
  return (
    <Suspense fallback={<VerificationResultsLoading />}>
      <VerificationResultsPageContent />
    </Suspense>
  );
}

export const dynamic = 'force-dynamic';
    
