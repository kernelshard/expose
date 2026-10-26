# Stage 1: Build static Go binary
FROM golang:alpine AS builder

WORKDIR /app

# Cache Go modules first (speeds up rebuilds)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Compile static binary: no cgo, stripped debug symbols (-s -w)
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X github.com/kernelshard/expose/internal/version.Version=${VERSION} -X github.com/kernelshard/expose/internal/version.GitCommit=${COMMIT} -X github.com/kernelshard/expose/internal/version.BuildDate=${DATE}" \
    -o /app/bin/expose \
    ./cmd/expose

# Stage 2: Minimal, secure runtime
FROM alpine:3.21

# Add standard root CA certificates and an unprivileged user
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -S expose && adduser -S expose -G expose

USER expose

# Copy only the compiled binary from builder
COPY --from=builder /app/bin/expose /usr/local/bin/expose

# 7890: Control plane, 8080: Public HTTP proxy
EXPOSE 7890 8080

ENTRYPOINT ["expose"]
CMD ["server", "--domain=localhost", "--control-port=7890", "--public-port=8080"]
