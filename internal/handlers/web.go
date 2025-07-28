package handlers

import (
	"html/template"
	"io/fs"
	"net/http"

	"server-panel/internal/templates"
)

// WebHandlers Web页面处理器
type WebHandlers struct {
	*Handlers
	tmpl *templates.Manager
}

// NewWebHandlers 创建Web处理器
func NewWebHandlers(h *Handlers, debug bool, templateFS fs.FS) *WebHandlers {
	return &WebHandlers{
		Handlers: h,
		tmpl:     templates.New(debug, templateFS),
	}
}

// Dashboard 仪表板页面
func (h *WebHandlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	// 如果是HTMX请求，只返回内容部分
	if r.Header.Get("HX-Request") == "true" {
		h.tmpl.RenderPartial(w, "dashboard", nil)
		return
	}

	// 否则返回完整页面
	data := templates.PageData{
		Title:     "系统概览",
		PageTitle: "仪表板",
	}

	h.tmpl.Render(w, "dashboard", data)
}

// SystemPage 系统监控页面
func (h *WebHandlers) SystemPage(w http.ResponseWriter, r *http.Request) {
	systemContent := `
	<div class="space-y-6">
		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">系统信息</h3>
			<div id="system-info" hx-get="/api/system/details" hx-trigger="load" hx-swap="innerHTML">
				<div class="animate-pulse">
					<div class="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
					<div class="h-4 bg-gray-200 rounded w-1/2 mb-2"></div>
					<div class="h-4 bg-gray-200 rounded w-2/3"></div>
				</div>
			</div>
		</div>

		<div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">进程列表</h3>
			<div id="process-list" hx-get="/api/system/processes" hx-trigger="load" hx-swap="innerHTML">
				<div class="animate-pulse">
					<div class="h-4 bg-gray-200 rounded w-full mb-2"></div>
					<div class="h-4 bg-gray-200 rounded w-full mb-2"></div>
					<div class="h-4 bg-gray-200 rounded w-full"></div>
				</div>
			</div>
		</div>
	</div>`

	// 如果是HTMX请求，只返回内容部分
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(systemContent))
		return
	}

	// 否则返回完整页面
	data := templates.PageData{
		Title:     "系统监控",
		PageTitle: "系统监控",
		Content:   template.HTML(systemContent),
	}

	h.tmpl.Render(w, "system", data)
}

// ProjectsPage 项目管理页面
func (h *WebHandlers) ProjectsPage(w http.ResponseWriter, r *http.Request) {
	// 如果是HTMX请求，只返回内容部分
	if r.Header.Get("HX-Request") == "true" {
		h.tmpl.RenderPartial(w, "projects", nil)
		return
	}

	// 否则返回完整页面
	data := templates.PageData{
		Title:     "项目管理",
		PageTitle: "项目管理",
	}

	h.tmpl.Render(w, "projects", data)
}

// EnvironmentPage 环境配置页面
func (h *WebHandlers) EnvironmentPage(w http.ResponseWriter, r *http.Request) {
	// 如果是HTMX请求，只返回内容部分
	if r.Header.Get("HX-Request") == "true" {
		h.tmpl.RenderPartial(w, "environment", nil)
		return
	}

	// 否则返回完整页面
	data := templates.PageData{
		Title:     "环境配置",
		PageTitle: "环境配置",
	}

	h.tmpl.Render(w, "environment", data)
}

// LoginPage 登录页面
func (h *WebHandlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// 处理登录逻辑
		h.handleLogin(w, r)
		return
	}
	
	// 简单的登录页面
	loginHTML := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>登录 - DigWis 面板</title>
    <link rel="stylesheet" href="/static/css/output.css">
</head>
<body class="bg-gray-50 flex items-center justify-center min-h-screen">
    <div class="max-w-md w-full space-y-8">
        <div>
            <div class="mx-auto h-12 w-12 flex items-center justify-center rounded-full bg-blue-100">
                <span class="text-2xl">🖥️</span>
            </div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                登录到 DigWis 面板
            </h2>
        </div>
        <form class="mt-8 space-y-6" method="POST">
            <div class="rounded-md shadow-sm -space-y-px">
                <div>
                    <input type="text" name="username" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm" 
                           placeholder="用户名">
                </div>
                <div>
                    <input type="password" name="password" required 
                           class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm" 
                           placeholder="密码">
                </div>
            </div>
            <div>
                <button type="submit" 
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    登录
                </button>
            </div>
        </form>
    </div>
</body>
</html>`
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(loginHTML))
}
