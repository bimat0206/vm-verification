package renderer

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"api_images_upload_render/config"

	"github.com/fogleman/gg"
)

type Layout struct {
	LayoutID        int64       `json:"layoutId"`
	LayoutType      *int        `json:"layoutType,omitempty"`      // Optional field from real data
	MachineModelID  *int        `json:"machineModelId,omitempty"`  // Optional field from real data
	SubLayoutList   []SubLayout `json:"subLayoutList"`
}

type SubLayout struct {
	TrayList []Tray `json:"trayList"`
}

type Tray struct {
	LayoutTrayID *int64 `json:"layoutTrayId,omitempty"` // Optional field from real data
	TrayCode     string `json:"trayCode"`
	TrayNo       int    `json:"trayNo"`
	SlotList     []Slot `json:"slotList"`
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

func RenderLayoutToBytes(layout Layout) ([]byte, error) {
	cfg := config.GetConfig()

	// Check if there are any trays to render
	if len(layout.SubLayoutList) == 0 || len(layout.SubLayoutList[0].TrayList) == 0 {
		return nil, fmt.Errorf("no trays found in layout")
	}

	trays := layout.SubLayoutList[0].TrayList
	numRows := len(trays)

	// Calculate the actual number of columns needed based on the JSON data
	actualColumns := 0
	for _, tray := range trays {
		for _, slot := range tray.SlotList {
			if slot.SlotNo > actualColumns {
				actualColumns = slot.SlotNo
			}
		}
	}
	
	// Ensure we have at least 1 column and don't exceed the maximum
	if actualColumns == 0 {
		actualColumns = 1
	}
	if actualColumns > cfg.NumColumns {
		actualColumns = cfg.NumColumns
	}

	canvasWidth := cfg.Padding*2 + float64(actualColumns)*cfg.CellWidth + float64(actualColumns-1)*cfg.CellSpacing
	canvasHeight := cfg.Padding*2 + cfg.TitlePadding + cfg.HeaderHeight +
		float64(numRows)*(cfg.CellHeight+cfg.FooterHeight) +
		float64(numRows-1)*cfg.RowSpacing +
		cfg.FooterHeight + cfg.MetadataHeight

	dc := gg.NewContext(int(canvasWidth*cfg.CanvasScale), int(canvasHeight*cfg.CanvasScale))
	dc.Scale(cfg.CanvasScale, cfg.CanvasScale)

	// Set white background
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// Load fonts
	titleFont, err := gg.LoadFontFace(cfg.BoldFontPath, cfg.TitleFontSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load title font: %v", err)
	}
	dc.SetFontFace(titleFont)

	// Draw title
	title := fmt.Sprintf("Kootoro Vending Machine Layout (ID: %d)", layout.LayoutID)
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored(title, canvasWidth/2, cfg.Padding, 0.5, 0.5)

	// Load column font
	columnFont, err := gg.LoadFontFace(cfg.FontPath, cfg.HeaderFontSize)
	if err != nil {
		return nil, fmt.Errorf("failed to load column font: %v", err)
	}
	dc.SetFontFace(columnFont)

	// Draw column numbers
	for col := 0; col < actualColumns; col++ {
		x := cfg.Padding + float64(col)*(cfg.CellWidth+cfg.CellSpacing) + cfg.CellWidth/2
		y := cfg.Padding + cfg.TitlePadding + cfg.HeaderHeight/2
		dc.DrawStringAnchored(fmt.Sprintf("%d", col+1), x, y, 0.5, 0.5)
	}

	// Draw rows
	for rowIdx, tray := range trays {
		rowLetter := tray.TrayCode
		rowY := cfg.Padding + cfg.TitlePadding + cfg.HeaderHeight + float64(rowIdx)*(cfg.CellHeight+cfg.FooterHeight+cfg.RowSpacing)

		if rowIdx > 0 {
			separatorY := rowY - cfg.RowSpacing/2
			dc.SetRGB(0.784, 0.784, 0.784)
			dc.SetLineWidth(1.0 / cfg.CanvasScale)
			dc.DrawLine(cfg.Padding, separatorY, canvasWidth-cfg.Padding, separatorY)
			dc.Stroke()
		}

		// Draw row letter
		dc.SetRGB(0, 0, 0)
		rowFont, err := gg.LoadFontFace(cfg.FontPath, 16.0)
		if err != nil {
			return nil, fmt.Errorf("failed to load row font: %v", err)
		}
		dc.SetFontFace(rowFont)
		dc.DrawStringAnchored(rowLetter, cfg.Padding-cfg.TextPadding, rowY+cfg.CellHeight/2, 1.0, 0.5)

		// Sort slots by slotNo
		sort.Slice(tray.SlotList, func(i, j int) bool {
			return tray.SlotList[i].SlotNo < tray.SlotList[j].SlotNo
		})

		// Load position font
		positionFont, err := gg.LoadFontFace(cfg.BoldFontPath, cfg.PositionFontSize)
		if err != nil {
			return nil, fmt.Errorf("failed to load position font: %v", err)
		}
		dc.SetFontFace(positionFont)

		for col := 0; col < actualColumns; col++ {
			slot := findSlotByNo(tray.SlotList, col+1)
			cellX := cfg.Padding + float64(col)*(cfg.CellWidth+cfg.CellSpacing)

			// Draw cell background
			dc.SetRGB(0.98, 0.98, 0.98)
			dc.DrawRectangle(cellX, rowY, cfg.CellWidth, cfg.CellHeight)
			dc.Fill()

			// Draw cell border
			dc.SetRGB(0.706, 0.706, 0.706)
			dc.SetLineWidth(1.0 / cfg.CanvasScale)
			dc.DrawRectangle(cellX, rowY, cfg.CellWidth, cfg.CellHeight)
			dc.Stroke()

			if slot != nil {
				// Draw position code
				positionCode := fmt.Sprintf("%s%d", rowLetter, col+1)
				dc.SetRGB(0, 0, 0.588)
				dc.DrawString(positionCode, cellX+8, rowY+16)

				// Product image
				imgX := cellX + (cfg.CellWidth-cfg.ImageSize)/2
				imgY := rowY + (cfg.CellHeight-cfg.ImageSize)/2 - 10

				// Load the image
				loadImageCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ImageLoadTimeout)*time.Second)
				img, err := loadImageFromURL(loadImageCtx, slot.ProductTemplateImage)
				cancel()
				if err != nil {
					log.Printf("Failed to load image for %s: %v", slot.Position, err)
					// Draw placeholder
					dc.SetRGB(0.941, 0.941, 0.941)
					dc.DrawRectangle(imgX, imgY, cfg.ImageSize, cfg.ImageSize)
					dc.Fill()
					dc.SetRGB(0.784, 0.784, 0.784)
					dc.SetLineWidth(0.5 / cfg.CanvasScale)
					dc.DrawRectangle(imgX, imgY, cfg.ImageSize, cfg.ImageSize)
					dc.Stroke()

					// Load placeholder font
					placeholderFont, err := gg.LoadFontFace(cfg.FontPath, 10.0)
					if err == nil {
						dc.SetFontFace(placeholderFont)
					}
					dc.SetRGB(0.588, 0.588, 0.588)
					dc.DrawStringAnchored("Image Unavailable", cellX+cfg.CellWidth/2, imgY+cfg.ImageSize/2, 0.5, 0.5)
				} else {
					// Scale and draw image
					dc.Push()
					dc.Translate(imgX, imgY)
					scaleX := cfg.ImageSize / float64(img.Bounds().Dx())
					scaleY := cfg.ImageSize / float64(img.Bounds().Dy())
					dc.Scale(scaleX, scaleY)
					dc.DrawImage(img, 0, 0)
					dc.Pop()
				}

				// Draw product name
				nameY := imgY + cfg.ImageSize + 15
				productFont, err := gg.LoadFontFace(cfg.FontPath, 12.0)
				if err != nil {
					return nil, fmt.Errorf("failed to load product font: %v", err)
				}
				dc.SetFontFace(productFont)
				dc.SetRGB(0, 0, 0)
				productName := strings.TrimSpace(slot.ProductTemplateName)
				if productName == "" {
					productName = "Sản phẩm"
				}
				maxWidth := cfg.CellWidth - 20
				lines := splitTextToLines(dc, productName, maxWidth)
				for i, line := range lines {
					lineY := nameY + float64(i)*18
					dc.DrawStringAnchored(line, cellX+cfg.CellWidth/2, lineY, 0.5, 0.5)
				}
			}
		}
	}

	// Draw footer
	footerFont, err := gg.LoadFontFace(cfg.BoldFontPath, 18.0)
	if err != nil {
		return nil, fmt.Errorf("failed to load footer font: %v", err)
	}
	dc.SetFontFace(footerFont)
	footerText := fmt.Sprintf("Kootoro Vending Machine Layout (ID: %d)", layout.LayoutID)
	footerY := canvasHeight - cfg.Padding/2 - cfg.MetadataHeight
	dc.SetRGB(0, 0, 0)
	dc.DrawStringAnchored(footerText, canvasWidth/2, footerY, 0.5, 0.5)

	// Draw metadata
	metadataFont, err := gg.LoadFontFace(cfg.FontPath, 12.0)
	if err != nil {
		return nil, fmt.Errorf("failed to load metadata font: %v", err)
	}
	dc.SetFontFace(metadataFont)
	now := time.Now()
	formattedDate := now.Format("Jan 02, 2006 15:04:05")
	metadataText := fmt.Sprintf("Generated at: %s", formattedDate)
	dc.SetRGB(0.392, 0.392, 0.392)
	dc.DrawStringAnchored(metadataText, canvasWidth/2, canvasHeight-10, 0.5, 0.5)

	// Encode the image to PNG
	var buf bytes.Buffer
	img := dc.Image()
	encoder := &png.Encoder{
		CompressionLevel: png.BestSpeed,
	}
	err = encoder.Encode(&buf, img)
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

	cacheDir := "/tmp/image_cache"
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return nil, err
	}

	hash := md5.Sum([]byte(url))
	cachedFile := filepath.Join(cacheDir, fmt.Sprintf("%x.png", hash))

	if _, err := os.Stat(cachedFile); err == nil {
		f, err := os.Open(cachedFile)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		img, _, err := image.Decode(f)
		return img, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Timeout: 20 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = os.WriteFile(cachedFile, data, 0644)
	if err != nil {
		log.Printf("Warning: Failed to write image to cache: %v", err)
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}
