
.PHONY: build

build:
	wails build

install: build
	mkdir -p ~/.local/share/applications ~/.local/share/icons/hicolor/256x256/apps ~/.local/bin/
	cp build/bin/cattail ~/.local/bin
	cp cattail.desktop ~/.local/share/applications/
	cp frontend/src/assets/images/icon.png ~/.local/share/icons/hicolor/256x256/apps/com.cattail.png
