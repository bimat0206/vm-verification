"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { Home, CheckSquare, BarChart3, UploadCloud, HeartPulse, Menu, ShieldAlert, ShieldCheck, ShieldQuestion } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Sheet, SheetContent, SheetTrigger } from "@/components/ui/sheet";
import React, { useState, useEffect } from "react";
import apiClient from "@/lib/api-client"; // To fetch API health

interface NavLinkItem {
  title: string;
  href: string;
  icon: React.ElementType;
  matchPaths?: string[];
}

const navLinks: NavLinkItem[] = [
  { title: "Home", href: "/", icon: Home, matchPaths: ["/"] },
  { title: "New Verification", href: "/verification/new", icon: CheckSquare, matchPaths: ["/verification/new"] },
  { title: "Results", href: "/verifications", icon: BarChart3, matchPaths: ["/verifications"] },
  // Removed: { title: "Lookup", href: "/verification/lookup", icon: Search, matchPaths: ["/verification/lookup"] },
  { title: "Upload", href: "/tools/image-upload", icon: UploadCloud, matchPaths: ["/tools/image-upload"] },
  { title: "Health", href: "/tools/health-check", icon: HeartPulse, matchPaths: ["/tools/health-check", "/tools/predictive-health"] },
];

const ApiStatusIndicator: React.FC = () => {
  const [status, setStatus] = useState<string>("Checking...");
  const [isOnline, setIsOnline] = useState<boolean | null>(null);

  useEffect(() => {
    let isMounted = true;
    const checkHealth = async () => {
      try {
        // apiClient.healthCheck() does not take a debug parameter by default.
        // If you want to debug this specific call, you might need to temporarily modify healthCheck or makeRequest.
        const health = await apiClient.healthCheck(); 
        if (isMounted) {
          if (health.status && health.status.toLowerCase().includes("healthy")) {
            setStatus("Connected"); 
            setIsOnline(true);
          } else if (health.status) { 
            setStatus(health.status); 
            setIsOnline(false); 
          } else { 
            setStatus("Unknown"); 
            setIsOnline(false);
          }
        }
      } catch (error) { 
        if (isMounted) {
          setStatus("Failed"); 
          setIsOnline(false);
        }
      }
    };

    checkHealth();
    const intervalId = setInterval(checkHealth, 30000); 

    return () => {
      isMounted = false;
      clearInterval(intervalId);
    };
  }, []);

  let IconComponent = ShieldQuestion; // Default icon
  let textColor = "text-yellow-400"; 
  let bgColor = "bg-yellow-500/20"; 
  let dotColor = "bg-yellow-400";

  if (isOnline === true) {
    IconComponent = ShieldCheck;
    textColor = "text-green-400";
    bgColor = "bg-green-500/20";
    dotColor = "bg-green-400";
  } else if (isOnline === false) {
    IconComponent = ShieldAlert;
    textColor = "text-red-400";
    bgColor = "bg-red-500/20";
    dotColor = "bg-red-400";
  }
  
  if (status === "Checking...") {
    bgColor = "bg-muted/30";
    textColor = "text-muted-foreground";
    dotColor = "bg-muted-foreground animate-pulse";
  }


  return (
    <div className={cn(
      "flex items-center space-x-2 px-3 py-1.5 rounded-full text-xs font-medium border border-border",
      bgColor,
      textColor
    )}>
      <span className={cn("w-2 h-2 rounded-full", dotColor)}></span>
      <span>{status}</span>
    </div>
  );
};


export function TopNavBar() {
  const pathname = usePathname();
  const [mobileMenuOpen, setMobileMenuOpen] = React.useState(false);

  const isActive = (item: NavLinkItem) => {
    if (item.matchPaths?.some(p => pathname === p || (p !== "/" && pathname.startsWith(p + "/")))) return true;
    if (item.href === "/" && item.matchPaths && item.matchPaths.includes("/")) return pathname === "/";
    return pathname === item.href;
  };

  const NavLinksContent = ({isMobile = false}: {isMobile?: boolean}) => (
    <>
      {navLinks.map((item) => {
        const active = isActive(item);
        return (
          <Link 
            key={item.href} 
            href={item.href}
            className={cn(
              "flex items-center text-sm font-medium text-primary-foreground transition-all duration-200 ease-in-out px-3 py-2 rounded-lg",
              isMobile ? "w-full justify-start text-base py-3 pl-4" : "hover:bg-white/10",
              active ? "bg-primary-gradient shadow-lg text-white" : 
              "opacity-75 hover:opacity-100",
            )}
            onClick={() => isMobile && setMobileMenuOpen(false)}
          >
            <item.icon className={cn("h-5 w-5", isMobile ? "mr-3" : "mr-2")} />
            {item.title}
          </Link>
        );
      })}
    </>
  );


  return (
    <nav className="fixed top-0 left-0 right-0 z-50 h-16 bg-translucent-near-black backdrop-blur-md shadow-lg border-b border-border">
      <div className="container mx-auto flex h-full items-center justify-between px-4 md:px-6">
        <Link href="/" className="flex items-center gap-2">
          {/* Removed legacyBehavior and <a> child, Link acts as <a> */}
          <span className="text-xl font-bold text-gradient-primary">
            Vending Machine Verification Hub
          </span>
        </Link>

        {/* Desktop Navigation */}
        <div className="hidden md:flex items-center space-x-1"> 
          <NavLinksContent />
        </div>

        <div className="flex items-center gap-4">
          <ApiStatusIndicator />
          {/* Mobile Navigation Trigger */}
          <div className="md:hidden">
            <Sheet open={mobileMenuOpen} onOpenChange={setMobileMenuOpen}>
              <SheetTrigger asChild>
                <Button variant="ghost" size="icon" className="text-primary-foreground hover:bg-white/10">
                  <Menu className="h-6 w-6" />
                  <span className="sr-only">Toggle Menu</span>
                </Button>
              </SheetTrigger>
              <SheetContent side="left" className="w-3/4 bg-background border-border p-6 pt-12">
                 <Link 
                    href="/" 
                    className="flex items-center gap-2 mb-8 pl-2"
                    onClick={() => setMobileMenuOpen(false)}
                  >
                    <span className="text-xl font-bold text-gradient-primary">
                    Vending Machine Verification Hub
                    </span>
                </Link>
                <div className="flex flex-col space-y-2">
                  <NavLinksContent isMobile={true} />
                </div>
              </SheetContent>
            </Sheet>
          </div>
        </div>
      </div>
    </nav>
  );
}
