# Build stage
FROM golang:1.23.5-alpine AS builder

# Install bash for build script
RUN apk add --no-cache bash

# Set working directory
WORKDIR /app

# Copy libraries first
COPY libraries/ ./libraries/

# Copy source directory
COPY source/ ./source/

# Set working directory to source
WORKDIR /app/source

# Run the build script (which builds WASM and copies files)
RUN chmod +x build.sh && \
    GOROOT=$(go env GOROOT) && \
    env GOOS=js GOARCH=wasm go build -o public/main.wasm main.go && \
    cp "${GOROOT}/misc/wasm/wasm_exec.js" public/wasm_exec.js

# Runtime stage for serving
FROM golang:1.23.5-alpine AS runtime

WORKDIR /app

# Copy everything from builder stage to ensure all dependencies are available
COPY --from=builder /app /app

# Set working directory to source
WORKDIR /app/source

# Expose port
EXPOSE 3000

# Run the server
CMD ["go", "run", "serve.go"] 