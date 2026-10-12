# -------- build stage --------
FROM golang:1.23-alpine AS builder
WORKDIR /src

# Copy go mod files. go.sum must be copied too: `go mod download` does not
# write the checksums that `go build` requires, so a go.mod-only copy fails
# with "missing go.sum entry".
COPY go.mod go.sum ./

# Download and verify dependencies against go.sum
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
