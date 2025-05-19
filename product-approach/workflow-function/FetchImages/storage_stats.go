package main

// getDetailedStorageMethodStats returns detailed storage method statistics
func getDetailedStorageMethodStats(results ParallelFetchResults) map[string]interface{} {
	stats := map[string]interface{}{
		"totalImages": 2,
		"distribution": map[string]interface{}{
			StorageMethodInline: map[string]interface{}{
				"count": 0,
				"images": []string{},
				"totalSize": int64(0),
			},
			StorageMethodS3Temporary: map[string]interface{}{
				"count": 0,
				"images": []string{},
				"savedResponseSize": int64(0),
			},
		},
	}
	
	// Analyze reference image
	if results.ReferenceMeta.StorageMethod != "" {
		method := results.ReferenceMeta.StorageMethod
		methodStats := stats["distribution"].(map[string]interface{})[method].(map[string]interface{})
		methodStats["count"] = methodStats["count"].(int) + 1
		methodStats["images"] = append(methodStats["images"].([]string), "reference")
		
		if method == StorageMethodInline {
			methodStats["totalSize"] = methodStats["totalSize"].(int64) + int64(len(results.ReferenceMeta.Base64Data))
		} else {
			// Estimate saved response size for S3 storage
			if results.ReferenceMeta.Size > 0 {
				savedSize := int64(float64(results.ReferenceMeta.Size) * Base64ExpansionFactor)
				methodStats["savedResponseSize"] = methodStats["savedResponseSize"].(int64) + savedSize
			}
		}
	}
	
	// Analyze checking image
	if results.CheckingMeta.StorageMethod != "" {
		method := results.CheckingMeta.StorageMethod
		methodStats := stats["distribution"].(map[string]interface{})[method].(map[string]interface{})
		methodStats["count"] = methodStats["count"].(int) + 1
		methodStats["images"] = append(methodStats["images"].([]string), "checking")
		
		if method == StorageMethodInline {
			methodStats["totalSize"] = methodStats["totalSize"].(int64) + int64(len(results.CheckingMeta.Base64Data))
		} else {
			// Estimate saved response size for S3 storage
			if results.CheckingMeta.Size > 0 {
				savedSize := int64(float64(results.CheckingMeta.Size) * Base64ExpansionFactor)
				methodStats["savedResponseSize"] = methodStats["savedResponseSize"].(int64) + savedSize
			}
		}
	}
	
	// Calculate efficiency metrics
	inlineCount := 0
	s3Count := 0
	totalInlineSize := int64(0)
	totalSavedSize := int64(0)
	
	// Count inline images
	if dist, ok := stats["distribution"].(map[string]interface{}); ok {
		if inline, ok := dist[StorageMethodInline].(map[string]interface{}); ok {
			inlineCount = inline["count"].(int)
			totalInlineSize = inline["totalSize"].(int64)
		}
		
		if s3, ok := dist[StorageMethodS3Temporary].(map[string]interface{}); ok {
			s3Count = s3["count"].(int)
			totalSavedSize = s3["savedResponseSize"].(int64)
		}
	}
	
	// Add efficiency metrics
	stats["efficiency"] = map[string]interface{}{
		"inlineCount":           inlineCount,
		"s3Count":               s3Count,
		"totalInlineSize":       totalInlineSize,
		"totalSavedSize":        totalSavedSize,
		"s3StoragePercentage":   float64(s3Count) / 2.0 * 100,
		"inlineStoragePercentage": float64(inlineCount) / 2.0 * 100,
	}
	
	return stats
}
