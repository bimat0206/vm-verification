
"use client";

import React, { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Loader2, Bot, Lightbulb, AlertCircle } from 'lucide-react';
import { predictiveHealthCheck } from '@/ai/flows/predictive-health-check'; // Genkit flow
import type { PredictiveHealthCheckInput, PredictiveHealthCheckOutput } from '@/lib/types'; // Use types from lib for consistency

const formSchema = z.object({
  historicalPerformanceData: z.string().min(50, { message: "Please provide substantial historical data (min 50 characters)." }),
});

type FormData = z.infer<typeof formSchema>;

export default function PredictiveHealthPage() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<PredictiveHealthCheckOutput | null>(null);
  const [error, setError] = useState<string | null>(null);

  const form = useForm<FormData>({
    resolver: zodResolver(formSchema),
  });

  const onSubmit = async (data: FormData) => {
    setLoading(true);
    setError(null);
    setResult(null);

    try {
      const input: PredictiveHealthCheckInput = {
        historicalPerformanceData: data.historicalPerformanceData,
      };
      const predictionOutput = await predictiveHealthCheck(input);
      setResult(predictionOutput);
    } catch (err: any) {
      console.error("Predictive Health Check Error:", err);
      setError(err.message || "An unexpected error occurred while performing the predictive health check.");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mx-auto py-8">
      <div className="flex items-center mb-8">
        <Bot className="w-10 h-10 mr-3 text-primary" />
        <h1 className="text-3xl font-bold font-headline">Predictive System Health Check</h1>
      </div>
      <p className="text-muted-foreground mb-6">
        Utilize AI to analyze historical performance data and predict potential system health degradations.
        Provide detailed historical data for the best results.
      </p>

      <Card>
        <form onSubmit={form.handleSubmit(onSubmit)}>
          <CardHeader>
            <CardTitle>Input Historical Data</CardTitle>
            <CardDescription>
              Paste historical performance data of the VMS system. Include metrics like verification success rates,
              API response times, resource utilization, error logs, etc. The more comprehensive the data, the better the prediction.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid w-full gap-1.5">
              <Label htmlFor="historicalPerformanceData">Historical Performance Data</Label>
              <Textarea
                id="historicalPerformanceData"
                placeholder="Example: 
                2023-10-01: API Latency Avg: 150ms, Verification Success: 99.5%, CPU Usage: 40%
                2023-10-02: API Latency Avg: 160ms, Verification Success: 99.2%, CPU Usage: 45%, Error Count: 5 (Type: Timeout)
                ..."
                rows={10}
                {...form.register('historicalPerformanceData')}
                className="font-code"
              />
              {form.formState.errors.historicalPerformanceData && (
                <p className="text-sm text-destructive">{form.formState.errors.historicalPerformanceData.message}</p>
              )}
            </div>
          </CardContent>
          <CardFooter>
            <Button type="submit" disabled={loading}>
              {loading ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Lightbulb className="mr-2 h-4 w-4" />}
              Analyze and Predict
            </Button>
          </CardFooter>
        </form>
      </Card>

      {loading && (
        <div className="mt-6 flex items-center justify-center text-muted-foreground">
          <Loader2 className="mr-2 h-5 w-5 animate-spin" />
          Analyzing data and generating prediction...
        </div>
      )}

      {error && (
        <Alert variant="destructive" className="mt-6">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Analysis Error</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {result && (
        <div className="mt-8 space-y-6">
          <Card className="bg-card">
            <CardHeader>
              <CardTitle className="flex items-center text-primary">
                <Bot className="w-6 h-6 mr-2" /> AI Prediction
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="whitespace-pre-wrap">{result.prediction}</p>
            </CardContent>
          </Card>

          <Card className="bg-card">
            <CardHeader>
              <CardTitle className="flex items-center text-accent">
                <Lightbulb className="w-6 h-6 mr-2" /> Recommendations
              </CardTitle>
            </CardHeader>
            <CardContent>
              <p className="whitespace-pre-wrap">{result.recommendations}</p>
            </CardContent>
          </Card>
        </div>
      )}
    </div>
  );
}
