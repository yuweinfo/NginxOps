# NginxOps

一个现代化的 Nginx 管理面板，提供 Web 界面来管理 Nginx 配置、分析访问日志、管理 SSL 证书等功能。

## 功能特性

### 站点管理
- 可视化创建和管理 Nginx 虚拟主机配置
- 支持反向代理、静态站点等多种站点类型
- 一键启用/禁用站点
- 自动生成 Nginx 配置文件

### 负载均衡
- 管理后端服务器池（Upstream）
- 支持轮询、IP哈希、最少连接等负载均衡策略
- 健康检查配置

### SSL 证书管理
- 支持 Let's Encrypt 和 ZeroSSL 自动申请
- 支持 DNS 验证（阿里云、腾讯云、Cloudflare）
- 证书自动续期
- 导入现有证书

### 日志分析
- 实时日志查看（WebSocket）
- 访问日志统计和分析
- IP 地理位置分析
- User-Agent 解析
- 响应时间、状态码统计

### 其他功能
- JWT 身份认证
- 操作审计日志
- 深色/浅色主题切换
- 响应式设计

## 技术栈

### 后端
- Java 21
- Spring Boot 3.4
- MyBatis-Plus
- PostgreSQL
- Flyway（数据库迁移）
- JWT 认证

### 前端
- React 18
- TypeScript
- Vite
- Tailwind CSS
- shadcn/ui
- ECharts / Recharts

### 基础设施
- Docker
- Supervisor
- Nginx

## 快速开始

### 使用 Docker Compose 部署（推荐）

```bash
# 1. 启动服务
docker compose up -d

# 2. 访问管理面板
# http://localhost:8899
# 默认账号: admin / admin123
```

服务启动时会自动完成：
- PostgreSQL 数据库初始化
- 数据库 Schema 迁移（Flyway）
- 默认管理员账号创建

### 环境要求

**Docker 部署**：
- Docker 20.10+
- Docker Compose 2.0+

**本地开发**：
- Java 21+
- Node.js 18+
- PostgreSQL 14+

> **提示**：默认使用预构建镜像 `docker.cnb.cool/yuweinfo/nginx:latest`。如需自定义配置（数据库密码、JWT密钥等），可复制 `.env.example` 为 `.env` 后修改，再启动服务。如需自行构建镜像，请运行 `docker compose build`。

## 配置说明

### 环境变量

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

主要配置项：

| 变量 | 说明 | 默认值 |
|------|------|--------|
| USE_EXTERNAL_DB | 是否使用外部数据库 | false |
| DB_HOST | 数据库主机 | 127.0.0.1 |
| DB_PORT | 数据库端口 | 5432 |
| DB_NAME | 数据库名称 | nginxops |
| DB_USER | 数据库用户 | postgres |
| DB_PASSWORD | 数据库密码 | - |
| JWT_SECRET | JWT 密钥（至少256位） | - |
| ADMIN_PORT | 管理面板端口 | 8899 |
| HTTP_PORT | HTTP 代理端口 | 80 |
| HTTPS_PORT | HTTPS 代理端口 | 443 |

### 使用外部数据库

```bash
# .env 文件配置
USE_EXTERNAL_DB=true
DB_HOST=your-db-host
DB_PORT=5432
DB_NAME=nginxops
DB_USER=your-username
DB_PASSWORD=your-password
```

## 开发指南

### 后端开发

```bash
cd backend

# 安装依赖并编译
mvn clean install

# 运行测试
mvn test

# 运行单个测试
mvn test -Dtest=ClassName

# 启动开发服务器（需要先配置数据库）
mvn spring-boot:run
```

后端默认运行在 `http://localhost:8080`

### 前端开发

```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 代码检查
npm run lint

# 构建生产版本
npm run build
```

前端开发服务器会自动代理 `/api` 请求到后端 `localhost:8080`。

### 数据库迁移

数据库迁移由 Flyway 管理，迁移文件位于：
```
backend/src/main/resources/db/migration/
```

命名规范：`V{version}__Description.sql`

## 部署说明

### Docker 部署脚本

```bash
./deploy.sh <command>

可用命令：
  init      初始化环境（创建 .env 文件）
  build     构建 Docker 镜像
  start     启动服务
  stop      停止服务
  restart   重启服务
  logs      查看日志
  backup    备份数据
  clean     清理所有容器和镜像
```

### 数据持久化

所有数据存储在 `/Data/Docker/DockerData/nginxops` 目录：
```
nginxops/
├── postgresql/    # 数据库文件
├── nginx/
│   ├── conf.d/    # 站点配置
│   └── ssl/       # SSL 证书
├── logs/          # 日志文件
│   ├── app/       # 应用日志
│   └── nginx/     # Nginx 日志
└── data/          # 其他数据
```

### 端口说明

- **8899**: 管理面板端口
- **80**: HTTP 业务代理端口
- **443**: HTTPS 业务代理端口

## 项目结构

```
.
├── backend/                 # 后端代码
│   ├── src/main/java/com/nginxops/
│   │   ├── config/         # 配置类
│   │   ├── controller/     # 控制器
│   │   ├── service/        # 服务层
│   │   ├── repository/     # 数据访问层
│   │   ├── model/          # 实体模型
│   │   ├── security/       # 安全相关
│   │   └── websocket/      # WebSocket 处理
│   └── src/main/resources/
│       ├── application.yml # 应用配置
│       └── db/migration/   # 数据库迁移
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── api/           # API 接口
│   │   ├── components/    # 组件
│   │   ├── views/         # 页面
│   │   ├── contexts/      # Context
│   │   └── hooks/         # 自定义 Hooks
│   └── vite.config.ts     # Vite 配置
├── docker/                 # Docker 相关文件
│   ├── entrypoint.sh      # 容器入口脚本
│   ├── supervisord.conf   # Supervisor 配置
│   └── nginx.conf         # Nginx 配置
├── sql/                    # SQL 脚本
├── Dockerfile              # Docker 镜像构建
├── docker-compose.yml      # Docker Compose 配置
└── deploy.sh               # 部署脚本
```

## API 文档

### 认证接口

- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出

### 站点管理

- `GET /api/sites` - 获取站点列表
- `POST /api/sites` - 创建站点
- `PUT /api/sites/{id}` - 更新站点
- `DELETE /api/sites/{id}` - 删除站点

### 证书管理

- `GET /api/certificates` - 获取证书列表
- `POST /api/certificates` - 申请证书
- `POST /api/certificates/import` - 导入证书

### 统计分析

- `GET /api/stats/summary` - 获取统计概览
- `GET /api/stats/hourly` - 获取小时级统计
- `GET /api/stats/top-ips` - 获取热门 IP

### WebSocket

- `/ws/logs` - 实时日志流

## 安全建议

1. **修改默认密码**：首次登录后立即修改 admin 密码
2. **配置 JWT 密钥**：使用强随机字符串作为 JWT_SECRET
3. **使用 HTTPS**：生产环境建议配置 HTTPS
4. **数据库安全**：使用强密码并限制数据库访问
5. **防火墙配置**：仅开放必要端口

## 故障排查

### 查看日志

```bash
# 查看所有日志
./deploy.sh logs

# 查看特定服务日志
./deploy.sh logs nginxops
```

### 常见问题

1. **服务无法启动**
   - 检查端口是否被占用
   - 检查 `.env` 配置是否正确
   - 查看日志获取详细错误信息

2. **数据库连接失败**
   - 确认数据库服务已启动
   - 检查数据库连接配置
   - 确认数据库用户权限

3. **证书申请失败**
   - 检查 DNS 解析是否正确
   - 确认 DNS API 密钥配置正确
   - 查看详细错误日志

## 许可证

[MIT License](LICENSE)

## 贡献

欢迎提交 Issue 和 Pull Request。
