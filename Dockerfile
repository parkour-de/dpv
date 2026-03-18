# Use the offical golang image to create a binary.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.26-alpine as builder

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
# Expecting to copy go.mod and if present go.sum.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
ARG VERSION
RUN go build -v -o /app/bin/membership -ldflags "-X main.version=${VERSION}" ./src/cmd/membership

FROM alpine:latest

WORKDIR /app

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/bin/membership /app/bin/membership

RUN mkdir -p /app/documents
VOLUME ["/app/cfg", "/app/documents"]

# Run the web service on container startup.
CMD ["/app/bin/membership"]