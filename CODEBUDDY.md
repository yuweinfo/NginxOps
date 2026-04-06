# CODEBUDDY.md This file provides guidance to CodeBuddy when working with code in this repository.

## Project Overview

NginxOps is a web-based Nginx management platform that provides a unified interface for managing Nginx sites, upstreams (load balancers), SSL certificates (with ACME support), and monitoring access logs. It packages Go backend, React frontend, PostgreSQL, and Nginx into a single Docker image managed by Supervisor.

## Development Commands

### Backend (Go)
```bash
cd backend
go build -o nginxops ./cmd/server           # Build server
go build -o migrate ./cmd/migrate           # Build migration tool
./nginxops                                 # Run server (port 8080)
./migrate -action up                       # Run migrations
./migrate -action down                     # Rollback migration
./migrate -action version                  # Check migration version
go mod download                            # Download dependencies
go run ./cmd/server                        # Run with hot reload (development)
```

### Frontend (React + Vite)
```bash
cd frontend
npm install                                # Install dependencies
npm run dev                                # Development server (hot reload)
npm run build                              # Production build (TypeScript + Vite)
npm run lint                               # Run ESLint
npm run preview                            # Preview production build
```

### Docker
```bash
docker-compose up --build                  # Build and start all services
docker build -t nginxops:latest .          # Build Docker image
```

## Architecture

### Backend Structure (Go + Gin)

**Layered Architecture**: Handler → Service → Repository → Database

- `cmd/server/main.go` - Application entry point, route definitions, dependency injection
- `cmd/migrate/main.go` - Database migration CLI tool
- `internal/config/` - YAML configuration loading (`/data/config.yml`), environment variable fallback
- `internal/database/` - PostgreSQL connection via GORM, migration execution using golang-migrate
- `internal/handler/` - HTTP handlers (Gin), request parsing, response formatting
- `internal/service/` - Business logic layer, coordinates repositories and external operations
- `internal/repository/` - Data access layer, GORM queries
- `internal/model/` - GORM model definitions
- `internal/middleware/` - JWT authentication (`AuthRequired`), CORS, audit logging
- `internal/websocket/` - WebSocket handler for real-time log streaming
- `pkg/` - Shared utilities: JWT, ACME (Let's Encrypt), Nginx config generation, HTTP response helpers

**Key Dependencies**: Gin (web), GORM (ORM), golang-migrate, lego (ACME), gorilla/websocket

### Frontend Structure (React + TypeScript + Vite)

- `src/main.tsx` - React entry point
- `src/App.tsx` - Route definitions, setup guard, theme management
- `src/views/` - Page components (Dashboard, Sites, LoadBalancer, Certificates, Logs, Control, Audit, Profile)
- `src/components/` - Reusable components (DnsProviderDialog, ProtectedRoute, ErrorBoundary)
- `src/components/ui/` - shadcn/ui components (Radix-based: Button, Dialog, Select, Toast, etc.)
- `src/api/` - API client modules using axios with JWT token injection
- `src/contexts/AuthContext.tsx` - Authentication state management
- `src/hooks/` - Custom hooks (toast, theme)

**Key Dependencies**: React Router, axios, Radix UI, Tailwind CSS, echarts, recharts

### Database Migrations

- Migrations stored in `backend/migrations/` with `{version}_{name}.up.sql` and `.down.sql` files
- Uses golang-migrate library
- Tables: users, sites, upstreams, certificates, certificate_requests, dns_providers, access_log, audit_log, ip_geo_cache, nginx_config_history

### Docker Deployment

The application runs as a unified container with Supervisor managing three processes:
1. **PostgreSQL** - Internal database (optional, can use external DB)
2. **Go Backend** - API server on port 8899
3. **Nginx** - Reverse proxy on ports 80/443, serves frontend static files

**First Run**: If `/data/config.yml` doesn't exist, the system enters setup mode where users configure database and admin credentials via web UI.

**Data Directory** (`/data/`):
- `config.yml` - Main configuration file
- `postgresql/` - PostgreSQL data
- `nginx/conf.d/` - Generated Nginx site configs
- `nginx/ssl/` - SSL certificates
- `logs/` - Application, Nginx, and PostgreSQL logs

### Authentication Flow

- JWT-based authentication with Bearer token
- Token stored in localStorage (`nginxops_token`)
- `AuthRequired()` middleware validates tokens on protected routes
- Token expiration: 24 hours (configurable)

### Nginx Configuration Generation

The backend generates Nginx configuration files in `/data/nginx/conf.d/`:
- Site configs with server blocks, SSL, proxy settings
- Upstream configs for load balancing
- Supports automatic certificate provisioning via ACME (Let's Encrypt, Aliyun DNS, Tencent Cloud DNS)

### API Response Pattern

All API endpoints return a unified JSON response structure:
```go
type ApiResponse struct {
    Success bool        `json:"success"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```
Use `pkg/response` helpers: `Success()`, `Error()`, `BadRequest()`, `Unauthorized()`, `Forbidden()`, `NotFound()`, `InternalError()`.

### Environment Variables

Configuration can be set via `/data/config.yml` or environment variables:
- `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD` - Database connection
- `JWT_SECRET` - JWT signing key (min 32 characters)
- `DATA_DIR` - Data directory path (default: `/data`)

### Setup Flow

On first run, if `/data/config.yml` doesn't exist, the system enters setup mode:
1. Frontend shows setup wizard at `/setup`
2. User configures database (internal/external) and admin credentials
3. `POST /api/setup/init` creates config file and initializes database
4. System restarts in normal mode with authentication enabled
