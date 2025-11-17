# Build stage
FROM golang:1.25-alpine AS builder

# Install the necessary utilities for the assembly
RUN apk add --no-cache git make

WORKDIR /build

# Copying go.mod and go.sum to cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copying the source code
COPY . .

# Building a binary file
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /build/telegrambot \
    ./cmd/telegrambot/main.go

# Runtime stage
FROM alpine:latest

# Installing ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Creating an unprivileged user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copying the binary file from the builder stage
COPY --from=builder /build/telegrambot .

# Copying static resources (localization and captcha images)
COPY --from=builder /build/internal/localization/locales ./internal/localization/locales
COPY --from=builder /build/assets/captcha ./assets/captcha

# Changing the file owner
RUN chown -R appuser:appuser /app

# Switching to an unprivileged user
USER appuser

# Launching the app
CMD ["/app/telegrambot"]
