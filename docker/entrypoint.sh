#!/bin/bash
# =====================================================
# NginxOps - 统一启动入口
# 配置文件: /data/config.yml
# =====================================================

set -e

DATA_DIR="${DATA_DIR:-/data}"
PGDATA="${PGDATA:-/data/postgresql}"
CONFIG_FILE="/data/config.yml"

echo "================================================"
echo "NginxOps Starting..."
echo "================================================"
echo "DATA_DIR: ${DATA_DIR}"
echo "CONFIG_FILE: ${CONFIG_FILE}"
echo "================================================"

# 创建必要的目录
mkdir -p ${DATA_DIR}/logs/app
mkdir -p ${DATA_DIR}/logs/nginx
mkdir -p ${DATA_DIR}/nginx/conf.d
mkdir -p ${DATA_DIR}/nginx/ssl
mkdir -p ${DATA_DIR}/postgresql
mkdir -p ${DATA_DIR}/data
mkdir -p ${DATA_DIR}/backups
mkdir -p /run/nginx
mkdir -p /var/log/supervisor

# 设置日志目录权限
chown -R postgres:postgres ${DATA_DIR}/logs
chmod 755 ${DATA_DIR}/logs

# 创建空的日志文件
touch ${DATA_DIR}/logs/app/stdout.log
chmod 644 ${DATA_DIR}/logs/app/stdout.log

# 启动日志转发服务
/usr/bin/tail -f ${DATA_DIR}/logs/app/stdout.log 2>/dev/null &
TAIL_PID=$!

trap "kill $TAIL_PID 2>/dev/null" EXIT

# 创建默认SSL证书
if [ ! -f ${DATA_DIR}/nginx/ssl/default.crt ]; then
    echo "Generating default SSL certificate..."
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
        -keyout ${DATA_DIR}/nginx/ssl/default.key \
        -out ${DATA_DIR}/nginx/ssl/default.crt \
        -subj "/CN=default" 2>/dev/null
fi

# =====================================================
# 检查配置文件是否存在
# =====================================================
if [ ! -f "${CONFIG_FILE}" ]; then
    echo "================================================"
    echo "Config file not found: ${CONFIG_FILE}"
    echo "System will enter setup mode."
    echo "Please access the web interface to complete setup."
    echo "================================================"
    
    # 设置 PostgreSQL 为手动启动模式（不自动启动，但保留配置）
    sed -i '/\[program:postgresql\]/,/^\[/ s/autostart=true/autostart=false/' /etc/supervisord.conf
    
    # 预先创建并设置 PostgreSQL 数据目录权限（setup 模式下也需要）
    # 这样 initdb 在以 postgres 用户运行时才能成功
    mkdir -p ${DATA_DIR}/postgresql
    chown -R postgres:postgres ${DATA_DIR}/postgresql 2>/dev/null || true
    chmod 700 ${DATA_DIR}/postgresql 2>/dev/null || true
else
    echo "Config file found, loading configuration..."
    
    # 解析配置文件获取数据库设置
    USE_EXTERNAL_DB=$(grep -E "^\s*useExternal:" ${CONFIG_FILE} | awk '{print $2}' || echo "false")
    
    if [ "${USE_EXTERNAL_DB}" = "true" ]; then
        echo "Using external database, PostgreSQL will not be started"
        sed -i '/^\[program:postgresql\]/,/^$/d' /etc/supervisord.conf
    else
        echo "Using internal database, PostgreSQL will be started by supervisor"
        
        # 确保 PostgreSQL 自动启动（setup 模式可能已将其设为 false）
        sed -i '/\[program:postgresql\]/,/^\[/ s/autostart=false/autostart=true/' /etc/supervisord.conf
        
        # 设置权限
        chown -R postgres:postgres ${DATA_DIR}/postgresql
        chmod 700 ${DATA_DIR}/postgresql
        
        # 创建PostgreSQL日志文件
        touch ${DATA_DIR}/logs/postgresql.log
        chown postgres:postgres ${DATA_DIR}/logs/postgresql.log
        chmod 644 ${DATA_DIR}/logs/postgresql.log
        
        # 初始化数据库集群
        if [ ! -d "${PGDATA}/base" ]; then
            echo "Initializing PostgreSQL database cluster..."
            su - postgres -c "initdb -D ${PGDATA} --encoding=UTF8 --locale=en_US.UTF-8"
            
            # 配置PostgreSQL
            cat >> ${PGDATA}/postgresql.conf << EOF
listen_addresses = '127.0.0.1'
port = 5432
log_directory = '${DATA_DIR}/logs'
log_filename = 'postgresql.log'
EOF
            
            # 配置访问权限
            cat >> ${PGDATA}/pg_hba.conf << EOF
local   all             all                                     trust
host    all             all             127.0.0.1/32            trust
EOF
        fi
        
        # 从配置文件读取数据库配置
        DB_NAME=$(grep -E "^\s*name:" ${CONFIG_FILE} | awk '{print $2}' || echo "nginxops")
        DB_USER=$(grep -E "^\s*user:" ${CONFIG_FILE} | awk '{print $2}' || echo "postgres")
        DB_PASSWORD=$(grep -E "^\s*password:" ${CONFIG_FILE} | awk '{print $2}' || echo "postgres")
        
        # 启动PostgreSQL
        echo "Starting PostgreSQL for initialization..."
        su - postgres -c "pg_ctl -D ${PGDATA} -l ${DATA_DIR}/logs/postgresql.log start"
        sleep 3
        
        # 创建数据库和用户（如果不存在）
        if ! su - postgres -c "psql -h 127.0.0.1 -d postgres -lqt | cut -d \| -f 1 | grep -qw ${DB_NAME}"; then
            echo "Creating database and user..."
            su - postgres -c "psql -h 127.0.0.1 -d postgres -c \"CREATE USER \\\"${DB_USER}\\\" WITH PASSWORD '${DB_PASSWORD}';\"" || true
            su - postgres -c "psql -h 127.0.0.1 -d postgres -c \"CREATE DATABASE \\\"${DB_NAME}\\\" OWNER \\\"${DB_USER}\\\";\"" || true
            su - postgres -c "psql -h 127.0.0.1 -d postgres -c \"GRANT ALL PRIVILEGES ON DATABASE \\\"${DB_NAME}\\\" TO \\\"${DB_USER}\\\";\"" || true
        fi
        
        echo "Database ready."
        
        # 停止PostgreSQL（让supervisor管理）
        su - postgres -c "pg_ctl -D ${PGDATA} stop"
    fi
fi

# 设置Nginx配置目录权限
chmod -R 755 ${DATA_DIR}/nginx

echo "================================================"
echo "Starting Supervisor..."
echo "================================================"

# 启动supervisor管理所有服务
exec /usr/bin/supervisord -c /etc/supervisord.conf
