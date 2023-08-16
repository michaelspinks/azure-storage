# syntax=docker/dockerfile:1

FROM golang:1.20.4 AS build-env

# Set destination for COPY
WORKDIR /app

COPY go.mod ./
run go mod download

# Build
COPY . .
RUN go build -o video-storage

FROM debian:bullseye-slim

# Install ca-certificates
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /
COPY --from=build-env /app/video-storage /video-storage

# Run
CMD ["/video-storage"]