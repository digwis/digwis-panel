package templates

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strings"
	"time"
)

// Manager 模板管理器
type Manager struct {
	templates map[string]*template.Template
	debug     bool
	fs        fs.FS
}

// PageData 页面数据结构
type PageData struct {
	Title      string
	PageTitle  string
	Content    template.HTML
	LastUpdate string
	User       interface{}
	Data       interface{}
}

// New 创建模板管理器
func New(debug bool, templateFS fs.FS) *Manager {
	m := &Manager{
		templates: make(map[string]*template.Template),
		debug:     debug,
		fs:        templateFS,
	}

	// 加载模板
	m.loadTemplates()

	return m
}

// loadTemplates 加载所有模板
func (m *Manager) loadTemplates() {
	// 创建模板函数映射
	funcMap := template.FuncMap{
		"dict": func(values ...interface{}) map[string]interface{} {
			dict := make(map[string]interface{})
			for i := 0; i < len(values); i += 2 {
				if i+1 < len(values) {
					dict[values[i].(string)] = values[i+1]
				}
			}
			return dict
		},
		"slice": func(values ...interface{}) []interface{} {
			return values
		},
	}

	// 布局模板
	layoutPath := "web/templates/layout.html"

	// 加载所有组件模板
	componentPaths := []string{
		"web/templates/components/stats-cards.html",
		"web/templates/components/loading.html",
		"web/templates/components/cpu-details.html",
		"web/templates/components/memory-details.html",
	}

	// 页面模板列表
	pages := map[string]string{
		"dashboard":   "web/templates/pages/dashboard.html",
		"projects":    "web/templates/pages/projects.html",
		"environment": "web/templates/pages/environment.html",
		"login":       "web/templates/login.html", // 保持原有的登录页面
	}

	for pageName, pagePath := range pages {
		// 检查文件是否存在
		if _, err := os.Stat(layoutPath); err != nil {
			continue
		}
		if _, err := os.Stat(pagePath); err != nil {
			continue
		}

		// 创建新模板并设置函数映射
		tmpl := template.New(pageName).Funcs(funcMap)

		// 解析布局模板
		tmpl, err := tmpl.ParseFiles(layoutPath)
		if err != nil {
			continue
		}

		// 解析所有组件模板
		for _, componentPath := range componentPaths {
			if _, err := os.Stat(componentPath); err == nil {
				tmpl, err = tmpl.ParseFiles(componentPath)
				if err != nil {
					if m.debug {
						fmt.Printf("Error parsing component template %s: %v\n", componentPath, err)
					}
					continue
				}
			}
		}

		// 解析页面模板
		tmpl, err = tmpl.ParseFiles(pagePath)
		if err != nil {
			if m.debug {
				fmt.Printf("Error parsing page template %s: %v\n", pagePath, err)
			}
			continue
		}

		m.templates[pageName] = tmpl
		if m.debug {
			fmt.Printf("Loaded template: %s\n", pageName)
		}
	}
}

// Render 渲染页面
func (m *Manager) Render(w http.ResponseWriter, page string, data PageData) error {
	// 设置默认值
	if data.LastUpdate == "" {
		data.LastUpdate = time.Now().Format("15:04:05")
	}

	// 调试模式下每次重新加载模板
	if m.debug {
		m.loadTemplates()
	}

	// 获取模板
	tmpl, exists := m.templates[page]
	if !exists {
		// 如果模板不存在，尝试重新加载一次
		m.loadTemplates()
		tmpl, exists = m.templates[page]
		if !exists {
			http.Error(w, fmt.Sprintf("模板 %s 未找到", page), http.StatusNotFound)
			return fmt.Errorf("template %s not found", page)
		}
	}

	// 如果Content为空，渲染页面模板内容
	if data.Content == "" {
		var buf strings.Builder
		err := tmpl.ExecuteTemplate(&buf, page, data)
		if err != nil {
			http.Error(w, fmt.Sprintf("渲染页面模板失败: %v", err), http.StatusInternalServerError)
			return err
		}
		data.Content = template.HTML(buf.String())
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 渲染layout模板
	return tmpl.ExecuteTemplate(w, "layout.html", data)
}

// RenderPartial 渲染部分模板（用于HTMX请求）
func (m *Manager) RenderPartial(w http.ResponseWriter, page string, data interface{}) error {
	// 如果是调试模式，重新加载模板
	if m.debug {
		m.loadTemplates()
	}

	tmpl, exists := m.templates[page]
	if !exists {
		// 如果模板不存在，尝试重新加载一次
		m.loadTemplates()
		tmpl, exists = m.templates[page]
		if !exists {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(fmt.Sprintf(`<div class="text-center py-8">
				<h3 class="text-lg font-semibold text-gray-900">模板 %s 未找到</h3>
				<p class="text-sm text-gray-600">请检查模板文件是否存在</p>
			</div>`, page)))
			return nil
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, page, data)
}

// RenderComponent 渲染组件模板
func (m *Manager) RenderComponent(w http.ResponseWriter, templateName, componentName string, data interface{}) error {
	// 如果是调试模式，重新加载模板
	if m.debug {
		m.loadTemplates()
	}

	tmpl, exists := m.templates[templateName]
	if !exists {
		// 如果模板不存在，尝试重新加载一次
		m.loadTemplates()
		tmpl, exists = m.templates[templateName]
		if !exists {
			return fmt.Errorf("template %s not found", templateName)
		}
	}

	// 设置响应头并渲染组件
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, componentName, data)
}

// RenderComponentFiber 渲染组件模板 (Fiber版本) - 已禁用
// func (m *Manager) RenderComponentFiber(c *fiber.Ctx, templateName, componentName string, data interface{}) error {
// 	// 如果是调试模式，重新加载模板
// 	if m.debug {
// 		m.loadTemplates()
// 	}

// 	tmpl, exists := m.templates[templateName]
// 	if !exists {
// 		// 如果模板不存在，尝试重新加载一次
// 		m.loadTemplates()
// 		tmpl, exists = m.templates[templateName]
// 		if !exists {
// 			return fmt.Errorf("template %s not found", templateName)
// 		}
// 	}

// 	// 设置响应头并渲染组件
// 	c.Set("Content-Type", "text/html; charset=utf-8")

// 	// 使用字符串缓冲区渲染模板
// 	var buf strings.Builder
// 	if err := tmpl.ExecuteTemplate(&buf, componentName, data); err != nil {
// 		return err
// 	}

// 	return c.SendString(buf.String())
// }

// RenderLayout 渲染布局模板
func (m *Manager) RenderLayout(w http.ResponseWriter, data PageData) error {
	// 设置默认值
	if data.LastUpdate == "" {
		data.LastUpdate = time.Now().Format("15:04:05")
	}

	// 加载布局模板
	layoutPath := "web/templates/layout.html"
	tmpl, err := template.ParseFiles(layoutPath)
	if err != nil {
		http.Error(w, "布局模板加载失败", http.StatusInternalServerError)
		return err
	}

	// 设置响应头
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// 渲染模板
	return tmpl.Execute(w, data)
}

// GetFS 获取文件系统
func (m *Manager) GetFS() fs.FS {
	return m.fs
}

// RenderJSON 渲染JSON响应
func (m *Manager) RenderJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// APIResponse API响应结构
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RenderAPIResponse 渲染API响应
func (m *Manager) RenderAPIResponse(w http.ResponseWriter, success bool, message string, data interface{}, err error) {
	response := APIResponse{
		Success: success,
		Message: message,
		Data:    data,
	}
	
	if err != nil {
		response.Error = err.Error()
	}
	
	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}
	
	m.RenderJSON(w, response)
}
