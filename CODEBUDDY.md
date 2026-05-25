# CODEBUDDY.md This file provides guidance to CodeBuddy when working with code in this repository.

## Project Overview

NginxOps is a web-based Nginx management platform that provides a unified interface for managing Nginx sites, upstreams (load balancers), SSL certificates (with ACME support), and monitoring access logs. It packages a Go backend, React frontend, and Nginx into a Docker image. PostgreSQL runs as a separate service.

## Development Commands

### Backend (Go)
```bash
cd backend
go build -o nginxops ./cmd/server           # Build server
go build -o migrate ./cmd/migrate           # Build migration CLI tool
./nginxops                                  # Run server (port 8080 by default)
./migrate -action up                        # Run migrations up
./migrate -action down                      # Rollback last migration
./migrate -action version                   # Check current migration version
go mod download                             # Download dependencies
go run ./cmd/server                         # Run server directly
```

**No tests exist in this project** — no `*_test.go`, `*.test.ts`, or `*.spec.ts` files are present.

### Frontend (React + Vite)
```bash
cd frontend
npm install                                 # Install dependencies
npm run dev                                 # Dev server on port 3000, proxies /api→localhost:8080, /ws→ws://localhost:8080
npm run build                               # Production build (tsc -b && vite build)
npm run lint                                # Run ESLint
npm run preview                             # Preview production build
```

Path alias: `@` → `./src`

### Docker
```bash
docker-compose up --build                   # Build and start all services (postgres + app)
docker build -t nginxops:latest .            # Build Docker image (3-stage: node→go→alpine)
```

## Architecture

### Backend Structure (Go + Gin)

**Layered Architecture**: Handler → Service → Repository → Database

- `cmd/server/main.go` — App entry point, route definitions, manual dependency injection via `New*Handler/New*Service/New*Repository` constructors
- `cmd/migrate/main.go` — Database migration CLI tool
- `internal/config/` — YAML config loading (`/data/config.yml`), environment variable fallback via `loadFromEnv()`
- `internal/database/` — PostgreSQL via GORM, auto-migration on startup, 30-retry connection with 1s interval
- `internal/handler/` — HTTP handlers (Gin), request parsing, response formatting
- `internal/service/` — Business logic; site/upstream services generate Nginx configs via `strings.Builder`; ACME service handles async certificate requests
- `internal/repository/` — Data access via GORM; returns raw GORM errors (no wrapping)
- `internal/model/` — GORM model definitions; JSON fields stored as `TEXT` in PostgreSQL
- `internal/middleware/` — `AuthRequired()` (JWT), `CORS()`, `AuditMiddleware()` (logs to audit_log table)
- `internal/websocket/` — Hub-and-spoke WebSocket for real-time log streaming (gorilla/websocket). **Note: `/ws/logs` has no authentication**
- `pkg/` — Shared utilities: JWT (`pkg/jwt`), ACME (`pkg/acme`), DNS providers (`pkg/dnsprovider`), Nginx config generation, HTTP response helpers (`pkg/response`)

**Two Server Modes** in `main.go`:
- **Setup mode** (when `/data/config.yml` absent): only `/api/health`, `/api/setup/*`, `/api/auth/login` routes
- **Main mode** (normal): full route tree with public and `AuthRequired()`-protected groups

**Key Dependencies**: Gin, GORM, golang-migrate, lego/v4 (ACME), gorilla/websocket

### Frontend Structure (React + TypeScript + Vite)

- `src/App.tsx` — Route definitions, `SetupGuard` (checks `/api/setup/status`), theme management (dark/light via localStorage)
- `src/views/` — Page components (Dashboard, Sites, LoadBalancer, Certificates, Logs, Control, AccessControl, Audit, Profile, MetricsAnalysis)
- `src/components/` — Reusable components (DnsProviderDialog, ProtectedRoute, ErrorBoundary)
- `src/components/ui/` — shadcn/ui components (Radix-based)
- `src/api/request.ts` — Axios instance with `baseURL: '/api'`, Bearer token injection, auto-redirect on 401/403
- `src/contexts/AuthContext.tsx` — Auth state management, token in `localStorage('nginxops_token')`

**Key Dependencies**: React Router, axios, Radix UI, Tailwind CSS, echarts, recharts

### API Response Pattern

All endpoints return unified JSON:
```go
type ApiResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```
Use `pkg/response` helpers: `Success()`, `SuccessWithMessage()`, `Error()`, `BadRequest()`, `Unauthorized()`, `Forbidden()`, `NotFound()`, `InternalError()`.

**Important anti-pattern**: Many handlers use `response.Error(c, 200, ...)` for business logic errors, returning HTTP 200 with `success: false`. Clients must check the `success` field, not just HTTP status.

### Configuration

Loaded from `/data/config.yml` with env var fallback. Full config structure:

```yaml
server:
  port: 8080                    # SERVER_PORT
database:
  host: localhost               # DB_HOST
  port: 5432                    # DB_PORT
  name: nginxops                # DB_NAME
  user: postgres                # DB_USER
  password: ""                  # DB_PASSWORD
  sslmode: disable
jwt:
  secret: ""                    # JWT_SECRET (min 32 chars)
  expiration: 86400000          # Milliseconds (default 24h)
nginx:
  config-path: /etc/nginx/nginx.conf
  conf-dir: /data/nginx/conf.d  # Auto-set by DATA_DIR
  ssl-path: /data/nginx/ssl     # Auto-set by DATA_DIR
  access-log: /data/logs/nginx/access.log  # Auto-set by DATA_DIR
  reload-command: nginx -s reload
  test-command: nginx -t
acme:
  account-key-path: /data/nginx/acme  # Auto-set by DATA_DIR
```

`DATA_DIR` env var auto-sets `conf-dir`, `ssl-path`, `access-log`, and `account-key-path`.

### Nginx Configuration Generation

Backend generates Nginx configs in `/data/nginx/conf.d/` using `strings.Builder` (not Go templates):

- **Site configs** (`site_service.go` `buildNginxConfig()`): 3 site types — `static` (root + try_files), `proxy` (location blocks from JSON, supports custom headers + websocket), `loadbalance` (proxy_pass to upstream). SSL sites get HTTP→HTTPS redirect + TLS config. File naming: `{domain_with_dots→underscores}.conf`
- **Upstream configs** (`upstream_service.go` `buildUpstreamConfig()`): Supports `round_robin`, `ip_hash`, `least_conn`. Optional health check (requires `nginx_upstream_check_module`). File naming: `upstream_{name}.conf`
- **Config validation** (`nginx_service.go` `ValidateConfig()`): Copies full conf.d to temp dir, runs `nginx -t`, falls back to `true` if nginx binary absent
- **Config history**: All changes archived to `nginx_config_history` table before overwriting

### ACME Certificate Management

- Uses `go-acme/lego/v4`, DNS-01 challenge only
- Supported issuers: `letsencrypt`, `letsencrypt-staging`, `zerossl`
- DNS providers: Aliyun, Tencent Cloud, Cloudflare (via lego + custom `pkg/dnsprovider/`)
- Certificate requests are **asynchronous** (goroutine with `sync.Map` dedup)
- Status flow: `pending` → `validating` → `issuing` → `completed`/`failed`

### Database Migrations

- Stored in `backend/migrations/` as `{6-digit-number}_{name}.{up|down}.sql`
- Uses golang-migrate; auto-runs on app startup (`database.RunMigrations()`)
- 9 migrations; 15 tables total including: users, sites, upstreams, certificates, certificate_requests, dns_providers, access_log, audit_log, ip_geo_cache, nginx_config_history, stats_summary, access_rules, access_rule_items, site_access_rules, country_cidrs
- Schema conventions: `BIGSERIAL` PKs, `TIMESTAMPTZ` timestamps, JSON as `TEXT`, index naming `idx_{table}_{column}`

### Docker Deployment

**Not Supervisor** — `docker/entrypoint.sh` starts nginx as daemon, then `exec /app/nginxops` (Go binary as PID 1).

- **Port 8899**: Management panel (serves frontend static files, proxies `/api/` and `/ws/` to Go on 8080)
- **Ports 80/443**: User site traffic (default server returns 444 to reject unknown domains)
- Default self-signed SSL cert generated at `/data/nginx/ssl/default.crt` on first run
- Symlinks persist Nginx data to `/data/`: `conf.d`, `ssl`, `/var/log/nginx`
- Healthcheck: `curl -f http://localhost:8899/api/health && curl -f http://localhost:80/`

**docker-compose.yml**: PostgreSQL 16-alpine + NginxOps app. Named volumes for `nginxops-postgres-data` and `nginxops-data`.

### Authentication Flow

- JWT Bearer token, stored in `localStorage('nginxops_token')`
- `AuthRequired()` middleware validates tokens; expiration default 24h (86400000ms, configurable)
- Setup flow: if `/data/config.yml` missing → frontend shows `/welcome` wizard → `POST /api/setup/init` creates config + initializes DB → system restarts in normal mode

### Access Control Architecture

Rule-based model (refactored in migration 005):
- `access_rules` — Reusable rule definitions (IP whitelist/blacklist, geo-based)
- `access_rule_items` — Rule entries (IP ranges, country codes)
- `site_access_rules` — Many-to-many association between sites and rules
- Nginx config snippets for access control generated per-site via `buildAccessControlConfig()`
