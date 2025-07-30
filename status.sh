#!/bin/bash

# OBC Coin API 状态检查脚本

PID_FILE="obc_coin_api.pid"
LOG_FILE="obc_coin_api.log"

echo "=== OBC Coin API 服务状态 ==="

# 检查 PID 文件
if [ ! -f "$PID_FILE" ]; then
    echo "状态: ❌ 未运行 (未找到 PID 文件)"
    exit 1
fi

# 读取 PID
PID=$(cat "$PID_FILE")

# 检查进程是否存在
if ps -p $PID > /dev/null 2>&1; then
    echo "状态: ✅ 运行中"
    echo "PID: $PID"
    
    # 获取进程信息
    echo "进程信息:"
    ps -p $PID -o pid,ppid,cmd,etime,pcpu,pmem
    
    # 检查端口
    echo ""
    echo "端口检查:"
    if lsof -i :8080 > /dev/null 2>&1; then
        echo "✅ 端口 8080 已监听"
        lsof -i :8080
    else
        echo "❌ 端口 8080 未监听"
    fi
    
    # 健康检查
    echo ""
    echo "健康检查:"
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ 服务响应正常"
    else
        echo "⚠️  服务无响应或健康检查失败"
    fi
    
    # 显示最近日志
    echo ""
    echo "最近日志 (最后 10 行):"
    if [ -f "$LOG_FILE" ]; then
        tail -10 "$LOG_FILE"
    else
        echo "未找到日志文件"
    fi
    
else
    echo "状态: ❌ 未运行 (进程 $PID 不存在)"
    echo "清理残留的 PID 文件..."
    rm -f "$PID_FILE"
    exit 1
fi

echo ""
echo "=== 可用命令 ==="
echo "启动服务: ./start.sh"
echo "停止服务: ./stop.sh"
echo "重启服务: ./restart.sh"
echo "查看日志: tail -f $LOG_FILE"