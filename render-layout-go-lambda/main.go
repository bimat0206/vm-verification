package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"io"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/fogleman/gg"
)

// Layout represents the structure of the JSON data
type Layout struct {
	LayoutID      int64       `json:"layoutId"`
	SubLayoutList []SubLayout `json:"subLayoutList"`
}

type SubLayout struct {
	TrayList []Tray `json:"trayList"`
}

type Tray struct {
	TrayCode string `json:"trayCode"`
	TrayNo   int    `json:"trayNo"`
	SlotList []Slot `json:"slotList"`
}

type Slot struct {
	VmLayoutSlotId       int    `json:"vmLayoutSlotId"`
	ProductId            int    `json:"productId"`
	ProductTemplateId    int    `json:"productTemplateId"`
	MaxQuantity          int    `json:"maxQuantity"`
	SlotNo               int    `json:"slotNo"`
	Status               int    `json:"status"`
	Position             string `json:"position"`
	SlotIndexCode        int    `json:"slotIndexCode"`
	CellNumber           int    `json:"cellNumber"`
	ProductTemplateName  string `json:"productTemplateName"`
	ProductTemplateImage string `json:"productTemplateImage"`
}

func handler(ctx context.Context, event events.EventBridgeEvent) error {
	// Extract detail as a map
	detail := make(map[string]interface{})
	if err := json.Unmarshal(event.Detail, &detail); err != nil {
		return fmt.Errorf("failed to parse event detail: %v", err)
	}
	
	// Access bucket and object information
	bucketInfo, ok := detail["bucket"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("bucket info not found in event")
	}
	
	objectInfo, ok := detail["object"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("object info not found in event")
	}
	
	bucket, ok := bucketInfo["name"].(string)
	if !ok {
		return fmt.Errorf("bucket name not found in event")
	}
	
	objectKey, ok := objectInfo["key"].(string)
	if !ok {
		return fmt.Errorf("object key not found in event")
	}

	// Check if the object is in the "raw" folder and is a JSON file
	if !strings.HasPrefix(objectKey, "raw/") || filepath.Ext(objectKey) != ".json" {
		return nil
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %v", err)
	}
	s3Client := s3.NewFromConfig(cfg)

	// Download the JSON file from S3
	getObjectOutput, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &objectKey,
	})
	if err != nil {
		return fmt.Errorf("failed to get object %s from bucket %s: %v", objectKey, bucket, err)
	}
	defer getObjectOutput.Body.Close()

	// Read the JSON data
	data, err := ioutil.ReadAll(getObjectOutput.Body)
	if err != nil {
		return fmt.Errorf("failed to read object data: %v", err)
	}

	// Parse the JSON data
	var layout Layout
	err = json.Unmarshal(data, &layout)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Render the layout to an image
	img, err := renderLayoutToBytes(layout)
	if err != nil {
		return fmt.Errorf("failed to render layout: %v", err)
	}

	// Save the image to the "processed" folder in S3
	processedKey := strings.Replace(objectKey, "raw/", "processed/", 1)
	processedKey = strings.TrimSuffix(processedKey, ".json") + ".png"
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &processedKey,
		Body:   bytes.NewReader(img),
	})
	if err != nil {
		return fmt.Errorf("failed to upload image to %s: %v", processedKey, err)
	}

	log.Printf("Generated image for layout %d and saved to %s", layout.LayoutID, processedKey)
	return nil
}

func renderLayoutToBytes(layout Layout) ([]byte, error) {
	const numColumns = 7
	const cellWidth = 150.0
	const cellHeight = 180.0
	const rowSpacing = 60.0
	const cellSpacing = 10.0
	const headerHeight = 40.0
	const footerHeight = 30.0
	const imageSize = 100.0
	const padding = 20.0
	const titlePadding = 40.0
	const textPadding = 5.0
	const metadataHeight = 20.0
	// Reduce scale to prevent memory issues
	scale := 5.0 // Changed from 6.0 to 2.0

	// Check if there are any trays to render
	if len(layout.SubLayoutList) == 0 || len(layout.SubLayoutList[0].TrayList) == 0 {
		return nil, fmt.Errorf("no trays found in layout")
	}

	trays := layout.SubLayoutList[0].TrayList
	numRows := len(trays)

	canvasWidth := padding*2 + float64(numColumns)*cellWidth + float64(numColumns-1)*cellSpacing
	canvasHeight := padding*2 + titlePadding + headerHeight +
		float64(numRows)*(cellHeight+footerHeight) +
		float64(numRows-1)*rowSpacing +
		footerHeight + metadataHeight

	// Check if canvas size is too large
	maxDimension := 4000.0 // Set a reasonable max dimension
	if canvasWidth*scale > maxDimension || canvasHeight*scale > maxDimension {
		log.Printf("Warning: Canvas size too large (%fx%f), capping dimensions", canvasWidth*scale, canvasHeight*scale)
		// Adjust scale to keep within max dimension
		scaleW := maxDimension / canvasWidth
		scaleH := maxDimension / canvasHeight
		scale = min(scale, min(scaleW, scaleH))
	}

	dc := gg.NewContext(int(canvasWidth*scale), int(canvasHeight*scale))
	dc.Scale(scale, scale)

	// Set background
	dc.SetRGB(1, 1, 1) // white
	dc.Clear()

	// Get font paths
	fontPath := "/app/fonts/DejaVuSans.ttf"
	boldFontPath := "/app/fonts/DejaVuSans-Bold.ttf"

	// Check if fonts exist, use fallback if not
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		fontPath = filepath.Join("fonts", "DejaVuSans.ttf")
	}
	if _, err := os.Stat(boldFontPath); os.IsNotExist(err) {
		boldFontPath = filepath.Join("fonts", "DejaVuSans-Bold.ttf")
		// If bold still doesn't exist, use regular font as fallback
		if _, err := os.Stat(boldFontPath); os.IsNotExist(err) {
			boldFontPath = fontPath
		}
	}

	// Check current directory for fonts
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		cwd, _ := os.Getwd()
		log.Printf("Looking for fonts in CWD: %s", cwd)
		files, _ := os.ReadDir(cwd)
		for _, file := range files {
			log.Printf("Found file: %s", file.Name())
		}
	}

	// Try to load fonts with smaller size to reduce memory usage
	titleFontSize := 18.0 * scale / 4
	if titleFontSize > 10 {
		titleFontSize = 10
	}
	
	titleFont, err := gg.LoadFontFace(boldFontPath, titleFontSize)
	if err != nil {
		log.Printf("Warning: Failed to load title font: %v", err)
		// Continue without custom font
	} else {
		dc.SetFontFace(titleFont)
	}

	// Draw title
	title := fmt.Sprintf("Kootoro Vending Machine Layout (ID: %d)", layout.LayoutID)
	dc.SetRGB(0, 0, 0) // black
	dc.DrawStringAnchored(title, canvasWidth/2, padding, 0.5, 0.5)

	// Load column font
	columnFont, err := gg.LoadFontFace(fontPath, 14)
	if err == nil {
		dc.SetFontFace(columnFont)
	}

	// Draw column numbers
	for col := 0; col < numColumns; col++ {
		x := padding + float64(col)*(cellWidth+cellSpacing) + cellWidth/2
		y := padding + titlePadding + headerHeight/2
		dc.DrawStringAnchored(fmt.Sprintf("%d", col+1), x, y, 0.5, 0.5)
	}

	// Load row font
	rowFont, err := gg.LoadFontFace(fontPath, 14)
	if err == nil {
		dc.SetFontFace(rowFont)
	}

	// Draw rows
	for rowIdx, tray := range trays {
		rowLetter := tray.TrayCode
		rowY := padding + titlePadding + headerHeight + float64(rowIdx)*(cellHeight+footerHeight+rowSpacing)

		if rowIdx > 0 {
			separatorY := rowY - rowSpacing/2
			dc.SetRGB(0.784, 0.784, 0.784) // rgb(200,200,200)
			dc.SetLineWidth(1.0 / scale)
			dc.DrawLine(padding, separatorY, canvasWidth-padding, separatorY)
			dc.Stroke()
		}

		// Draw row letter
		dc.SetRGB(0, 0, 0)
		dc.DrawStringAnchored(rowLetter, padding-textPadding, rowY+cellHeight/2, 1.0, 0.5) // right-aligned

		// Sort slots by slotNo
		sort.Slice(tray.SlotList, func(i, j int) bool {
			return tray.SlotList[i].SlotNo < tray.SlotList[j].SlotNo
		})

		// Load position font
		positionFont, err := gg.LoadFontFace(boldFontPath, 14)
		if err == nil {
			dc.SetFontFace(positionFont)
		}

		for col := 0; col < numColumns; col++ {
			slot := findSlotByNo(tray.SlotList, col+1)
			cellX := padding + float64(col)*(cellWidth+cellSpacing)

			// Draw cell background
			dc.SetRGB(0.98, 0.98, 0.98) // rgb(250,250,250)
			dc.DrawRectangle(cellX, rowY, cellWidth, cellHeight)
			dc.Fill()
			
			// Draw cell border
			dc.SetRGB(0.706, 0.706, 0.706) // rgb(180,180,180)
			dc.SetLineWidth(1.0 / scale)
			dc.DrawRectangle(cellX, rowY, cellWidth, cellHeight)
			dc.Stroke()

			if slot != nil {
				// Draw position code
				positionCode := fmt.Sprintf("%s%d", rowLetter, col+1)
				dc.SetRGB(0, 0, 0.588) // rgb(0,0,150)
				dc.DrawString(positionCode, cellX+8, rowY+16)

				// Draw image placeholder
				imgX := cellX + (cellWidth-imageSize)/2
				imgY := rowY + (cellHeight-imageSize)/2 - 10
				
				// Try to load the image with timeout and error handling
				loadImageCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				img, err := loadImageFromURL(loadImageCtx, slot.ProductTemplateImage)
				cancel()
				
				if err != nil {
					log.Printf("Failed to load image for %s: %v", slot.Position, err)
					// Draw placeholder
					dc.SetRGB(0.941, 0.941, 0.941) // rgb(240,240,240)
					dc.DrawRectangle(imgX, imgY, imageSize, imageSize)
					dc.Fill()
					dc.SetRGB(0.784, 0.784, 0.784) // rgb(200,200,200)
					dc.SetLineWidth(0.5 / scale)
					dc.DrawRectangle(imgX, imgY, imageSize, imageSize)
					dc.Stroke()
					
					// Load placeholder font
					placeholderFont, err := gg.LoadFontFace(fontPath, 14)
					if err == nil {
						dc.SetFontFace(placeholderFont)
					}
					
					dc.SetRGB(0.588, 0.588, 0.588) // rgb(150,150,150)
					dc.DrawStringAnchored("Sản phẩm", cellX+cellWidth/2, imgY+imageSize/2, 0.5, 0.5)
				} else {
					// Scale and draw image
					dc.Push()
					dc.Translate(imgX, imgY)
					scaleX := imageSize / float64(img.Bounds().Dx())
					scaleY := imageSize / float64(img.Bounds().Dy())
					dc.Scale(scaleX, scaleY)
					dc.DrawImage(img, 0, 0)
					dc.Pop()
				}

				// Load product font
				productFont, err := gg.LoadFontFace(fontPath, 14)
				if err == nil {
					dc.SetFontFace(productFont)
				}

				// Draw product name
				nameY := imgY + imageSize + 15
				dc.SetRGB(0, 0, 0)
				productName := strings.TrimSpace(slot.ProductTemplateName)
				if productName == "" {
					productName = "Sản phẩm"
				}
				maxWidth := cellWidth - 20
				lines := splitTextToLines(dc, productName, maxWidth)
				for i, line := range lines {
					lineY := nameY + float64(i)*18
					dc.DrawStringAnchored(line, cellX+cellWidth/2, lineY, 0.5, 0.5)
				}
			}
		}
	}

	// Load metadata font
	metadataFont, err := gg.LoadFontFace(fontPath, 14)
	if err == nil {
		dc.SetFontFace(metadataFont)
	}

	// Draw footer
	footerText := fmt.Sprintf("Kootoro Vending Machine Layout (ID: %d)", layout.LayoutID)
	footerY := canvasHeight - padding/2 - metadataHeight
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored(footerText, canvasWidth/2, footerY, 0.5, 0.5)

	// Draw metadata
	now := time.Now()
	formattedDate := now.Format("Jan 02, 2006 15:04:05")
	metadataText := fmt.Sprintf("Generated at: %s", formattedDate)
	dc.SetRGB(0.392, 0.392, 0.392) // rgb(100,100,100)
	dc.DrawStringAnchored(metadataText, canvasWidth/2, canvasHeight-10, 0.5, 0.5)

	// Encode the image to PNG with better compression
	var buf bytes.Buffer
	encoder := png.Encoder{CompressionLevel: png.BestCompression} // Changed from NoCompression to DefaultCompression
	err = encoder.Encode(&buf, dc.Image())
	if err != nil {
		return nil, fmt.Errorf("failed to encode PNG: %v", err)
	}

	return buf.Bytes(), nil
}

func findSlotByNo(slots []Slot, slotNo int) *Slot {
	for _, slot := range slots {
		if slot.SlotNo == slotNo {
			return &slot
		}
	}
	return nil
}

func splitTextToLines(dc *gg.Context, text string, maxWidth float64) []string {
	words := strings.Split(text, " ")
	if len(words) == 0 {
		return []string{""}
	}
	lines := []string{}
	currentLine := words[0]
	for _, word := range words[1:] {
		testLine := currentLine + " " + word
		w, _ := dc.MeasureString(testLine)
		if w <= maxWidth {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	if len(lines) > 2 {
		lines[1] = lines[1][:len(lines[1])-3] + "..."
		return lines[:2]
	}
	return lines
}

func loadImageFromURL(ctx context.Context, url string) (image.Image, error) {
	if url == "" {
		return nil, fmt.Errorf("empty image URL")
	}

	// Create a temp directory for caching
	tempDir := "/tmp/image_cache"
	err := os.MkdirAll(tempDir, 0755)
	if err != nil {
		return nil, err
	}

	// Create a cache key based on the URL
	hash := md5.Sum([]byte(url))
	cachedFile := filepath.Join(tempDir, fmt.Sprintf("%x.png", hash))

	// Check if the image is already in the cache
	if _, err := os.Stat(cachedFile); err == nil {
		// Load from cache
		f, err := os.Open(cachedFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		return img, err
	}

	// Download the image with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	
	// Limit read to 5MB to prevent memory issues
	data, err := ioutil.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))
	if err != nil {
		return nil, err
	}

	// Save to cache
	err = ioutil.WriteFile(cachedFile, data, 0644)
	if err != nil {
		log.Printf("Warning: Failed to write image to cache: %v", err)
		// Continue anyway, just won't be cached
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

// Helper function for Go 1.24
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func main() {
	lambda.Start(handler)
}