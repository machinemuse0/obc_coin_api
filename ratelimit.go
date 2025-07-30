package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

// IPRateLimit IP限流器结构
type IPRateLimit struct {
	mu           sync.RWMutex
	clients      map[string]*ClientLimiter
	cleanupTimer *time.Timer
}

// ClientLimiter 客户端限流器
type ClientLimiter struct {
	secondTokens int       // 每秒令牌数
	minuteTokens int       // 每分钟令牌数
	lastSecond   time.Time // 上次秒级重置时间
	lastMinute   time.Time // 上次分钟级重置时间
	lastAccess   time.Time // 最后访问时间
}

// NewIPRateLimit 创建新的IP限流器
func NewIPRateLimit() *IPRateLimit {
	rl := &IPRateLimit{
		clients: make(map[string]*ClientLimiter),
	}
	
	// 启动清理定时器，每5分钟清理一次不活跃的客户端
	rl.cleanupTimer = time.AfterFunc(5*time.Minute, rl.cleanup)
	
	return rl
}

// cleanup 清理不活跃的客户端（超过10分钟未访问）
func (rl *IPRateLimit) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	for ip, client := range rl.clients {
		if now.Sub(client.lastAccess) > 10*time.Minute {
			delete(rl.clients, ip)
		}
	}
	
	// 重新设置清理定时器
	rl.cleanupTimer.Reset(5 * time.Minute)
}

// Allow 检查IP是否允许访问
func (rl *IPRateLimit) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	
	// 获取或创建客户端限流器
	client, exists := rl.clients[ip]
	if !exists {
		client = &ClientLimiter{
			secondTokens: 1,  // 新用户初始有1个秒级令牌
			minuteTokens: 10, // 新用户初始有10个分钟级令牌
			lastSecond:   now,
			lastMinute:   now,
			lastAccess:   now,
		}
		rl.clients[ip] = client
	} else {
		client.lastAccess = now
		
		// 检查秒级限流 - 每秒重置为1个令牌
		if now.Sub(client.lastSecond) >= time.Second {
			client.secondTokens = 1
			client.lastSecond = now
		}
		
		// 检查分钟级限流 - 每分钟重置为10个令牌
		if now.Sub(client.lastMinute) >= time.Minute {
			client.minuteTokens = 10
			client.lastMinute = now
		}
	}
	
	// 检查是否有可用令牌
	if client.secondTokens > 0 && client.minuteTokens > 0 {
		client.secondTokens--
		client.minuteTokens--
		return true
	}
	
	return false
}

// 全局限流器实例
var globalRateLimit = NewIPRateLimit()

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取客户端IP
		clientIP := r.Header.Get("X-Real-IP")
		if clientIP == "" {
			clientIP = r.Header.Get("X-Forwarded-For")
		}
		if clientIP == "" {
			clientIP = r.RemoteAddr
		}
		
		// 简化IP地址（去掉端口号）
		if idx := strings.LastIndex(clientIP, ":"); idx != -1 {
			clientIP = clientIP[:idx]
		}
		
		// 检查限流
		if !globalRateLimit.Allow(clientIP) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			response := TokenResponse{
				Success: false,
				Message: "请求过于频繁，请稍后再试",
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// TokenAddRateLimitMiddleware 专门用于 /api/token/add 的限流中间件
func TokenAddRateLimitMiddleware(next http.Handler) http.Handler {
	return RateLimitMiddleware(next)
}