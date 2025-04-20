/**
 * renderer.js - Canvas rendering functionality with improved fallback
 */

const fs = require('fs');
const { CANVAS } = require('./config');
const { log } = require('./common');
const { fetchImage, isValidImageUrl } = require('./services');
const { splitTextToLines } = require('./utils');

// Track canvas availability
let canvasAvailable = true;
let pngjs;
let canvasModule;

// Lazy-load canvas module with fallback
function getCanvasModule() {
  if (canvasModule) return canvasModule;
  
  if (!canvasAvailable) {
    log('Canvas module unavailable, using fallback renderer');
    return null;
  }
  
  try {
    canvasModule = require('canvas');
    log('Canvas module loaded successfully');
    return canvasModule;
  } catch (err) {
    log('Failed to load canvas module:', err.message);
    canvasAvailable = false;
    
    // Try to load pngjs as fallback
    try {
      pngjs = require('pngjs');
      log('Loaded pngjs as fallback renderer');
    } catch (pngErr) {
      log('Failed to load pngjs fallback:', pngErr.message);
    }
    
    return null;
  }
}

/**
 * Create fallback "canvas" when canvas module is unavailable
 */
function createFallbackCanvas(width, height) {
  log('Creating fallback canvas', { width, height });
  
  // Create a basic PNG with pngjs if available
  if (pngjs) {
    const { PNG } = pngjs;
    try {
      const png = new PNG({ width, height });
      // Fill with white
      for (let y = 0; y < height; y++) {
        for (let x = 0; x < width; x++) {
          const idx = (width * y + x) << 2;
          png.data[idx] = 255;       // R
          png.data[idx + 1] = 255;   // G
          png.data[idx + 2] = 255;   // B
          png.data[idx + 3] = 255;   // A
        }
      }
      
      return {
        width,
        height,
        getContext: () => ({
          scale: () => {},
          fillStyle: '',
          fillRect: () => {},
          fillText: () => {},
          textAlign: '',
          textBaseline: '',
          font: '',
          imageSmoothingEnabled: true,
          measureText: () => ({ width: 100 }),
          strokeStyle: '',
          lineWidth: 1,
          strokeRect: () => {},
          beginPath: () => {},
          moveTo: () => {},
          lineTo: () => {},
          stroke: () => {},
          drawImage: () => {}
        }),
        toBuffer: () => PNG.sync.write(png)
      };
    } catch (pngErr) {
      log('Error creating PNG:', pngErr.message);
    }
  }
  
  // Basic object that mimics canvas when all else fails
  return {
    width,
    height,
    getContext: () => ({
      scale: () => {},
      fillStyle: '',
      fillRect: () => {},
      fillText: () => {},
      textAlign: '',
      textBaseline: '',
      font: '',
      imageSmoothingEnabled: true,
      measureText: () => ({ width: 100 }),
      strokeStyle: '',
      lineWidth: 1,
      strokeRect: () => {},
      beginPath: () => {},
      moveTo: () => {},
      lineTo: () => {},
      stroke: () => {},
      drawImage: () => {}
    }),
    toBuffer: () => Buffer.from(''),
  };
}

/**
 * Create canvas with fallback mechanism
 */
function createSafeCanvas(width, height) {
  const canvas = getCanvasModule();
  if (!canvas) {
    return createFallbackCanvas(width, height);
  }
  
  try {
    const { createCanvas } = canvas;
    log('Creating canvas', { width, height });
    return createCanvas(width, height);
  } catch (err) {
    log('Error creating canvas:', err.message);
    canvasAvailable = false;
    return createFallbackCanvas(CANVAS.fallbackWidth, CANVAS.fallbackHeight);
  }
}

/**
 * Main render function with improved fallback
 */
async function renderLayout(layout) {
  log('renderLayout: started', { layoutId: layout.layoutId });

  try {
    // Check for canvas availability early
    getCanvasModule();
    
    // Get layout trays
    const trays = layout.subLayoutList?.[0]?.trayList || [];
    log(`Found ${trays.length} trays in layout`);

    // Check maximum tray limit
    if (trays.length > CANVAS.maxTrays) {
      throw new Error(`Layout exceeds maximum tray limit of ${CANVAS.maxTrays}`);
    }

    // Calculate canvas dimensions
    const numColumns = 7; // Fixed at 7 columns
    const numRows = Math.ceil(trays.length / numColumns) || 1;
    const { cell, row, header, footer, padding, titlePadding, metadataHeight } = CANVAS;
    
    // Limit size to prevent crashes
    const canvasWidth = Math.min(padding * 2 + numColumns * cell.width + (numColumns - 1) * cell.spacing, 2000);
    const canvasHeight = Math.min(
      padding * 2 +
      titlePadding +
      header.height +
      numRows * (cell.height + footer.height) +
      (numRows - 1) * row.spacing +
      footer.height +
      metadataHeight, 
      2000
    );

    // Create canvas with fallback support
    const canvas = createSafeCanvas(canvasWidth * CANVAS.scale, canvasHeight * CANVAS.scale);
    const ctx = canvas.getContext('2d');
    
    if (canvasAvailable) {
      ctx.scale(CANVAS.scale, CANVAS.scale);
      ctx.imageSmoothingEnabled = true;
      
      // Draw background
      ctx.fillStyle = 'white';
      ctx.fillRect(0, 0, canvasWidth, canvasHeight);
      
      // Set default text alignment
      ctx.textAlign = 'center';
      ctx.textBaseline = 'middle';
      
      // Draw components only if canvas is available
      await drawTitle(ctx, layout, canvasWidth, padding);
      await drawColumnHeaders(ctx, numColumns, padding, titlePadding, header.height);
      await drawRows(ctx, trays, numColumns, padding, titlePadding, header.height);
      await drawFooter(ctx, layout, canvasWidth, canvasHeight, padding, metadataHeight);
    }

    log('renderLayout: finished drawing, returning PNG buffer');
    return canvas.toBuffer('image/png');
  } catch (renderErr) {
    log('Critical error in renderLayout:', renderErr.message);
    return createErrorCanvas(renderErr, layout);
  }
}

/**
 * Draw the layout title
 */
async function drawTitle(ctx, layout, canvasWidth, padding) {
  if (!canvasAvailable) return;
  
  const title = `Kootoro Vending Machine Layout (ID: ${layout.layoutId || 'Unknown'})`;
  ctx.font = 'bold 18px Arial';
  ctx.fillStyle = 'black';
  ctx.fillText(title, canvasWidth / 2, padding);
}

/**
 * Draw column headers
 */
async function drawColumnHeaders(ctx, numColumns, padding, titlePadding, headerHeight) {
  if (!canvasAvailable) return;
  
  ctx.font = '14px Arial';
  ctx.fillStyle = 'black';
  
  for (let col = 0; col < numColumns; col++) {
    const x = padding + col * (CANVAS.cell.width + CANVAS.cell.spacing) + CANVAS.cell.width / 2;
    const y = padding + titlePadding + headerHeight / 2;
    ctx.fillText(`${col + 1}`, x, y);
  }
}

/**
 * Draw all rows and their slots
 */
async function drawRows(ctx, trays, numColumns, padding, titlePadding, headerHeight) {
  if (!canvasAvailable) return;
  
  for (let rowIdx = 0; rowIdx < trays.length; rowIdx++) {
    const tray = trays[rowIdx];
    const rowLetter = tray.trayCode || String.fromCharCode(65 + rowIdx);
    const rowY = padding + titlePadding + headerHeight + 
                 rowIdx * (CANVAS.cell.height + CANVAS.footer.height + CANVAS.row.spacing);

    // Draw row separator line if not first row
    if (rowIdx > 0) {
      const separatorY = rowY - CANVAS.row.spacing / 2;
      ctx.strokeStyle = 'rgb(200, 200, 200)';
      ctx.lineWidth = 1.0 / CANVAS.scale;
      ctx.beginPath();
      ctx.moveTo(padding, separatorY);
      ctx.lineTo(padding + numColumns * (CANVAS.cell.width + CANVAS.cell.spacing) - CANVAS.cell.spacing, separatorY);
      ctx.stroke();
    }

    // Draw row label
    ctx.font = '16px Arial';
    ctx.fillStyle = 'black';
    ctx.textAlign = 'right';
    ctx.fillText(rowLetter, padding - CANVAS.textPadding, rowY + CANVAS.cell.height / 2);
    ctx.textAlign = 'center';

    // Get slots for this tray and sort by slot number
    const slots = tray.slotList ? tray.slotList.sort((a, b) => a.slotNo - b.slotNo) : [];

    // Process images in smaller batches - only if canvas is available
    if (canvasAvailable) {
      // Fetch images in smaller batches to reduce memory usage
      const imagePromises = [];
      for (const slot of slots) {
        if (slot.productTemplateImage && isValidImageUrl(slot.productTemplateImage)) {
          imagePromises.push(fetchImage(slot.productTemplateImage));
        } else {
          imagePromises.push(Promise.resolve(null));
        }
      }
      
      const batchSize = 3; // Smaller batch size for ARM64
      const imageBuffers = [];
      
      for (let i = 0; i < imagePromises.length; i += batchSize) {
        const batchPromises = imagePromises.slice(i, i + batchSize);
        const batchResults = await Promise.all(batchPromises);
        imageBuffers.push(...batchResults);
      }

      // Draw each cell in the row
      for (let col = 0; col < numColumns; col++) {
        await drawCell(ctx, slots, rowLetter, col, rowY, imageBuffers);
      }
    }
  }
}

/**
 * Draw an individual cell
 */
async function drawCell(ctx, slots, rowLetter, col, rowY, imageBuffers) {
  if (!canvasAvailable) return;
  
  const slot = slots.find((s) => s.slotNo === col + 1);
  const cellX = CANVAS.padding + col * (CANVAS.cell.width + CANVAS.cell.spacing);

  // Draw cell background and border
  ctx.fillStyle = 'rgb(250, 250, 250)';
  ctx.fillRect(cellX, rowY, CANVAS.cell.width, CANVAS.cell.height);
  ctx.strokeStyle = 'rgb(180, 180, 180)';
  ctx.lineWidth = 1.0 / CANVAS.scale;
  ctx.strokeRect(cellX, rowY, CANVAS.cell.width, CANVAS.cell.height);

  if (!slot) return;
  
  const positionCode = `${rowLetter}${col + 1}`;
  
  // Draw position code
  ctx.textAlign = 'left';
  ctx.font = 'bold 14px Arial';
  ctx.fillStyle = 'rgb(0, 0, 150)';
  ctx.fillText(positionCode, cellX + 8, rowY + 16);
  ctx.textAlign = 'center';

  // Calculate image position
  const imgX = cellX + (CANVAS.cell.width - CANVAS.image.size) / 2;
  const imgY = rowY + (CANVAS.cell.height - CANVAS.image.size) / 2 - 10;
  
  // Get image buffer
  const imageBuffer = imageBuffers[slots.indexOf(slot)];
  
  if (imageBuffer) {
    try {
      const canvas = getCanvasModule();
      if (canvas) {
        const { loadImage } = canvas;
        const img = await loadImage(imageBuffer);
        ctx.drawImage(img, imgX, imgY, CANVAS.image.size, CANVAS.image.size);
      }
    } catch (err) {
      log('Failed to load image:', { error: err.message });
      drawImagePlaceholder(ctx, cellX, imgX, imgY);
    }
  } else {
    drawImagePlaceholder(ctx, cellX, imgX, imgY);
  }

  // Draw product name
  const nameY = imgY + CANVAS.image.size + 15;
  ctx.font = '12px Arial';
  ctx.fillStyle = 'black';
  
  let productName = slot.productTemplateName ? slot.productTemplateName.trim() : '';
  if (productName === '') productName = 'Sản phẩm';
  
  const maxWidth = CANVAS.cell.width - 20;
  const lines = splitTextToLines(ctx, productName, maxWidth);
  
  for (let i = 0; i < lines.length; i++) {
    const lineY = nameY + i * 18;
    ctx.fillText(lines[i], cellX + CANVAS.cell.width / 2, lineY);
  }
}

/**
 * Draw a placeholder for missing images
 */
function drawImagePlaceholder(ctx, cellX, imgX, imgY) {
  if (!canvasAvailable) return;
  
  ctx.fillStyle = 'rgb(240, 240, 240)';
  ctx.fillRect(imgX, imgY, CANVAS.image.size, CANVAS.image.size);
  ctx.strokeStyle = 'rgb(200, 200, 200)';
  ctx.lineWidth = 0.5 / CANVAS.scale;
  ctx.strokeRect(imgX, imgY, CANVAS.image.size, CANVAS.image.size);
  ctx.font = '10px Arial';
  ctx.fillStyle = 'rgb(150, 150, 150)';
  ctx.fillText('Image Unavailable', cellX + CANVAS.cell.width / 2, imgY + CANVAS.image.size / 2);
}

/**
 * Draw the footer with layout ID and timestamp
 */
async function drawFooter(ctx, layout, canvasWidth, canvasHeight, padding, metadataHeight) {
  if (!canvasAvailable) return;
  
  const footerText = `Kootoro Vending Machine Layout (ID: ${layout.layoutId || 'Unknown'})`;
  const footerY = canvasHeight - padding / 2 - metadataHeight;
  
  ctx.font = 'bold 18px Arial';
  ctx.fillStyle = 'black';
  ctx.fillText(footerText, canvasWidth / 2, footerY);

  const now = new Date();
  const formattedDate = now.toISOString().replace('T', ' ').substring(0, 19);
  const metadataText = `Generated at: ${formattedDate}`;
  
  ctx.font = '12px Arial';
  ctx.fillStyle = 'rgb(100, 100, 100)';
  ctx.fillText(metadataText, canvasWidth / 2, canvasHeight - 10);
}

/**
 * Create an error canvas when rendering fails
 */
function createErrorCanvas(error, layout) {
  try {
    // If canvas is available, create an error canvas
    const canvas = getCanvasModule();
    if (canvas && canvasAvailable) {
      const { createCanvas } = canvas;
      const errorCanvas = createCanvas(800, 400);
      const errorCtx = errorCanvas.getContext('2d');
      
      errorCtx.fillStyle = 'white';
      errorCtx.fillRect(0, 0, 800, 400);
      
      errorCtx.font = 'bold 20px Arial';
      errorCtx.fillStyle = 'red';
      errorCtx.textAlign = 'center';
      errorCtx.fillText('Error rendering layout', 400, 100);
      
      errorCtx.font = '16px Arial';
      errorCtx.fillStyle = 'black';
      errorCtx.fillText(`Layout ID: ${layout.layoutId || 'Unknown'}`, 400, 150);
      errorCtx.fillText(`Error: ${error.message}`, 400, 200);
      errorCtx.fillText(`Time: ${new Date().toISOString()}`, 400, 250);
      
      return errorCanvas.toBuffer('image/png');
    }
    
    // If canvas isn't available, use fallback method
    return createFallbackCanvas(800, 400).toBuffer();
    
  } catch (canvasErr) {
    log('Failed to create error canvas:', canvasErr.message);
    
    // Fallback to pngjs
    if (pngjs) {
      try {
        const { PNG } = pngjs;
        const png = new PNG({ width: 400, height: 200 });
        
        // Fill with white
        for (let y = 0; y < 200; y++) {
          for (let x = 0; x < 400; x++) {
            const idx = (400 * y + x) << 2;
            // White background
            png.data[idx] = 255;
            png.data[idx + 1] = 255;
            png.data[idx + 2] = 255;
            png.data[idx + 3] = 255;
            
            // Add red border
            if (x < 5 || x > 395 || y < 5 || y > 195) {
              png.data[idx] = 255;     // R
              png.data[idx + 1] = 0;   // G
              png.data[idx + 2] = 0;   // B
            }
          }
        }
        
        return PNG.sync.write(png);
      } catch (pngErr) {
        log('Failed to create error PNG:', pngErr.message);
      }
    }
    
    // Last resort - empty buffer
    return Buffer.from('');
  }
}

module.exports = {
  renderLayout
};