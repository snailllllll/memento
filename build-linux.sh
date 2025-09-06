#!/bin/bash
# 针对 Linux 5.10.134-18.al8.x86_64 内核的交叉编译脚本
# 在 macOS 上编译，目标运行环境：Linux 5.10.134-18.al8.x86_64
set -e

# 项目根目录
PROJECT_ROOT=$(cd "$(dirname "$0")" && pwd)
cd "$PROJECT_ROOT"

# 清理旧构建
# rm -rf bin/linux
mkdir -p bin/linux

# 设置交叉编译环境变量
export CGO_ENABLED=0                    # 禁用 CGO，避免 glibc 版本问题
export GOOS=linux                       # 目标操作系统
export GOARCH=amd64                     # 目标架构（x86_64）
export GOAMD64=v3                       # 针对较新 CPU 优化，兼容 5.10 内核

# 构建参数优化
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-s -w -extldflags '-static'"  # 静态链接，减小体积

echo "🚀 开始为 Linux 5.10.134-18.al8.x86_64 交叉编译..."
echo "构建时间: $BUILD_TIME"
echo "目标内核: 5.10.134-18.al8.x86_64"

# 执行构建
go build \
    -a \
    -installsuffix cgo \
    -ldflags "$LDFLAGS" \
    -tags netgo \
    -o bin/linux/memento_backend_linux_amd64 \
    ./main.go

# 验证构建结果
if [ -f bin/linux/memento_backend_linux_amd64 ]; then
    echo "✅ 编译成功！"
    echo "📁 文件位置: bin/linux/memento_backend_linux_amd64"
    
    # 检查文件信息
    if command -v file >/dev/null 2>&1; then
        echo "📋 文件信息:"
        file bin/linux/memento_backend_linux_amd64
    fi
    
    # 检查是否为静态链接
    if command -v otool >/dev/null 2>&1; then
        echo "🔗 链接检查:"
        otool -L bin/linux/memento_backend_linux_amd64 | head -5
    fi
    
    # 显示文件大小
    ls -lh bin/linux/memento_backend_linux_amd64
    
    echo ""
    echo "🎯 该可执行文件可在 Linux 5.10.134-18.al8.x86_64 上直接运行"
    echo "💡 复制到目标服务器: scp bin/linux/memento_backend_linux_amd64 user@server:/path/"
else
    echo "❌ 编译失败"
    exit 1
fi