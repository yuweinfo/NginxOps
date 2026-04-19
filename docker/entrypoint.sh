#!/bin/bash
# =====================================================
# NginxOps - 统一启动入口
# 配置文件: /data/config.yml
# =====================================================

set -e

DATA_DIR="${DATA_DIR:-/data}"
PGDATA="${PGDATA:-/data/postgresql}"
CONFIG_FILE="${DATA_DIR}/config.yml"

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
    
    # setup模式下，将暂时禁用PostgreSQL
    # 注意：这里只是设置一个标志，真正的配置在下面的TEMP_SUPERVISOR_CONF中处理
    SETUP_MODE=true
    
    # 在setup模式下，我们假定使用内部数据库（PostgreSQL），但不自动启动它
    # 等待setup流程完成后再启动
    USE_EXTERNAL_DB=false
    
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
        # 设置标志，不在supervisor配置中包含PostgreSQL
        USE_EXTERNAL_DB=true
    else
        echo "Using internal database, PostgreSQL will be started by supervisor"
        # 设置标志，在supervisor配置中包含PostgreSQL
        USE_EXTERNAL_DB=false
        
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

# 将环境变量替换到supervisor配置中
echo "================================================"
echo "Preparing Supervisor configuration with environment variables..."
echo "================================================"

# 创建临时supervisor配置
TEMP_SUPERVISOR_CONF="/tmp/supervisord.conf"
cat > ${TEMP_SUPERVISOR_CONF} << EOF
[supervisord]
nodaemon=true
logfile=/var/log/supervisor/supervisord.log
pidfile=/var/run/supervisord.pid
user=root
logfile_maxbytes=50MB
logfile_backups=10
loglevel=info

[unix_http_server]
file=/var/run/supervisor.sock

[supervisorctl]
serverurl=unix:///var/run/supervisor.sock

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

EOF

# 判断是否启动PostgreSQL
# 非setup模式且使用内部数据库 → 自动启动 (autostart=true)
# setup模式 → 由setup handler启动 (autostart=false)
# 使用外部数据库 → 不启动
if [ "${USE_EXTERNAL_DB}" != "true" ]; then
    POSTGRES_AUTOSTART="true"
    if [ "${SETUP_MODE}" = "true" ]; then
        POSTGRES_AUTOSTART="false"
    fi
    echo "Configuring PostgreSQL (autostart=${POSTGRES_AUTOSTART})"
    cat >> ${TEMP_SUPERVISOR_CONF} << EOF
[program:postgresql]
command=/bin/bash -c "PGDATA=${PGDATA} /usr/bin/postgres -D ${PGDATA}"
user=postgres
autostart=${POSTGRES_AUTOSTART}
autorestart=true
stdout_logfile=${DATA_DIR}/logs/postgresql-stdout.log
stderr_logfile=${DATA_DIR}/logs/postgresql-stderr.log
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
priority=10
startsecs=10
startretries=5
environment=PGDATA="${PGDATA}",PGPORT="5432",DATA_DIR="${DATA_DIR}"

EOF
else
    echo "Using external database, PostgreSQL will NOT be started"
fi

# 添加Go后端和Nginx配置
# Go后端在setup模式下也能启动（提供Web UI），只是不连接数据库
cat >> ${TEMP_SUPERVISOR_CONF} << EOF
[program:go-backend]
command=/app/nginxops
directory=/app
autostart=true
autorestart=true
stdout_logfile=${DATA_DIR}/logs/app/stdout.log
stdout_logfile_maxbytes=50MB
stdout_logfile_backups=10
stdout_syslog=true
redirect_stderr=true
priority=20
startsecs=10
startretries=5
environment=DATA_DIR="${DATA_DIR}"

[program:nginx]
command=/usr/sbin/nginx -g "daemon off;"
autostart=true
autorestart=true
stdout_logfile=${DATA_DIR}/logs/nginx/supervisor-stdout.log
stderr_logfile=${DATA_DIR}/logs/nginx/supervisor-stderr.log
priority=30
startsecs=5
startretries=3
environment=DATA_DIR="${DATA_DIR}"
EOF

echo "================================================"
echo "Starting Supervisor..."
echo "================================================"

# 启动supervisor管理所有服务
exec /usr/bin/supervisord -c ${TEMP_SUPERVISOR_CONF}
