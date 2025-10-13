#!/bin/bash

# Build script for naturalTime.js
# This creates a browserify bundle that works with goja

cd "$(dirname "$0")"

echo "Installing dependencies..."
npm install

echo "Building JavaScript bundle..."
mkdir -p dist
npx browserify naturalTime.js --standalone naturaltime > dist/naturalTime.bundle.js

echo "Build complete! The bundle is in dist/naturalTime.bundle.js"
