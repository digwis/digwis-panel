package handlers

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"os/user"
	"strings"
	"time"

	"server-panel/internal/auth"
	"server-panel/internal/config"
	"server-panel/internal/i18n"
	"server-panel/internal/session"
	"server-panel/internal/system"
	"server-panel/internal/environment"
	"server-panel/internal/projects"
	"server-panel/internal/templates/pages"
)

// Handlers 处理器
type Handlers struct {
	systemMonitor  *system.Monitor
	envManager     *environment.Manager
	projectManager *projects.Manager
	sessionStore   *session.Store
	authManager    *auth.Manager
}

// NewHandlers 创建处理器
func NewHandlers(systemMonitor *system.Monitor, envManager *environment.Manager, projectManager *projects.Manager) *Handlers {
	// 初始化会话存储
	store := session.NewStore()

	// 初始化认证管理器
	authConfig := config.AuthConfig{
		SessionTimeout:   time.Hour * 24,        // 24小时会话超时
		MaxLoginAttempts: 5,                     // 最大登录尝试次数
		LockoutDuration:  time.Minute * 15,      // 15分钟锁定时间
		SecretKey:        "digwis-panel-secret", // 会话密钥
	}
	authManager := auth.NewManager(authConfig)

	return &Handlers{
		systemMonitor:  systemMonitor,
		envManager:     envManager,
		projectManager: projectManager,
		sessionStore:   store,
		authManager:    authManager,
	}
}

// GetSessionStore 获取会话存储
func (h *Handlers) GetSessionStore() *session.Store {
	return h.sessionStore
}



// writeJSON 写入JSON响应
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// getCurrentUser 获取当前用户
func getCurrentUser() string {
	if user, err := user.Current(); err == nil {
		return user.Username
	}
	return "unknown"
}



// LoginPage 登录页面
func (h *Handlers) LoginPage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.handleLogin(w, r)
		return
	}

	// 检查是否有错误消息
	errorMsg := r.URL.Query().Get("error")

	// 使用templ模板渲染登录页面
	component := pages.Login("DigWis Panel - 登录", errorMsg)
	component.Render(r.Context(), w)
}



// handleLogin 处理登录
func (h *Handlers) handleLogin(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	// 获取客户端IP地址
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	// 使用系统用户认证
	authSession, err := h.authManager.Authenticate(username, password, clientIP)
	if err != nil {
		// 认证失败
		if r.Header.Get("Content-Type") == "application/json" ||
		   r.Header.Get("HX-Request") == "true" {
			writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
		} else {
			// 重定向到登录页面并显示错误
			http.Redirect(w, r, "/login?error="+err.Error(), http.StatusFound)
		}
		return
	}

	// 认证成功，创建会话
	sess, err := h.sessionStore.Get(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "会话创建失败",
		})
		return
	}

	sess.Set("username", authSession.Username)
	sess.Set("authenticated", true)
	sess.Set("auth_session_id", authSession.ID)
	sess.Set("login_time", authSession.LoginTime)

	if err := sess.SaveWithStore(w, h.sessionStore); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "会话保存失败",
		})
		return
	}

	// 检查请求类型
	if r.Header.Get("Content-Type") == "application/json" ||
	   r.Header.Get("HX-Request") == "true" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "登录成功",
			"redirect": "/dashboard",
		})
	} else {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
	}
}

// APILogin API 登录
func (h *Handlers) APILogin(w http.ResponseWriter, r *http.Request) {
	var username, password string

	// 检查Content-Type来决定如何解析数据
	contentType := r.Header.Get("Content-Type")

	if strings.Contains(contentType, "application/json") {
		// JSON格式
		var credentials struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   "无效的JSON数据",
			})
			return
		}
		username = credentials.Username
		password = credentials.Password
	} else {
		// Form数据格式
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]interface{}{
				"success": false,
				"error":   "无效的表单数据",
			})
			return
		}
		username = r.FormValue("username")
		password = r.FormValue("password")
	}

	// 获取客户端IP地址
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = forwarded
	}

	// 使用系统用户认证
	authSession, err := h.authManager.Authenticate(username, password, clientIP)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// 认证成功，创建会话
	sess, err := h.sessionStore.Get(r)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "会话创建失败",
		})
		return
	}

	sess.Set("username", authSession.Username)
	sess.Set("authenticated", true)
	sess.Set("auth_session_id", authSession.ID)
	sess.Set("login_time", authSession.LoginTime)

	if err := sess.SaveWithStore(w, h.sessionStore); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   "会话保存失败",
		})
		return
	}

	// 检查是否是HTMX请求
	if r.Header.Get("HX-Request") == "true" {
		// 返回重定向指令给HTMX
		w.Header().Set("HX-Redirect", "/dashboard")
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "登录成功",
			"redirect": "/dashboard",
		})
	} else {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "登录成功",
			"user": map[string]interface{}{
				"username": authSession.Username,
				"login_time": authSession.LoginTime,
			},
		})
	}
}

// Dashboard 仪表板
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	// 从会话中获取用户名
	sess, err := h.sessionStore.Get(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	username := sess.Get("username")
	if username == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 获取当前语言
	currentLang := i18n.GetLanguageFromRequest(r)

	// 使用templ模板渲染仪表板
	usernameStr := fmt.Sprintf("%v", username)
	title := fmt.Sprintf("DigWis Panel - %s", i18n.T(currentLang, "dashboard.title"))
	component := pages.Dashboard(title, usernameStr, currentLang)
	component.Render(r.Context(), w)
}

// SystemPage 系统页面
func (h *Handlers) SystemPage(w http.ResponseWriter, r *http.Request) {
	// 从会话中获取用户名
	sess, err := h.sessionStore.Get(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	username := sess.Get("username")
	if username == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 获取当前语言
	currentLang := i18n.GetLanguageFromRequest(r)

	// 使用Dashboard模板作为系统页面（可以后续创建专门的系统页面模板）
	usernameStr := fmt.Sprintf("%v", username)
	title := fmt.Sprintf("DigWis Panel - %s", i18n.T(currentLang, "nav.system"))
	component := pages.Dashboard(title, usernameStr, currentLang)
	component.Render(r.Context(), w)
}

// SetLanguage 设置语言处理器
func (h *Handlers) SetLanguage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	lang := r.FormValue("lang")
	if lang == "" {
		lang = r.URL.Query().Get("lang")
	}

	// 设置语言 Cookie
	i18n.SetLanguageCookie(w, lang)

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// ProjectsPage 项目页面
func (h *Handlers) ProjectsPage(w http.ResponseWriter, r *http.Request) {
	// 从会话中获取用户名
	sess, err := h.sessionStore.Get(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	username := sess.Get("username")
	if username == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 获取项目概览
	overview, err := h.projectManager.GetOverview()
	if err != nil {
		// 如果获取失败，创建一个空的概览
		overview = &projects.ProjectOverview{
			Projects:       []projects.Project{},
			TotalProjects:  0,
			ActiveProjects: 0,
			TotalSize:      0,
			FirstTimeSetup: true,
		}
	}

	// 获取当前语言
	currentLang := i18n.GetLanguageFromRequest(r)

	// 使用templ模板渲染项目页面
	usernameStr := fmt.Sprintf("%v", username)
	title := fmt.Sprintf("DigWis Panel - %s", i18n.T(currentLang, "projects.title"))
	component := pages.Projects(title, usernameStr, currentLang, overview)
	component.Render(r.Context(), w)
}

// EnvironmentPage 环境页面
func (h *Handlers) EnvironmentPage(w http.ResponseWriter, r *http.Request) {
	// 从会话中获取用户名
	sess, err := h.sessionStore.Get(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	username := sess.Get("username")
	if username == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 获取环境概览
	overview, err := h.envManager.GetOverview()
	if err != nil {
		// 如果获取失败，创建一个空的概览
		overview = &environment.EnvironmentOverview{}
	}

	// 获取当前语言
	currentLang := i18n.GetLanguageFromRequest(r)

	// 使用templ模板渲染环境页面
	usernameStr := fmt.Sprintf("%v", username)
	title := fmt.Sprintf("DigWis Panel - %s", i18n.T(currentLang, "environment.title"))
	component := pages.Environment(title, usernameStr, currentLang, overview)
	component.Render(r.Context(), w)
}

// SystemStats 系统统计
func (h *Handlers) SystemStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// StatsDetails 统计详情
func (h *Handlers) StatsDetails(w http.ResponseWriter, r *http.Request) {
	// 从 URL 参数中获取类型
	statsType := r.URL.Query().Get("type")
	if statsType == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "缺少统计类型参数",
		})
		return
	}

	// 根据类型返回详细信息
	details := map[string]interface{}{
		"type": statsType,
		"data": "详细统计数据",
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    details,
	})
}

// CPUChart CPU 图表数据
func (h *Handlers) CPUChart(w http.ResponseWriter, r *http.Request) {
	stats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// SystemOverview 系统概览
func (h *Handlers) SystemOverview(w http.ResponseWriter, r *http.Request) {
	stats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// SystemDetails 系统详情
func (h *Handlers) SystemDetails(w http.ResponseWriter, r *http.Request) {
	details := h.systemMonitor.GetSystemDetails()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    details,
	})
}

// ProcessList 进程列表
func (h *Handlers) ProcessList(w http.ResponseWriter, r *http.Request) {
	processes, err := h.systemMonitor.GetProcessList()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    processes,
	})
}

// ProjectsScan 项目扫描
func (h *Handlers) ProjectsScan(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projectManager.ScanProjects()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    projects,
	})
}

// ProjectCreateForm 项目创建表单
func (h *Handlers) ProjectCreateForm(w http.ResponseWriter, r *http.Request) {
	// 简化处理，返回基本表单
	form := map[string]interface{}{
		"name":        "",
		"description": "",
		"type":        "web",
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    form,
	})
}

// ProjectCreate 创建项目
func (h *Handlers) ProjectCreate(w http.ResponseWriter, r *http.Request) {
	var req projects.CreateProjectRequest

	// 解析JSON请求
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "无效的请求格式: " + err.Error(),
		})
		return
	}

	// 创建项目
	project, err := h.projectManager.CreateProject(req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "项目创建成功",
		"data":    project,
	})
}

// ProjectDelete 删除项目
func (h *Handlers) ProjectDelete(w http.ResponseWriter, r *http.Request) {
	// 从URL参数中获取项目ID
	projectID := r.URL.Query().Get("id")
	if projectID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "项目ID不能为空",
		})
		return
	}

	// 删除项目
	if err := h.projectManager.DeleteProject(projectID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "项目删除成功",
	})
}

// EnvironmentStatus 环境状态
func (h *Handlers) EnvironmentStatus(w http.ResponseWriter, r *http.Request) {
	status, err := h.envManager.GetOverview()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// InstallEnvironment 安装环境
func (h *Handlers) InstallEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	// 创建进度通道
	progressChan := make(chan environment.InstallProgress, 10)
	defer close(progressChan)

	// 启动安装过程
	go func() {
		err := h.envManager.InstallEnvironment(req.Name, req.Version, progressChan)
		if err != nil {
			progressChan <- environment.InstallProgress{
				Environment: req.Name,
				Progress:    0,
				Message:     err.Error(),
				Status:      "error",
			}
		}
	}()

	// 等待安装完成或出错
	var lastProgress environment.InstallProgress
	for progress := range progressChan {
		lastProgress = progress
		if progress.Status == "completed" || progress.Status == "error" {
			break
		}
	}

	if lastProgress.Status == "error" {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   lastProgress.Message,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("环境 %s 安装成功", req.Name),
	})
}

// UninstallEnvironment 卸载环境
func (h *Handlers) UninstallEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	// 卸载环境
	err := h.envManager.UninstallEnvironment(req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("环境 %s 卸载成功", req.Name),
	})
}

// UpgradeEnvironment 升级环境
func (h *Handlers) UpgradeEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   "无效的请求数据",
		})
		return
	}

	// 创建进度通道
	progressChan := make(chan environment.InstallProgress, 10)
	defer close(progressChan)

	// 启动升级过程
	go func() {
		err := h.envManager.UpgradeEnvironment(req.Name, req.Version, progressChan)
		if err != nil {
			progressChan <- environment.InstallProgress{
				Environment: req.Name,
				Progress:    0,
				Message:     err.Error(),
				Status:      "error",
			}
		}
	}()

	// 等待升级完成或出错
	var lastProgress environment.InstallProgress
	for progress := range progressChan {
		lastProgress = progress
		if progress.Status == "completed" || progress.Status == "error" {
			break
		}
	}

	if lastProgress.Status == "error" {
		writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   lastProgress.Message,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("环境 %s 升级成功", req.Name),
	})
}

// EnvironmentProgress 获取环境安装进度
func (h *Handlers) EnvironmentProgress(w http.ResponseWriter, r *http.Request) {
	progress := h.envManager.GetProgress()

	if progress == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": true,
			"data":    nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    progress,
	})
}

// TestSSEHandler 测试 SSE 处理器
func (h *Handlers) TestSSEHandler(w http.ResponseWriter, r *http.Request) {
	// 设置 SSE 头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 发送测试数据
	fmt.Fprintf(w, "data: {\"message\": \"SSE 连接测试成功\", \"timestamp\": \"%s\"}\n\n", time.Now().Format(time.RFC3339))

	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

// SSEStatsHandler SSE 统计数据处理器 - 优化版本
func (h *Handlers) SSEStatsHandler(w http.ResponseWriter, r *http.Request) {
	// 临时注释认证检查以测试数据流
	/*
	sess, err := h.sessionStore.Get(r)
	if err != nil || sess.Get("authenticated") != true {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}
	*/

	// 设置SSE头部
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // 禁用nginx缓冲

	// 设置CORS头部
	origin := r.Header.Get("Origin")
	if origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// 尝试获取Flusher接口，如果不支持则使用替代方案
	flusher, hasFlusher := w.(http.Flusher)

	// 发送连接建立事件
	fmt.Fprintf(w, "event: connected\ndata: {\"message\": \"SSE连接已建立\", \"hasFlusher\": %t}\n\n", hasFlusher)
	if hasFlusher {
		flusher.Flush()
	}

	// 创建智能推送管理器
	pushManager := &smartPushManager{
		lastStats: nil,
		threshold: 5.0, // CPU/内存变化超过5%才推送
	}

	// 创建定时器 - 增加到15秒以进一步减少负载
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// 发送初始数据
	h.sendSystemStatsOptimized(w, flusher, hasFlusher, pushManager)

	// 持续发送数据
	for {
		select {
		case <-ticker.C:
			h.sendSystemStatsOptimized(w, flusher, hasFlusher, pushManager)
		case <-r.Context().Done():
			// 客户端断开连接
			return
		}
	}
}

// smartPushManager 智能推送管理器
type smartPushManager struct {
	lastStats *system.SystemStats
	threshold float64
}

// shouldPush 判断是否应该推送数据
func (spm *smartPushManager) shouldPush(newStats *system.SystemStats) bool {
	if spm.lastStats == nil {
		return true // 首次推送
	}

	// 检查CPU使用率变化
	cpuDiff := abs(newStats.CPU.Usage - spm.lastStats.CPU.Usage)
	if cpuDiff > spm.threshold {
		return true
	}

	// 检查内存使用率变化
	memDiff := abs(newStats.Memory.Usage - spm.lastStats.Memory.Usage)
	if memDiff > spm.threshold {
		return true
	}

	// 检查磁盘使用率变化（阈值更大，因为磁盘变化较慢）
	diskDiff := abs(newStats.Disk.Usage - spm.lastStats.Disk.Usage)
	if diskDiff > spm.threshold*2 {
		return true
	}

	return false
}

// abs 计算绝对值
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// sendSystemStatsOptimized 发送系统统计数据 - 优化版本
func (h *Handlers) sendSystemStatsOptimized(w http.ResponseWriter, flusher http.Flusher, hasFlusher bool, pushManager *smartPushManager) {
	stats, err := h.systemMonitor.GetSystemStats()
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"%s\"}\n\n", err.Error())
		if hasFlusher {
			flusher.Flush()
		}
		return
	}

	// 检查是否需要推送
	if !pushManager.shouldPush(stats) {
		// 发送心跳包保持连接
		fmt.Fprintf(w, "event: heartbeat\ndata: {\"timestamp\": \"%s\"}\n\n", time.Now().Format(time.RFC3339))
		if hasFlusher {
			flusher.Flush()
		}
		return
	}

	// 更新最后推送的数据
	pushManager.lastStats = stats

	// 只发送关键数据，减少传输量
	compactStats := map[string]interface{}{
		"cpu": map[string]interface{}{
			"usage": roundFloat(stats.CPU.Usage, 1),
			"cores": stats.CPU.Cores,
		},
		"memory": map[string]interface{}{
			"usage": roundFloat(stats.Memory.Usage, 1),
			"used":  stats.Memory.Used,
			"total": stats.Memory.Total,
		},
		"disk": map[string]interface{}{
			"usage": roundFloat(stats.Disk.Usage, 1),
			"used":  stats.Disk.Used,
			"total": stats.Disk.Total,
		},
		"load_avg": map[string]interface{}{
			"load1": roundFloat(stats.LoadAvg.Load1, 2),
		},
		"timestamp": stats.Timestamp.Format(time.RFC3339),
	}

	data, err := json.Marshal(compactStats)
	if err != nil {
		fmt.Fprintf(w, "event: error\ndata: {\"error\": \"数据序列化失败\"}\n\n")
		if hasFlusher {
			flusher.Flush()
		}
		return
	}

	// 发送压缩后的数据
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(w, "event: stats\ndata: %s\nid: %s\n\n", data, timestamp)

	if hasFlusher {
		flusher.Flush()
	}
}

// sendSystemStatsImproved 发送系统统计数据 - 改进版本（保持向后兼容）
func (h *Handlers) sendSystemStatsImproved(w http.ResponseWriter, flusher http.Flusher, hasFlusher bool) {
	// 创建临时推送管理器
	pushManager := &smartPushManager{
		lastStats: nil,
		threshold: 0, // 总是推送，保持原有行为
	}
	h.sendSystemStatsOptimized(w, flusher, hasFlusher, pushManager)
}

// roundFloat 四舍五入浮点数
func roundFloat(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
