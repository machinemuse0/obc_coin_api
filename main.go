package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 加载配置文件
	if err := LoadConfig("config.yaml"); err != nil {
		log.Printf("加载配置文件失败: %v，使用默认配置", err)
	}

	// 检查 BFC 目录是否存在
	bfcDir := GetBFCDirectory()
	if _, err := os.Stat(bfcDir); os.IsNotExist(err) {
		log.Fatalf("BFC 目录不存在: %s，请检查配置文件中的 bfc.directory 路径", bfcDir)
	} else if err != nil {
		log.Fatalf("检查 BFC 目录时发生错误: %v", err)
	}
	log.Printf("BFC 目录检查通过: %s", bfcDir)

	r := chi.NewRouter()

	// 基础中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 设置路由
	r.Route("/api", func(r chi.Router) {
		r.Route("/token", func(r chi.Router) {
			// 为 /add 路由添加限流中间件
			r.With(TokenAddRateLimitMiddleware).Post("/add", addToken)
			r.Post("/publish", publishToken)
		})
	})

	// 启动服务器
	serverAddr := GetServerAddress()
	log.Printf("服务器启动在地址: %s", serverAddr)
	log.Printf("代币模板路径: %s", GetCoinTemplatePath())
	log.Printf("BFC 目录: %s", GetBFCDirectory())
	log.Printf("BFC 二进制路径: %s", GetBFCBinaryPath())
	log.Printf("Benfen RPC URL: %s", GetBenfenRPCURL())
	log.Printf("Benfen RPC 超时: %d 秒", GetBenfenRPCTimeout())
	log.Printf("Benfen RPC 重试次数: %d", GetBenfenRPCRetryCount())
	
	// 启动定时清理任务
	startCleanupScheduler()
	
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", AppConfig.Server.Port), r))
}
