package middleware

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"server-panel/internal/session"
)

// 使用路由器包中的类型定义
import "server-panel/internal/router"

type HandlerFunc = router.HandlerFunc
type MiddlewareFunc = router.MiddlewareFunc

// Logger 日志中间件
func Logger() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应记录器
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next(recorder, r)

			duration := time.Since(start)
			log.Printf("%s %s %d %v %s",
				r.Method,
				r.URL.Path,
				recorder.statusCode,
				duration,
				r.RemoteAddr,
			)
		}
	}
}

// Recovery 恢复中间件
func Recovery() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					log.Printf("Panic recovered: %v\n%s", err, debug.Stack())

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, `{"success": false, "error": "Internal server error"}`)
				}
			}()

			next(w, r)
		}
	}
}

// CORS 跨域中间件
func CORS() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}
}

// responseRecorder 响应记录器
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Flush 实现 http.Flusher 接口
func (r *responseRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现 http.Hijacker 接口
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// Auth 认证中间件
func Auth(store *session.Store) MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 检查会话
			sess, err := store.Get(r)
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			// 检查是否已登录
			if sess.Get("authenticated") != true {
				redirectToLogin(w, r)
				return
			}

			next(w, r)
		}
	}
}

// redirectToLogin 重定向到登录页面
func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	// 如果是 API 请求，返回 JSON 错误
	if strings.HasPrefix(r.URL.Path, "/api/") {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `{"success": false, "error": "未授权访问"}`)
		return
	}

	// 否则重定向到登录页面
	http.Redirect(w, r, "/login", http.StatusFound)
}

// Performance 性能监控中间件 - 优化版本
func Performance() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// 创建响应记录器
			recorder := &responseRecorder{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next(recorder, r)

			duration := time.Since(start)
			method := r.Method
			path := r.URL.Path
			status := recorder.statusCode
			clientIP := getClientIP(r)

			// 记录所有请求的性能（简化版本）
			log.Printf("%s %s %d %v %s",
				method, path, status, duration, clientIP)

			// 记录慢请求警告
			if duration > 50*time.Millisecond {
				log.Printf("⚠️ 慢请求警告: %s %s - %v (来自: %s)",
					method, path, duration, clientIP)
			}

			// 记录超慢请求错误
			if duration > 200*time.Millisecond {
				log.Printf("❌ 超慢请求: %s %s - %v (来自: %s)",
					method, path, duration, clientIP)
			}

			// 添加性能头部
			w.Header().Set("X-Response-Time", duration.String())
		}
	}
}

// RateLimit 简单的速率限制中间件
func RateLimit() MiddlewareFunc {
	// 简单的内存存储，生产环境应使用Redis
	requests := make(map[string][]time.Time)
	var mutex sync.Mutex

	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			defer mutex.Unlock()

			clientIP := getClientIP(r)
			now := time.Now()

			// 清理过期记录
			if times, exists := requests[clientIP]; exists {
				var validTimes []time.Time
				for _, t := range times {
					if now.Sub(t) < time.Minute {
						validTimes = append(validTimes, t)
					}
				}
				requests[clientIP] = validTimes
			}

			// 检查请求频率（每分钟最多60个请求）
			if len(requests[clientIP]) >= 60 {
				w.WriteHeader(http.StatusTooManyRequests)
				fmt.Fprintf(w, `{"success": false, "error": "请求过于频繁"}`)
				return
			}

			// 记录当前请求
			requests[clientIP] = append(requests[clientIP], now)

			next(w, r)
		}
	}
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	// 检查X-Forwarded-For头
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// 检查X-Real-IP头
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}
