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
	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"database"`
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