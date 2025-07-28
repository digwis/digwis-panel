package i18n

import (
	"net/http"
	"strings"
)

// Translator 翻译器
type Translator struct {
	translations map[string]map[string]string
	defaultLang  string
}

// NewTranslator 创建新的翻译器
func NewTranslator() *Translator {
	t := &Translator{
		translations: make(map[string]map[string]string),
		defaultLang:  "zh",
	}
	
	// 加载内置翻译
	t.loadBuiltinTranslations()
	
	return t
}

// loadBuiltinTranslations 加载内置翻译
func (t *Translator) loadBuiltinTranslations() {
	// 中文翻译
	t.translations["zh"] = map[string]string{
		// 导航
		"nav.dashboard":    "仪表盘",
		"nav.system":       "系统监控",
		"nav.projects":     "项目管理",
		"nav.environment":  "环境配置",
		"nav.logout":       "退出登录",
		"nav.subtitle":     "服务器管理面板",
		"nav.admin":        "管理员",
		
		// 项目管理
		"projects.title":           "项目管理",
		"projects.subtitle":        "管理您的项目和应用程序",
		"projects.welcome.title":   "欢迎使用项目管理",
		"projects.welcome.message": "在 /var/www/ 中未找到项目。您想创建第一个项目吗？",
		"projects.create.first":    "创建第一个项目",
		"projects.skip.now":        "暂时跳过",
		"projects.new_project":     "新建项目",
		"projects.refresh":         "刷新",
		"projects.total":           "总项目",
		"projects.active":          "活跃项目",
		"projects.total_size":      "总大小",
		"projects.backups":         "备份",
		"projects.no_projects":     "暂无项目",
		"projects.get_started":     "通过创建新项目开始使用",
		"projects.create_title":    "创建新项目",
		"projects.project_name":    "项目名称",
		"projects.domain":          "域名",
		"projects.create_database": "创建数据库",
		"projects.enable_ssl":      "启用 SSL",
		"projects.enable_backup":   "启用自动备份",
		"projects.create_project":  "创建项目",
		
		// 环境管理
		"environment.title":           "环境管理",
		"environment.subtitle":        "管理您的服务器环境和依赖项",
		"environment.welcome.title":   "欢迎使用环境管理",
		"environment.welcome.message": "看起来这是您第一次使用。您想安装推荐的环境堆栈吗？",
		"environment.install.all":     "安装全部",
		"environment.skip.now":        "暂时跳过",
		"environment.recommended":     "推荐环境堆栈",
		"environment.refresh":         "刷新",
		"environment.version":         "版本",
		"environment.port":            "端口",
		"environment.status":          "状态",
		"environment.status.installed":     "已安装",
		"environment.status.not_installed": "未安装",
		"environment.status.installing":    "安装中",
		"environment.status.running":       "运行中",
		"environment.status.stopped":       "已停止",
		"environment.action.start":         "启动",
		"environment.action.stop":          "停止",
		"environment.action.restart":       "重启",
		"environment.action.install":       "安装",
		"environment.action.uninstall":     "卸载",
		
		// 通用
		"common.loading":    "加载中...",
		"common.error":      "错误",
		"common.success":    "成功",
		"common.warning":    "警告",
		"common.info":       "信息",
		"common.close":      "关闭",
		"common.cancel":     "取消",
		"common.confirm":    "确认",
		"common.save":       "保存",
		"common.delete":     "删除",
		"common.edit":       "编辑",
		"common.view":       "查看",
		"common.create":     "创建",
		"common.update":     "更新",
		"common.test":       "测试SSE连接",
		
		// 登录
		"login.title":       "登录",
		"login.username":    "用户名",
		"login.password":    "密码",
		"login.submit":      "登录",
		"login.welcome":     "欢迎回来",
		
		// 仪表盘
		"dashboard.title":   "仪表盘",
		"dashboard.welcome": "欢迎使用 DigWis Panel",

		// 系统监控
		"system.cpu.usage":      "CPU 使用率",
		"system.memory.usage":   "内存使用率",
		"system.disk.usage":     "磁盘使用率",
		"system.network.traffic": "网络流量",
		"system.cpu.trend":      "CPU 使用趋势",
		"system.memory.details": "内存详细信息",
		"system.connecting":     "连接中...",
		"system.status.normal":  "系统正常",


	}
	
	// 英文翻译
	t.translations["en"] = map[string]string{
		// Navigation
		"nav.dashboard":    "Dashboard",
		"nav.system":       "System Monitor",
		"nav.projects":     "Project Management",
		"nav.environment":  "Environment",
		"nav.logout":       "Logout",
		"nav.subtitle":     "Server Management Panel",
		"nav.admin":        "Administrator",
		
		// Project Management
		"projects.title":           "Project Management",
		"projects.subtitle":        "Manage your projects and applications",
		"projects.welcome.title":   "Welcome to Project Management",
		"projects.welcome.message": "No projects found in /var/www/. Would you like to create your first project?",
		"projects.create.first":    "Create First Project",
		"projects.skip.now":        "Skip for Now",
		"projects.new_project":     "New Project",
		"projects.refresh":         "Refresh",
		"projects.total":           "Total Projects",
		"projects.active":          "Active Projects",
		"projects.total_size":      "Total Size",
		"projects.backups":         "Backups",
		"projects.no_projects":     "No projects",
		"projects.get_started":     "Get started by creating a new project",
		"projects.create_title":    "Create New Project",
		"projects.project_name":    "Project Name",
		"projects.domain":          "Domain",
		"projects.create_database": "Create database",
		"projects.enable_ssl":      "Enable SSL",
		"projects.enable_backup":   "Enable automatic backup",
		"projects.create_project":  "Create Project",
		
		// Environment Management
		"environment.title":           "Environment Management",
		"environment.subtitle":        "Manage your server environment and dependencies",
		"environment.welcome.title":   "Welcome to Environment Management",
		"environment.welcome.message": "It looks like this is your first time here. Would you like to install the recommended environment stack?",
		"environment.install.all":     "Install All",
		"environment.skip.now":        "Skip for Now",
		"environment.recommended":     "Recommended Environment Stack",
		"environment.refresh":         "Refresh",
		"environment.version":         "Version",
		"environment.port":            "Port",
		"environment.status":          "Status",
		"environment.status.installed":     "Installed",
		"environment.status.not_installed": "Not Installed",
		"environment.status.installing":    "Installing",
		"environment.status.running":       "Running",
		"environment.status.stopped":       "Stopped",
		"environment.action.start":         "Start",
		"environment.action.stop":          "Stop",
		"environment.action.restart":       "Restart",
		"environment.action.install":       "Install",
		"environment.action.uninstall":     "Uninstall",
		
		// Common
		"common.loading":    "Loading...",
		"common.error":      "Error",
		"common.success":    "Success",
		"common.warning":    "Warning",
		"common.info":       "Info",
		"common.close":      "Close",
		"common.cancel":     "Cancel",
		"common.confirm":    "Confirm",
		"common.save":       "Save",
		"common.delete":     "Delete",
		"common.edit":       "Edit",
		"common.view":       "View",
		"common.create":     "Create",
		"common.update":     "Update",
		"common.test":       "Test SSE Connection",
		
		// Login
		"login.title":       "Login",
		"login.username":    "Username",
		"login.password":    "Password",
		"login.submit":      "Sign In",
		"login.welcome":     "Welcome Back",
		
		// Dashboard
		"dashboard.title":   "Dashboard",
		"dashboard.welcome": "Welcome to DigWis Panel",

		// System Monitor
		"system.cpu.usage":      "CPU Usage",
		"system.memory.usage":   "Memory Usage",
		"system.disk.usage":     "Disk Usage",
		"system.network.traffic": "Network Traffic",
		"system.cpu.trend":      "CPU Usage Trend",
		"system.memory.details": "Memory Details",
		"system.connecting":     "Connecting...",
		"system.status.normal":  "System Normal",


	}
}

// T 翻译函数
func (t *Translator) T(lang, key string) string {
	if lang == "" {
		lang = t.defaultLang
	}
	
	if translations, exists := t.translations[lang]; exists {
		if translation, exists := translations[key]; exists {
			return translation
		}
	}
	
	// 回退到默认语言
	if lang != t.defaultLang {
		if translations, exists := t.translations[t.defaultLang]; exists {
			if translation, exists := translations[key]; exists {
				return translation
			}
		}
	}
	
	// 如果都没找到，返回键名
	return key
}

// GetLanguageFromRequest 从请求中获取语言
func (t *Translator) GetLanguageFromRequest(r *http.Request) string {
	// 1. 检查 URL 参数
	if lang := r.URL.Query().Get("lang"); lang != "" {
		if t.IsValidLanguage(lang) {
			return lang
		}
	}
	
	// 2. 检查 Cookie
	if cookie, err := r.Cookie("language"); err == nil {
		if t.IsValidLanguage(cookie.Value) {
			return cookie.Value
		}
	}
	
	// 3. 检查 Accept-Language 头
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		langs := strings.Split(acceptLang, ",")
		for _, lang := range langs {
			lang = strings.TrimSpace(strings.Split(lang, ";")[0])
			if strings.HasPrefix(lang, "zh") {
				return "zh"
			}
			if strings.HasPrefix(lang, "en") {
				return "en"
			}
		}
	}
	
	// 4. 返回默认语言
	return t.defaultLang
}

// IsValidLanguage 检查是否是有效的语言
func (t *Translator) IsValidLanguage(lang string) bool {
	_, exists := t.translations[lang]
	return exists
}

// GetAvailableLanguages 获取可用的语言列表
func (t *Translator) GetAvailableLanguages() []Language {
	return []Language{
		{Code: "zh", Name: "中文", NativeName: "中文"},
		{Code: "en", Name: "English", NativeName: "English"},
	}
}

// Language 语言信息
type Language struct {
	Code       string `json:"code"`
	Name       string `json:"name"`
	NativeName string `json:"native_name"`
}

// SetLanguageCookie 设置语言 Cookie
func (t *Translator) SetLanguageCookie(w http.ResponseWriter, lang string) {
	if t.IsValidLanguage(lang) {
		cookie := &http.Cookie{
			Name:     "language",
			Value:    lang,
			Path:     "/",
			MaxAge:   365 * 24 * 60 * 60, // 1年
			HttpOnly: false, // 允许 JavaScript 访问
			SameSite: http.SameSiteLaxMode,
		}
		http.SetCookie(w, cookie)
	}
}

// 全局翻译器实例
var GlobalTranslator = NewTranslator()

// T 全局翻译函数
func T(lang, key string) string {
	return GlobalTranslator.T(lang, key)
}

// GetLanguageFromRequest 全局语言检测函数
func GetLanguageFromRequest(r *http.Request) string {
	return GlobalTranslator.GetLanguageFromRequest(r)
}

// SetLanguageCookie 全局设置语言 Cookie 函数
func SetLanguageCookie(w http.ResponseWriter, lang string) {
	GlobalTranslator.SetLanguageCookie(w, lang)
}
