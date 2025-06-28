
"use client";

import React, { useState, useEffect, useCallback } from 'react';
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ScrollArea } from "@/components/ui/scroll-area";
import { Folder, FileText as FileIcon, ArrowUp, Home, AlertCircle, RotateCw, Eye, Loader2, Image as ImageIconLucide } from "lucide-react"; // Added ImageIconLucide
import apiClient from '@/lib/api-client';
import type { BrowserItem, BrowserResponse, BucketType, ApiError } from '@/lib/types'; // Updated types

import NextImage from 'next/image'; // Renamed to avoid conflict
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { cn } from '@/lib/utils';
// Removed appConfigInstance import as bucket name comes from API or is implicit

interface S3ImageBrowserProps {
  bucketType: BucketType;
  // onImageSelect from guide passes BrowserItem, which contains path (S3 key) and name.
  // The S3 URI is constructed by the parent if needed.
  onImageSelect: (item: BrowserItem) => void; 
  initialPath?: string;
  title?: string;
}

export function S3ImageBrowser({ bucketType, onImageSelect, initialPath = '', title }: S3ImageBrowserProps) {
  const [currentPath, setCurrentPath] = useState(initialPath);
  const [parentPath, setParentPath] = useState<string | undefined>(undefined);
  const [pathInput, setPathInput] = useState(initialPath);
  const [items, setItems] = useState<BrowserItem[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  

  const [previewImageUrl, setPreviewImageUrl] = useState<string | null>(null);
  const [isPreviewLoading, setIsPreviewLoading] = useState(false);

  const fetchItems = useCallback(async (path: string) => {
    setLoading(true);
    setError(null);
    try {
      const response: BrowserResponse = await apiClient.browseFolder(bucketType, path, true); 
      setItems(response.items || []);
      setCurrentPath(response.currentPath); 
      setParentPath(response.parentPath);
      setPathInput(response.currentPath);
    } catch (err: any) {
      const apiErr = err as ApiError;
      setError(`Failed to browse S3: ${apiErr.message || 'Unknown error'}`);
      // toast({ variant: "destructive", title: "Error browsing S3", description: apiErr.message });
      console.error("Error browsing S3", apiErr.message);
      setItems([]); 
    } finally {
      setLoading(false);
    }
  }, [bucketType]);

  useEffect(() => {
    fetchItems(currentPath);
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [bucketType]); // Fetch when bucketType changes or on initial load

  const handlePathInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPathInput(e.target.value);
  };

  const handleBrowsePath = () => {
    fetchItems(pathInput.trim());
  };

  const handleItemClick = (item: BrowserItem) => {
    if (item.type === 'folder') {
      fetchItems(item.path); // item.path is the S3 key/prefix for the directory
    } else {
      onImageSelect(item); // Pass the full BrowserItem
      // toast({ title: "File Selected", description: item.name });
      console.log("File Selected", item.name);
    }
  };

  const handlePreviewClick = async (e: React.MouseEvent, item: BrowserItem) => {
    e.stopPropagation(); // Prevent item click from triggering
    if (item.type === 'image' || item.type === 'file') { // Can attempt preview for 'file' if it's an image type
        setIsPreviewLoading(true);
        setPreviewImageUrl(null);
        try {
            const url = await apiClient.getImageUrl(item.path, bucketType, true);
            setPreviewImageUrl(url);
        } catch (previewError: any) {
            // toast({ variant: "destructive", title: "Preview Error", description: `Could not load preview: ${previewError.message}`});
            console.error("Preview Error", `Could not load preview: ${previewError.message}`);
            setPreviewImageUrl(null);
        } finally {
            setIsPreviewLoading(false);
        }
    } else {
        // toast({ title: "Cannot Preview", description: "Preview is only available for image files."});
        console.warn("Cannot Preview", "Preview is only available for image files.");
    }
  };


  const navigateUp = () => {
    if (parentPath !== undefined) {
      fetchItems(parentPath);
    }
  };

  const navigateRoot = () => {
    fetchItems('');
  };

  const filteredItems = items.filter(item =>
    item.name.toLowerCase().includes(searchTerm.toLowerCase())
  );

  const getIconForItemType = (type: BrowserItem['type']) => {
    switch (type) {
      case 'folder': return <Folder className="w-5 h-5 mr-2 text-accent" />;
      case 'image': return <ImageIconLucide className="w-5 h-5 mr-2 text-muted-foreground" />;
      case 'file': return <FileIcon className="w-5 h-5 mr-2 text-muted-foreground" />;
      default: return <FileIcon className="w-5 h-5 mr-2 text-muted-foreground" />;
    }
  };


  return (
    <Card className="w-full shadow-md bg-card border-border">
      <CardHeader className="p-4">
        <CardTitle className="flex justify-between items-center text-primary-foreground">
          <span>{title || `S3 Browser (${bucketType})`}</span>
          <Button onClick={() => fetchItems(currentPath)} variant="ghost" size="sm" disabled={loading} className="text-muted-foreground hover:text-primary-foreground">
            <RotateCw className={`w-4 h-4 ${loading ? 'animate-spin' : ''}`} />
          </Button>
        </CardTitle>
        <div className="flex gap-2 mt-2">
          <Input
            type="text"
            value={pathInput}
            onChange={handlePathInputChange}
            onKeyPress={(e) => e.key === 'Enter' && handleBrowsePath()}
            placeholder="Enter S3 path (e.g. folder/subfolder)"
            className="flex-grow bg-input border-input text-foreground"
          />
          <Button onClick={handleBrowsePath} disabled={loading} className="btn-gradient">
            {loading ? <Loader2 className="animate-spin w-4 h-4" /> : <Eye className="w-4 h-4" />}
             <span className="ml-2 hidden sm:inline">Browse</span>
          </Button>
        </div>
        <div className="flex gap-2 mt-2 items-center">
          <Button onClick={navigateRoot} variant="outline" size="sm" disabled={loading || !currentPath} className="border-input hover:bg-accent/10"><Home className="w-4 h-4 mr-1" /> Root</Button>
          <Button onClick={navigateUp} variant="outline" size="sm" disabled={loading || parentPath === undefined} className="border-input hover:bg-accent/10"><ArrowUp className="w-4 h-4 mr-1" /> Up</Button>
          <Input
            type="text"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            placeholder="Search current folder..."
            className="flex-grow bg-input border-input text-foreground"
          />
        </div>
      </CardHeader>
      <CardContent className="p-4">
        {error && <div className="text-destructive p-2 flex items-center bg-destructive/10 rounded-md"><AlertCircle className="w-4 h-4 mr-2"/> {error}</div>}
        <ScrollArea className="h-64 border border-border rounded-md p-2 bg-background">
          {loading && !filteredItems.length ? ( 
            <div className="flex items-center justify-center h-full text-muted-foreground"><Loader2 className="w-5 h-5 animate-spin mr-2"/>Loading items...</div>
          ) : !filteredItems.length && !loading ? (
            <p className="text-muted-foreground text-center p-4">No items found in '{currentPath || "root"}'.</p>
          ) : (
            <ul className="space-y-1">
              {filteredItems.map((item) => (
                <li key={item.path}> {/* Use item.path (S3 key) as key */}
                  <Button
                    variant="ghost"
                    className={cn("w-full justify-start text-left h-auto py-2 px-3 text-foreground hover:bg-accent/20",
                                 item.type === 'folder' && "font-medium"
                    )}
                    onClick={() => handleItemClick(item)}
                  >
                    <div className="flex items-center w-full">
                      {getIconForItemType(item.type)}
                      <span className="flex-grow truncate">{item.name}</span>
                      {item.type === 'image' && ( 
                        <Dialog>
                          <DialogTrigger asChild>
                            <Button variant="outline" size="sm" className="ml-2 text-xs h-6 px-2 border-input hover:bg-accent/10" onClick={(e) => handlePreviewClick(e, item)}>Preview</Button>
                          </DialogTrigger>
                          <DialogContent className="max-w-xl bg-card border-border">
                            <DialogHeader>
                              <DialogTitle className="text-primary-foreground">{item.name}</DialogTitle>
                            </DialogHeader>
                            <div className="mt-4 max-h-[70vh] overflow-auto flex items-center justify-center bg-background rounded">
                              {isPreviewLoading ? <Loader2 className="w-8 h-8 animate-spin text-primary" /> :
                               previewImageUrl ? (
                                // Use regular img tag for presigned URLs to avoid Next.js processing issues
                                previewImageUrl.includes('amazonaws.com') || previewImageUrl.includes('s3.') ? (
                                  <img
                                    src={previewImageUrl}
                                    alt={`Preview of ${item.name}`}
                                    className="max-w-full max-h-full object-contain"
                                    onError={(e) => {
                                      console.error('Failed to load image:', previewImageUrl);
                                      (e.target as HTMLImageElement).style.display = 'none';
                                    }}
                                    data-ai-hint="file preview"
                                  />
                                ) : (
                                  <NextImage src={previewImageUrl} alt={`Preview of ${item.name}`} width={600} height={400} className="object-contain" unoptimized={true} data-ai-hint="file preview"/>
                                )
                              ) : (
                                <p className="text-muted-foreground">Could not load preview.</p>
                              )}
                            </div>
                          </DialogContent>
                        </Dialog>
                      )}
                    </div>
                  </Button>
                </li>
              ))}
            </ul>
          )}
        </ScrollArea>
      </CardContent>
    </Card>
  );
}
