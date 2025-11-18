# -------- build stage --------
FROM golang:1.23-alpine AS builder
WORKDIR /src

# Copy go mod files
COPY go.mod ./

# Download dependencies and generate go.sum
# Using go.sum checksum database for verification
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a \
    -o /bin/ping .

# -------- runtime stage --------
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /bin/ping /ping
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/ping"]
