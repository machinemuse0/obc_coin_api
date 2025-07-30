package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// 加载配置文件
	if err := LoadConfig("config.yaml"); err != nil {
		log.Printf("加载配置文件失败: %v，使用默认配置", err)
	}

	r := chi.NewRouter()

	// 基础中间件
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// 设置路由
	r.Route("/api", func(r chi.Router) {
		r.Route("/token", func(r chi.Router) {
			r.Post("/add", addToken)
		})
	})

	// 启动服务器
	serverAddr := GetServerAddress()
	log.Printf("服务器启动在地址: %s", serverAddr)
	log.Printf("代币模板路径: %s", GetCoinTemplatePath())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", AppConfig.Server.Port), r))
}