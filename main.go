package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"server-panel/internal/auth"
	"server-panel/internal/config"
	"server-panel/internal/handlers"
	"server-panel/internal/middleware"
	"server-panel/internal/projects"
	"server-panel/internal/system"

	"github.com/gorilla/mux"
)

func main() {
	// 命令行参数
	var (
		port     = flag.String("port", "8443", "服务器端口")
		configPath = flag.String("config", "/etc/server-panel/config.yaml", "配置文件路径")
		debug    = flag.Bool("debug", false, "调试模式")
	)
	flag.Parse()

	// 检查是否以root权限运行
	if os.Geteuid() != 0 {
		log.Fatal("此程序需要root权限运行")
	}

	// 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Printf("加载配置文件失败，使用默认配置: %v", err)
		cfg = config.Default()
	}

	if *debug {
		cfg.Debug = true
	}

	// 初始化系统监控
	systemMonitor := system.NewMonitor()

	// 初始化项目管理
	projectManager := projects.NewManager()

	// 初始化认证系统
	authManager := auth.NewManager(cfg.Auth)

	// 创建路由器
	router := mux.NewRouter()

	// 中间件
	router.Use(middleware.Logging)
	router.Use(middleware.Recovery)
	router.Use(middleware.CORS)

	// 静态文件服务
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./web/static/"))))

	// 创建处理器
	h := handlers.New(authManager, systemMonitor, projectManager, cfg)

	// 路由设置
	setupRoutes(router, h)

	// 启动服务器
	addr := ":" + *port

	// 检查是否有SSL证书
	certFile := "/etc/server-panel/server.crt"
	keyFile := "/etc/server-panel/server.key"

	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			fmt.Printf("🚀 服务器面板启动在 https://localhost%s\n", addr)
			fmt.Println("🔒 使用HTTPS安全连接")
		} else {
			fmt.Printf("🚀 服务器面板启动在 http://localhost%s\n", addr)
			fmt.Println("⚠️  使用HTTP连接（建议配置HTTPS）")
		}
	} else {
		fmt.Printf("🚀 服务器面板启动在 http://localhost%s\n", addr)
		fmt.Println("⚠️  使用HTTP连接（建议配置HTTPS）")
	}

	fmt.Println("📝 使用系统用户账户登录（需要管理员权限）")

	if cfg.Debug {
		fmt.Println("🐛 调试模式已启用")
	}

	// 优雅关闭
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\n🛑 正在关闭服务器...")
		os.Exit(0)
	}()

	// 尝试HTTPS，如果证书不存在则使用HTTP
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			log.Fatal(http.ListenAndServeTLS(addr, certFile, keyFile, router))
		}
	}

	log.Fatal(http.ListenAndServe(addr, router))
}

func setupRoutes(router *mux.Router, h *handlers.Handlers) {
	// 公开路由
	router.HandleFunc("/", h.Home).Methods("GET")
	router.HandleFunc("/login", h.Login).Methods("GET", "POST")
	router.HandleFunc("/api/login", h.APILogin).Methods("POST")

	// 需要认证的路由
	auth := router.PathPrefix("/").Subrouter()
	auth.Use(h.RequireAuth)

	// 仪表板
	auth.HandleFunc("/dashboard", h.Dashboard).Methods("GET")
	
	// 环境管理
	auth.HandleFunc("/environment", h.Environment).Methods("GET")
	auth.HandleFunc("/api/environment/install", h.InstallEnvironment).Methods("POST")
	auth.HandleFunc("/api/environment/status", h.EnvironmentStatus).Methods("GET")
	auth.HandleFunc("/api/environment/uninstall", h.UninstallEnvironment).Methods("POST")
	auth.HandleFunc("/api/environment/upgrade", h.UpgradeEnvironment).Methods("POST")

	// PHP插件管理
	auth.HandleFunc("/api/php/extensions", h.PHPExtensions).Methods("GET")
	auth.HandleFunc("/api/php/extensions/enable", h.EnablePHPExtension).Methods("POST")
	auth.HandleFunc("/api/php/extensions/disable", h.DisablePHPExtension).Methods("POST")

	// 项目管理
	auth.HandleFunc("/projects", h.Projects).Methods("GET")
	auth.HandleFunc("/api/projects/scan", h.ScanProjects).Methods("GET")
	auth.HandleFunc("/api/projects/databases", h.ProjectDatabases).Methods("GET")
	auth.HandleFunc("/api/projects/create", h.CreateProject).Methods("POST")

	// 系统监控
	auth.HandleFunc("/system", h.SystemInfo).Methods("GET")
	auth.HandleFunc("/api/system/stats", h.SystemStats).Methods("GET")
	auth.HandleFunc("/api/system/details", h.SystemDetails).Methods("GET")
	auth.HandleFunc("/api/system/processes", h.ProcessList).Methods("GET")
	auth.HandleFunc("/api/system/services", h.ServiceList).Methods("GET")

	// 文件管理
	auth.HandleFunc("/files", h.FileManager).Methods("GET")
	auth.HandleFunc("/api/files/list", h.FileList).Methods("GET")
	auth.HandleFunc("/api/files/upload", h.FileUpload).Methods("POST")
	auth.HandleFunc("/api/files/delete", h.FileDelete).Methods("DELETE")

	// 日志查看
	auth.HandleFunc("/logs", h.LogViewer).Methods("GET")
	auth.HandleFunc("/api/logs/read", h.ReadLogs).Methods("GET")

	// 设置
	auth.HandleFunc("/settings", h.Settings).Methods("GET", "POST")

	// SSL证书管理
	auth.HandleFunc("/api/ssl/status", h.SSLStatus).Methods("GET")
	auth.HandleFunc("/api/ssl/self-signed", h.GenerateSelfSigned).Methods("POST")
	auth.HandleFunc("/api/ssl/letsencrypt", h.RequestLetsEncrypt).Methods("POST")
	auth.HandleFunc("/api/ssl/renew", h.RenewCertificate).Methods("POST")
	auth.HandleFunc("/api/ssl/delete", h.DeleteCertificate).Methods("DELETE")

	// 退出登录
	auth.HandleFunc("/logout", h.Logout).Methods("POST")
}
