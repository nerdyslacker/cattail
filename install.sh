#!/bin/bash

# Install wails
go install github.com/wailsapp/wails/v2/cmd/wails@v2.8.0

# Check if all dependencies are installed 
wails doctor

# Clone the repository 
git clone https://github.com/nerdyslacker/cattail
cd cattail

# Build the app
wails build

# Install on system
mkdir -p ~/.local/share/applications ~/.local/share/icons/hicolor/256x256/apps ~/.local/bin/
cp build/bin/cattail ~/.local/bin
cp cattail.desktop ~/.local/share/applications/
cp frontend/src/assets/images/icon.png ~/.local/share/icons/hicolor/256x256/apps/com.cattail.png