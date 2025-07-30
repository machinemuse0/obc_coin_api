package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TokenRequest 定义添加代币的请求结构
type TokenRequest struct {
	Icon        string `json:"icon"`
	Symbol      string `json:"symbol"`
	Decimal     int    `json:"decimal"`
	Name        string `json:"name"`
	CustomInfo  string `json:"custom_info"`
	Description string `json:"description,omitempty"`
}

// TokenResponse 定义响应结构
type TokenResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// addToken 处理添加代币的请求
func addToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 解析请求体
	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := TokenResponse{
			Success: false,
			Message: "无效的请求格式",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 验证必填字段
	if req.Symbol == "" || req.Name == "" {
		response := TokenResponse{
			Success: false,
			Message: "Symbol 和 Name 字段不能为空",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 处理模板文件替换
	outputFile, err := processTemplate(req)
	if err != nil {
		response := TokenResponse{
			Success: false,
			Message: fmt.Sprintf("模板处理失败: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := TokenResponse{
		Success: true,
		Message: "代币添加成功",
		Data: map[string]interface{}{
			"request":     req,
			"output_file": outputFile,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// processTemplate 处理模板文件替换
func processTemplate(req TokenRequest) (string, error) {
	// 获取原始模板目录路径
	originalTemplatePath := GetCoinTemplatePath()
	
	// 生成唯一的复制目录名
	timestamp := time.Now().Unix()
	newDirName := fmt.Sprintf("coin_tmp_%d", timestamp)
	newTemplatePath := filepath.Join(filepath.Dir(originalTemplatePath), newDirName)
	
	// 复制模板目录
	if err := copyDir(originalTemplatePath, newTemplatePath); err != nil {
		return "", fmt.Errorf("复制模板目录失败: %v", err)
	}
	
	// 获取模板文件路径（从原始目录读取）
	templatePath := filepath.Join(originalTemplatePath, "sources", "fast_coin.move")
	
	// 读取模板文件
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("读取模板文件失败: %v", err)
	}

	// 准备替换变量
	content := string(templateContent)
	
	// 替换变量
	content = strings.ReplaceAll(content, "DECIMALTMP", strconv.Itoa(req.Decimal))
	content = strings.ReplaceAll(content, "SYMBOLTMP", req.Symbol)
	content = strings.ReplaceAll(content, "NAMETMP", req.Name)
	
	// 处理 DESCRIPTIONTMP
	description := req.Description
	if description == "" {
		description = req.Name // 如果没有描述，使用名称作为描述
	}
	content = strings.ReplaceAll(content, "DESCRIPTIONTMP", description)
	
	// 处理 JSONTMP - 直接使用 custom_info 字段（它本身就是 JSON 字符串）
	// 对 custom_info 中的双引号进行转义
	customInfoEscaped := strings.ReplaceAll(req.CustomInfo, "\"", "\\\"")
	content = strings.ReplaceAll(content, "JSONTMP", customInfoEscaped)

	// 生成输出文件路径（在复制的目录中）
	outputDir := filepath.Join(newTemplatePath, "sources")
	outputFile := filepath.Join(outputDir, "fast_coin_1.move")
	
	// 写入输出文件
	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("写入输出文件失败: %v", err)
	}

	return outputFile, nil
}

// copyDir 递归复制目录
func copyDir(src, dst string) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// 遍历并复制每个条目
	for _, entry := range entries {
		// 跳过 build 目录
		if entry.IsDir() && entry.Name() == "build" {
			continue
		}
		
		// 跳过 fast_coin.move 文件
		if !entry.IsDir() && entry.Name() == "fast_coin.move" {
			continue
		}
		
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 递归复制子目录
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// 复制文件
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile 复制单个文件
func copyFile(src, dst string) error {
	// 打开源文件
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 获取源文件信息
	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	// 创建目标文件
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// 设置文件权限
	return os.Chmod(dst, srcInfo.Mode())
}