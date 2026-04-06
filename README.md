# NginxOps

一个简洁高效的 Nginx 运维管理平台，提供 Web 界面管理 Nginx 站点、负载均衡、SSL 证书和访问日志。

## 功能特性

- **站点管理** - 可视化创建和管理 Nginx 站点配置
- **负载均衡** - 配置 Upstream 实现后端服务器负载均衡
- **SSL 证书** - 支持证书导入和 ACME 自动申请（Let's Encrypt、阿里云 DNS、腾讯云 DNS）
- **日志分析** - 实时查看访问日志，支持地理位置解析
- **配置管理** - 在线编辑 Nginx 配置，支持配置历史回溯
- **审计日志** - 记录所有操作，便于追溯

## 快速开始

创建 `docker-compose.yml` 文件：

```yaml
services:
  nginxops:
    image: nginxops:latest
    build:
      context: .
      dockerfile: Dockerfile
    container_name: nginxops
    restart: unless-stopped
    ports:
      - "${ADMIN_PORT:-8899}:8899"
      - "${HTTP_PORT:-80}:80"
      - "${HTTPS_PORT:-443}:443"
    volumes:
      - nginxops-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8899/api/health"]
      interval: 30s
      timeout: 10s
      start_period: 120s
      retries: 3
    networks:
      - nginxops-net

networks:
  nginxops-net:
    driver: bridge
    name: nginxops-net

volumes:
  nginxops-data:
    name: nginxops-data
```

```bash
# 从源码构建并启动
docker-compose up -d --build

# 访问管理面板
# http://localhost:8899
```

首次启动会进入初始化向导，按提示配置数据库和管理员账号即可。

### 端口说明

| 端口 | 用途 |
|------|------|
| 8899 | 管理面板 |
| 80 | HTTP 代理 |
| 443 | HTTPS 代理 |

## 技术栈

**后端**
- Go 1.24 + Gin
- PostgreSQL + GORM
- golang-migrate（数据库迁移）
- lego（ACME 证书管理）

**前端**
- React 18 + TypeScript
- Vite
- shadcn/ui + Tailwind CSS
- echarts + recharts

**部署**
- Docker + Supervisor
- Nginx

## 本地开发

### 后端

```bash
cd backend
go mod download
go run ./cmd/server
```

### 前端

```bash
cd frontend
npm install
npm run dev
```

## 许可证

[MIT License](LICENSE)
