'use client';

import { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Progress } from '@/components/ui/progress';
import { Badge } from '@/components/ui/badge';
import { Clock, CheckCircle, XCircle, AlertCircle, RefreshCw } from 'lucide-react';

interface VerificationMonitorProps {
  verificationId: string;
  onComplete?: (verification: any) => void;
  onError?: (error: string) => void;
}

export function VerificationMonitor({ verificationId, onComplete, onError }: VerificationMonitorProps) {
  const [status, setStatus] = useState<string>('CHECKING');
  const [progress, setProgress] = useState(0);
  const [attempt, setAttempt] = useState(0);
  const [maxAttempts, setMaxAttempts] = useState(80);
  const [estimatedTimeRemaining, setEstimatedTimeRemaining] = useState(0);
  const [lastChecked, setLastChecked] = useState<Date>(new Date());
  const [error, setError] = useState<string | null>(null);
  const [isMonitoring, setIsMonitoring] = useState(false);

  const formatTime = (seconds: number): string => {
    if (seconds < 60) return `${seconds}s`;
    const minutes = Math.floor(seconds / 60);
    const remainingSeconds = seconds % 60;
    return `${minutes}m ${remainingSeconds}s`;
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'COMPLETED':
      case 'CORRECT':
        return 'bg-green-500';
      case 'INCORRECT':
        return 'bg-yellow-500';
      case 'ERROR':
      case 'FAILED':
        return 'bg-red-500';
      case 'PENDING':
      case 'PROCESSING':
        return 'bg-blue-500';
      default:
        return 'bg-gray-500';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'COMPLETED':
      case 'CORRECT':
        return <CheckCircle className="w-4 h-4" />;
      case 'INCORRECT':
        return <AlertCircle className="w-4 h-4" />;
      case 'ERROR':
      case 'FAILED':
        return <XCircle className="w-4 h-4" />;
      case 'PENDING':
      case 'PROCESSING':
        return <RefreshCw className="w-4 h-4 animate-spin" />;
      default:
        return <Clock className="w-4 h-4" />;
    }
  };

  const checkVerificationStatus = async () => {
    try {
      setIsMonitoring(true);
      setError(null);
      
      const { default: apiClient } = await import('@/lib/api-client');
      
      // Start polling with progress updates
      const verification = await apiClient.pollVerificationResults(
        verificationId,
        (verification, nextInterval, progressInfo) => {
          setStatus(verification.verificationStatus);
          setLastChecked(new Date());
          
          if (progressInfo) {
            setAttempt(progressInfo.attempt);
            setMaxAttempts(progressInfo.maxAttempts);
            setProgress(progressInfo.progressPercent);
            setEstimatedTimeRemaining(progressInfo.estimatedTimeRemaining);
          }
        },
        80, // maxAttempts
        true // debug
      );
      
      // Verification completed
      setStatus(verification.verificationStatus);
      setProgress(100);
      setEstimatedTimeRemaining(0);
      
      if (onComplete) {
        onComplete(verification);
      }
      
    } catch (error: any) {
      console.error('Verification monitoring error:', error);
      setError(error.message);
      setProgress(0);
      
      if (onError) {
        onError(error.message);
      }
    } finally {
      setIsMonitoring(false);
    }
  };

  const startMonitoring = () => {
    checkVerificationStatus();
  };

  const isComplete = ['COMPLETED', 'CORRECT', 'INCORRECT', 'ERROR', 'FAILED'].includes(status);
  const isProcessing = ['PENDING', 'PROCESSING', 'CHECKING'].includes(status);

  return (
    <Card className="w-full max-w-2xl mx-auto">
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {getStatusIcon(status)}
          Verification Monitor
        </CardTitle>
        <p className="text-sm text-muted-foreground">
          ID: <code className="font-mono text-xs">{verificationId}</code>
        </p>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Status Badge */}
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium">Status:</span>
          <Badge className={`${getStatusColor(status)} text-white`}>
            {status}
          </Badge>
        </div>

        {/* Progress Bar */}
        {isProcessing && (
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Progress</span>
              <span>{Math.round(progress)}%</span>
            </div>
            <Progress value={progress} className="w-full" />
            <div className="flex justify-between text-xs text-muted-foreground">
              <span>Attempt {attempt} of {maxAttempts}</span>
              {estimatedTimeRemaining > 0 && (
                <span>~{formatTime(estimatedTimeRemaining)} remaining</span>
              )}
            </div>
          </div>
        )}

        {/* Last Checked */}
        <div className="text-xs text-muted-foreground">
          Last checked: {lastChecked.toLocaleTimeString()}
        </div>

        {/* Error Message */}
        {error && (
          <div className="p-3 bg-red-50 border border-red-200 rounded-md">
            <p className="text-sm text-red-800">{error}</p>
          </div>
        )}

        {/* Action Buttons */}
        <div className="flex gap-2">
          {!isMonitoring && !isComplete && (
            <Button onClick={startMonitoring} className="flex-1">
              {isProcessing ? 'Resume Monitoring' : 'Start Monitoring'}
            </Button>
          )}
          
          {isMonitoring && (
            <Button disabled className="flex-1">
              <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
              Monitoring...
            </Button>
          )}
          
          {isComplete && (
            <Button variant="outline" onClick={() => window.location.reload()} className="flex-1">
              Refresh Page
            </Button>
          )}
        </div>

        {/* Help Text */}
        {isProcessing && (
          <div className="text-xs text-muted-foreground bg-blue-50 p-3 rounded-md">
            <p className="font-medium mb-1">Verification in progress</p>
            <p>
              Complex verifications can take 5-15 minutes to complete. The system is analyzing images, 
              comparing layouts, and generating detailed reports. Please be patient.
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
