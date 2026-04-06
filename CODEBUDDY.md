# CODEBUDDY.md This file provides guidance to CodeBuddy when working with code in this repository.

## Project Overview

NginxOps is a web-based management interface for Nginx configuration, log analysis, and SSL certificate management. It consists of a Spring Boot backend and React frontend, deployed as a unified Docker image.

## Common Commands

### Backend (Java/Spring Boot)

```bash
# Build backend
cd backend && mvn clean package -DskipTests

# Run tests
cd backend && mvn test

# Run single test class
cd backend && mvn test -Dtest=ClassName

# Run single test method
cd backend && mvn test -Dtest=ClassName#methodName
```

### Frontend (React/TypeScript)

```bash
# Install dependencies
cd frontend && npm install

# Development server (proxies /api to localhost:8080)
cd frontend && npm run dev

# Build for production
cd frontend && npm run build

# Lint
cd frontend && npm run lint
```

### Docker Deployment

```bash
# Initialize environment (creates .env from .env.example)
./deploy.sh init

# Build Docker image
./deploy.sh build

# Start services
./deploy.sh start

# Stop services
./deploy.sh stop

# View logs
./deploy.sh logs [service]
```

## Architecture

### Backend Structure (Spring Boot 3.4 + Java 21)

```
backend/src/main/java/com/nginxpanel/
├── config/          # Security, WebSocket, CORS configuration
├── controller/      # REST API endpoints
├── service/         # Business logic layer
│   └── dns/         # DNS provider implementations (Strategy pattern)
├── repository/      # MyBatis-Plus mappers
├── model/entity/    # JPA entities
├── dto/             # Data transfer objects
├── parser/          # Nginx config and log parsers
├── security/        # JWT authentication filter and utilities
├── aspect/          # AOP aspects (audit logging)
└── websocket/       # WebSocket handlers for real-time log streaming
```

**Key Dependencies**: MyBatis-Plus (ORM), Flyway (migrations), jjwt (authentication), acme4j (Let's Encrypt), Alibaba/Tencent Cloud DNS SDKs.

**Database**: PostgreSQL with Flyway migrations in `backend/src/main/resources/db/migration/`. Schema reference in `sql/init.sql`.

**Authentication**: JWT-based stateless authentication. Tokens stored in localStorage, validated via `JwtAuthenticationFilter`. Public endpoints: `/api/auth/**`, `/api/health`, `/ws/**`.

### Frontend Structure (React 18 + TypeScript + Vite)

```
frontend/src/
├── api/             # Axios-based API client modules
├── components/      # Reusable UI components (shadcn/ui based)
│   └── ui/          # shadcn/ui primitives
├── contexts/        # React contexts (AuthContext)
├── hooks/           # Custom hooks
├── layouts/         # Page layout components
├── views/           # Page-level components (routes)
└── lib/             # Utility functions
```

**Key Features**:
- Dashboard: Statistics and charts (echarts, recharts)
- Sites: Nginx virtual host configuration
- LoadBalancer: Upstream/server pool management
- Certificates: SSL certificate management with ACME
- Logs: Real-time log viewing via WebSocket
- Control: Nginx reload/status control
- Audit: Operation audit log viewing

**API Client**: Centralized in `api/request.ts` with JWT token injection and 401/403 redirect handling.

### Docker Deployment

Single unified image containing:
1. **PostgreSQL** - Internal database (optional, can use external)
2. **Java Backend** - Spring Boot on port 8080
3. **Nginx** - Reverse proxy + business proxy on ports 80/443
4. **Frontend** - Static files served by Nginx

Managed by Supervisor (`docker/supervisord.conf`). Entry point (`docker/entrypoint.sh`) handles database initialization; Flyway handles schema migration on Spring Boot startup.

**Data Directory** (`/data`):
- `postgresql/` - Database files
- `nginx/conf.d/` - Site configurations
- `nginx/ssl/` - SSL certificates
- `logs/` - Application and Nginx logs
- `data/` - Application data (ACME keys, etc.)

### Configuration

Environment variables (see `.env.example`):
- `USE_EXTERNAL_DB` - Toggle internal/external PostgreSQL
- `DB_HOST/PORT/NAME/USER/PASSWORD` - Database connection
- `JWT_SECRET` - JWT signing key (256+ bits)
- `JAVA_OPTS` - JVM options

Backend config: `backend/src/main/resources/application.yml` with environment variable substitution.

Frontend dev proxy: `frontend/vite.config.ts` proxies `/api` → `localhost:8080`, `/ws` → WebSocket.
