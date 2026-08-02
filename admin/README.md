# AdminTemplate — RBAC Admin Backend Template

Universal admin backend starter kit with three-level RBAC, JWT auth, operation logs, file storage, and Excel/CSV export. Production-grade, Docker-first, clone-and-go.

## Stack

| Layer | Tech |
|---|---|
| Backend | Go 1.23+ · Gin · GORM v2 |
| Frontend | Vue 3 · Element Plus · Pinia · Vite |
| Database | MySQL 8 |
| Auth | JWT (Access 2h + Refresh 7d, with reuse detection) |
| Storage | Local disk default · pluggable adapter interface |
| API Docs | Swagger (swaggo) at `/swagger/index.html` |
| Deployment | Docker Compose (3 services: mysql · backend · frontend-nginx) |

## Features

- **RBAC**: User → Role → Menu with menu + button/API-level permission granularity (`user:create`, `role:assign`)
- **Operation log**: middleware + annotation-driven capture of `(user, action, ip, ua, request body, response status, latency)`
- **File upload**: local storage with adapter interface — swap to OSS/MinIO by implementing `Storage`
- **Excel/CSV export**: generic exporter for any list endpoint
- **Pagination + fuzzy search + batch operations**: standard across all resources
- **Auto-migrate + seed data**: first startup creates admin/admin123 with full menu tree

## Quick Start

```bash
git clone <this-repo>
cd AdminTemplate
docker compose up -d
```

- Frontend: http://localhost:8081 (default, set `FRONTEND_PORT` to change)
- Backend API: http://localhost:8080/api/v1 (default, set `BACKEND_PORT` to change)
- Swagger: http://localhost:8080/swagger/index.html
- Login: `admin` / `admin123` (change immediately in production)

#### Changing the port

The backend port is fully driven by the `BACKEND_PORT` env var. To move the backend off 8080:

```bash
BACKEND_PORT=9000 FRONTEND_PORT=9081 docker compose up -d
```

Both the host port mapping, container-side `SERVER_PORT`, `STORAGE_BASE_URL`, AND the admin frontend's nginx `proxy_pass` URLs follow it — no file edits needed. The admin frontend's `nginx.conf` is a template; `envsubst` renders `${BACKEND_PORT}` at container start.

The uni-app mini-program reads the API URL from local storage (configurable in "我的 → 服务器设置"), so it also adapts without a rebuild. For production builds, the default API URL is set at build time via the `VITE_API_BASE_URL` env var:

```bash
VITE_API_BASE_URL=https://api.example.com:8082 npm run build:h5
```

Without it, the default is `http://localhost:8080`.

## Layout

```
AdminTemplate/
├── backend/                  # Go service (Gin + GORM)
│   ├── cmd/server/main.go
│   ├── internal/{config,model,dto,router,middleware,handler,service,repository,pkg,seed}
│   ├── configs/config.example.yaml
│   ├── docs/                 # swag-generated
│   ├── storage/uploads/      # uploaded files (gitignored)
│   ├── Dockerfile
│   └── go.mod
├── frontend/                 # Vue 3 SPA
│   ├── src/{api,stores,router,views,components,directives,utils,layouts,styles}
│   ├── public/
│   ├── nginx.conf            # reverse-proxy to backend
│   ├── vite.config.ts
│   └── Dockerfile
├── docker-compose.yml
└── README.md
```

## Development (without Docker)

```bash
# Backend
cd backend
cp configs/config.example.yaml configs/config.local.yaml
# edit DB host, JWT secret, etc.
go run ./cmd/server

# Frontend
cd frontend
npm install
npm run dev      # http://localhost:5173
```

## Customization

- **Add new resource**: model → DTO → repository → service → handler → router → seed menu; auto-pickup permission middleware
- **Swap storage**: implement `pkg/storage.Storage` interface (Put / Get / Delete / GetSignedURL)
- **Add permission code**: define as `resource:action` in menu table → use `@Permission("...")` annotation on handler

## License

MIT