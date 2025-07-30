#!/bin/bash

# OBC Coin API 停止脚本

PID_FILE="obc_coin_api.pid"
LOG_FILE="obc_coin_api.log"

# 检查 PID 文件是否存在
if [ ! -f "$PID_FILE" ]; then
    echo "未找到 PID 文件，服务可能未运行"
    exit 1
fi

# 读取 PID
PID=$(cat "$PID_FILE")

# 检查进程是否存在
if ! ps -p $PID > /dev/null 2>&1; then
    echo "进程 $PID 不存在，清理 PID 文件"
    rm -f "$PID_FILE"
    exit 1
fi

echo "正在停止 OBC Coin API 服务 (PID: $PID)..."

# 发送 TERM 信号
kill $PID

# 等待进程结束
for i in {1..10}; do
    if ! ps -p $PID > /dev/null 2>&1; then
        echo "✅ 服务已成功停止"
        rm -f "$PID_FILE"
        exit 0
    fi
    echo "等待进程结束... ($i/10)"
    sleep 1
done

# 如果进程仍然存在，强制终止
echo "进程未响应 TERM 信号，强制终止..."
kill -9 $PID

# 再次检查
if ! ps -p $PID > /dev/null 2>&1; then
    echo "✅ 服务已强制停止"
    rm -f "$PID_FILE"
else
    echo "❌ 无法停止服务，请手动处理"
    exit 1
fi