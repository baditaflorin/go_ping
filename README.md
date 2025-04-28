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

## 🗺️ Endpoint

| Method | Path | Response |
|--------|------|----------|
| `GET`  | `/ping` | `pong\n` (`200 OK`) |

---

## 🚀 Quick Start (Local)

```bash
# native run
make run               # or: go run main.go
curl localhost:8080/ping   # → pong

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
| `run` | `go run main.go` |
| `docker-buildx` | Multi‑arch build & push to GHCR |
| `docker-run` | `docker run -p 8080:8080 ping:latest` |
| `compose-up` | `docker compose up --build` |
| `clean` | Delete local image |

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

