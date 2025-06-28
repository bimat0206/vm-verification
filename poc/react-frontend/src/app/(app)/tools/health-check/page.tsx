
"use client";

import React, { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert';
import apiClient from '@/lib/api-client';
import { useAppConfig } from '@/config';
import type { ApiError } from '@/lib/types';
import { Loader2, CheckCircle, AlertCircle, Terminal, Server } from 'lucide-react';
import { cn } from '@/lib/utils';

interface HealthStatus {
  status: string;
  services?: Record<string, string>;
  details?: any;
  request?: any;
  statusCode?: number;
}

export default function HealthCheckPage() {
  const { config: appConfig, isLoading: configLoading } = useAppConfig();
  const [loading, setLoading] = useState<'direct' | 'system' | false>(false);
  const [healthResult, setHealthResult] = useState<HealthStatus | null>(null);
  const [error, setError] = useState<string | null>(null);

  const configSource = "Environment Variables / Default Config";

  const handleDirectApiCall = async () => {
    setLoading('direct');
    setError(null);
    setHealthResult(null);
    const endpoint = '/api/health';
    const requestDetails = {
      method: 'GET',
      url: `${appConfig.apiBaseUrl}${endpoint}`,
      headers: { 'X-Api-Key': '********' }
    };

    try {
      const response = await fetch(requestDetails.url, {
        method: requestDetails.method,
        headers: { 'X-Api-Key': appConfig.apiKey }
      });
      const responseBody = await response.json().catch(() => response.text());
      setHealthResult({
        status: response.ok ? 'Healthy (Direct)' : 'Unhealthy (Direct)',
        details: responseBody,
        request: requestDetails,
        statusCode: response.status,
      });
    } catch (err: any) {
      setError(`Direct API call failed: ${err.message}`);
      setHealthResult({
        status: 'Error (Direct)',
        details: { error: err.message, stack: err.stack },
        request: requestDetails,
        statusCode: err.statusCode || 500,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleSystemHealthCheck = async () => {
    setLoading('system');
    setError(null);
    setHealthResult(null);
    try {
      const result = await apiClient.healthCheck(true);
      setHealthResult(result);
    } catch (err) {
      const apiErr = err as ApiError;
      setError(`System health check failed: ${apiErr.message}`);
      setHealthResult({ status: 'Error (System)', details: apiErr.details || apiErr.message });
    } finally {
      setLoading(false);
    }
  };

  const getStatusColor = (status?: string): string => {
    if (!status) return 'text-muted-foreground';
    if (status.toLowerCase().includes('healthy')) return 'text-success';
    if (status.toLowerCase().includes('unhealthy') || status.toLowerCase().includes('degraded')) return 'text-yellow-400'; // Using a common Tailwind yellow
    if (status.toLowerCase().includes('error')) return 'text-destructive';
    return 'text-primary-foreground'; // Default for card content
  };

  // Show loading state while config is loading
  if (configLoading) {
    return (
      <div className="container mx-auto py-12 px-4">
        <div className="flex items-center justify-center min-h-[50vh]">
          <div className="text-center">
            <Loader2 className="w-8 h-8 animate-spin mx-auto mb-4 text-primary" />
            <p className="text-lg text-muted-foreground">Loading configuration...</p>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto py-12 px-4 space-y-10">
      <h1 className="text-4xl md:text-5xl font-bold text-gradient-primary text-center mb-2">System Health Check</h1>
      <p className="text-center text-lg text-secondary-foreground max-w-2xl mx-auto mb-10">
        Monitor the status of your VMS and its dependent services.
      </p>

      <Card className="shadow-xl border-border bg-card">
        <CardHeader className="p-6">
          <CardTitle className="text-2xl font-headline text-primary-foreground">Configuration Source</CardTitle>
        </CardHeader>
        <CardContent className="p-6 pt-0">
          <p className="text-secondary-foreground">System configuration is loaded from: <strong>{configSource}</strong>.</p>
        </CardContent>
      </Card>

      <Card className="shadow-xl border-border bg-card">
        <CardHeader className="p-6">
            <CardTitle className="text-2xl font-headline text-primary-foreground">Perform Health Checks</CardTitle>
        </CardHeader>
        <CardContent className="p-6 pt-0">
            <div className="grid md:grid-cols-2 gap-6">
                <Button onClick={handleDirectApiCall} disabled={!!loading} className="w-full py-6 text-lg btn-gradient">
                    {loading === 'direct' ? <Loader2 className="mr-2 h-5 w-5 animate-spin" /> : <Terminal className="mr-2 h-5 w-5" />}
                    Direct API Call to /health
                </Button>
                <Button onClick={handleSystemHealthCheck} disabled={!!loading} className="w-full py-6 text-lg btn-gradient">
                    {loading === 'system' ? <Loader2 className="mr-2 h-5 w-5 animate-spin" /> : <Server className="mr-2 h-5 w-5" />}
                    Check System Health (via API Client)
                </Button>
            </div>
        </CardContent>
      </Card>


      {error && (
        <Alert variant="destructive" className="shadow-lg">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {healthResult && (
        <Card className="shadow-xl border-border bg-card">
          <CardHeader className="p-6">
            <CardTitle className={cn("flex items-center text-xl font-headline", getStatusColor(healthResult.status))}>
              {healthResult.status.toLowerCase().includes('healthy') && <CheckCircle className="mr-2 h-6 w-6" />}
              {(healthResult.status.toLowerCase().includes('unhealthy') || healthResult.status.toLowerCase().includes('degraded') || healthResult.status.toLowerCase().includes('error')) && <AlertCircle className="mr-2 h-6 w-6" />}
              Overall Status: {healthResult.status}
              {healthResult.statusCode && ` (HTTP ${healthResult.statusCode})`}
            </CardTitle>
            <CardDescription className="text-secondary-foreground">Detailed health check results below.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4 p-6 pt-0">
            {healthResult.request && (
              <div>
                <h3 className="font-semibold text-lg text-primary-foreground mb-1">Request Details (Direct Call):</h3>
                <pre className="bg-muted p-3 rounded-md text-xs overflow-auto font-code text-muted-foreground">
                  {JSON.stringify(healthResult.request, null, 2)}
                </pre>
              </div>
            )}
            {healthResult.details && (
              <div>
                <h3 className="font-semibold text-lg text-primary-foreground mb-1">Response/Details:</h3>
                <pre className="bg-muted p-3 rounded-md text-xs overflow-auto font-code text-muted-foreground">
                  {JSON.stringify(healthResult.details, null, 2)}
                </pre>
              </div>
            )}
            {healthResult.services && Object.keys(healthResult.services).length > 0 && (
              <div>
                <h3 className="font-semibold text-lg text-primary-foreground mb-1">Backend Services Status:</h3>
                <ul className="list-disc list-inside pl-4 space-y-1 text-secondary-foreground">
                  {Object.entries(healthResult.services).map(([service, status]) => (
                    <li key={service} className={getStatusColor(status)}>
                      <strong>{service}:</strong> {status}
                    </li>
                  ))}
                </ul>
              </div>
            )}
          </CardContent>
        </Card>
      )}
    </div>
  );
}
