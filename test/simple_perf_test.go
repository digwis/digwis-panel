package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// 简单的性能测试工具
func main() {
	serverURL := "http://localhost:9090"
	
	log.Printf("🧪 DigWis Panel 页面访问速度测试")
	log.Printf("服务器地址: %s", serverURL)
	log.Printf("模拟低配VPS环境测试")
	log.Printf("========================================")
	
	// 检查服务器状态
	if !checkServer(serverURL) {
		log.Fatalf("❌ 服务器未运行，请先启动 DigWis Panel")
	}
	
	// 测试页面列表
	pages := []struct {
		name string
		url  string
	}{
		{"首页重定向", "/"},
		{"仪表板页面", "/dashboard"},
		{"系统监控页面", "/system"},
		{"项目管理页面", "/projects"},
		{"环境管理页面", "/environment"},
	}
	
	// 测试静态资源
	staticResources := []struct {
		name string
		url  string
	}{
		{"CSS样式文件", "/static/css/output.css"},
		{"HTMX库文件", "/static/js/htmx.min.js"},
		{"Alpine.js库文件", "/static/js/alpine.min.js"},
	}
	
	// 测试API接口
	apiEndpoints := []struct {
		name string
		url  string
	}{
		{"系统统计API", "/api/stats"},
		{"系统概览API", "/api/system/overview"},
		{"系统详情API", "/api/system/details"},
		{"进程列表API", "/api/system/processes"},
	}
	
	log.Printf("\n📄 页面加载测试")
	log.Printf("----------------------------------------")
	testPages(serverURL, pages)
	
	log.Printf("\n🎨 静态资源加载测试")
	log.Printf("----------------------------------------")
	testStaticResources(serverURL, staticResources)
	
	log.Printf("\n🔌 API接口响应测试")
	log.Printf("----------------------------------------")
	testAPIEndpoints(serverURL, apiEndpoints)
	
	log.Printf("\n🔄 连续访问测试（模拟用户操作）")
	log.Printf("----------------------------------------")
	testContinuousAccess(serverURL)
	
	log.Printf("\n📊 测试完成！")
}

// 检查服务器状态
func checkServer(serverURL string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(serverURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound
}

// 测试页面加载
func testPages(serverURL string, pages []struct{ name, url string }) {
	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil // 允许重定向
		},
	}
	
	for _, page := range pages {
		testURL := serverURL + page.url
		
		start := time.Now()
		resp, err := client.Get(testURL)
		duration := time.Since(start)
		
		if err != nil {
			log.Printf("❌ %s: 请求失败 - %v (耗时: %v)", page.name, err, duration)
			continue
		}
		
		// 读取响应体以获取完整加载时间
		bodyStart := time.Now()
		body, err := io.ReadAll(resp.Body)
		bodyDuration := time.Since(bodyStart)
		resp.Body.Close()
		
		totalDuration := time.Since(start)
		
		if err != nil {
			log.Printf("⚠️  %s: 响应读取失败 - %v (连接耗时: %v)", page.name, err, duration)
			continue
		}
		
		status := "✅"
		if totalDuration > 2*time.Second {
			status = "⚠️"
		} else if totalDuration > 5*time.Second {
			status = "❌"
		}
		
		log.Printf("%s %s: %d字节, 状态码:%d, 连接:%v, 读取:%v, 总计:%v", 
			status, page.name, len(body), resp.StatusCode, duration, bodyDuration, totalDuration)
		
		// 页面间稍作停顿
		time.Sleep(500 * time.Millisecond)
	}
}

// 测试静态资源
func testStaticResources(serverURL string, resources []struct{ name, url string }) {
	client := &http.Client{Timeout: 15 * time.Second}
	
	for _, resource := range resources {
		testURL := serverURL + resource.url
		
		start := time.Now()
		resp, err := client.Get(testURL)
		duration := time.Since(start)
		
		if err != nil {
			log.Printf("❌ %s: 请求失败 - %v (耗时: %v)", resource.name, err, duration)
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		totalDuration := time.Since(start)
		
		if err != nil {
			log.Printf("⚠️  %s: 读取失败 - %v (连接耗时: %v)", resource.name, err, duration)
			continue
		}
		
		status := "✅"
		if totalDuration > 1*time.Second {
			status = "⚠️"
		} else if totalDuration > 3*time.Second {
			status = "❌"
		}
		
		// 检查是否启用了压缩
		encoding := resp.Header.Get("Content-Encoding")
		cacheControl := resp.Header.Get("Cache-Control")
		
		log.Printf("%s %s: %d字节, 状态码:%d, 耗时:%v, 压缩:%s, 缓存:%s", 
			status, resource.name, len(body), resp.StatusCode, totalDuration, encoding, cacheControl)
		
		time.Sleep(200 * time.Millisecond)
	}
}

// 测试API接口
func testAPIEndpoints(serverURL string, apis []struct{ name, url string }) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	for _, api := range apis {
		testURL := serverURL + api.url
		
		start := time.Now()
		resp, err := client.Get(testURL)
		duration := time.Since(start)
		
		if err != nil {
			log.Printf("❌ %s: 请求失败 - %v (耗时: %v)", api.name, err, duration)
			continue
		}
		
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		totalDuration := time.Since(start)
		
		if err != nil {
			log.Printf("⚠️  %s: 读取失败 - %v (连接耗时: %v)", api.name, err, duration)
			continue
		}
		
		status := "✅"
		if totalDuration > 1*time.Second {
			status = "⚠️"
		} else if totalDuration > 3*time.Second {
			status = "❌"
		}
		
		// 检查响应头
		responseTime := resp.Header.Get("X-Response-Time")
		contentType := resp.Header.Get("Content-Type")
		
		log.Printf("%s %s: %d字节, 状态码:%d, 耗时:%v, 服务器响应时间:%s, 类型:%s", 
			status, api.name, len(body), resp.StatusCode, totalDuration, responseTime, contentType)
		
		time.Sleep(1 * time.Second)
	}
}

// 连续访问测试
func testContinuousAccess(serverURL string) {
	client := &http.Client{Timeout: 10 * time.Second}
	
	// 模拟用户连续操作：访问仪表板 -> 查看系统监控 -> 查看进程列表
	operations := []struct {
		name string
		url  string
		wait time.Duration
	}{
		{"打开仪表板", "/dashboard", 1 * time.Second},
		{"获取系统统计", "/api/stats", 500 * time.Millisecond},
		{"切换到系统页面", "/system", 1 * time.Second},
		{"获取进程列表", "/api/system/processes", 500 * time.Millisecond},
		{"获取系统详情", "/api/system/details", 500 * time.Millisecond},
		{"返回仪表板", "/dashboard", 1 * time.Second},
	}
	
	log.Printf("模拟用户连续操作...")
	totalStart := time.Now()
	
	for i, op := range operations {
		testURL := serverURL + op.url
		
		start := time.Now()
		resp, err := client.Get(testURL)
		duration := time.Since(start)
		
		if err != nil {
			log.Printf("❌ 操作%d - %s: 失败 - %v (耗时: %v)", i+1, op.name, err, duration)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			totalDuration := time.Since(start)
			
			status := "✅"
			if totalDuration > 2*time.Second {
				status = "⚠️"
			}
			
			log.Printf("%s 操作%d - %s: %d字节, 耗时:%v", 
				status, i+1, op.name, len(body), totalDuration)
		}
		
		// 等待用户思考时间
		time.Sleep(op.wait)
	}
	
	totalDuration := time.Since(totalStart)
	log.Printf("🏁 连续操作总耗时: %v", totalDuration)
}
