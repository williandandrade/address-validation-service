# Build stage
FROM golang:1.25-alpine AS builder

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
    ldconfig /usr/local/lib 2>/dev/null || true && \
    rm -rf /tmp/libpostal

# Copy go mod files first for better caching
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build with gopostal tag (enables real gopostal parser via CGO)
RUN CGO_ENABLED=1 GOOS=linux go build -tags gopostal -ldflags="-w -s" -o /address-validation-service ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies including libpostal shared libraries
RUN apk add --no-cache ca-certificates tzdata libstdc++ libgcc

# Copy libpostal libraries and data from builder
COPY --from=builder /usr/local/lib/libpostal* /usr/local/lib/
COPY --from=builder /usr/local/share/libpostal /usr/local/share/libpostal

# Update library cache
RUN ldconfig /usr/local/lib 2>/dev/null || true

# Create non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy binary from builder
COPY --from=builder /address-validation-service .

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
