# -------- build stage --------
FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN CGO_ENABLED=0 GOSUMDB=off go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOSUMDB=off go build -o /bin/ping .

# -------- runtime stage --------
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /bin/ping /ping
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/ping"]
