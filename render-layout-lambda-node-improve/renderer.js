/**
 * renderer.js - Canvas rendering functionality for vending machine layouts
 * 
 * Handles canvas initialization and all rendering operations.
 */

const { CANVAS } = require('./config');
const { log } = require('./common');
const { fetchImage, isValidImageUrl } = require('./services');
const { splitTextToLines } = require('./utils');

// Lazy-load canvas module
let canvasModule;
function getCanvasModule() {
  if (!canvasModule) {
    try {
      canvasModule = require('canvas');
      log('Canvas module loaded successfully');
    } catch (err) {
      log('Failed to load canvas module:', err.message, err.stack);
      throw new Error('Canvas module initialization failed');
    }
  }
  return canvasModule;
}

/**
 * Create canvas with fallback to smaller size if creation fails
 * 
 * @param {number} width - Canvas width
 * @param {number} height - Canvas height
 * @returns {Canvas} - Canvas object
 */
function createSafeCanvas(width, height) {
  const { createCanvas } = getCanvasModule();
  try {
    log('Creating canvas', { width, height });
    return createCanvas(width, height);
  } catch (err) {
    log('Error creating canvas:', err.message, err.stack);
    log('Creating fallback canvas');
    return createCanvas(CANVAS.fallbackWidth, CANVAS.fallbackHeight);
  }
}

/**
 * Main function to render a layout to a PNG image
 * 
 * @param {Object} layout - Layout data
 * @returns {Buffer} - PNG image buffer
 */
async function renderLayout(layout) {
  log('renderLayout: started', { layoutId: layout.layoutId });

  try {
    // Get layout trays
    const trays = layout.subLayoutList?.[0]?.trayList || [];
    log(`Found ${trays.length} trays in layout`);

    // Check maximum tray limit
    if (trays.length > CANVAS.maxTrays) {
      throw new Error(`Layout exceeds maximum tray limit of ${CANVAS.maxTrays}`);
    }

    // Calculate canvas dimensions
    const numColumns = 7; // Fixed at 7 columns based on Kootoro vending machine spec
    const numRows = Math.ceil(trays.length / numColumns) || 1;
    const { cell, row, header, footer, padding, titlePadding, metadataHeight } = CANVAS;
    
    // Reduce canvas size to avoid memory issues
    const canvasWidth = Math.min(padding * 2 + numColumns * cell.width + (numColumns - 1) * cell.spacing, 3000);
    const canvasHeight = Math.min(
      padding * 2 +
      titlePadding +
      header.height +
      numRows * (cell.height + footer.height) +
      (numRows - 1) * row.spacing +
      footer.height +
      metadataHeight, 
      3000
    );

    // Create canvas with scaling
    const canvas = createSafeCanvas(canvasWidth * CANVAS.scale, canvasHeight * CANVAS.scale);
    const ctx = canvas.getContext('2d');
    ctx.scale(CANVAS.scale, CANVAS.scale);
    ctx.imageSmoothingEnabled = true;
    
    // Draw background
    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);

    // Set default text alignment
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';

    // Draw title
    await drawTitle(ctx, layout, canvasWidth, padding);
    
    // Draw column headers
    await drawColumnHeaders(ctx, numColumns, padding, titlePadding, header.height);
    
    // Draw rows and slots
    await drawRows(ctx, trays, numColumns, padding, titlePadding, header.height);
    
    // Draw footer
    await drawFooter(ctx, layout, canvasWidth, canvasHeight, padding, metadataHeight);

    log('renderLayout: finished drawing, returning PNG buffer');
    return canvas.toBuffer('image/png');
  } catch (renderErr) {
    log('Critical error in renderLayout:', renderErr.message, renderErr.stack);
    return createErrorCanvas(renderErr, layout);
  }
}

/**
 * Draw the layout title
 */
async function drawTitle(ctx, layout, canvasWidth, padding) {
  const title = `Kootoro Vending Machine Layout (ID: ${layout.layoutId || 'Unknown'})`;
  ctx.font = 'bold 18px Arial';
  ctx.fillStyle = 'black';
  ctx.fillText(title, canvasWidth / 2, padding);
}

/**
 * Draw column headers
 */
async function drawColumnHeaders(ctx, numColumns, padding, titlePadding, headerHeight) {
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

    // Fetch all images in parallel with a limit to avoid memory issues
    const imagePromises = [];
    for (const slot of slots) {
      if (slot.productTemplateImage && isValidImageUrl(slot.productTemplateImage)) {
        imagePromises.push(fetchImage(slot.productTemplateImage));
      } else {
        log(`Invalid or missing image URL for product ${slot.productTemplateName}`, { position: slot.position });
        imagePromises.push(Promise.resolve(null));
      }
    }
    
    // Process images in smaller batches to reduce memory usage
    const batchSize = 5;
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

/**
 * Draw an individual cell
 */
async function drawCell(ctx, slots, rowLetter, col, rowY, imageBuffers) {
  const slot = slots.find((s) => s.slotNo === col + 1);
  const cellX = CANVAS.padding + col * (CANVAS.cell.width + CANVAS.cell.spacing);

  // Draw cell background and border
  ctx.fillStyle = 'rgb(250, 250, 250)';
  ctx.fillRect(cellX, rowY, CANVAS.cell.width, CANVAS.cell.height);
  ctx.strokeStyle = 'rgb(180, 180, 180)';
  ctx.lineWidth = 1.0 / CANVAS.scale;
  ctx.strokeRect(cellX, rowY, CANVAS.cell.width, CANVAS.cell.height);

  // If there's a product in this slot
  if (slot) {
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
        // Draw image
        log('renderLayout: loading image', { positionCode });
        const { loadImage } = getCanvasModule();
        const img = await loadImage(imageBuffer);
        ctx.drawImage(img, imgX, imgY, CANVAS.image.size, CANVAS.image.size);
        log('renderLayout: image loaded and drawn', { positionCode });
      } catch (err) {
        // Draw placeholder if image loading fails
        log('Failed to load image:', { positionCode, error: err.message });
        drawImagePlaceholder(ctx, cellX, imgX, imgY);
      }
    } else {
      // Draw placeholder if no image
      log('renderLayout: no image available', { positionCode });
      drawImagePlaceholder(ctx, cellX, imgX, imgY);
    }

    // Draw product name (supporting Vietnamese characters)
    const nameY = imgY + CANVAS.image.size + 15;
    ctx.font = '12px Arial';
    ctx.fillStyle = 'black';
    
    // Handle product name, using a default if empty
    let productName = slot.productTemplateName ? slot.productTemplateName.trim() : '';
    if (productName === '') productName = 'Sản phẩm'; // Default Vietnamese product name
    
    // Handle text wrapping
    const maxWidth = CANVAS.cell.width - 20;
    const lines = splitTextToLines(ctx, productName, maxWidth);
    
    for (let i = 0; i < lines.length; i++) {
      const lineY = nameY + i * 18;
      ctx.fillText(lines[i], cellX + CANVAS.cell.width / 2, lineY);
    }
  }
}

/**
 * Draw a placeholder for missing images
 */
function drawImagePlaceholder(ctx, cellX, imgX, imgY) {
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
    const { createCanvas } = getCanvasModule();
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
  } catch (canvasErr) {
    log('Failed to create error canvas:', canvasErr.message);
    // Return an empty buffer if even the error canvas fails
    return Buffer.from('');
  }
}

module.exports = {
  renderLayout
};