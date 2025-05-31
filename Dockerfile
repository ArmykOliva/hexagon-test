# Build stage
FROM golang:1.23.5-alpine AS builder

# Install bash for build script
RUN apk add --no-cache bash

# Set working directory
WORKDIR /app

# Copy go modules and libraries first for better caching
COPY source/go.mod source/go.sum* ./source/
COPY libraries/ ./libraries/

# Download dependencies
WORKDIR /app/source
RUN go mod download

# Copy source code
COPY source/ ./

# Create public directory if it doesn't exist
RUN mkdir -p public

# Build the WASM application
RUN GOROOT=$(go env GOROOT) && \
    env GOOS=js GOARCH=wasm go build -o public/main.wasm main.go && \
    cp "${GOROOT}/misc/wasm/wasm_exec.js" public/wasm_exec.js

# Runtime stage
FROM golang:1.23.5-alpine AS runtime

WORKDIR /app

# Copy the built application and serve.go
COPY --from=builder /app/source/public ./public
COPY --from=builder /app/source/serve.go ./

# Expose port
EXPOSE 3000

# Run the server
CMD ["go", "run", "serve.go"] 