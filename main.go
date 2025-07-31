package main

import (
	"embed"
	"flag"
	"log"

	"server-panel/internal/server"
)

//go:embed assets/*
var staticFiles embed.FS

func main() {
	// 命令行参数
	var (
		host  = flag.String("host", "0.0.0.0", "服务器监听地址")
		port  = flag.String("port", "8080", "服务器端口")
		debug = flag.Bool("debug", false, "启用调试模式")
	)
	flag.Parse()

	// 创建服务器配置
	config := &server.Config{
		Host:        *host,
		Port:        *port,
		Debug:       *debug,
		StaticFiles: staticFiles,
	}

	// 创建服务器
	srv := server.New(config)

	// 启动服务器
	log.Printf("🖥️  DigWis Panel 启动中... (优化版本)")
	if err := srv.StartWithGracefulShutdown(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
