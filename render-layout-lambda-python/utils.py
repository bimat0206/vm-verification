import io
import os
import math
import logging
import requests
from datetime import datetime
from PIL import Image, ImageDraw, ImageFont

logger = logging.getLogger(__name__)

def split_text_to_lines(draw, text, max_width, font):
    """Split text into lines that fit within max_width"""
    if not text:
        return ['']
    
    # Check if the entire text fits
    text_width = draw.textlength(text, font=font)
    if text_width <= max_width:
        return [text]
    
    words = text.split(' ')
    if not words:
        return ['']
    
    lines = []
    current_line = words[0]
    
    for i in range(1, len(words)):
        word = words[i]
        test_line = f"{current_line} {word}"
        test_width = draw.textlength(test_line, font=font)
        
        if test_width <= max_width:
            current_line = test_line
        else:
            lines.append(current_line)
            current_line = word
    
    if current_line:
        lines.append(current_line)
    
    # Truncate to 2 lines with ellipsis if needed
    if len(lines) > 2:
        if len(lines[1]) > 3:
            lines[1] = f"{lines[1][:-3]}..."
        return lines[:2]
    
    return lines

def load_and_resize_image(url, target_size):
    """Load an image from URL and resize it while maintaining aspect ratio"""
    try:
        # Get image data
        response = requests.get(url, timeout=10)
        response.raise_for_status()
        
        # Open image from binary data
        img = Image.open(io.BytesIO(response.content))
        
        # Calculate resize dimensions while maintaining aspect ratio
        width, height = img.size
        if width > height:
            new_width = target_size
            new_height = int(height * (target_size / width))
        else:
            new_height = target_size
            new_width = int(width * (target_size / height))
        
        # Resize with high quality
        img = img.resize((new_width, new_height), Image.LANCZOS)
        
        # Convert to RGBA if needed
        if img.mode != 'RGBA':
            img = img.convert('RGBA')
        
        return img
    except Exception as e:
        logger.error(f"Error processing image from {url}: {e}")
        # Return a placeholder image
        placeholder = Image.new('RGBA', (int(target_size), int(target_size)), color=(200, 200, 200, 128))
        draw = ImageDraw.Draw(placeholder)
        draw.text((10, target_size/2), "Image\nError", fill=(80, 80, 80))
        return placeholder

def render_layout(layout, get_font_func=None):
    """Render layout to PNG image with improved text and image alignment"""
    logger.info(f"renderLayout: started - layoutId: {layout.get('layoutId')}")
    
    try:
        # Define layout parameters
        num_columns = 7
        cell_width, cell_height = 150, 180
        row_spacing, cell_spacing = 60, 10
        header_height, footer_height = 40, 30
        image_size = 100
        padding = 20
        title_padding = 40
        text_padding = 5
        metadata_height = 20
        
        logger.info("Processing layout for trays")
        # Safer extraction of trayList
        subLayoutList = layout.get('subLayoutList', [])
        trays = []
        if subLayoutList and len(subLayoutList) > 0:
            trays = subLayoutList[0].get('trayList', [])
        logger.info(f"Found {len(trays)} trays in layout")
        
        # Calculate dimensions
        num_rows = max(len(trays), 1)  # Ensure at least 1 row
        canvas_width = padding * 2 + num_columns * cell_width + (num_columns - 1) * cell_spacing
        canvas_height = (
            padding * 2 +
            title_padding +
            header_height +
            num_rows * (cell_height + footer_height) +
            (num_rows - 1) * row_spacing +
            footer_height +
            metadata_height
        )
        
        scale = 3.0  # High resolution scale factor
        logger.info(f"renderLayout: creating canvas - width: {canvas_width}, height: {canvas_height}, scale: {scale}")
        
        # Create the canvas with white background
        try:
            img = Image.new('RGB', (int(canvas_width * scale), int(canvas_height * scale)), color='white')
            logger.info("Canvas created successfully")
        except Exception as canvas_err:
            logger.error(f"Error creating canvas: {canvas_err}")
            # Fallback to smaller canvas
            img = Image.new('RGB', (800, 600), color='white')
            logger.info("Created fallback smaller canvas")
            
        draw = ImageDraw.Draw(img)
        
        # Load fonts using get_font_func if provided or fallback
        if get_font_func:
            try:
                title_font = get_font_func(18)
                header_font = get_font_func(14)
                position_font = get_font_func(14)
                cell_font = get_font_func(12)
                small_font = get_font_func(10)
                row_font = get_font_func(16)
            except Exception as font_err:
                logger.error(f"Error loading fonts: {font_err}")
                # Use default fonts
                title_font = ImageFont.load_default()
                header_font = ImageFont.load_default()
                position_font = ImageFont.load_default()
                cell_font = ImageFont.load_default()
                small_font = ImageFont.load_default()
                row_font = ImageFont.load_default()
        else:
            # Default to system fonts if no font function provided
            title_font = ImageFont.load_default()
            header_font = ImageFont.load_default()
            position_font = ImageFont.load_default()
            cell_font = ImageFont.load_default()
            small_font = ImageFont.load_default()
            row_font = ImageFont.load_default()
        
        # Title with layout ID
        layout_id = layout.get('layoutId', 'Unknown')
        title = f"Kootoro Vending Machine Layout (ID: {layout_id})"
        title_width = draw.textlength(title, font=title_font)
        draw.text(
            (canvas_width * scale / 2 - title_width / 2, padding * scale),
            title,
            fill='black',
            font=title_font
        )
        
        # Draw column numbers
        for col in range(num_columns):
            x = padding * scale + col * (cell_width + cell_spacing) * scale + cell_width * scale / 2
            y = (padding + title_padding + header_height / 2) * scale
            col_text = f"{col + 1}"
            col_width = draw.textlength(col_text, font=header_font)
            draw.text(
                (x - col_width / 2, y - header_font.size / 2),
                col_text,
                fill='black',
                font=header_font
            )
        
        # Draw rows and slots
        for row_idx, tray in enumerate(trays):
            # Get row letter (A, B, C...) from tray code or index
            row_letter = tray.get('trayCode') or chr(65 + row_idx)  # Fallback to A, B, C, etc.
            row_y = (padding + title_padding + header_height + row_idx * (cell_height + footer_height + row_spacing)) * scale
            
            # Draw row separator
            if row_idx > 0:
                separator_y = row_y - (row_spacing / 2) * scale
                draw.line(
                    [(padding * scale, separator_y), ((canvas_width - padding) * scale, separator_y)],
                    fill=(200, 200, 200),
                    width=2  # Increased width for better visibility
                )
            
            # Draw row label
            row_text = row_letter
            row_width = draw.textlength(row_text, font=row_font)
            draw.text(
                ((padding - text_padding) * scale - row_width, row_y + cell_height * scale / 2 - row_font.size / 2),
                row_text,
                fill='black',
                font=row_font
            )
            
            # Get slots for this tray
            slots = []
            if tray.get('slotList'):
                slots = sorted(tray.get('slotList', []), key=lambda s: s.get('slotNo', 0))
            
            # Draw slots
            for col in range(num_columns):
                # Find slot for this position
                slot = next((s for s in slots if s.get('slotNo') == col + 1), None)
                cell_x = padding * scale + col * (cell_width + cell_spacing) * scale
                
                # Draw cell border (slightly darker for better visibility)
                draw.rectangle(
                    [(cell_x, row_y), (cell_x + cell_width * scale, row_y + cell_height * scale)],
                    outline=(200, 200, 200),
                    width=2  # Increased width for better visibility
                )
                
                # Draw position label (e.g., A1, B2)
                position_text = f"{row_letter}{col + 1}"
                position_width = draw.textlength(position_text, font=position_font)
                draw.text(
                    (cell_x + text_padding * scale, row_y + text_padding * scale),
                    position_text,
                    fill='blue',
                    font=position_font
                )
                
                # Draw product details if slot exists
                if slot and slot.get('product'):
                    product = slot.get('product', {})
                    product_name = product.get('name', '')
                    product_image_url = product.get('imageUrl', '')
                    
                    # Draw product image if URL exists
                    if product_image_url:
                        try:
                            # Load and resize product image
                            product_img = load_and_resize_image(product_image_url, image_size * scale)
                            
                            # Calculate center position for the image
                            img_x = cell_x + (cell_width * scale / 2) - (product_img.width / 2)
                            img_y = row_y + (cell_height * scale * 0.4) - (product_img.height / 2)
                            
                            # Paste with proper alpha channel handling
                            if product_img.mode == 'RGBA':
                                # Create mask from alpha channel
                                mask = product_img.split()[3] if len(product_img.split()) > 3 else None
                                img.paste(product_img, (int(img_x), int(img_y)), mask)
                            else:
                                img.paste(product_img, (int(img_x), int(img_y)))
                                
                            logger.info(f"Added product image at {position_text}")
                        except Exception as img_err:
                            logger.error(f"Error loading product image at {position_text}: {img_err}")
                    
                    # Draw product name with improved text alignment
                    if product_name:
                        # Split text into lines that fit
                        lines = split_text_to_lines(draw, product_name, cell_width * scale * 0.9, cell_font)
                        line_height = cell_font.size * 1.2  # Add some line spacing
                        
                        # Calculate total text block height
                        text_block_height = len(lines) * line_height
                        
                        # Position text below the image
                        text_y = row_y + (cell_height * scale * 0.7)
                        
                        for line in lines:
                            # Calculate width of this specific line
                            line_width = draw.textlength(line, font=cell_font)
                            # Center this line horizontally
                            text_x = cell_x + (cell_width * scale / 2) - (line_width / 2)
                            
                            draw.text(
                                (text_x, text_y),
                                line,
                                fill='black',
                                font=cell_font
                            )
                            text_y += line_height  # Move to next line position
        
        # Add footer with timestamp
        metadata_text = f"Generated at: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}"
        metadata_width = draw.textlength(metadata_text, font=small_font)
        draw.text(
            (canvas_width * scale / 2 - metadata_width / 2, (canvas_height - metadata_height / 2) * scale - small_font.size / 2),
            metadata_text,
            fill='black',
            font=small_font
        )
        
        # Convert to PNG
        png_buffer = io.BytesIO()
        img.save(png_buffer, format='PNG', optimize=True)
        png_buffer.seek(0)
        logger.info(f"Successfully rendered layout {layout_id} to PNG")
        return png_buffer.getvalue()
        
    except Exception as e:
        logger.error(f"Error rendering layout: {e}")
        # Create a simple error image
        error_img = Image.new('RGB', (800, 600), color='white')
        draw = ImageDraw.Draw(error_img)
        draw.text((100, 100), f"Error rendering layout: {str(e)}", fill='red')
        error_buffer = io.BytesIO()
        error_img.save(error_buffer, format='PNG')
        error_buffer.seek(0)
        return error_buffer.getvalue()

def write_file_local(filename, data):
    """Write data to a local file"""
    try:
        # Create directory if it doesn't exist
        directory = os.path.dirname(filename)
        if directory and not os.path.exists(directory):
            os.makedirs(directory, exist_ok=True)
            
        with open(filename, 'wb') as f:
            f.write(data)
        logger.info(f"Wrote {len(data)} bytes to {filename}")
        return True
    except Exception as e:
        logger.error(f"Error writing to file {filename}: {e}")
        return False