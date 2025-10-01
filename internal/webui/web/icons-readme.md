# PWA Icons

This directory contains icon files for the Progressive Web App.

## Required Icons:

- icon-192.png (192x192px) - Standard app icon
- icon-512.png (512x512px) - High-res app icon
- icon-apple-touch.png (180x180px) - Apple touch icon

## Icon Guidelines:

- Use the banned.video branding colors (#ff6b35, #f7931e)
- Include the "🚫" emoji or similar banned symbol
- Make sure icons work on both light and dark backgrounds
- Follow platform icon guidelines for rounded corners/masking

## Generating Icons:

You can create these icons using:

1. Design tools (Figma, Sketch, Canva)
2. Online PWA icon generators
3. Command line tools like ImageMagick

Example ImageMagick command to resize:

```bash
# Create 192px icon
convert source-icon.png -resize 192x192 icon-192.png

# Create 512px icon
convert source-icon.png -resize 512x512 icon-512.png

# Create Apple touch icon
convert source-icon.png -resize 180x180 icon-apple-touch.png
```
