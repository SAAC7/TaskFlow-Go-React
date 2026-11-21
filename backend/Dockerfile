# ---- Build stage ----
# If `go mod tidy` raises the "go" line in go.mod, bump this version too.
FROM golang:1.23-alpine AS build
WORKDIR /src

# Download dependencies first so Docker caches this layer.
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---- Runtime stage: tiny image, no shell, non-root user ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
