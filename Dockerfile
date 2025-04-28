# -------- build stage --------
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/ping .

# -------- runtime stage --------
FROM gcr.io/distroless/static:nonroot
COPY --from=builder /bin/ping /ping
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/ping"]
