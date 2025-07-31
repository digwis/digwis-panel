package server

import (
	"compress/gzip"
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"server-panel/internal/router"
	"server-panel/internal/middleware"
	"server-panel/internal/handlers"
	"server-panel/internal/system"
	"server-panel/internal/environment"
	"server-panel/internal/projects"
)

// Server 原生 HTTP 服务器
type Server struct {
	httpServer *http.Server
	router     *router.Router
	config     *Config
}

// Config 服务器配置
type Config struct {
	Host         string        `json:"host"`
	Port         string        `json:"port"`
	Debug        bool          `json:"debug"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
	StaticFiles  embed.FS      `json:"-"` // 嵌入的静态文件
}

// DefaultConfig 默认配置 - 优化版本
func DefaultConfig() *Config {
	return &Config{
		Host:         "127.0.0.1",
		Port:         "8080",
		Debug:        false,
		ReadTimeout:  30 * time.Second,  // 增加读取超时，适应VPS环境
		WriteTimeout: 30 * time.Second,  // 增加写入超时，适应SSE长连接
		IdleTimeout:  120 * time.Second, // 增加空闲超时，减少连接重建
	}
}

// New 创建新的服务器实例
func New(config *Config) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	// 创建路由器
	r := router.New()

	// 添加全局中间件
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.Performance()) // 性能监控
	r.Use(middleware.CORS())
	// 注意：RateLimit中间件可以根据需要启用
	// r.Use(middleware.RateLimit())

	// 初始化组件
	systemMonitor := system.NewMonitor()
	envManager := environment.NewManager()
	projectManager := projects.NewManager()

	// 初始化处理器
	h := handlers.NewHandlers(systemMonitor, envManager, projectManager)

	// 注册路由
	registerRoutes(r, h, config)

	// 创建 HTTP 服务器 - 优化配置
	httpServer := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler:        r,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		IdleTimeout:    config.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB

		// 优化TCP连接设置
		ReadHeaderTimeout: 10 * time.Second, // 防止慢速攻击

		// 启用HTTP/2支持（如果可用）
		// TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	return &Server{
		httpServer: httpServer,
		router:     r,
		config:     config,
	}
}

// registerRoutes 注册所有路由
func registerRoutes(r *router.Router, h *handlers.Handlers, config *Config) {
	// 静态文件服务 - 优化版本，支持压缩和缓存
	var staticHandler http.Handler

	// 尝试使用嵌入的文件系统
	if config != nil && config.StaticFiles != (embed.FS{}) {
		assetsFS, err := fs.Sub(config.StaticFiles, "assets")
		if err != nil {
			log.Printf("警告: 无法创建嵌入文件系统: %v，回退到文件系统", err)
			staticHandler = createOptimizedStaticHandler(http.Dir("./assets"))
		} else {
			staticHandler = createOptimizedStaticHandler(http.FS(assetsFS))
			log.Printf("✅ 使用嵌入式静态文件系统（已优化）")
		}
	} else {
		// 回退到文件系统
		staticHandler = createOptimizedStaticHandler(http.Dir("./assets"))
		log.Printf("⚠️  使用文件系统静态文件 (./assets) - 已优化")
	}

	// 添加静态文件处理器到router的NotFound处理器中
	originalNotFound := r.NotFound
	r.NotFound = func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/static/") {
			// 移除 /static/ 前缀
			req.URL.Path = strings.TrimPrefix(req.URL.Path, "/static")
			staticHandler.ServeHTTP(w, req)
			return
		}
		originalNotFound(w, req)
	}

	// 公开路由
	r.GET("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/dashboard", http.StatusFound)
	})
	r.GET("/login", h.LoginPage)
	r.POST("/login", h.LoginPage)
	r.POST("/api/login", h.APILogin)
	r.POST("/api/set-language", h.SetLanguage)

	// 测试SSE端点（无需认证）
	r.GET("/api/test/sse", h.TestSSEHandler)

	// SSE端点（标准库实现）
	r.GET("/api/sse/stats", h.SSEStatsHandler)

	// 需要认证的路由组
	protected := r.Group("/")
	protected.Use(middleware.Auth(h.GetSessionStore()))
	{
		// Web页面路由
		protected.GET("/dashboard", h.Dashboard)
		protected.GET("/system", h.SystemPage)
		protected.GET("/projects", h.ProjectsPage)
		protected.GET("/environment", h.EnvironmentPage)

		// API路由 - 系统监控
		protected.GET("/api/stats", h.SystemStats)
		protected.GET("/api/stats/:type/details", h.StatsDetails)
		protected.GET("/api/stats/cpu/chart", h.CPUChart)
		protected.GET("/api/system/overview", h.SystemOverview)
		protected.GET("/api/system/details", h.SystemDetails)
		protected.GET("/api/system/processes", h.ProcessList)

		// SSE路由已移到公开路由中（带自定义认证）

		// API路由 - 项目管理
		protected.GET("/api/projects/scan", h.ProjectsScan)
		protected.GET("/api/projects/create-form", h.ProjectCreateForm)
		protected.POST("/api/projects/create", h.ProjectCreate)
		protected.DELETE("/api/projects/:id/delete", h.ProjectDelete)

		// API路由 - 环境管理
		protected.GET("/api/environment/status", h.EnvironmentStatus)
		protected.POST("/api/environment/install", h.InstallEnvironment)
		protected.POST("/api/environment/uninstall", h.UninstallEnvironment)
		protected.POST("/api/environment/upgrade", h.UpgradeEnvironment)
		protected.GET("/api/environment/progress", h.EnvironmentProgress)
	}
}

// Start 启动服务器
func (s *Server) Start() error {
	// 打印启动信息
	s.printStartupInfo()

	log.Printf("🚀 DigWis Panel 启动在 http://%s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// StartWithGracefulShutdown 启动服务器并支持优雅关闭
func (s *Server) StartWithGracefulShutdown() error {
	// 启动服务器
	go func() {
		if err := s.Start(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 正在关闭服务器...")

	// 创建关闭上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭失败: %v", err)
		return err
	}

	log.Println("✅ 服务器已安全关闭")
	return nil
}

// Stop 停止服务器
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// createOptimizedStaticHandler 创建优化的静态文件处理器
func createOptimizedStaticHandler(fileSystem http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置缓存头
		setCacheHeaders(w, r.URL.Path)

		// 检查是否支持gzip压缩
		if shouldCompress(r.URL.Path) && acceptsGzip(r) {
			// 尝试提供gzip压缩版本
			if serveCompressed(w, r, fileSystem) {
				return
			}
		}

		// 提供原始文件
		http.FileServer(fileSystem).ServeHTTP(w, r)
	})
}

// setCacheHeaders 设置缓存头
func setCacheHeaders(w http.ResponseWriter, path string) {
	ext := filepath.Ext(path)

	switch ext {
	case ".css", ".js":
		// CSS和JS文件缓存1小时
		w.Header().Set("Cache-Control", "public, max-age=3600")
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico":
		// 图片文件缓存1天
		w.Header().Set("Cache-Control", "public, max-age=86400")
	case ".woff", ".woff2", ".ttf", ".eot":
		// 字体文件缓存1周
		w.Header().Set("Cache-Control", "public, max-age=604800")
	default:
		// 其他文件缓存1小时
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}

	// 添加ETag支持
	w.Header().Set("ETag", fmt.Sprintf(`"%x"`, time.Now().Unix()))
}

// shouldCompress 判断文件是否应该压缩
func shouldCompress(path string) bool {
	ext := filepath.Ext(path)
	compressibleTypes := map[string]bool{
		".css":  true,
		".js":   true,
		".html": true,
		".json": true,
		".xml":  true,
		".svg":  true,
	}
	return compressibleTypes[ext]
}

// acceptsGzip 检查客户端是否支持gzip
func acceptsGzip(r *http.Request) bool {
	return strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")
}

// serveCompressed 提供gzip压缩的文件
func serveCompressed(w http.ResponseWriter, r *http.Request, fileSystem http.FileSystem) bool {
	file, err := fileSystem.Open(r.URL.Path)
	if err != nil {
		return false
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil || stat.IsDir() {
		return false
	}

	// 设置gzip头
	w.Header().Set("Content-Encoding", "gzip")
	w.Header().Set("Content-Type", getContentType(r.URL.Path))

	// 创建gzip写入器
	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()

	// 复制文件内容到gzip写入器
	_, err = io.Copy(gzipWriter, file)
	return err == nil
}

// getContentType 获取内容类型
func getContentType(path string) string {
	ext := filepath.Ext(path)
	contentTypes := map[string]string{
		".css":  "text/css; charset=utf-8",
		".js":   "application/javascript; charset=utf-8",
		".html": "text/html; charset=utf-8",
		".json": "application/json; charset=utf-8",
		".svg":  "image/svg+xml",
		".png":  "image/png",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".gif":  "image/gif",
		".ico":  "image/x-icon",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}
	return "application/octet-stream"
}

// printStartupInfo 打印启动信息
func (s *Server) printStartupInfo() {
	fmt.Println("🖥️  DigWis Panel (原生 Go 版本)")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Printf("🌐 地址: %s\n", s.httpServer.Addr)
	fmt.Printf("📁 静态文件: 嵌入式 (embed)\n")
	fmt.Printf("🐛 调试模式: %v\n", s.config.Debug)
	fmt.Printf("⏱️  读取超时: %v\n", s.config.ReadTimeout)
	fmt.Printf("⏱️  写入超时: %v\n", s.config.WriteTimeout)
	fmt.Printf("⏱️  空闲超时: %v\n", s.config.IdleTimeout)
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println("📊 可用功能:")
	fmt.Println("  • 系统监控")
	fmt.Println("  • 项目管理")
	fmt.Println("  • 环境管理")
	fmt.Println("  • 实时数据推送")
	fmt.Println(strings.Repeat("=", 40))
	fmt.Println("🔧 使用原生 Go (net/http)")
}

// GetRouter 获取路由器实例
func (s *Server) GetRouter() *router.Router {
	return s.router
}

// GetConfig 获取配置
func (s *Server) GetConfig() *Config {
	return s.config
}
