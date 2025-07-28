# --- Stage 1: Go builder ---
FROM golang:1.24.5-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o filename-fixer ./cmd/client/main.go

# --- Stage 2: Extract runtime dependencies ---
FROM alpine:3.20 AS deps

# Install runtime tools - Alpine package names
RUN apk add --no-cache \
    jq \
    ca-certificates \
    # poppler-utils \
    # tesseract-ocr \
    # pandoc \
    && rm -rf /var/cache/apk/*

# --- Stage 3: Slim final image ---
FROM alpine:3.20 AS final

# Copy CA certificates from deps stage for TLS connections
COPY --from=deps /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy only runtime tools from deps image
COPY --from=deps /usr/bin/jq /usr/bin/jq
# COPY --from=deps /usr/bin/pdftotext /usr/bin/pdftotext
# COPY --from=deps /usr/bin/tesseract /usr/bin/tesseract
# COPY --from=deps /usr/share/tessdata /usr/share/tessdata
# COPY --from=deps /usr/bin/pandoc /usr/bin/pandoc

#  Copy libraries from deps image
COPY --from=deps /lib /lib/
COPY --from=deps /usr/lib /usr/lib/

# Copy the Go binary
COPY --from=builder /app/filename-fixer /usr/local/bin/filename-fixer

ENTRYPOINT ["filename-fixer"]