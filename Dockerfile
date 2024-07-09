# Use Arch Linux as the base image
FROM archlinux:latest

# Update the system and install necessary dependencies
RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm \
    wget \
    git \
    gcc \
    make \
    pkg-config \
    gtk3 \
    webkit2gtk \
    go \
    npm \
    && pacman -Scc --noconfirm

ENV PATH="/root/go/bin:${PATH}"

# Install Wails
RUN go install github.com/wailsapp/wails/v2/cmd/wails@v2.8.0

# Set the working directory in the container
WORKDIR /app

# Copy the project files into the container
COPY . .

# Make the build script executable
RUN chmod +x build.sh

# Set the entrypoint to the build script
ENTRYPOINT ["./build.sh"]