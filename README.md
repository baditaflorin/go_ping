# Ping Service

A dead‑simple, production‑grade **“ping / pong”** HTTP endpoint you can drop behind any load‑balancer or health‑check.  Written in plain Go (no third‑party deps) and shipped as a minimal **distroless** container.

---

## ✨ Features

| What | Why it matters |
|------|----------------|
| **`/ping` ⇒ `pong`** | Single‑purpose readiness & liveness probe. |
| **Tiny binary ≈ 2 MB** | Fast image pulls, low RAM / CPU (64 Mi / 50 m by default). |
| **Distroless, non‑root** | Minimal attack surface, drops most CVEs on day 1. |
| **Multi‑arch image** | `linux/amd64` **and** `linux/arm64` in the same tag—runs anywhere. |
| **Docker‑first workflow** | No external PaaS required; perfect for one‑box VPS setups.

---

## 🗺️ Endpoints

| Method | Path | Response | Purpose |
|--------|------|----------|---------|
| `GET`  | `/` | `pong\n` (`200 OK`) | Health check / readiness probe |
| `GET`  | `/health` | `{"status":"healthy"}` (`200 OK`) | JSON health endpoint |
| `GET`  | `/metrics` | Prometheus metrics (`200 OK`) | Prometheus-compatible metrics scrape endpoint |

---

## 📊 Observability & Metrics

### Prometheus Metrics

All metrics are exposed at the `/metrics` endpoint in Prometheus text format. The following metrics are collected:

#### HTTP Metrics
- **`http_requests_total`** (Counter): Total number of HTTP requests received
- **`http_request_duration_seconds`** (Histogram): HTTP request latency with buckets (0.005s, 0.01s, 0.025s, 0.05s, 0.1s, 0.25s, 0.5s, 1s, 2.5s, 5s, 10s)
- **`http_request_size_bytes`** (Histogram): HTTP request payload size
- **`http_response_size_bytes`** (Histogram): HTTP response payload size
- **`http_errors_total`** (Counter): Total number of HTTP 5xx errors
- **`http_requests_active`** (Gauge): Number of currently active HTTP requests

#### Application Metrics (extensible)
- **`background_jobs_total`** (Counter): Background job execution count
- **`background_job_duration_seconds`** (Histogram): Background job latency
- **`background_job_errors_total`** (Counter): Background job error count
- **`api_calls_total`** (Counter): External API call count
- **`api_call_duration_seconds`** (Histogram): External API call latency
- **`api_call_errors_total`** (Counter): External API call error count
- **`file_processes_total`** (Counter): File/CSV/TSV processing operations
- **`file_process_duration_seconds`** (Histogram): File processing latency
- **`file_process_bytes_total`** (Counter): Total bytes processed
- **`file_process_errors_total`** (Counter): File processing error count

### Correlation IDs (Request Tracing)

Every request is assigned a **correlation ID** (UUID) to enable end-to-end request tracing across your system. Correlation IDs flow through logs, metrics labels (where appropriate), and outgoing API calls.

#### How It Works

1. **Incoming Request**: The middleware checks for:
   - `X-Request-ID` header (takes priority)
   - `X-Correlation-ID` header (fallback)
   - Generates a new UUID v4 if neither is present

2. **Request Processing**: The correlation ID is:
   - Stored in the request context (`ping/observability.CorrelationID`)
   - Included in all structured logs
   - Exposed back in the response as `X-Correlation-ID` header

3. **Propagation**: When making downstream API calls, include the correlation ID:
   ```go
   correlationID := observability.GetCorrelationID(ctx)
   // Add to outgoing request headers
   outgoingReq.Header.Set("X-Correlation-ID", correlationID)
   ```

#### Example Usage

```bash
# Request with explicit correlation ID
curl -H "X-Request-ID: my-trace-123" http://localhost:8080/

# Response includes the same ID
# Headers: X-Correlation-ID: my-trace-123

# Without explicit ID, server generates one
curl http://localhost:8080/
# Headers: X-Correlation-ID: 550e8400-e29b-41d4-a716-446655440000
```

### Scraping with Prometheus

Add this job to your Prometheus `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'ping-service'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
    scrape_timeout: 10s
```

### Architecture & Design Patterns

The observability layer is implemented following SOLID principles:

- **Single Responsibility**: `observability/` package owns all Prometheus collectors
- **Dependency Inversion**: Business logic is independent of Prometheus details
- **Open/Closed**: Add new metrics by extending the `Metrics` struct, not modifying existing code
- **Middleware Pattern**: `RequestInstrumentationMiddleware` keeps instrumentation cross-cutting
- **Context-Based Correlation**: Correlation IDs flow through `context.Context` (idiomatic Go)

---

## 🚀 Quick Start (Local)

```bash
# native run
make run               # or: go run main.go
curl localhost:8080/                       # → pong
curl localhost:8080/health                 # → {"status":"healthy"}
curl localhost:8080/metrics                # → Prometheus metrics

# docker run (buildx multi‑arch)
make docker-buildx      # cross‑build + push if GHCR creds present
make docker-run         # exposes :8080
```

> **Requirements**  
> • Go ≥ 1.22 (if running locally)  
> • Docker ≥ 23.0 with **buildx** & QEMU (comes standard in Docker Desktop)

---

## 📦 Build & Push a Multi‑Arch Image

```bash
export IMAGE="ghcr.io/baditaflorin/ping:latest"

docker buildx build \
  --platform linux/amd64,linux/arm64/v8 \
  -t $IMAGE \
  --push .
```

### Make the package public (once)

1. Open **https://github.com/users/baditaflorin/packages/container/ping**
2. **Package settings → Make public**

Anyone (or any CI) can now pull without a PAT:

```bash
docker pull ghcr.io/baditaflorin/ping:latest
```

---

## ☁️ Deploy on Your Hetzner VM

### 1 · Pull & run

```bash
ssh root@domain.com     # your server

docker pull ghcr.io/<GH_USER>/ping:latest

docker run -d \
  --name ping \
  --restart unless-stopped \
  -e PORT=8080 \
  --network bridge \
  ghcr.io/<GH_USER>/ping:latest
```

*(Replace with `docker compose up -d` if you keep a compose stack.)*

### 2 · Expose via **Nginx Proxy Manager**

| Field | Value                      |
|-------|----------------------------|
| **Domain** | `ping.domain.com`          |
| **Scheme** | `http`                     |
| **Forward host** | `ping` (container name)    |
| **Forward port** | `8080`                     |
| **SSL** | Request Let’s Encrypt cert |

After NPM reloads:

```bash
curl https://ping.domain.com/ping
pong
```

---

Run it on the VM:

```bash
ssh root@domain.com
cd /opt/ping            # folder where docker-compose.yml lives
docker compose up -d
```

Compose will pull the latest image (or build if you add `build: .`) and keep it running just like the earlier `docker run` command.

---

## 🛠️ Make Targets

| Target | Does |
|--------|------|
| `build` | Build the application binary |
| `run` | `go run main.go` with observability enabled |
| `test` | Run all tests (metrics, correlation IDs, handlers) |
| `docker-build` | Build local Docker image with metrics support |
| `docker-buildx` | Multi‑arch build & push to GHCR with metrics |
| `docker-run` | `docker run -p 8080:8080 ping:latest` with `/metrics` exposed |
| `docker-compose-up` | `docker compose up --build` with Prometheus service |
| `clean` | Delete local image and binary |

---

## 🤖 CI / CD (GitHub Actions mini‑sample)

```yaml
name: Build & Push
on:
  push:
    branches: [main]

env:
  IMAGE: ghcr.io/${{ github.repository }}:latest

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: ${{ env.IMAGE }}
          platforms: linux/amd64,linux/arm64/v8
```

Add a deploy step (`ssh`, `rsync`, `docker pull && docker restart ping`) to complete the pipeline.

---

## 🛡️ Hardening Notes

* **Distroless static** base; no shell, no package manager.
* Runs as **UID 65532** (non‑root) and drops all caps.
* `CGO_ENABLED=0` cuts libc, shrinks the binary.

---

## 👫 Contributing

PRs & issues welcome—open tickets for flags, metrics, or JSON output.

---

## 📄 License

MIT © 2025 Vivi & contributors

