// predictive-health-check.ts
'use server';

/**
 * @fileOverview This file defines a Genkit flow for predictive health checks of the VMS system.
 *
 * - predictiveHealthCheck - A function that initiates the predictive health check process.
 * - PredictiveHealthCheckInput - The input type for the predictiveHealthCheck function.
 * - PredictiveHealthCheckOutput - The return type for the predictiveHealthCheck function.
 */

import {ai} from '@/ai/genkit';
import {z} from 'genkit';

const PredictiveHealthCheckInputSchema = z.object({
  historicalPerformanceData: z.string().describe('Historical performance data of the VMS system, including metrics like verification success rates, API response times, and resource utilization.'),
});
export type PredictiveHealthCheckInput = z.infer<typeof PredictiveHealthCheckInputSchema>;

const PredictiveHealthCheckOutputSchema = z.object({
  prediction: z.string().describe('A prediction of potential system health degradations based on the historical performance data.'),
  recommendations: z.string().describe('Recommendations for preventative measures to minimize downtime.'),
});
export type PredictiveHealthCheckOutput = z.infer<typeof PredictiveHealthCheckOutputSchema>;

export async function predictiveHealthCheck(input: PredictiveHealthCheckInput): Promise<PredictiveHealthCheckOutput> {
  return predictiveHealthCheckFlow(input);
}

const predictiveHealthCheckPrompt = ai.definePrompt({
  name: 'predictiveHealthCheckPrompt',
  input: {schema: PredictiveHealthCheckInputSchema},
  output: {schema: PredictiveHealthCheckOutputSchema},
  prompt: `You are an expert system administrator tasked with predicting potential system health degradations based on historical performance data.

  Analyze the following historical performance data of the VMS system and provide a prediction of potential system health degradations, along with recommendations for preventative measures to minimize downtime.

  Historical Performance Data:
  {{historicalPerformanceData}}

  Prediction:
  
  Recommendations:
  `,
});

const predictiveHealthCheckFlow = ai.defineFlow(
  {
    name: 'predictiveHealthCheckFlow',
    inputSchema: PredictiveHealthCheckInputSchema,
    outputSchema: PredictiveHealthCheckOutputSchema,
  },
  async input => {
    const {output} = await predictiveHealthCheckPrompt(input);
    return output!;
  }
);
