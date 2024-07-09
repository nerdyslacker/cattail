.PHONY: all docker-build docker-run docker-cp build install local-install

# Main target that runs all steps
all: docker-build docker-run docker-cp local-install

# Build the Docker image
docker-build:
	docker build -t wails-cattail-builder .

# Run the Docker container
docker-run:
	docker run --name wails-cattail-build wails-cattail-builder

# Copy built files from the container
docker-cp:
	mkdir -p ./build
	docker cp wails-cattail-build:/app/build/. ./build/
	docker rm wails-cattail-build

# Build target (run inside Docker)
build:
	wails build

# Install target (now only used inside Docker)
install: build

# Local install target (run on host system)
local-install:
	mkdir -p ~/.local/share/applications ~/.local/share/icons/hicolor/256x256/apps ~/.local/bin/
	cp build/bin/cattail ~/.local/bin
	cp cattail.desktop ~/.local/share/applications/
	cp frontend/src/assets/images/icon.png ~/.local/share/icons/hicolor/256x256/apps/com.cattail.png