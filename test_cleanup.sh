#!/bin/bash

# 测试清理功能的脚本
# 创建一些模拟的过期临时目录来测试清理功能

echo "=== OBC Coin API 清理功能测试 ==="

# 获取模板目录的父目录
TEMPLATE_DIR="/Users/mofei/work/obc/obc/bridge/move/tokens/coin_tmp"
PARENT_DIR=$(dirname "$TEMPLATE_DIR")

echo "模板目录: $TEMPLATE_DIR"
echo "父目录: $PARENT_DIR"

# 创建一些测试用的过期目录
echo "\n创建测试用的过期目录..."

# 创建15分钟前的目录
OLD_TIMESTAMP=$(($(date +%s) - 900))  # 15分钟前
OLD_DIR="$PARENT_DIR/coin_tmp_$OLD_TIMESTAMP"
mkdir -p "$OLD_DIR/sources"
echo "创建过期目录: $OLD_DIR"

# 创建5分钟前的目录（不应该被清理）
NEW_TIMESTAMP=$(($(date +%s) - 300))  # 5分钟前
NEW_DIR="$PARENT_DIR/coin_tmp_$NEW_TIMESTAMP"
mkdir -p "$NEW_DIR/sources"
echo "创建新目录: $NEW_DIR"

# 创建一个非coin_tmp格式的目录（不应该被清理）
OTHER_DIR="$PARENT_DIR/other_tmp_$OLD_TIMESTAMP"
mkdir -p "$OTHER_DIR"
echo "创建其他格式目录: $OTHER_DIR"

echo "\n当前临时目录列表:"
ls -la "$PARENT_DIR" | grep tmp

echo "\n等待清理功能运行..."
echo "提示: 清理功能每10分钟运行一次，或者重启服务器立即触发清理"
echo "\n可以通过以下命令查看服务器日志:"
echo "tail -f nohup.out"

echo "\n测试完成后，可以再次运行以下命令查看清理结果:"
echo "ls -la \"$PARENT_DIR\" | grep tmp"