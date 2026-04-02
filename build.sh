#!/bin/bash

# Magic-ORM 代码生成器构建脚本

set -e

echo "🔨 开始构建 Magic-ORM 代码生成器..."

# 设置版本号
VERSION="1.0.0"
BINARY_NAME="gendb"

# 创建 bin 目录
mkdir -p bin

# 构建当前平台的版本
echo "📦 构建当前平台版本..."
go build -o bin/${BINARY_NAME} -ldflags="-s -w" ./cmd/gendb

echo "✅ 构建完成！"
echo ""
echo "📍 可执行文件位置:"
echo "   - ./bin/${BINARY_NAME}"
echo ""
echo "💡 使用方法:"
echo "   1. 添加到 PATH: export PATH=\$PATH:\$(pwd)/bin"
echo "   2. 直接使用：./bin/${BINARY_NAME} user product"
echo "   3. 查看帮助：./bin/${BINARY_NAME} -h"
echo ""

# 如果是 Windows 系统
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" ]]; then
    echo "🪟 Windows 用户可以将 bin/${BINARY_NAME}.exe 添加到系统 PATH"
fi
