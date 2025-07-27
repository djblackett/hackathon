# --- Stage 1: Go builder ---
FROM golang:1.24.5-bookworm AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o filename-fixer ./cmd/client/main.go

# --- Stage 2: Extract runtime dependencies ---
FROM debian:bookworm-slim AS deps

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        #  poppler-utils tesseract-ocr pandoc \
    && rm -rf /var/lib/apt/lists/*

# --- Stage 3: Slim final image ---
FROM debian:bookworm-slim AS final

# Copy only runtime tools from deps image
# COPY --from=deps /usr/bin/pdftotext /usr/bin/pdftotext
# COPY --from=deps /usr/bin/tesseract /usr/bin/tesseract
# COPY --from=deps /usr/bin/pandoc /usr/bin/pandoc
COPY --from=deps /usr/lib /usr/lib/    
COPY --from=deps /lib /lib/             
# COPY --from=deps /usr/share/tessdata /usr/share/tessdata

# Copy the Go binary
COPY --from=builder /app/filename-fixer /usr/local/bin/filename-fixer

ENTRYPOINT ["filename-fixer"]
# ENTRYPOINT ["/bin/sh"]
