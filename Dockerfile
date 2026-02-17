ARG PARSER=regex

# --- Gopostal builder (only used when PARSER=gopostal) ---
FROM golang:1.25-alpine AS gopostal-builder

WORKDIR /app

# Install build dependencies including libpostal requirements
RUN apk add --no-cache git ca-certificates gcc g++ make curl autoconf automake libtool pkgconfig

# Install libpostal
RUN git clone https://github.com/openvenues/libpostal /tmp/libpostal && \
    cd /tmp/libpostal && \
    ./bootstrap.sh && \
    ./configure --datadir=/usr/local/share/libpostal && \
    make -j$(nproc) && \
    make install && \
    ldconfig /usr/local/lib && \
    rm -rf /tmp/libpostal

# Stage libpostal runtime files for the final image
RUN mkdir -p /runtime/lib /runtime/share && \
    cp /usr/local/lib/libpostal* /runtime/lib/ && \
    cp -r /usr/local/share/libpostal /runtime/share/

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .
RUN cp configs/.env.example configs/.env

# Build with gopostal tag (enables real gopostal parser via CGO)
RUN CGO_ENABLED=1 GOOS=linux go build -tags gopostal -ldflags="-w -s" -o /address-validation-service ./cmd/server

# --- Regex builder (default, no libpostal) ---
FROM golang:1.25-alpine AS regex-builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

# Create empty runtime dirs (no libpostal needed)
RUN mkdir -p /runtime/lib /runtime/share

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .
RUN cp configs/.env.example configs/.env

# Build without gopostal tag — pure Go, no CGO needed
RUN CGO_ENABLED=0 go build -ldflags="-w -s" -o /address-validation-service ./cmd/server

# --- Select builder based on PARSER arg ---
FROM ${PARSER}-builder AS builder

# --- Runtime ---
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata libstdc++ libgcc

# Copy libpostal libraries and data (empty dirs for regex builds, populated for gopostal)
COPY --from=builder /runtime/lib/ /usr/local/lib/
COPY --from=builder /runtime/share/ /usr/local/share/

# Update library cache (needed for gopostal, harmless for regex)
RUN ldconfig /usr/local/lib 2>/dev/null || true

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /address-validation-service .

# Copy configs directory
COPY --from=builder /app/configs ./configs

# Set ownership
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1

# Run the application
ENTRYPOINT ["./address-validation-service"]
