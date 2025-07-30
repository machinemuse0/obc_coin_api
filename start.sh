#!/bin/bash

# OBC Coin API 启动脚本

PID_FILE="obc_coin_api.pid"
LOG_FILE="obc_coin_api.log"

# 检查是否已经在运行
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p $PID > /dev/null 2>&1; then
        echo "服务已经在运行中 (PID: $PID)"
        echo "如需重启，请先运行 ./stop.sh"
        exit 1
    else
        echo "发现残留的 PID 文件，正在清理..."
        rm -f "$PID_FILE"
    fi
fi

echo "正在启动 OBC Coin API 服务..."

# 检查 Go 环境
if ! command -v go &> /dev/null; then
    echo "错误: 未找到 Go 环境，请确保已安装 Go"
    exit 1
fi

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "错误: 未找到配置文件 config.yaml"
    exit 1
fi

# 启动服务
nohup go run . > "$LOG_FILE" 2>&1 &
PID=$!

# 保存 PID
echo $PID > "$PID_FILE"

# 等待服务启动
sleep 3

# 检查服务是否成功启动
if ps -p $PID > /dev/null 2>&1; then
    echo "✅ 服务启动成功!"
    echo "PID: $PID"
    echo "日志文件: $LOG_FILE"
    echo "配置检查:"
    
    # 检查服务是否响应
    sleep 2
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ 服务健康检查通过"
    else
        echo "⚠️  服务可能还在启动中，请查看日志: tail -f $LOG_FILE"
    fi
    
    echo ""
    echo "使用方法:"
    echo "  查看日志: tail -f $LOG_FILE"
    echo "  停止服务: ./stop.sh"
    echo "  重启服务: ./restart.sh"
else
    echo "❌ 服务启动失败，请查看日志: $LOG_FILE"
    rm -f "$PID_FILE"
    exit 1
fi