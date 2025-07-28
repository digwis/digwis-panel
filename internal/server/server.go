package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
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

// DefaultConfig 默认配置
func DefaultConfig() *Config {
	return &Config{
		Host:         "127.0.0.1",
		Port:         "8080",
		Debug:        false,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
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
	r.Use(middleware.CORS())

	// 初始化组件
	systemMonitor := system.NewMonitor()
	envManager := environment.NewManager()
	projectManager := projects.NewManager()

	// 初始化处理器
	h := handlers.NewHandlers(systemMonitor, envManager, projectManager)

	// 注册路由
	registerRoutes(r, h, config)

	// 创建 HTTP 服务器
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", config.Host, config.Port),
		Handler:      r,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		IdleTimeout:  config.IdleTimeout,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	return &Server{
		httpServer: httpServer,
		router:     r,
		config:     config,
	}
}

// registerRoutes 注册所有路由
func registerRoutes(r *router.Router, h *handlers.Handlers, config *Config) {
	// 静态文件服务 - 优先使用嵌入的文件系统
	var staticHandler http.Handler

	// 尝试使用嵌入的文件系统
	if config != nil && config.StaticFiles != (embed.FS{}) {
		assetsFS, err := fs.Sub(config.StaticFiles, "assets")
		if err != nil {
			log.Printf("警告: 无法创建嵌入文件系统: %v，回退到文件系统", err)
			staticHandler = http.StripPrefix("/static/", http.FileServer(http.Dir("./assets")))
		} else {
			staticHandler = http.StripPrefix("/static/", http.FileServer(http.FS(assetsFS)))
			log.Printf("✅ 使用嵌入式静态文件系统")
		}
	} else {
		// 回退到文件系统
		staticHandler = http.StripPrefix("/static/", http.FileServer(http.Dir("./assets")))
		log.Printf("⚠️  使用文件系统静态文件 (./assets)")
	}

	// 添加静态文件处理器到router的NotFound处理器中
	originalNotFound := r.NotFound
	r.NotFound = func(w http.ResponseWriter, req *http.Request) {
		if strings.HasPrefix(req.URL.Path, "/static/") {
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
