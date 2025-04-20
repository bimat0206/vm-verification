const { S3Client, GetObjectCommand, PutObjectCommand } = require('@aws-sdk/client-s3');
const crypto = require('crypto');
const fs = require('fs');
const path = require('path');

// Lazy-load canvas to isolate potential crashes
let canvasModule;

try {
  canvasModule = require('canvas');
} catch (err) {
  console.error('Failed to load canvas module:', err.message, err.stack);
  throw new Error('Canvas module initialization failed');
} // <-- Add this closing brace


const { createCanvas, loadImage, registerFont } = canvasModule;

// Dynamic import for node-fetch
const fetch = (...args) => import('node-fetch').then(({ default: fetch }) => fetch(...args));

// Initialize S3 client
const s3Client = new S3Client({ region: process.env.AWS_REGION || 'us-east-1' });

// Logging Module
const capturedLogs = [];
function log(...args) {
  const msg = args
    .map((a) => (typeof a === 'string' ? a : JSON.stringify(a, null, 2)))
    .join(' ');
  const entry = `${new Date().toISOString()} ${msg}`;
  console.log(entry);
  capturedLogs.push(entry);
}

// Font Setup Module
function setupFonts() {
  try {
    process.env.FONTCONFIG_PATH = '/tmp/fontconfig';
    process.env.FONTCONFIG_FILE = '/tmp/fonts.conf';

    if (!fs.existsSync('/tmp/fontconfig')) {
      fs.mkdirSync('/tmp/fontconfig', { recursive: true });
      fs.chmodSync('/tmp/fontconfig', '777');
    }

    if (!fs.existsSync('/tmp/fonts.conf')) {
      const fontsConfContent = `<?xml version="1.0"?>
<!DOCTYPE fontconfig SYSTEM "fonts.dtd">
<fontconfig>
  <dir>/app/fonts</dir>
  <cachedir>/tmp/fontconfig</cachedir>
</fontconfig>`;
      fs.writeFileSync('/tmp/fonts.conf', fontsConfContent.trim());
    }

    const fontPaths = [
      '/app/fonts/dejavu-sans.ttf',
      '/app/fonts/arial.ttf',
      '/var/task/fonts/arial.ttf',
      `${process.env.LAMBDA_TASK_ROOT}/fonts/arial.ttf`,
      `${__dirname}/fonts/arial.ttf`,
      '/opt/fonts/arial.ttf'
    ];

    let fontRegistered = false;
    for (const fontPath of fontPaths) {
      try {
        log(`Attempting to register font from path: ${fontPath}`);
        if (fs.existsSync(fontPath)) {
          registerFont(fontPath, { family: 'Arial' });
          log(`Successfully registered font from ${fontPath}`);
          fontRegistered = true;
          break;
        } else {
          log(`Font file not found at ${fontPath}`);
        }
      } catch (fontErr) {
        log(`Failed to register font from ${fontPath}: ${fontErr.message}`);
      }
    }

    if (!fontRegistered) {
      log('Warning: Could not register Arial font, will use default system fonts');
    }
  } catch (err) {
    log('Font setup failed:', err.message, err.stack);
    throw err;
  }
}

// S3 Operations Module
async function downloadJsonFromS3(bucket, key) {
  try {
    log('Downloading JSON from S3', { bucket, key });
    const response = await s3Client.send(new GetObjectCommand({ Bucket: bucket, Key: key }));
    log('Received S3 object response');

    const streamToBuffer = async (stream) => {
      log('Converting S3 stream to buffer');
      const chunks = [];
      for await (const chunk of stream) {
        chunks.push(chunk);
      }
      const buffer = Buffer.concat(chunks);
      log('Stream converted to buffer', { bufferLength: buffer.length });
      return buffer;
    };

    const contentBuffer = await streamToBuffer(response.Body);
    const layout = JSON.parse(contentBuffer.toString('utf-8'));
    log('Parsed layout JSON', { layoutId: layout.layoutId });
    return layout;
  } catch (err) {
    log('Failed to download or parse JSON from S3:', err.message, err.stack);
    throw err;
  }
}

async function uploadToS3(bucket, key, body, contentType) {
  try {
    log('Uploading to S3', { bucket, key });
    await s3Client.send(
      new PutObjectCommand({
        Bucket: bucket,
        Key: key,
        Body: body,
        ContentType: contentType,
      })
    );
    log(`Uploaded to ${key}`);
  } catch (err) {
    log('Failed to upload to S3:', err.message, err.stack);
    throw err;
  }
}

// Event Parsing Module
function parseEvent(event) {
  try {
    if (!event) {
      log('Event is undefined');
      throw new Error('Event is undefined');
    }

    log('Parsing event for bucket/key', event);
    let bucket, key;

    if (event.detail && event.detail.bucket && event.detail.object) {
      bucket = event.detail.bucket.name;
      key = event.detail.object.key;
      log('Detected EventBridge S3 event structure', { bucket, key });
    } else if (event.Records && event.Records[0]) {
      bucket = event.Records[0].s3.bucket.name;
      key = decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, ' '));
      log('Detected S3 Put event structure', { bucket, key });
    } else {
      log('Event does not contain S3 object info', event);
      throw new Error('Event does not contain S3 object info');
    }

    if (!key.startsWith('raw/') || !key.endsWith('.json')) {
      log('Not a raw JSON file, skipping.', { key });
      return { status: 'skipped', reason: 'Not a raw JSON file' };
    }

    return { bucket, key };
  } catch (err) {
    log('Failed to parse event:', err.message, err.stack);
    throw err;
  }
}

// Canvas Rendering Module
async function renderLayout(layout) {
  log('renderLayout: started', { layoutId: layout.layoutId });

  try {
    const numColumns = 7;
    const cellWidth = 150,
      cellHeight = 180,
      rowSpacing = 60,
      cellSpacing = 10;
    const headerHeight = 40,
      footerHeight = 30,
      imageSize = 100,
      padding = 20,
      titlePadding = 40,
      textPadding = 5,
      metadataHeight = 20;

    log('Processing layout for trays');
    const trays = layout.subLayoutList?.[0]?.trayList || [];
    log(`Found ${trays.length} trays in layout`);

    const MAX_TRAYS = 50;
    if (trays.length > MAX_TRAYS) {
      throw new Error(`Layout exceeds maximum tray limit of ${MAX_TRAYS}`);
    }

    const numRows = Math.max(trays.length, 1);
    const canvasWidth = padding * 2 + numColumns * cellWidth + (numColumns - 1) * cellSpacing;
    const canvasHeight =
      padding * 2 +
      titlePadding +
      headerHeight +
      numRows * (cellHeight + footerHeight) +
      (numRows - 1) * rowSpacing +
      footerHeight +
      metadataHeight;

    const scale = 2.0;
    log('renderLayout: creating canvas', { canvasWidth, canvasHeight, scale });

    let canvas;
    try {
      canvas = createCanvas(canvasWidth * scale, canvasHeight * scale);
      log('Canvas created successfully');
    } catch (canvasErr) {
      log('Error creating canvas:', canvasErr.message, canvasErr.stack);
      canvas = createCanvas(800, 600);
      log('Created fallback smaller canvas');
    }

    const ctx = canvas.getContext('2d');
    ctx.scale(scale, scale);
    ctx.imageSmoothingEnabled = true;
    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);

    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';

    // Title
    const title = `Kootoro Vending Machine Layout (ID: ${layout.layoutId || 'Unknown'})`;
    ctx.font = 'bold 18px Arial';
    ctx.fillStyle = 'black';
    ctx.fillText(title, canvasWidth / 2, padding);

    // Column numbers
    ctx.font = '14px Arial';
    for (let col = 0; col < numColumns; col++) {
      const x = padding + col * (cellWidth + cellSpacing) + cellWidth / 2;
      const y = padding + titlePadding + headerHeight / 2;
      ctx.fillText(`${col + 1}`, x, y);
    }

    // Rows and slots
    for (let rowIdx = 0; rowIdx < trays.length; rowIdx++) {
      const tray = trays[rowIdx];
      const rowLetter = tray.trayCode || String.fromCharCode(65 + rowIdx);
      const rowY = padding + titlePadding + headerHeight + rowIdx * (cellHeight + footerHeight + rowSpacing);

      if (rowIdx > 0) {
        const separatorY = rowY - rowSpacing / 2;
        ctx.strokeStyle = 'rgb(200, 200, 200)';
        ctx.lineWidth = 1.0 / scale;
        ctx.beginPath();
        ctx.moveTo(padding, separatorY);
        ctx.lineTo(canvasWidth - padding, separatorY);
        ctx.stroke();
      }

      ctx.font = '16px Arial';
      ctx.fillStyle = 'black';
      ctx.textAlign = 'right';
      ctx.fillText(rowLetter, padding - textPadding, rowY + cellHeight / 2);
      ctx.textAlign = 'center';

      const slots = tray.slotList ? tray.slotList.sort((a, b) => a.slotNo - b.slotNo) : [];

      const fetchImage = async (url) => {
        try {
          const response = await fetch(url, {
            signal: AbortSignal.timeout(5000),
            headers: { 'User-Agent': 'Mozilla/5.0 Vending Machine Layout Generator', 'Accept': 'image/*' }
          });
          if (!response.ok) throw new Error(`HTTP status ${response.status}`);
          return await response.buffer();
        } catch (err) {
          log('Failed to fetch image:', { url, error: err.message });
          return null;
        }
      };

      const imagePromises = slots.map(slot => slot.productTemplateImage ? fetchImage(slot.productTemplateImage) : Promise.resolve(null));
      const imageBuffers = await Promise.all(imagePromises);

      for (let col = 0; col < numColumns; col++) {
        const slot = slots.find((s) => s.slotNo === col + 1);
        const cellX = padding + col * (cellWidth + cellSpacing);

        ctx.fillStyle = 'rgb(250, 250, 250)';
        ctx.fillRect(cellX, rowY, cellWidth, cellHeight);
        ctx.strokeStyle = 'rgb(180, 180, 180)';
        ctx.lineWidth = 1.0 / scale;
        ctx.strokeRect(cellX, rowY, cellWidth, cellHeight);

        if (slot) {
          const positionCode = `${rowLetter}${col + 1}`;
          ctx.textAlign = 'left';
          ctx.font = 'bold 14px Arial';
          ctx.fillStyle = 'rgb(0, 0, 150)';
          ctx.fillText(positionCode, cellX + 8, rowY + 16);
          ctx.textAlign = 'center';

          const imgX = cellX + (cellWidth - imageSize) / 2;
          const imgY = rowY + (cellHeight - imageSize) / 2 - 10;
          const imageBuffer = imageBuffers[slots.indexOf(slot)];
          if (imageBuffer) {
            try {
              log('renderLayout: loading image', { positionCode });
              const img = await loadImage(imageBuffer);
              ctx.drawImage(img, imgX, imgY, imageSize, imageSize);
              log('renderLayout: image loaded and drawn', { positionCode });
            } catch (err) {
              log('renderLayout: failed to load image:', { positionCode, error: err.message });
              ctx.fillStyle = 'rgb(240, 240, 240)';
              ctx.fillRect(imgX, imgY, imageSize, imageSize);
              ctx.strokeStyle = 'rgb(200, 200, 200)';
              ctx.lineWidth = 0.5 / scale;
              ctx.strokeRect(imgX, imgY, imageSize, imageSize);
              ctx.font = '10px Arial';
              ctx.fillStyle = 'rgb(150, 150, 150)';
              ctx.fillText('Image Unavailable', cellX + cellWidth / 2, imgY + imageSize / 2);
            }
          } else {
            log('renderLayout: no image available', { positionCode });
            ctx.fillStyle = 'rgb(240, 240, 240)';
            ctx.fillRect(imgX, imgY, imageSize, imageSize);
            ctx.strokeStyle = 'rgb(200, 200, 200)';
            ctx.lineWidth = 0.5 / scale;
            ctx.strokeRect(imgX, imgY, imageSize, imageSize);
            ctx.font = '10px Arial';
            ctx.fillStyle = 'rgb(150, 150, 150)';
            ctx.fillText('Image Unavailable', cellX + cellWidth / 2, imgY + imageSize / 2);
          }

          const nameY = imgY + imageSize + 15;
          ctx.font = '12px Arial';
          ctx.fillStyle = 'black';
          let productName = slot.productTemplateName ? slot.productTemplateName.trim() : '';
          if (productName === '') productName = 'Sản phẩm';
          const maxWidth = cellWidth - 20;
          const lines = splitTextToLines(ctx, productName, maxWidth);
          for (let i = 0; i < lines.length; i++) {
            const lineY = nameY + i * 18;
            ctx.fillText(lines[i], cellX + cellWidth / 2, lineY);
          }
        }
      }
    }

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

    log('renderLayout: finished drawing, returning PNG buffer');
    return canvas.toBuffer('image/png');
  } catch (renderErr) {
    log('Critical error in renderLayout:', renderErr.message, renderErr.stack);
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
    errorCtx.fillText(`Error: ${renderErr.message}`, 400, 200);
    errorCtx.fillText(`Time: ${new Date().toISOString()}`, 400, 250);
    return errorCanvas.toBuffer('image/png');
  }
}

function splitTextToLines(ctx, text, maxWidth) {
  if (!text) return [''];
  if (ctx.measureText(text).width <= maxWidth) return [text];
  const words = text.split(' ');
  if (words.length === 0) return [''];
  const lines = [];
  let currentLine = words[0];
  for (let i = 1; i < words.length; i++) {
    const word = words[i];
    const testLine = `${currentLine} ${word}`;
    if (ctx.measureText(testLine).width <= maxWidth) {
      currentLine = testLine;
    } else {
      lines.push(currentLine);
      currentLine = word;
    }
  }
  if (currentLine) lines.push(currentLine);
  if (lines.length > 2) {
    lines[1] = `${lines[1].slice(0, -3)}...`;
    return lines.slice(0, 2);
  }
  return lines;
}

// Main Handler
exports.handler = async (event) => {
  log('Lambda handler started');
  try {
    // Initialize fonts
    setupFonts();

    // Parse event
    const parseResult = parseEvent(event);
    if (parseResult.status === 'skipped') {
      return parseResult;
    }
    const { bucket, key } = parseResult;

    // Download and parse JSON
    const layout = await downloadJsonFromS3(bucket, key);

    // Render PNG
    const pngBuffer = await renderLayout(layout);

    // Upload PNG
    const layoutId = layout.layoutId || crypto.randomBytes(4).toString('hex');
    const outputKey = `rendered-layout/layout_${layoutId}.png`;
    await uploadToS3(bucket, outputKey, pngBuffer, 'image/png');

    // Upload logs
    const logKey = `logs/layout_${layoutId}_${Date.now()}.log`;
    await uploadToS3(bucket, logKey, capturedLogs.join('\n'), 'text/plain');

    log('Lambda handler completed successfully');
    return { status: 'success', outputKey };
  } catch (err) {
    log('Error in Lambda handler:', err.message, err.stack);
    try {
      const targetBucket = process.env.LOG_BUCKET || process.env.BUCKET;
      const fallbackKey = `logs/error_${Date.now()}.log`;
      if (targetBucket) {
        await uploadToS3(targetBucket, fallbackKey, capturedLogs.join('\n') + `\n${err.stack}`, 'text/plain');
      } else {
        log('No S3 bucket specified for error log upload, skipping');
      }
    } catch (uploadErr) {
      log('Failed to upload error log:', uploadErr.message);
    }
    throw err;
  }
};