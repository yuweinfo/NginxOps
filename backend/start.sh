#!/bin/bash

# NginxOps Go Backend 启动脚本

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}======================================${NC}"
echo -e "${GREEN}   NginxOps Go Backend${NC}"
echo -e "${GREEN}======================================${NC}"

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo -e "${YELLOW}Warning: config.yaml not found, using default config${NC}"
fi

# 检查Go环境
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}Go version: $(go version)${NC}"

# 编译项目
echo -e "${YELLOW}Building...${NC}"
if ! go build -o nginxops ./cmd/server; then
    echo -e "${RED}Build failed${NC}"
    exit 1
fi

echo -e "${GREEN}Build successful${NC}"

# 启动服务
echo -e "${GREEN}Starting server...${NC}"
./nginxops
