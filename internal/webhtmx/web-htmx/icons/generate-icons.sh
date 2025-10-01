#!/bin/bash

# Generate PWA icons from SVG
# This script requires ImageMagick (convert command)

BASE_SVG="icon-base.svg"
SIZES=(72 96 128 144 152 192 384 512)

echo "Generating PWA icons from $BASE_SVG..."

for size in "${SIZES[@]}"; do
    output="icon-${size}x${size}.png"
    
    if command -v convert &> /dev/null; then
        convert "$BASE_SVG" -resize "${size}x${size}" "$output"
        echo "Generated $output"
    else
        echo "ImageMagick not found. Please install it to generate PNG icons."
        echo "You can also use online tools to convert the SVG to PNG sizes: ${SIZES[*]}"
        break
    fi
done

echo "Icon generation complete!"
echo ""
echo "To install ImageMagick:"
echo "  Ubuntu/Debian: sudo apt install imagemagick"
echo "  macOS: brew install imagemagick"
echo "  Windows: Download from https://imagemagick.org/script/download.php"