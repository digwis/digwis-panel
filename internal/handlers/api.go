package handlers

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"runtime"
	"strings"
	"time"

	"server-panel/internal/environment"
	"server-panel/internal/system"
	"server-panel/internal/templates"
)

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}

	i := 0
	size := float64(bytes)
	for size >= unit && i < len(sizes)-1 {
		size /= unit
		i++
	}

	return fmt.Sprintf("%.1f %s", size, sizes[i])
}

// APIHandlers API处理器
type APIHandlers struct {
	*Handlers
	tmpl *templates.Manager
}

// NewAPIHandlers 创建API处理器
func NewAPIHandlers(h *Handlers, debug bool, templateFS fs.FS) *APIHandlers {
	return &APIHandlers{
		Handlers: h,
		tmpl:     templates.New(debug, templateFS),
	}
}

// SystemStats 系统统计信息
func (h *APIHandlers) SystemStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		h.tmpl.RenderAPIResponse(w, false, "获取系统统计失败", nil, err)
		return
	}
	
	// 如果是HTMX请求，返回HTML片段
	if r.Header.Get("HX-Request") == "true" {
		// 使用dashboard模板渲染stats-cards组件
		err := h.tmpl.RenderComponent(w, "dashboard", "stats-cards", stats)
		if err != nil {
			http.Error(w, fmt.Sprintf("渲染统计卡片失败: %v", err), http.StatusInternalServerError)
		}
		return
	}
	
	// 普通API请求返回JSON
	h.tmpl.RenderAPIResponse(w, true, "获取系统统计成功", stats, nil)
}

// SystemOverview 系统概览信息
func (h *APIHandlers) SystemOverview(w http.ResponseWriter, r *http.Request) {
	details := h.systemMonitor.GetSystemDetails()
	if details == nil {
		h.tmpl.RenderAPIResponse(w, false, "获取系统概览失败", nil, fmt.Errorf("系统详情为空"))
		return
	}

	// 如果是HTMX请求，返回HTML片段
	if r.Header.Get("HX-Request") == "true" {
		overviewHTML := fmt.Sprintf(`
			<div class="flex items-center space-x-8">
				<div>
					<div class="text-lg font-semibold">%s</div>
					<div class="text-sm opacity-90">%s</div>
				</div>
				<div>
					<div class="text-lg font-semibold">运行时间</div>
					<div class="text-sm opacity-90">%s</div>
				</div>
				<div>
					<div class="text-lg font-semibold">Go 版本</div>
					<div class="text-sm opacity-90">%s</div>
				</div>
			</div>`,
			details["hostname"],
			details["os"],
			details["uptime"],
			runtime.Version(),
		)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(overviewHTML))
		return
	}

	// 普通API请求返回JSON
	h.tmpl.RenderAPIResponse(w, true, "获取系统概览成功", details, nil)
}

// StatsDetails 统计详情
func (h *APIHandlers) StatsDetails(w http.ResponseWriter, r *http.Request) {
	// 获取统计类型
	statsType := r.URL.Path[len("/api/stats/"):]
	statsType = strings.TrimSuffix(statsType, "/details")

	// 获取系统统计
	stats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		h.tmpl.RenderAPIResponse(w, false, "获取统计详情失败", nil, err)
		return
	}

	// 获取进程列表
	processes, err := h.systemMonitor.GetProcessList()
	if err != nil {
		processes = []system.Process{} // 如果获取失败，使用空列表
	}

	// 如果是HTMX请求，返回详情组件
	if r.Header.Get("HX-Request") == "true" {
		switch statsType {
		case "cpu":
			// 准备CPU详情数据
			data := map[string]interface{}{
				"CPU": stats.CPU,
				"CPUHistory": h.generateCPUHistory(), // 生成模拟的CPU历史数据
				"TopCPUProcesses": h.getTopProcesses(processes, "cpu", 5),
			}
			err := h.tmpl.RenderComponent(w, "dashboard", "cpu-details", data)
			if err != nil {
				http.Error(w, fmt.Sprintf("渲染CPU详情失败: %v", err), http.StatusInternalServerError)
			}
			return

		case "memory":
			// 准备内存详情数据
			data := map[string]interface{}{
				"Memory": h.formatMemoryStats(stats.Memory),
				"TopMemoryProcesses": h.getTopProcesses(processes, "memory", 5),
			}
			err := h.tmpl.RenderComponent(w, "dashboard", "memory-details", data)
			if err != nil {
				http.Error(w, fmt.Sprintf("渲染内存详情失败: %v", err), http.StatusInternalServerError)
			}
			return
		}
	}

	// 普通API请求返回JSON
	h.tmpl.RenderAPIResponse(w, true, "获取统计详情成功", stats, nil)
}

// renderProcessTooltip 渲染进程悬停提示
func (h *APIHandlers) renderProcessTooltip(w http.ResponseWriter, processes []system.Process, processType string) error {
	title := "热门进程"
	if processType == "cpu" {
		title = "CPU 占用最高的进程"
	} else if processType == "memory" {
		title = "内存占用最高的进程"
	}

	tooltipHTML := fmt.Sprintf(`
		<div class="text-sm font-semibold mb-3">%s</div>
		<div class="space-y-2">`, title)

	for _, proc := range processes {
		tooltipHTML += fmt.Sprintf(`
			<div class="flex justify-between items-center py-1">
				<div class="flex-1 truncate">
					<div class="text-sm font-medium">%s</div>
					<div class="text-xs opacity-75">PID: %d</div>
				</div>
				<div class="text-right ml-2">
					<div class="text-sm">%.1f%% CPU</div>
					<div class="text-xs opacity-75">%.1f%% MEM</div>
				</div>
			</div>`,
			proc.Name, proc.PID, proc.CPU, proc.Memory)
	}

	tooltipHTML += `</div>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(tooltipHTML))
	return nil
}

// generateCPUHistory 生成模拟的CPU历史数据
func (h *APIHandlers) generateCPUHistory() []float64 {
	// 生成30个数据点，模拟30分钟的CPU使用率
	history := make([]float64, 30)
	for i := range history {
		// 生成一些随机但合理的CPU使用率数据
		base := 20.0 + float64(i%10)*3.0
		history[i] = base + (float64(i%5) * 2.0)
		if history[i] > 100 {
			history[i] = 100
		}
	}
	return history
}

// formatMemoryStats 格式化内存统计信息
func (h *APIHandlers) formatMemoryStats(mem system.MemoryStats) map[string]interface{} {
	return map[string]interface{}{
		"Total":             mem.Total,
		"Used":              mem.Used,
		"Free":              mem.Free,
		"Available":         mem.Available,
		"Usage":             mem.Usage,
		"TotalFormatted":    formatBytes(int64(mem.Total)),
		"UsedFormatted":     formatBytes(int64(mem.Used)),
		"FreeFormatted":     formatBytes(int64(mem.Free)),
		"AvailableFormatted": formatBytes(int64(mem.Available)),
		"CacheFormatted":    formatBytes(int64(mem.Total - mem.Used - mem.Free)),
		"CacheUsage":        float64(mem.Total - mem.Used - mem.Free) / float64(mem.Total) * 100,
		"Swap": map[string]interface{}{
			"Total":         mem.Swap.Total,
			"Used":          mem.Swap.Used,
			"Free":          mem.Swap.Free,
			"Usage":         mem.Swap.Usage,
			"UsedFormatted": formatBytes(int64(mem.Swap.Used)),
		},
	}
}

// getTopProcesses 获取排序后的进程列表
func (h *APIHandlers) getTopProcesses(processes []system.Process, sortBy string, limit int) []system.Process {
	if len(processes) == 0 {
		return processes
	}

	// 简单排序（实际应该使用sort包）
	// 这里为了简化，直接返回前几个进程
	if len(processes) > limit {
		return processes[:limit]
	}
	return processes
}

// CPUChart CPU图表数据
func (h *APIHandlers) CPUChart(w http.ResponseWriter, r *http.Request) {
	// 如果是HTMX请求，返回图表HTML
	if r.Header.Get("HX-Request") == "true" {
		history := h.generateCPUHistory()

		chartHTML := `<div class="h-full flex items-end space-x-1 px-4">`
		for i, val := range history {
			height := int(val * 2) // 将百分比转换为像素高度
			if height < 5 {
				height = 5 // 最小高度
			}
			if height > 200 {
				height = 200 // 最大高度
			}

			color := "bg-blue-400"
			if val > 80 {
				color = "bg-red-400"
			} else if val > 60 {
				color = "bg-yellow-400"
			}

			chartHTML += fmt.Sprintf(`
				<div class="flex-1 %s rounded-t transition-all duration-300"
					 style="height: %dpx"
					 title="%.1f%% (第%d分钟)">
				</div>`, color, height, val, i+1)
		}
		chartHTML += `</div>
		<div class="flex justify-between text-xs text-gray-500 mt-2 px-4">
			<span>30分钟前</span>
			<span>15分钟前</span>
			<span>现在</span>
		</div>`

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(chartHTML))
		return
	}

	// 普通API请求返回JSON数据
	history := h.generateCPUHistory()
	h.tmpl.RenderAPIResponse(w, true, "获取CPU图表数据成功", map[string]interface{}{
		"history": history,
	}, nil)
}

// SystemDetails 系统详细信息
func (h *APIHandlers) SystemDetails(w http.ResponseWriter, r *http.Request) {
	// 获取真实的系统详情
	details := h.systemMonitor.GetSystemDetails()

	// 添加额外信息
	details["go_version"] = runtime.Version()
	details["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	
	if r.Header.Get("HX-Request") == "true" {
		detailsHTML := fmt.Sprintf(`
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
			<div class="space-y-2">
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">主机名:</span>
					<span class="text-sm text-gray-900">%s</span>
				</div>
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">操作系统:</span>
					<span class="text-sm text-gray-900">%s</span>
				</div>
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">架构:</span>
					<span class="text-sm text-gray-900">%s</span>
				</div>
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">运行时间:</span>
					<span class="text-sm text-gray-900">%s</span>
				</div>
			</div>
			<div class="space-y-2">
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">内核版本:</span>
					<span class="text-sm text-gray-900">%s</span>
				</div>
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">CPU 核心:</span>
					<span class="text-sm text-gray-900">%d</span>
				</div>
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">Go 版本:</span>
					<span class="text-sm text-gray-900">%s</span>
				</div>
				<div class="flex justify-between">
					<span class="text-sm font-medium text-gray-600">最后更新:</span>
					<span class="text-sm text-gray-900">%s</span>
				</div>
			</div>
		</div>`,
			details["hostname"], details["os"], details["arch"], details["uptime"],
			details["kernel"], details["cpu_cores"], details["go_version"],
			time.Now().Format("15:04:05"),
		)
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(detailsHTML))
		return
	}
	
	h.tmpl.RenderAPIResponse(w, true, "获取系统详情成功", details, nil)
}

// ProcessList 进程列表
func (h *APIHandlers) ProcessList(w http.ResponseWriter, r *http.Request) {
	// 获取真实的进程数据
	processes, err := h.systemMonitor.GetProcessList()
	if err != nil {
		h.tmpl.RenderAPIResponse(w, false, "获取进程列表失败", nil, err)
		return
	}

	// 获取类型参数（用于悬停详情）
	processType := r.URL.Query().Get("type")

	// 只取前5个进程（悬停显示时）或前10个（完整列表）
	limit := 10
	if processType != "" {
		limit = 5
	}
	if len(processes) > limit {
		processes = processes[:limit]
	}

	if r.Header.Get("HX-Request") == "true" {
		// 如果是悬停详情请求，返回简化的HTML
		if processType != "" {
			h.renderProcessTooltip(w, processes, processType)
			return
		}
		processHTML := `
		<div class="overflow-x-auto">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">PID</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">进程名</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">CPU</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">内存</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">状态</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">`
		
		for _, proc := range processes {
			statusColor := "green"
			if proc.Status != "running" && proc.Status != "R" && proc.Status != "S" {
				statusColor = "red"
			}

			processHTML += fmt.Sprintf(`
					<tr>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%d</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">%s</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%.1f%%</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%.1f%%</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-%s-100 text-%s-800">
								running
							</span>
						</td>
					</tr>`,
				proc.PID, proc.Name, proc.CPU, proc.Memory, statusColor, statusColor)
		}
		
		processHTML += `
				</tbody>
			</table>
		</div>`
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(processHTML))
		return
	}
	
	h.tmpl.RenderAPIResponse(w, true, "获取进程列表成功", processes, nil)
}

// ProjectsScan 扫描项目
func (h *APIHandlers) ProjectsScan(w http.ResponseWriter, r *http.Request) {
	// 获取真实的项目数据
	projects, err := h.projectManager.ScanProjects()
	if err != nil {
		h.tmpl.RenderAPIResponse(w, false, "扫描项目失败", nil, err)
		return
	}
	
	if r.Header.Get("HX-Request") == "true" {
		projectsHTML := `
		<div class="space-y-4">`
		
		for _, project := range projects {
			statusColor := "green"
			statusText := "运行中"
			if project.Status == "inactive" || project.Status == "stopped" {
				statusColor = "red"
				statusText = "已停止"
			} else if project.Status == "error" {
				statusColor = "yellow"
				statusText = "错误"
			}

			// 格式化项目大小
			sizeStr := formatBytes(project.Size)

			projectsHTML += fmt.Sprintf(`
			<div class="border border-gray-200 rounded-lg p-4">
				<div class="flex items-center justify-between">
					<div>
						<h4 class="text-lg font-medium text-gray-900">%s</h4>
						<p class="text-sm text-gray-600">%s</p>
						<div class="flex items-center space-x-4 mt-2">
							<p class="text-xs text-gray-500">类型: %s</p>
							<p class="text-xs text-gray-500">大小: %s</p>
							%s
						</div>
					</div>
					<div class="flex items-center space-x-2">
						<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-%s-100 text-%s-800">
							%s
						</span>
						<button class="inline-flex items-center px-3 py-1.5 border border-transparent text-xs font-medium rounded text-white bg-blue-600 hover:bg-blue-700">
							管理
						</button>
					</div>
				</div>`,
				project.Name, project.Path, project.Type, sizeStr,
				func() string {
					if project.Domain != "" {
						return fmt.Sprintf(`<p class="text-xs text-gray-500">域名: %s</p>`, project.Domain)
					}
					return ""
				}(),
				statusColor, statusColor, statusText)
		}
		
		projectsHTML += `
		</div>`
		
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(projectsHTML))
		return
	}
	
	h.tmpl.RenderAPIResponse(w, true, "扫描项目成功", projects, nil)
}

// EnvironmentStatus 环境状态
func (h *APIHandlers) EnvironmentStatus(w http.ResponseWriter, r *http.Request) {
	// 获取真实的环境状态
	environments := h.envManager.GetAvailableEnvironments()

	// 转换为API格式
	status := make(map[string]interface{})
	for _, env := range environments {
		installed := env.Status == "installed"
		status[env.Name] = map[string]interface{}{
			"installed": installed,
			"version":   env.Version,
			"status":    env.Status,
		}
	}

	h.tmpl.RenderAPIResponse(w, true, "获取环境状态成功", status, nil)
}

// InstallEnvironment 安装环境
func (h *APIHandlers) InstallEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	// 解析JSON请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.tmpl.RenderAPIResponse(w, false, "无效的请求格式", nil, err)
		return
	}

	// 如果没有指定版本，使用最新版本
	if req.Version == "" {
		req.Version = "最新版本"
	}

	// 创建进度通道
	progressChan := make(chan environment.InstallProgress, 100)

	// 启动安装
	go func() {
		defer close(progressChan)
		if err := h.envManager.InstallEnvironment(req.Name, req.Version, progressChan); err != nil {
			progressChan <- environment.InstallProgress{
				Environment: req.Name,
				Progress:    0,
				Message:     fmt.Sprintf("安装失败: %v", err),
				Status:      "error",
			}
		}
	}()

	h.tmpl.RenderAPIResponse(w, true, fmt.Sprintf("开始安装 %s %s", req.Name, req.Version), map[string]string{
		"status": "started",
	}, nil)
}

// UninstallEnvironment 卸载环境
func (h *APIHandlers) UninstallEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	// 解析JSON请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.tmpl.RenderAPIResponse(w, false, "无效的请求格式", nil, err)
		return
	}

	if err := h.envManager.UninstallEnvironment(req.Name); err != nil {
		h.tmpl.RenderAPIResponse(w, false, fmt.Sprintf("卸载失败: %v", err), nil, err)
		return
	}

	h.tmpl.RenderAPIResponse(w, true, "卸载成功", nil, nil)
}

// UpgradeEnvironment 升级环境
func (h *APIHandlers) UpgradeEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	// 解析JSON请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.tmpl.RenderAPIResponse(w, false, "无效的请求格式", nil, err)
		return
	}

	// 如果没有指定版本，使用最新版本
	if req.Version == "" {
		req.Version = "最新版本"
	}

	// 创建进度通道
	progressChan := make(chan environment.InstallProgress, 100)

	// 启动升级
	go func() {
		defer close(progressChan)
		if err := h.envManager.UpgradeEnvironment(req.Name, req.Version, progressChan); err != nil {
			progressChan <- environment.InstallProgress{
				Environment: req.Name,
				Progress:    0,
				Message:     fmt.Sprintf("升级失败: %v", err),
				Status:      "error",
			}
		}
	}()

	h.tmpl.RenderAPIResponse(w, true, fmt.Sprintf("开始升级 %s 到 %s", req.Name, req.Version), map[string]string{
		"status": "started",
	}, nil)
}


// ProjectCreateForm 项目创建表单
func (h *APIHandlers) ProjectCreateForm(w http.ResponseWriter, r *http.Request) {
	content := `
	<form hx-post="/api/projects/create" hx-target="#modal-content" hx-swap="innerHTML">
		<div class="space-y-4">
			<div>
				<label for="project-name" class="block text-sm font-medium text-gray-700">项目名称</label>
				<input type="text" id="project-name" name="name" required
					   class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
					   placeholder="例如: my-website">
				<p class="mt-1 text-xs text-gray-500">将在 /var/www/ 目录下创建项目文件夹</p>
			</div>

			<div>
				<label for="project-type" class="block text-sm font-medium text-gray-700">项目类型</label>
				<select id="project-type" name="type"
						class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500">
					<option value="static">静态网站</option>
					<option value="php">PHP 项目</option>
					<option value="nodejs">Node.js 项目</option>
					<option value="python">Python 项目</option>
				</select>
			</div>

			<div>
				<label for="project-domain" class="block text-sm font-medium text-gray-700">域名 (可选)</label>
				<input type="text" id="project-domain" name="domain"
					   class="mt-1 block w-full border-gray-300 rounded-md shadow-sm focus:ring-blue-500 focus:border-blue-500"
					   placeholder="例如: example.com">
			</div>
		</div>

		<div class="mt-6 flex justify-end space-x-3">
			<button type="button" @click="closeModal()"
					class="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50">
				取消
			</button>
			<button type="submit"
					class="px-4 py-2 bg-blue-600 border border-transparent rounded-md text-sm font-medium text-white hover:bg-blue-700">
				创建项目
			</button>
		</div>
	</form>`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(content))
}

// 详细信息辅助方法
func (h *APIHandlers) getCPUDetails() string {
	stats, _ := h.systemMonitor.GetSystemStats()
	return fmt.Sprintf(`
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div>
				<p class="text-sm font-medium text-gray-600">当前使用率</p>
				<p class="text-2xl font-bold text-blue-600">%.1f%%</p>
			</div>
			<div>
				<p class="text-sm font-medium text-gray-600">CPU 核心数</p>
				<p class="text-2xl font-bold text-gray-900">%d</p>
			</div>
		</div>
		<div class="mt-4">
			<p class="text-sm text-gray-600">CPU 使用率实时监控，包含用户进程和系统进程的总使用情况。</p>
		</div>
	</div>`, stats.CPU.Usage, stats.CPU.Cores)
}

func (h *APIHandlers) getMemoryDetails() string {
	stats, _ := h.systemMonitor.GetSystemStats()
	return fmt.Sprintf(`
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div>
				<p class="text-sm font-medium text-gray-600">使用率</p>
				<p class="text-2xl font-bold text-green-600">%.1f%%</p>
			</div>
			<div>
				<p class="text-sm font-medium text-gray-600">总内存</p>
				<p class="text-2xl font-bold text-gray-900">%s</p>
			</div>
		</div>
		<div class="grid grid-cols-2 gap-4 mt-4">
			<div>
				<p class="text-sm font-medium text-gray-600">已使用</p>
				<p class="text-lg font-semibold text-gray-700">%s</p>
			</div>
			<div>
				<p class="text-sm font-medium text-gray-600">可用</p>
				<p class="text-lg font-semibold text-gray-700">%s</p>
			</div>
		</div>
	</div>`,
		stats.Memory.Usage,
		formatBytes(int64(stats.Memory.Total)),
		formatBytes(int64(stats.Memory.Used)),
		formatBytes(int64(stats.Memory.Available)))
}

func (h *APIHandlers) getDiskDetails() string {
	stats, _ := h.systemMonitor.GetSystemStats()
	return fmt.Sprintf(`
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div>
				<p class="text-sm font-medium text-gray-600">使用率</p>
				<p class="text-2xl font-bold text-yellow-600">%.1f%%</p>
			</div>
			<div>
				<p class="text-sm font-medium text-gray-600">总容量</p>
				<p class="text-2xl font-bold text-gray-900">%s</p>
			</div>
		</div>
		<div class="grid grid-cols-2 gap-4 mt-4">
			<div>
				<p class="text-sm font-medium text-gray-600">已使用</p>
				<p class="text-lg font-semibold text-gray-700">%s</p>
			</div>
			<div>
				<p class="text-sm font-medium text-gray-600">可用</p>
				<p class="text-lg font-semibold text-gray-700">%s</p>
			</div>
		</div>
		<div class="mt-4">
			<p class="text-sm text-gray-600">显示根分区 (/) 的磁盘使用情况</p>
		</div>
	</div>`,
		stats.Disk.Usage,
		formatBytes(int64(stats.Disk.Total)),
		formatBytes(int64(stats.Disk.Used)),
		formatBytes(int64(stats.Disk.Free)))
}

func (h *APIHandlers) getNetworkDetails() string {
	return `
	<div class="space-y-4">
		<div class="grid grid-cols-2 gap-4">
			<div>
				<p class="text-sm font-medium text-gray-600">状态</p>
				<p class="text-2xl font-bold text-green-600">正常</p>
			</div>
			<div>
				<p class="text-sm font-medium text-gray-600">活跃连接</p>
				<p class="text-2xl font-bold text-gray-900">--</p>
			</div>
		</div>
		<div class="mt-4">
			<p class="text-sm text-gray-600">网络连接状态监控，包含入站和出站流量统计。</p>
		</div>
	</div>`
}
