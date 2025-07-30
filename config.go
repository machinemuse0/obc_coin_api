package main

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config 定义应用配置结构
type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	CoinTemplatePath string `yaml:"coin_template_path"`
	BFC              struct {
		Directory  string `yaml:"directory"`
		BinaryPath string `yaml:"binary_path"`
	} `yaml:"bfc"`
	BenfenRPC struct {
		URL        string `yaml:"url"`
		Timeout    int    `yaml:"timeout"`
		RetryCount int    `yaml:"retry_count"`
	} `yaml:"benfen_rpc"`
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
	Cleanup struct {
		IntervalMinutes  int `yaml:"interval_minutes"`
		RetentionMinutes int `yaml:"retention_minutes"`
	} `yaml:"cleanup"`
	Log struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"log"`
}

// 全局配置变量
var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	// 解析 YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	AppConfig = &config
	return nil
}

// GetCoinTemplatePath 获取代币模板路径
func GetCoinTemplatePath() string {
	if AppConfig != nil {
		return AppConfig.CoinTemplatePath
	}
	return "./templates/coin_template.json" // 默认值
}

// GetServerAddress 获取服务器地址
func GetServerAddress() string {
	if AppConfig != nil {
		return fmt.Sprintf("%s:%d", AppConfig.Server.Host, AppConfig.Server.Port)
	}
	return "localhost:8080" // 默认值
}

// GetBFCDirectory 获取 BFC 目录路径
func GetBFCDirectory() string {
	if AppConfig != nil {
		return AppConfig.BFC.Directory
	}
	return "/usr/local/bfc" // 默认值
}

// GetBFCBinaryPath 获取 BFC 二进制文件路径
func GetBFCBinaryPath() string {
	if AppConfig != nil {
		return AppConfig.BFC.BinaryPath
	}
	return "/usr/local/bfc/bfc" // 默认值
}

// GetBenfenRPCURL 获取 Benfen RPC URL
func GetBenfenRPCURL() string {
	if AppConfig != nil {
		return AppConfig.BenfenRPC.URL
	}
	return "https://rpc.benfen.org" // 默认值
}

// GetBenfenRPCTimeout 获取 Benfen RPC 超时时间
func GetBenfenRPCTimeout() int {
	if AppConfig != nil {
		return AppConfig.BenfenRPC.Timeout
	}
	return 30 // 默认值
}

// GetBenfenRPCRetryCount 获取 Benfen RPC 重试次数
func GetBenfenRPCRetryCount() int {
	if AppConfig != nil {
		return AppConfig.BenfenRPC.RetryCount
	}
	return 3 // 默认重试3次
}

// GetCleanupIntervalMinutes 获取清理任务执行间隔（分钟）
func GetCleanupIntervalMinutes() int {
	if AppConfig != nil {
		return AppConfig.Cleanup.IntervalMinutes
	}
	return 10 // 默认10分钟
}

// GetCleanupRetentionMinutes 获取目录保留时间（分钟）
func GetCleanupRetentionMinutes() int {
	if AppConfig != nil {
		return AppConfig.Cleanup.RetentionMinutes
	}
	return 10 // 默认保留10分钟
}
