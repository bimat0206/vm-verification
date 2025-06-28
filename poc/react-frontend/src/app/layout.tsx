
import type { Metadata } from 'next';
import './globals.css';
import { Toaster } from "@/components/ui/toaster";
import { TopNavBar } from '@/components/layout/top-nav-bar';

export const metadata: Metadata = {
  title: 'VMS Frontend - Gradient Shift',
  description: 'Vending Machine Verification System with a new futuristic look.',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" className="dark" suppressHydrationWarning={true}>
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet" />
        <link href="https://fonts.googleapis.com/css2?family=Source+Code+Pro:wght@400;500&display=swap" rel="stylesheet" />
      </head>
      <body
        className="font-body antialiased bg-background text-foreground flex flex-col min-h-screen"
        suppressHydrationWarning={true}
      >
        <TopNavBar />
        <main className="flex-1 pt-16">
          {children}
        </main>
        <Toaster />
      </body>
    </html>
  );
}
