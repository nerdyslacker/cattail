# Use Void Linux as the base image
FROM ghcr.io/void-linux/void-glibc:latest

# Update the system and install necessary dependencies
RUN xbps-install -Syu && \
    xbps-install -y \
    wget \
    git \
    gcc \
    make \
    pkg-config \
    gtk+3-devel \
    webkit2gtk-devel \
    go \
    nodejs \
    && xbps-remove -O

ENV PATH="/root/go/bin:${PATH}"

# Install Wails
RUN go install github.com/wailsapp/wails/v2/cmd/wails@v2.8.0

# Set the working directory in the container
WORKDIR /app

# Copy the project files into the container
COPY . .

# Create build.sh script
RUN echo '#!/bin/sh\n\
# Run the build\n\
make build\n\
# Exit the script (and container) after building\n\
exit 0' > build.sh

# Make the build script executable
RUN chmod +x build.sh

# Set the entrypoint to the build script
ENTRYPOINT ["/bin/sh", "/app/build.sh"]