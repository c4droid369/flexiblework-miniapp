# ----- build stage -----
# Match the local Go version (currently 1.25.x). Using an older Alpine image
# triggers Go's toolchain auto-download — fine on a host with good network,
# but unreliable in CI/Alpine. Pin to the same major version as go.mod's
# `go` directive to skip the toolchain fetch entirely.
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Build.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ----- runtime stage -----
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app && \
    mkdir -p /app/storage/uploads && \
    chown -R app:app /app

WORKDIR /app
COPY --from=build /out/server /app/server

USER app
EXPOSE 8080

ENV SERVER_HOST=0.0.0.0 \
    SERVER_PORT=8080 \
    TZ=Asia/Shanghai

ENTRYPOINT ["/app/server"]