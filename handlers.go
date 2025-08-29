package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
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
	Success bool        `json:"success"`
	Message string      `json:"message"`
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

	// 获取项目目录（复制的模板目录）
	projectDir := filepath.Dir(filepath.Dir(outputFile)) // 从 sources/fast_coin_1.move 回到项目根目录

	// 编译 Move 项目
	compileOutput, err := compileMoveProject(projectDir)
	if err != nil {
		response := TokenResponse{
			Success: false,
			Message: fmt.Sprintf("编译失败: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 打印编译输出
	fmt.Printf("编译输出:\n%s\n", compileOutput)

	// 解析编译输出
	modules, dependencies, err := parseCompileOutput(compileOutput)
	if err != nil {
		response := TokenResponse{
			Success: false,
			Message: fmt.Sprintf("解析编译输出失败: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := TokenResponse{
		Success: true,
		Message: "代币添加和编译成功",
		Data: map[string]interface{}{
			"request": req,
			// "output_file":    outputFile,
			"compile_output": compileOutput,
			"modules":        modules,
			"dependencies":   dependencies,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func parseCompileOutput(compileOutput string) ([]string, []string, error) {
	// 查找JSON开始的位置（第一个{）
	start := strings.Index(compileOutput, "{")
	if start == -1 {
		return nil, nil, fmt.Errorf("无法在编译输出中找到JSON数据")
	}

	// 提取JSON部分
	jsonStr := compileOutput[start:]

	// 解析JSON
	var result struct {
		Modules      []string `json:"modules"`
		Dependencies []string `json:"dependencies"`
		Digest       []int    `json:"digest"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, nil, fmt.Errorf("解析JSON失败: %v", err)
	}

	return result.Modules, result.Dependencies, nil
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

// compileMoveProject 编译 Move 项目
func compileMoveProject(projectDir string) (string, error) {
	// 获取 BFC 二进制文件路径
	bfcBinaryPath := GetBFCBinaryPath()

	// 构建命令
	cmd := exec.Command(bfcBinaryPath, "move", "build", "--dump-bytecode-as-base64")
	cmd.Dir = projectDir

	// 执行命令并获取输出
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("编译失败: %v, 输出: %s", err, string(output))
	}

	return string(output), nil
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

// PublishRequest 定义发布请求的结构
type PublishRequest struct {
	Sender          string        `json:"sender"`
	CompiledModules []interface{} `json:"compiled_modules"`
	Dependencies    []interface{} `json:"dependencies"`
	GasBudget       string        `json:"gas_budget"`
}

// publishToken 处理发布代币的请求，转发到 Benfen RPC
func publishToken(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 解析请求体
	var req PublishRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := TokenResponse{
			Success: false,
			Message: "无效的请求格式",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 构建完整的 JSON-RPC 请求
	rpcRequest := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "unsafe_publish",
		"params":  []interface{}{req.Sender, req.CompiledModules, req.Dependencies, nil, req.GasBudget},
	}

	// 序列化请求体
	requestBody, err := json.Marshal(rpcRequest)
	if err != nil {
		response := TokenResponse{
			Success: false,
			Message: "序列化请求失败",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 获取 Benfen RPC URL
	rpcURL := GetBenfenRPCURL()

	// 创建转发请求
	httpReq, err := http.NewRequest("POST", rpcURL, bytes.NewBuffer(requestBody))
	if err != nil {
		response := TokenResponse{
			Success: false,
			Message: "创建转发请求失败",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 设置请求头
	httpReq.Header.Set("Accept", "*/*")
	httpReq.Header.Set("Accept-Encoding", "gzip, deflate, br")
	httpReq.Header.Set("Connection", "keep-alive")
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "PostmanRuntime-ApipostRuntime/1.1.0")

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		response := TokenResponse{
			Success: false,
			Message: fmt.Sprintf("转发请求失败: %v", err),
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		response := TokenResponse{
			Success: false,
			Message: "读取响应失败",
		}
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// 设置响应状态码和头部
	w.WriteHeader(resp.StatusCode)
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// 直接转发响应体
	w.Write(respBody)
}
