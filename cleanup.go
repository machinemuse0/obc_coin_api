package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

// cleanupOldDirectories 清理超过指定时间的临时目录
func cleanupOldDirectories() {
	// 获取模板目录的父目录
	originalTemplatePath := GetCoinTemplatePath()
	parentDir := filepath.Dir(originalTemplatePath)

	// 读取父目录下的所有文件和目录
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		log.Printf("清理任务: 读取目录失败 %s: %v", parentDir, err)
		return
	}

	// 正则表达式匹配 coin_tmp_时间戳 格式的目录
	tmpDirPattern := regexp.MustCompile(`^coin_tmp_(\\d+)$`)
	currentTime := time.Now().Unix()
	// 从配置文件获取保留时间
	retentionMinutes := GetCleanupRetentionMinutes()
	cleanupThreshold := int64(retentionMinutes * 60) // 转换为秒

	cleanedCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 检查目录名是否匹配模式
		matches := tmpDirPattern.FindStringSubmatch(entry.Name())
		if len(matches) != 2 {
			continue
		}

		// 解析时间戳
		timestamp, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			log.Printf("清理任务: 解析时间戳失败 %s: %v", entry.Name(), err)
			continue
		}

		// 检查是否超过清理阈值
		if currentTime-timestamp > cleanupThreshold {
			dirPath := filepath.Join(parentDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				log.Printf("清理任务: 删除目录失败 %s: %v", dirPath, err)
			} else {
				age := time.Duration(currentTime-timestamp) * time.Second
				log.Printf("清理任务: 成功删除过期目录 %s (存在时间: %v)", entry.Name(), age)
				cleanedCount++
			}
		}
	}

	if cleanedCount > 0 {
		log.Printf("清理任务完成: 共清理 %d 个过期目录", cleanedCount)
	} else {
		log.Printf("清理任务完成: 无需清理的目录")
	}
}

// startCleanupScheduler 启动定时清理任务
func startCleanupScheduler() {
	// 从配置文件获取清理间隔和保留时间
	intervalMinutes := GetCleanupIntervalMinutes()
	retentionMinutes := GetCleanupRetentionMinutes()
	log.Printf("启动定时清理任务: 每%d分钟清理一次超过%d分钟的临时目录", intervalMinutes, retentionMinutes)
	
	// 立即执行一次清理
	go cleanupOldDirectories()
	
	// 设置定时器，使用配置的间隔时间
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	go func() {
		for range ticker.C {
			cleanupOldDirectories()
		}
	}()
}