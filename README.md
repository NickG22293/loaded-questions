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

## Running both services together

```bash
docker compose up --build
```

- Frontend: `http://localhost:3000`
- Backend API: `http://localhost:8080`
