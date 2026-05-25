# Loaded Questions

A real-time multiplayer web game based on the board game *Loaded Questions*. Built with Go (backend) and React + TypeScript (frontend).

## Dev environment

The `.devcontainer` folder defines a containerized development environment with Go and Node pre-installed. Requires the [Dev Containers](https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-containers) VS Code extension.

```
cd .devcontainer
docker build -t dev-go-node .
```

Then **Reopen in Container** via the VS Code command palette.

---

## Backend

**Prerequisites:** Go 1.21+

### Local development

```bash
cd backend
go run .
```

The server starts on `http://localhost:8080`.

### Build

```bash
cd backend
go build -o server .
./server
```

### Test

```bash
cd backend
go test ./...

# With race detector
go test ./... -race
```

### Docker

```bash
# Build
docker build -t loaded-questions-backend ./backend

# Run
docker run -p 8080:8080 loaded-questions-backend
```

---

## Frontend

**Prerequisites:** Node 18+

### Local development

```bash
cd ui/frontend
npm install
npm run dev
```

The dev server starts on `http://localhost:5173` and proxies `/api` requests to the backend at `http://localhost:8080`. Run the backend first.

### Build

```bash
cd ui/frontend
npm install
npm run build
```

Production output is written to `ui/frontend/dist/`.

### Test

```bash
cd ui/frontend
npm test

# Watch mode
npm run test:watch
```

### Docker

```bash
# Build (serves via Nginx, proxies /api to the backend)
docker build -t loaded-questions-frontend ./ui/frontend

# Run
docker run -p 3000:80 loaded-questions-frontend
```

---

## Helm chart

The chart lives at `helm/loaded-questions/` and targets any nginx-ingress Kubernetes cluster.

### Push to GHCR (OCI registry)

Helm 3.8+ supports OCI registries natively — no plugin needed.

```bash
# 1. Authenticate (once per machine; uses a GitHub PAT with write:packages scope)
echo $GITHUB_TOKEN | helm registry login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin

# 2. Package the chart
helm package helm/loaded-questions
# Produces: loaded-questions-0.1.0.tgz

# 3. Push to GHCR
helm push loaded-questions-0.1.0.tgz oci://ghcr.io/YOUR_GITHUB_USERNAME/helm
```

The chart is now available at `oci://ghcr.io/YOUR_GITHUB_USERNAME/helm/loaded-questions`.

> **GHCR visibility** — new packages default to private. Go to your GitHub profile →
> **Packages** → `helm/loaded-questions` → **Package settings** → make it public if
> you want others to pull it without authenticating.

### Install from GHCR

```bash
helm install loaded-questions \
  oci://ghcr.io/YOUR_GITHUB_USERNAME/helm/loaded-questions \
  --version 0.1.0 \
  --set ingress.host=questions.yourdomain.com \
  --set backend.image.repository=ghcr.io/YOUR_GITHUB_USERNAME/loaded-questions-backend \
  --set backend.image.tag=latest \
  --set frontend.image.repository=ghcr.io/YOUR_GITHUB_USERNAME/loaded-questions-frontend \
  --set frontend.image.tag=latest
```

### Upgrade

Bump `version` in `helm/loaded-questions/Chart.yaml`, then:

```bash
helm package helm/loaded-questions
helm push loaded-questions-<NEW_VERSION>.tgz oci://ghcr.io/YOUR_GITHUB_USERNAME/helm
helm upgrade loaded-questions oci://ghcr.io/YOUR_GITHUB_USERNAME/helm/loaded-questions --version <NEW_VERSION>
```

---

## Running both services together

```bash
docker compose up --build
```

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
