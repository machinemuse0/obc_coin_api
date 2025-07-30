#!/bin/bash

# OBC Coin API 重启脚本

echo "正在重启 OBC Coin API 服务..."

# 停止服务
echo "步骤 1: 停止当前服务"
./stop.sh

# 等待一秒
sleep 1

# 启动服务
echo "步骤 2: 启动服务"
./start.sh