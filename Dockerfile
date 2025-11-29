# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o novauserbot .

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata git bash

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/novauserbot .
COPY --from=builder /build/locales ./locales

# Create data directory
RUN mkdir -p /app/data

# Set timezone
ENV TZ=UTC

# Run as non-root user
RUN adduser -D -u 1000 nova
RUN chown -R nova:nova /app
USER nova

CMD ["./novauserbot"]
