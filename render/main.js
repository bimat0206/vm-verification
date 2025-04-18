const { createCanvas, loadImage, registerFont } = require('canvas');
const fetch = (...args) => import('node-fetch').then(({default: fetch}) => fetch(...args));
const fs = require('fs').promises;
const path = require('path');
const crypto = require('crypto');

// Register custom font (e.g., Arial) for Unicode support
registerFont('arial.ttf', { family: 'Arial' });

// Main function to process all JSON files
async function main() {
    const inputDir = 'input';
    
    try {
        await fs.mkdir(inputDir, { recursive: true });
        const files = await fs.readdir(inputDir);
        for (const file of files) {
            if (file.toLowerCase().endsWith('.json')) {
                const filePath = path.join(inputDir, file);
                try {
                    await processJSONFile(filePath);
                } catch (err) {
                    console.log(`Error processing ${filePath}: ${err.message}`);
                }
            }
        }
        console.log('Processing complete.');
    } catch (err) {
        console.error(`Failed to read input directory: ${err.message}`);
        process.exit(1);
    }
}

// Process a single JSON file
async function processJSONFile(filePath) {
    const data = await fs.readFile(filePath, 'utf8');
    const layout = JSON.parse(data);

    const now = new Date();
    const dateDir = now.toISOString().split('T')[0];
    const timeDir = now.toTimeString().split(' ')[0].replace(/:/g, '-');
    const outputSubDir = path.join('output', dateDir, timeDir);
    await fs.mkdir(outputSubDir, { recursive: true });

    const outputFile = path.join(outputSubDir, `layout_${layout.layoutId}.png`);
    await renderLayout(layout, outputFile);

    console.log(`Generated image for layout ${layout.layoutId} at ${outputFile}`);
}

// Render the layout to a PNG file
async function renderLayout(layout, outputFile) {
    const numColumns = 7;
    const cellWidth = 150;
    const cellHeight = 180;
    const rowSpacing = 60;
    const cellSpacing = 10;
    const headerHeight = 40;
    const footerHeight = 30;
    const imageSize = 100;
    const padding = 20;
    const titlePadding = 40;
    const textPadding = 5;

    const trays = layout.subLayoutList[0].trayList;
    const numRows = trays.length;

    const canvasWidth = padding * 2 + numColumns * cellWidth + (numColumns - 1) * cellSpacing;
    const canvasHeight = padding * 2 + titlePadding + headerHeight +
        numRows * (cellHeight + footerHeight) +
        (numRows - 1) * rowSpacing +
        footerHeight;

    const scale = 4.0;
    const canvas = createCanvas(canvasWidth * scale, canvasHeight * scale);
    const ctx = canvas.getContext('2d');
    ctx.scale(scale, scale);
    ctx.imageSmoothingEnabled = true; // Enable anti-aliasing

    ctx.fillStyle = 'white';
    ctx.fillRect(0, 0, canvasWidth, canvasHeight);

    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';

    const title = `Kootoro Vending Machine Layout (ID: ${layout.layoutId})`;
    console.log(`Rendering title: ${title}`);
    ctx.font = '18px Arial';
    ctx.fillStyle = 'black';
    ctx.fillText(title, canvasWidth / 2, padding);

    ctx.font = '14px Arial';
    for (let col = 0; col < numColumns; col++) {
        const x = padding + col * (cellWidth + cellSpacing) + cellWidth / 2;
        const y = padding + titlePadding + headerHeight / 2;
        ctx.fillText(`${col + 1}`, x, y);
    }

    for (let rowIdx = 0; rowIdx < trays.length; rowIdx++) {
        const tray = trays[rowIdx];
        const rowLetter = tray.trayCode;
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

        const slots = tray.slotList.sort((a, b) => a.slotNo - b.slotNo);

        for (let col = 0; col < numColumns; col++) {
            const slot = slots.find(s => s.slotNo === col + 1);
            const cellX = padding + col * (cellWidth + cellSpacing);

            ctx.fillStyle = 'rgb(250, 250, 250)';
            ctx.fillRect(cellX, rowY, cellWidth, cellHeight);
            ctx.strokeStyle = 'rgb(230, 230, 230)';
            ctx.lineWidth = 0.5 / scale;
            ctx.strokeRect(cellX, rowY, cellWidth, cellHeight);

            if (slot) {
                const imgX = cellX + (cellWidth - imageSize) / 2;
                const imgY = rowY + (cellHeight - imageSize) / 2; // Center vertically
                try {
                    const img = await loadImageFromUrl(slot.productTemplateImage);
                    ctx.drawImage(img, imgX, imgY, imageSize, imageSize);
                } catch (err) {
                    console.log(`Failed to load image for ${slot.position}: ${err.message}`);
                    ctx.fillStyle = 'rgb(240, 240, 240)';
                    ctx.fillRect(imgX, imgY, imageSize, imageSize);
                    ctx.strokeStyle = 'rgb(200, 200, 200)';
                    ctx.lineWidth = 0.5 / scale;
                    ctx.strokeRect(imgX, imgY, imageSize, imageSize);
                    ctx.font = '10px Arial';
                    ctx.fillStyle = 'rgb(150, 150, 150)';
                    ctx.fillText('Image Unavailable', cellX + cellWidth / 2, imgY + imageSize / 2);
                }

                // Repositioned position code below the image
                const posY = rowY + (cellHeight - imageSize) / 2 + imageSize + 10; // 10px below image
                const cleanPosition = slot.position.length >= 2 ? slot.position.slice(0, 2) : slot.position;

                ctx.font = '14px Arial';
                const posWidth = ctx.measureText(cleanPosition).width + 20;
                ctx.fillStyle = 'rgb(220, 230, 255)'; // Light blue background
                const posHeight = 24;
                const posX = cellX + cellWidth / 2 - posWidth / 2;
                roundRect(ctx, posX, posY - posHeight / 2, posWidth, posHeight, 5);
                ctx.fill();

                ctx.fillStyle = 'rgb(0, 0, 150)'; // Dark blue text instead of white
                ctx.fillText(cleanPosition, cellX + cellWidth / 2, posY);

                // Product name below position code
                const nameY = posY + posHeight / 2 + 10; // 10px below position code
                ctx.font = '12px Arial';
                ctx.fillStyle = 'black';
                let productName = slot.productTemplateName.trim();
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

    const footerText = `Kootoro Vending Machine Layout (ID: ${layout.layoutId})`;
    console.log(`Rendering footer: ${footerText}`);
    const footerY = canvasHeight - padding / 2;
    ctx.font = '18px Arial';
    ctx.fillStyle = 'black';
    ctx.fillText(footerText, canvasWidth / 2, footerY);

    const out = require('fs').createWriteStream(outputFile);
    const stream = canvas.createPNGStream();
    stream.pipe(out);
    await new Promise(resolve => out.on('finish', resolve));
}

// Helper function to draw rounded rectangles
function roundRect(ctx, x, y, width, height, radius) {
    ctx.beginPath();
    ctx.moveTo(x + radius, y);
    ctx.lineTo(x + width - radius, y);
    ctx.quadraticCurveTo(x + width, y, x + width, y + radius);
    ctx.lineTo(x + width, y + height - radius);
    ctx.quadraticCurveTo(x + width, y + height, x + width - radius, y + height);
    ctx.lineTo(x + radius, y + height);
    ctx.quadraticCurveTo(x, y + height, x, y + height - radius);
    ctx.lineTo(x, y + radius);
    ctx.quadraticCurveTo(x, y, x + radius, y);
    ctx.closePath();
    ctx.fill();
}

// Helper function to split text into multiple lines
function splitTextToLines(ctx, text, maxWidth) {
    const width = ctx.measureText(text).width;
    if (width <= maxWidth) return [text];

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
        lines[1] = lines[1].slice(0, -3) + '...';
        return lines.slice(0, 2);
    }

    return lines;
}

// Function to download and cache product images
async function loadImageFromUrl(url) {
    if (!url) throw new Error('empty image URL');

    const cacheDir = 'image_cache';
    await fs.mkdir(cacheDir, { recursive: true });

    const urlHash = crypto.createHash('md5').update(url).digest('hex');
    const cachedFile = path.join(cacheDir, `${urlHash}.png`);

    try {
        const cachedData = await fs.readFile(cachedFile);
        return await loadImage(cachedData);
    } catch (err) {
        const response = await fetch(url, { signal: AbortSignal.timeout(20000) });
        if (!response.ok) throw new Error(`HTTP status ${response.status}`);

        const buffer = await response.buffer();
        await fs.writeFile(cachedFile, buffer);
        console.log(`Cached image to: ${cachedFile}`);
        return await loadImage(buffer);
    }
}

// Run the main function
main().catch(err => {
    console.error(err);
    process.exit(1);
});