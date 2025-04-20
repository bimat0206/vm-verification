package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"time"

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
	// Extract bucket and object key from the event
	bucket := event.Detail["bucket"].(map[string]interface{})["name"].(string)
	objectKey := event.Detail["object"].(map[string]interface{})["key"].(string)

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
	img, err := renderLayout(layout)
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

func renderLayout(layout Layout) ([]byte, error) {
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
	const scale = 4.0

	trays := layout.SubLayoutList[0].TrayList
	numRows := len(trays)

	canvasWidth := padding*2 + float64(numColumns)*cellWidth + float64(numColumns-1)*cellSpacing
	canvasHeight := padding*2 + titlePadding + headerHeight +
		float64(numRows)*(cellHeight+footerHeight) +
		float64(numRows-1)*rowSpacing +
		footerHeight + metadataHeight

	dc := gg.NewContext(int(canvasWidth*scale), int(canvasHeight*scale))
	dc.Scale(scale, scale)

	// Set background
	dc.SetRGB(1, 1, 1) // white
	dc.Clear()

	// Load fonts
	titleFont, err := gg.LoadFontFace("arialbd.ttf", 18)
	if err != nil {
		return nil, fmt.Errorf("failed to load title font: %v", err)
	}
	columnFont, err := gg.LoadFontFace("arial.ttf", 14)
	if err != nil {
		return nil, fmt.Errorf("failed to load column font: %v", err)
	}
	rowFont, err := gg.LoadFontFace("arial.ttf", 16)
	if err != nil {
		return nil, fmt.Errorf("failed to load row font: %v", err)
	}
	positionFont, err := gg.LoadFontFace("arialbd.ttf", 14)
	if err != nil {
		return nil, fmt.Errorf("failed to load position font: %v", err)
	}
	productFont, err := gg.LoadFontFace("arial.ttf", 12)
	if err != nil {
		return nil, fmt.Errorf("failed to load product font: %v", err)
	}
	placeholderFont, err := gg.LoadFontFace("arial.ttf", 10)
	if err != nil {
		return nil, fmt.Errorf("failed to load placeholder font: %v", err)
	}
	metadataFont, err := gg.LoadFontFace("arial.ttf", 12)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata font: %v", err)
	}

	// Draw title
	title := fmt.Sprintf("Kootoro Vending Machine Layout (ID: %d)", layout.LayoutID)
	dc.SetFontFace(titleFont)
	dc.SetRGB(0, 0, 0) // black
	dc.DrawStringAnchored(title, canvasWidth/2, padding, 0.5, 0.5)

	// Draw column numbers
	dc.SetFontFace(columnFont)
	for col := 0; col < numColumns; col++ {
		x := padding + float64(col)*(cellWidth+cellSpacing) + cellWidth/2
		y := padding + titlePadding + headerHeight/2
		dc.DrawStringAnchored(fmt.Sprintf("%d", col+1), x, y, 0.5, 0.5)
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
		dc.SetFontFace(rowFont)
		dc.SetRGB(0, 0, 0)
		dc.DrawStringAnchored(rowLetter, padding-textPadding, rowY+cellHeight/2, 1.0, 0.5) // right-aligned

		// Sort slots by slotNo
		sort.Slice(tray.SlotList, func(i, j int) bool {
			return tray.SlotList[i].SlotNo < tray.SlotList[j].SlotNo
		})

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
				dc.SetFontFace(positionFont)
				dc.SetRGB(0, 0, 0.588) // rgb(0,0,150)
				dc.DrawString(positionCode, cellX+8, rowY+16)

				// Draw image
				imgX := cellX + (cellWidth-imageSize)/2
				imgY := rowY + (cellHeight-imageSize)/2 - 10
				img, err := loadImageFromURL(slot.ProductTemplateImage)
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
					dc.SetFontFace(placeholderFont)
					dc.SetRGB(0.588, 0.588, 0.588) // rgb(150,150,150)
					dc.DrawStringAnchored("Image Unavailable", cellX+cellWidth/2, imgY+imageSize/2, 0.5, 0.5)
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

				// Draw product name
				nameY := imgY + imageSize + 15
				dc.SetFontFace(productFont)
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

	// Draw footer
	footerText := fmt.Sprintf("Kootoro Vending Machine Layout (ID: %d)", layout.LayoutID)
	footerY := canvasHeight - padding/2 - metadataHeight
	dc.SetFontFace(titleFont)
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored(footerText, canvasWidth/2, footerY, 0.5, 0.5)

	// Draw metadata
	now := time.Now()
	formattedDate := now.Format("Jan 02, 2006 15:04:05")
	metadataText := fmt.Sprintf("Generated at: %s", formattedDate)
	dc.SetFontFace(metadataFont)
	dc.SetRGB(0.392, 0.392, 0.392) // rgb(100,100,100)
	dc.DrawStringAnchored(metadataText, canvasWidth/2, canvasHeight-10, 0.5, 0.5)

	// Encode the image to PNG
	var buf bytes.Buffer
	err = dc.EncodePNG(&buf)
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

func loadImageFromURL(url string) (image.Image, error) {
	if url == "" {
		return nil, fmt.Errorf("empty image URL")
	}

	// Download the image with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %v", err)
	}

	// Decode the image
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}
	return img, nil
}

func main() {
	lambda.Start(handler)
}