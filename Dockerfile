# ----- 1. Build stage: minimal Go build environment -----
FROM golang:1.23.4-alpine AS build

WORKDIR /app

# Copy module files first (for caching)
COPY go.mod ./
RUN go mod download

# Copy the rest of the files
COPY . .

# Necessary to build statically-linked binary
ENV CGO_ENABLED=0
# Build the Go application
RUN go build -o go-load-lab cmd/server/main.go

# ----- 2. Final stage (Distroless) -----
FROM gcr.io/distroless/base-debian10

# Set environment variables so the Go app can see them
ENV HUB_LINK="https://hub.docker.com/repository/docker/joobypm/go-load-lab"

WORKDIR /app

# Copy the compiled binary
COPY --from=build /app/go-load-lab /app/go-load-lab

# Copy static files
COPY static ./static

EXPOSE 8080
USER nonroot:nonroot

# Launch the binary
ENTRYPOINT ["/app/go-load-lab"]
