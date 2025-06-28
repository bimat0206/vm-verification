
"use client";

import React from 'react';
import Image from 'next/image';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge'; 
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '@/components/ui/accordion';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { CheckCircle, XCircle, AlertTriangle, Clock, Copy, Eye, FileJson, MessageSquare, Loader2 } from 'lucide-react';
import type { Verification } from '@/lib/types';
import { format, parseISO, isValid } from 'date-fns';

import { cn } from '@/lib/utils';


interface VerificationCardProps {
  verification: Verification;
}

// Ensure StatusPill uses verificationStatus from Verification type
const StatusIndicator: React.FC<{ status?: Verification['verificationStatus'] }> = ({ status }) => {
  if (!status) return <Badge variant="outline" className={cn("bg-muted text-muted-foreground")}><AlertTriangle className="w-4 h-4 mr-1" /> UNKNOWN</Badge>;
  
  let icon = <AlertTriangle className="w-4 h-4 mr-1" />;
  let badgeClass = "bg-orange-600 hover:bg-orange-700"; 

  switch (status) {
    case 'CORRECT':
      icon = <CheckCircle className="w-4 h-4 mr-1" />;
      badgeClass = "bg-green-600 hover:bg-green-700";
      break;
    case 'INCORRECT':
      icon = <XCircle className="w-4 h-4 mr-1" />;
      badgeClass = "bg-destructive hover:bg-destructive/90";
      break;
    case 'PENDING':
      icon = <Clock className="w-4 h-4 mr-1" />;
      badgeClass = "bg-yellow-500 hover:bg-yellow-600";
      break;
    case 'PROCESSING': 
      icon = <Loader2 className="w-4 h-4 mr-1 animate-spin" />;
      badgeClass = "bg-blue-500 hover:bg-blue-600";
      break;
    case 'COMPLETED': 
      icon = <CheckCircle className="w-4 h-4 mr-1" />;
      badgeClass = "bg-purple-500 hover:bg-purple-600";
      break;
    case 'ERROR': 
      icon = <AlertTriangle className="w-4 h-4 mr-1" />;
      badgeClass = "bg-orange-600 hover:bg-orange-700";
      break;
    default: 
      badgeClass = "bg-muted hover:bg-muted/80";
      break;
  }
  return <Badge variant="default" className={cn(badgeClass, "text-primary-foreground")}>{icon} {status}</Badge>;
};

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


export function VerificationCard({ verification }: VerificationCardProps) {
  

  const copyToClipboard = (text: string, type: string) => {
    navigator.clipboard.writeText(text)
      .then(() => console.log(`${type} Copied!`, text) /* toast({ title: `${type} Copied!`, description: text }) */)
      .catch(() => console.error(`Failed to copy ${type}`) /* toast({ variant: "destructive", title: `Failed to copy ${type}` }) */);
  };

  return (
    <Card className="w-full shadow-lg overflow-hidden bg-card">
      <CardHeader className="bg-card-foreground/5 p-4">
        <div className="flex justify-between items-start">
          <div>
            <CardTitle className="text-lg font-headline mb-1 text-primary-foreground">ID: {verification.verificationId}</CardTitle>
            <CardDescription className="text-xs text-muted-foreground">
              Machine: {verification.vendingMachineId || 'N/A'}
            </CardDescription>
          </div>
          <StatusIndicator status={verification.verificationStatus} />
        </div>
      </CardHeader>
      <CardContent className="p-4 space-y-3 text-card-foreground">
        <div className="text-sm space-y-1">
          <p><strong>Verified At:</strong> {formatDateSafe(verification.verificationAt)}</p>
          {verification.overallAccuracy !== undefined && <p><strong>Accuracy:</strong> {formatPercentageValue(verification.overallAccuracy)}</p>}
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <ImageDisplay title="Reference Image" s3Url={verification.referenceImageUrl} s3Path={verification.referenceImageUrl || 'N/A'} />
          <ImageDisplay title="Checking Image" s3Url={verification.checkingImageUrl} s3Path={verification.checkingImageUrl || 'N/A'}/>
        </div>
        
        <Accordion type="multiple" className="w-full">
          {(verification.turn1ProcessedPath || verification.turn2ProcessedPath || verification.llmAnalysis) && (
            <AccordionItem value="llm-analysis-info">
              <AccordionTrigger className="text-base hover:no-underline text-primary-foreground">
                <div className="flex items-center">
                  <MessageSquare className="w-5 h-5 mr-2 text-accent" />
                  LLM Analysis Info
                </div>
              </AccordionTrigger>
              <AccordionContent className="prose prose-sm dark:prose-invert max-w-none p-2 bg-background/50 rounded-md border border-border">
                {verification.llmAnalysis && <p className="text-xs text-muted-foreground">Direct Analysis: Available (view in details)</p>}
                {verification.turn1ProcessedPath && <p className="text-xs text-muted-foreground">Turn 1 Path: {verification.turn1ProcessedPath}</p>}
                {verification.turn2ProcessedPath && <p className="text-xs text-muted-foreground">Turn 2 Path: {verification.turn2ProcessedPath}</p>}
                {!verification.llmAnalysis && !(verification.turn1ProcessedPath || verification.turn2ProcessedPath) && <p className="text-xs text-muted-foreground italic">No direct analysis or S3 paths found on this record. Full details may be available in expanded view.</p>}
                 <p className="text-xs text-muted-foreground italic mt-1">Full analysis available in details view.</p>
              </AccordionContent>
            </AccordionItem>
          )}
          {verification.rawData && (
             <AccordionItem value="raw-data">
              <AccordionTrigger className="text-base hover:no-underline text-primary-foreground">
                <div className="flex items-center">
                 <FileJson className="w-5 h-5 mr-2 text-accent" />
                  Raw Data
                </div>
              </AccordionTrigger>
              <AccordionContent className="p-2 bg-background/50 rounded-md border border-border">
                <pre className="text-xs whitespace-pre-wrap font-code overflow-auto max-h-60 text-muted-foreground">{JSON.stringify(verification.rawData, null, 2)}</pre>
              </AccordionContent>
            </AccordionItem>
          )}
        </Accordion>

      </CardContent>
      <CardFooter className="bg-card-foreground/5 p-4 flex justify-end">
        <Button variant="outline" size="sm" onClick={() => copyToClipboard(verification.verificationId, "Verification ID")}>
          <Copy className="w-4 h-4 mr-2" /> Copy ID
        </Button>
      </CardFooter>
    </Card>
  );
}

const ImageDisplay: React.FC<{ title: string; s3Url?: string; s3Path: string }> = ({ title, s3Url, s3Path }) => {
  
  const effectiveS3Url = s3Url || "https://placehold.co/600x400.png?text=Image+Not+Available";
  
  const copyS3Path = (e: React.MouseEvent) => {
    e.stopPropagation(); 
    navigator.clipboard.writeText(s3Path)
      .then(() => console.log("S3 Path Copied!", s3Path) /* toast({ title: "S3 Path Copied!", description: s3Path }) */)
      .catch(() => console.error("Failed to copy S3 Path") /* toast({ variant: "destructive", title: "Failed to copy S3 Path" }) */);
  };

  return (
    <div className="space-y-2">
      <h4 className="font-semibold text-sm text-primary-foreground">{title}</h4>
      <Dialog>
        <DialogTrigger asChild>
          <div className="relative w-full h-48 rounded-md overflow-hidden cursor-pointer border border-border hover:opacity-80 transition-opacity">
            {/* Use regular img tag for S3 presigned URLs to avoid Next.js processing issues */}
            {effectiveS3Url.includes('amazonaws.com') || effectiveS3Url.includes('s3.') ? (
              <img src={effectiveS3Url} alt={title} className="w-full h-full object-contain bg-muted" data-ai-hint="machine product" />
            ) : (
              <Image src={effectiveS3Url} alt={title} fill style={{objectFit: "contain"}} className="bg-muted" unoptimized={true} data-ai-hint="machine product" />
            )}
            <div className="absolute inset-0 bg-black/20 flex items-center justify-center opacity-0 hover:opacity-100 transition-opacity">
              <Eye className="w-8 h-8 text-white" />
            </div>
          </div>
        </DialogTrigger>
        <DialogContent className="max-w-3xl max-h-[90vh] bg-card border-border">
          <DialogHeader>
            <DialogTitle className="text-primary-foreground">{title}</DialogTitle>
          </DialogHeader>
          <div className="mt-4 overflow-auto max-h-[75vh]">
             {/* Use regular img tag for S3 presigned URLs to avoid Next.js processing issues */}
             {effectiveS3Url.includes('amazonaws.com') || effectiveS3Url.includes('s3.') ? (
               <img src={effectiveS3Url} alt={title} className="max-w-full max-h-full object-contain mx-auto" data-ai-hint="machine product" />
             ) : (
               <Image src={effectiveS3Url} alt={title} width={800} height={600} className="object-contain mx-auto" unoptimized={true} data-ai-hint="machine product"/>
             )}
          </div>
        </DialogContent>
      </Dialog>
      <div className="text-xs text-muted-foreground truncate flex items-center">
        Path: {s3Path}
        <Button variant="ghost" size="icon" className="ml-1 h-5 w-5 text-muted-foreground hover:text-primary-foreground" onClick={copyS3Path}>
            <Copy className="w-3 h-3" />
        </Button>
      </div>
    </div>
  );
};
