package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"server-panel/internal/auth"
	"server-panel/internal/config"
	"server-panel/internal/environment"
	"server-panel/internal/projects"
	"server-panel/internal/ssl"
	"server-panel/internal/system"
)

// Handlers 处理器集合
type Handlers struct {
	auth        *auth.Manager
	system      *system.Monitor
	environment *environment.Manager
	projects    *projects.Manager
	ssl         *ssl.Manager
	config      *config.Config
}

// New 创建处理器
func New(authManager *auth.Manager, systemMonitor *system.Monitor, projectManager *projects.Manager, cfg *config.Config) *Handlers {
	return &Handlers{
		auth:        authManager,
		system:      systemMonitor,
		environment: environment.NewManager(),
		projects:    projectManager,
		ssl:         ssl.NewManager(),
		config:      cfg,
	}
}

// Home 首页
func (h *Handlers) Home(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusFound)
}

// Login 登录页面
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>服务器管理面板 - 登录</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .login-container {
            background: white;
            padding: 40px;
            border-radius: 12px;
            box-shadow: 0 15px 35px rgba(0,0,0,0.1);
            width: 100%;
            max-width: 400px;
        }
        .logo {
            text-align: center;
            margin-bottom: 30px;
        }
        .logo h1 {
            color: #333;
            font-size: 28px;
            font-weight: 600;
        }
        .logo p {
            color: #666;
            margin-top: 5px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #333;
            font-weight: 500;
        }
        input[type="text"], input[type="password"] {
            width: 100%;
            padding: 12px 16px;
            border: 2px solid #e1e5e9;
            border-radius: 8px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        input[type="text"]:focus, input[type="password"]:focus {
            outline: none;
            border-color: #667eea;
        }
        .btn {
            width: 100%;
            padding: 12px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        .btn:hover {
            transform: translateY(-2px);
        }
        .btn:disabled {
            opacity: 0.6;
            cursor: not-allowed;
            transform: none;
        }
        .error {
            background: #fee;
            color: #c33;
            padding: 12px;
            border-radius: 8px;
            margin-bottom: 20px;
            border: 1px solid #fcc;
        }
        .info {
            background: #f0f9ff;
            color: #0369a1;
            padding: 12px;
            border-radius: 8px;
            margin-top: 20px;
            border: 1px solid #bae6fd;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <div class="logo">
            <h1>🚀 服务器面板</h1>
            <p>系统管理控制台</p>
        </div>
        
        {{if .Error}}
        <div class="error">{{.Error}}</div>
        {{end}}
        
        <form method="post" id="loginForm">
            <div class="form-group">
                <label for="username">用户名</label>
                <input type="text" id="username" name="username" required placeholder="输入系统用户名">
            </div>
            
            <div class="form-group">
                <label for="password">密码</label>
                <input type="password" id="password" name="password" required placeholder="输入密码">
            </div>
            
            <button type="submit" class="btn" id="loginBtn">
                <span id="btnText">登录</span>
                <span id="btnLoading" style="display: none;">登录中...</span>
            </button>
        </form>
        
        <div class="info">
            <strong>提示：</strong>使用具有管理员权限的系统用户账户登录<br>
            支持的管理员组：sudo、wheel、admin 或 root 用户
        </div>
    </div>

    <script>
        document.getElementById('loginForm').addEventListener('submit', function(e) {
            const btn = document.getElementById('loginBtn');
            const btnText = document.getElementById('btnText');
            const btnLoading = document.getElementById('btnLoading');
            
            btn.disabled = true;
            btnText.style.display = 'none';
            btnLoading.style.display = 'inline';
        });
    </script>
</body>
</html>`

		t, _ := template.New("login").Parse(tmpl)
		data := map[string]string{
			"Error": r.URL.Query().Get("error"),
		}
		t.Execute(w, data)
		return
	}

	// 处理POST请求
	username := r.FormValue("username")
	password := r.FormValue("password")
	
	if username == "" || password == "" {
		http.Redirect(w, r, "/login?error=请输入用户名和密码", http.StatusFound)
		return
	}

	// 获取客户端IP
	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	// 认证用户
	session, err := h.auth.Authenticate(username, password, clientIP)
	if err != nil {
		http.Redirect(w, r, "/login?error="+err.Error(), http.StatusFound)
		return
	}

	// 设置会话cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 在生产环境中应该设置为true
		MaxAge:   int(h.config.Auth.SessionTimeout.Seconds()),
	})

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// APILogin API登录
func (h *Handlers) APILogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	clientIP := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		clientIP = strings.Split(forwarded, ",")[0]
	}

	session, err := h.auth.Authenticate(req.Username, req.Password, clientIP)
	if err != nil {
		h.jsonError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.jsonSuccess(w, map[string]interface{}{
		"session_id": session.ID,
		"user":       session,
	})
}

// RequireAuth 认证中间件
func (h *Handlers) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				h.jsonError(w, "未授权访问", http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/login", http.StatusFound)
			}
			return
		}

		session, err := h.auth.ValidateSession(cookie.Value)
		if err != nil {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				h.jsonError(w, "会话无效", http.StatusUnauthorized)
			} else {
				http.Redirect(w, r, "/login", http.StatusFound)
			}
			return
		}

		// 将会话信息添加到请求上下文
		r.Header.Set("X-User-ID", session.UserID)
		r.Header.Set("X-Username", session.Username)

		next.ServeHTTP(w, r)
	})
}

// jsonSuccess 返回成功响应
func (h *Handlers) jsonSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

// jsonError 返回错误响应
func (h *Handlers) jsonError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   message,
	})
}

// Dashboard 仪表板
func (h *Handlers) Dashboard(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Username")

	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>服务器管理面板 - 仪表板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f8fafc;
            color: #334155;
        }
        .header {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 1rem 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            color: #1e293b;
            font-size: 1.5rem;
        }
        .user-info {
            display: flex;
            align-items: center;
            gap: 1rem;
        }
        .nav {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 0 2rem;
        }
        .nav ul {
            list-style: none;
            display: flex;
            gap: 2rem;
        }
        .nav a {
            display: block;
            padding: 1rem 0;
            text-decoration: none;
            color: #64748b;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }
        .nav a:hover, .nav a.active {
            color: #3b82f6;
            border-bottom-color: #3b82f6;
        }
        .container {
            max-width: 1200px;
            margin: 2rem auto;
            padding: 0 2rem;
        }
        .grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        .card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }
        .card h3 {
            margin-bottom: 1rem;
            color: #1e293b;
        }
        .stat {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.5rem;
        }
        .stat-value {
            font-weight: 600;
            color: #3b82f6;
        }
        .progress {
            width: 100%;
            height: 8px;
            background: #e2e8f0;
            border-radius: 4px;
            overflow: hidden;
            margin-top: 0.5rem;
        }
        .progress-bar {
            height: 100%;
            background: linear-gradient(90deg, #3b82f6, #1d4ed8);
            transition: width 0.3s;
        }
        .btn {
            background: #3b82f6;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 6px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            transition: background 0.2s;
        }
        .btn:hover {
            background: #2563eb;
        }
        .btn-danger {
            background: #ef4444;
        }
        .btn-danger:hover {
            background: #dc2626;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 服务器管理面板</h1>
        <div class="user-info">
            <span>欢迎, {{.Username}}</span>
            <form method="post" action="/logout" style="display: inline;">
                <button type="submit" class="btn btn-danger">退出</button>
            </form>
        </div>
    </div>

    <nav class="nav">
        <ul>
            <li><a href="/dashboard" class="active">仪表板</a></li>
            <li><a href="/environment">环境管理</a></li>
            <li><a href="/projects">项目管理</a></li>
        </ul>
    </nav>

    <div class="container">
        <div class="grid" id="statsGrid">
            <div class="card">
                <h3>💻 系统信息</h3>
                <div id="systemInfo">加载中...</div>
            </div>

            <div class="card">
                <h3>💾 内存使用</h3>
                <div id="memoryStats">加载中...</div>
            </div>

            <div class="card">
                <h3>💿 磁盘使用</h3>
                <div id="diskStats">加载中...</div>
            </div>

            <div class="card">
                <h3>🌐 网络流量</h3>
                <div id="networkStats">加载中...</div>
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>🔧 快速操作</h3>
                <div style="display: flex; gap: 1rem; flex-wrap: wrap;">
                    <a href="/environment" class="btn">安装环境</a>
                    <a href="/system" class="btn">系统监控</a>
                    <a href="/files" class="btn">文件管理</a>
                    <a href="/logs" class="btn">查看日志</a>
                </div>
            </div>

            <div class="card">
                <h3>📈 系统负载</h3>
                <div id="loadStats">加载中...</div>
            </div>
        </div>
    </div>

    <script>
        // 加载系统统计信息
        function loadStats() {
            fetch('/api/system/stats')
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        updateStats(data.data);
                    }
                })
                .catch(error => console.error('Error:', error));
        }

        function updateStats(stats) {
            // 系统信息
            document.getElementById('systemInfo').innerHTML = ` + "`" + `
                <div class="stat">
                    <span>主机名</span>
                    <span class="stat-value">${stats.hostname}</span>
                </div>
                <div class="stat">
                    <span>操作系统</span>
                    <span class="stat-value">${stats.os}</span>
                </div>
                <div class="stat">
                    <span>内核版本</span>
                    <span class="stat-value">${stats.kernel}</span>
                </div>
                <div class="stat">
                    <span>运行时间</span>
                    <span class="stat-value">${stats.uptime}</span>
                </div>
                <div class="stat">
                    <span>CPU核心</span>
                    <span class="stat-value">${stats.cpu.cores}</span>
                </div>
            ` + "`" + `;

            // 内存使用
            const memUsage = stats.memory.usage.toFixed(1);
            document.getElementById('memoryStats').innerHTML = ` + "`" + `
                <div class="stat">
                    <span>使用率</span>
                    <span class="stat-value">${memUsage}%</span>
                </div>
                <div class="progress">
                    <div class="progress-bar" style="width: ${memUsage}%"></div>
                </div>
                <div class="stat">
                    <span>已用</span>
                    <span class="stat-value">${formatBytes(stats.memory.used)}</span>
                </div>
                <div class="stat">
                    <span>总计</span>
                    <span class="stat-value">${formatBytes(stats.memory.total)}</span>
                </div>
            ` + "`" + `;

            // 磁盘使用
            const diskUsage = stats.disk.usage.toFixed(1);
            document.getElementById('diskStats').innerHTML = ` + "`" + `
                <div class="stat">
                    <span>使用率</span>
                    <span class="stat-value">${diskUsage}%</span>
                </div>
                <div class="progress">
                    <div class="progress-bar" style="width: ${diskUsage}%"></div>
                </div>
                <div class="stat">
                    <span>已用</span>
                    <span class="stat-value">${formatBytes(stats.disk.used)}</span>
                </div>
                <div class="stat">
                    <span>总计</span>
                    <span class="stat-value">${formatBytes(stats.disk.total)}</span>
                </div>
            ` + "`" + `;

            // 网络流量
            document.getElementById('networkStats').innerHTML = ` + "`" + `
                <div class="stat">
                    <span>接收</span>
                    <span class="stat-value">${formatBytes(stats.network.bytes_received)}</span>
                </div>
                <div class="stat">
                    <span>发送</span>
                    <span class="stat-value">${formatBytes(stats.network.bytes_sent)}</span>
                </div>
            ` + "`" + `;

            // 系统负载
            document.getElementById('loadStats').innerHTML = ` + "`" + `
                <div class="stat">
                    <span>1分钟</span>
                    <span class="stat-value">${stats.load_avg.load1.toFixed(2)}</span>
                </div>
                <div class="stat">
                    <span>5分钟</span>
                    <span class="stat-value">${stats.load_avg.load5.toFixed(2)}</span>
                </div>
                <div class="stat">
                    <span>15分钟</span>
                    <span class="stat-value">${stats.load_avg.load15.toFixed(2)}</span>
                </div>
            ` + "`" + `;
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        // 页面加载时获取统计信息
        loadStats();

        // 每30秒刷新一次
        setInterval(loadStats, 30000);
    </script>
</body>
</html>`

	t, _ := template.New("dashboard").Parse(tmpl)
	data := map[string]string{
		"Username": username,
	}
	t.Execute(w, data)
}

// SystemStats API获取系统统计
func (h *Handlers) SystemStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.system.GetSystemStats()
	if err != nil {
		h.jsonError(w, "获取系统统计失败", http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, stats)
}

// SystemDetails API获取详细系统信息
func (h *Handlers) SystemDetails(w http.ResponseWriter, r *http.Request) {
	details := h.system.GetSystemDetails()
	h.jsonSuccess(w, details)
}

// Environment 环境管理页面
func (h *Handlers) Environment(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Username")

	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>环境管理 - 服务器管理面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f8fafc;
            color: #334155;
        }
        .header {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 1rem 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            color: #1e293b;
            font-size: 1.5rem;
        }
        .user-info {
            display: flex;
            align-items: center;
            gap: 1rem;
        }
        .nav {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 0 2rem;
        }
        .nav ul {
            list-style: none;
            display: flex;
            gap: 2rem;
        }
        .nav a {
            display: block;
            padding: 1rem 0;
            text-decoration: none;
            color: #64748b;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }
        .nav a:hover, .nav a.active {
            color: #3b82f6;
            border-bottom-color: #3b82f6;
        }
        .container {
            max-width: 1200px;
            margin: 2rem auto;
            padding: 0 2rem;
        }
        .env-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 1.5rem;
        }
        .env-card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            border: 1px solid #e2e8f0;
        }
        .env-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
        }
        .env-title {
            font-size: 1.25rem;
            font-weight: 600;
            color: #1e293b;
        }
        .status {
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.875rem;
            font-weight: 500;
        }
        .status.installed {
            background: #dcfce7;
            color: #166534;
        }
        .status.not-installed {
            background: #fef2f2;
            color: #991b1b;
        }
        .status.installing {
            background: #fef3c7;
            color: #92400e;
        }
        .env-description {
            color: #64748b;
            margin-bottom: 1rem;
            line-height: 1.5;
        }
        .env-version {
            font-size: 0.875rem;
            color: #64748b;
            margin-bottom: 1rem;
        }
        .env-actions {
            display: flex;
            gap: 0.5rem;
        }
        .btn {
            padding: 0.5rem 1rem;
            border: none;
            border-radius: 6px;
            cursor: pointer;
            font-size: 0.875rem;
            font-weight: 500;
            transition: all 0.2s;
            text-decoration: none;
            display: inline-block;
            text-align: center;
        }
        .btn-primary {
            background: #3b82f6;
            color: white;
        }
        .btn-primary:hover {
            background: #2563eb;
        }
        .btn-danger {
            background: #ef4444;
            color: white;
        }
        .btn-danger:hover {
            background: #dc2626;
        }
        .btn:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }
        .progress-modal {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            display: none;
            align-items: center;
            justify-content: center;
            z-index: 1000;
        }
        .progress-content {
            background: white;
            padding: 2rem;
            border-radius: 8px;
            max-width: 500px;
            width: 90%;
        }
        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e2e8f0;
            border-radius: 4px;
            overflow: hidden;
            margin: 1rem 0;
        }
        .progress-fill {
            height: 100%;
            background: #3b82f6;
            transition: width 0.3s;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 服务器管理面板</h1>
        <div class="user-info">
            <span>欢迎, {{.Username}}</span>
            <form method="post" action="/logout" style="display: inline;">
                <button type="submit" class="btn btn-danger">退出</button>
            </form>
        </div>
    </div>

    <nav class="nav">
        <ul>
            <li><a href="/dashboard">仪表板</a></li>
            <li><a href="/environment" class="active">环境管理</a></li>
            <li><a href="/projects">项目管理</a></li>
        </ul>
    </nav>

    <div class="container">
        <h2 style="margin-bottom: 2rem;">环境管理</h2>
        <div class="env-grid" id="environmentGrid">
            加载中...
        </div>
    </div>

    <!-- 版本选择模态框 -->
    <div class="progress-modal" id="versionModal">
        <div class="progress-content">
            <h3 id="versionTitle">选择版本</h3>
            <div style="margin: 1rem 0;">
                <label for="versionSelect" style="display: block; margin-bottom: 0.5rem;">选择要安装的版本：</label>
                <select id="versionSelect" style="width: 100%; padding: 0.5rem; border: 1px solid #ddd; border-radius: 4px;">
                    <option value="">加载中...</option>
                </select>
            </div>
            <div style="text-align: center; margin-top: 1rem;">
                <button class="btn btn-primary" onclick="confirmVersionAction()" id="confirmBtn">确认</button>
                <button class="btn" onclick="closeVersionModal()" style="margin-left: 0.5rem;">取消</button>
            </div>
        </div>
    </div>

    <!-- PHP插件管理模态框 -->
    <div class="progress-modal" id="phpExtensionsModal">
        <div class="progress-content" style="max-width: 800px; width: 95%;">
            <h3>PHP插件管理</h3>
            <div id="phpExtensionsList" style="max-height: 400px; overflow-y: auto; margin: 1rem 0;">
                加载中...
            </div>
            <div style="text-align: center; margin-top: 1rem;">
                <button class="btn" onclick="closePHPExtensionsModal()">关闭</button>
            </div>
        </div>
    </div>

    <!-- 安装进度模态框 -->
    <div class="progress-modal" id="progressModal">
        <div class="progress-content">
            <h3 id="progressTitle">安装中...</h3>
            <div class="progress-bar">
                <div class="progress-fill" id="progressFill"></div>
            </div>
            <div id="progressMessage">准备安装...</div>
            <div style="text-align: center; margin-top: 1rem;">
                <button class="btn btn-danger" onclick="closeProgressModal()">关闭</button>
            </div>
        </div>
    </div>

    <script>
        let environments = [];
        let currentAction = '';
        let currentEnvironment = '';

        function loadEnvironments() {
            fetch('/api/environment/status')
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        environments = data.data;
                        renderEnvironments();
                    }
                })
                .catch(error => console.error('Error:', error));
        }

        function renderEnvironments() {
            const grid = document.getElementById('environmentGrid');
            grid.innerHTML = environments.map(env => ` + "`" + `
                <div class="env-card">
                    <div class="env-header">
                        <div class="env-title">${env.display_name}</div>
                        <div class="status ${env.status}">${getStatusText(env.status)}</div>
                    </div>
                    <div class="env-description">${env.description}</div>
                    ${env.version ? ` + "`" + `<div class="env-version">版本: ${env.version}</div>` + "`" + ` : ''}
                    <div class="env-actions">
                        ${env.status === 'not_installed' ?
                            ` + "`" + `<button class="btn btn-primary" onclick="showVersionModal('${env.name}', 'install')">安装</button>` + "`" + ` :
                            ` + "`" + `
                            <button class="btn btn-primary" onclick="showVersionModal('${env.name}', 'upgrade')">升级</button>
                            ${env.name === 'php' ? ` + "`" + `<button class="btn" onclick="showPHPExtensions()" style="background: #10b981; color: white;">插件管理</button>` + "`" + ` : ''}
                            <button class="btn btn-danger" onclick="uninstallEnvironment('${env.name}')">卸载</button>
                            ` + "`" + `
                        }
                    </div>
                </div>
            ` + "`" + `).join('');
        }

        function getStatusText(status) {
            const statusMap = {
                'installed': '已安装',
                'not_installed': '未安装',
                'installing': '安装中',
                'error': '错误'
            };
            return statusMap[status] || status;
        }

        function showVersionModal(name, action) {
            currentEnvironment = name;
            currentAction = action;

            const env = environments.find(e => e.name === name);
            if (!env) return;

            document.getElementById('versionTitle').textContent = ` + "`" + `${action === 'install' ? '安装' : '升级'} ${env.display_name}` + "`" + `;
            document.getElementById('confirmBtn').textContent = action === 'install' ? '安装' : '升级';

            // 填充版本选择
            const select = document.getElementById('versionSelect');
            select.innerHTML = '';

            if (env.available_versions && env.available_versions.length > 0) {
                env.available_versions.forEach(version => {
                    const option = document.createElement('option');
                    option.value = version;
                    option.textContent = version;
                    if (version === '最新版本') {
                        option.selected = true;
                    }
                    select.appendChild(option);
                });
            } else {
                const option = document.createElement('option');
                option.value = '最新版本';
                option.textContent = '最新版本';
                option.selected = true;
                select.appendChild(option);
            }

            document.getElementById('versionModal').style.display = 'flex';
        }

        function closeVersionModal() {
            document.getElementById('versionModal').style.display = 'none';
        }

        function confirmVersionAction() {
            const selectedVersion = document.getElementById('versionSelect').value;
            if (!selectedVersion) {
                alert('请选择版本');
                return;
            }

            closeVersionModal();

            if (currentAction === 'install') {
                installEnvironment(currentEnvironment, selectedVersion);
            } else if (currentAction === 'upgrade') {
                upgradeEnvironment(currentEnvironment, selectedVersion);
            }
        }

        function installEnvironment(name, version) {
            const env = environments.find(e => e.name === name);
            if (!env) return;

            document.getElementById('progressTitle').textContent = ` + "`" + `安装 ${env.display_name} ${version}` + "`" + `;
            document.getElementById('progressFill').style.width = '0%';
            document.getElementById('progressMessage').textContent = '准备安装...';
            document.getElementById('progressModal').style.display = 'flex';

            fetch('/api/environment/install', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: name, version: version })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    // 模拟进度更新
                    simulateProgress(name, '安装');
                } else {
                    alert('安装失败: ' + data.error);
                    closeProgressModal();
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('安装请求失败');
                closeProgressModal();
            });
        }

        function upgradeEnvironment(name, version) {
            const env = environments.find(e => e.name === name);
            if (!env) return;

            document.getElementById('progressTitle').textContent = ` + "`" + `升级 ${env.display_name} 到 ${version}` + "`" + `;
            document.getElementById('progressFill').style.width = '0%';
            document.getElementById('progressMessage').textContent = '准备升级...';
            document.getElementById('progressModal').style.display = 'flex';

            fetch('/api/environment/upgrade', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: name, version: version })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    // 模拟进度更新
                    simulateProgress(name, '升级');
                } else {
                    alert('升级失败: ' + data.error);
                    closeProgressModal();
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('升级请求失败');
                closeProgressModal();
            });
        }

        function uninstallEnvironment(name) {
            if (!confirm('确定要卸载此环境吗？')) return;

            fetch('/api/environment/uninstall', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: name })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('卸载成功');
                    loadEnvironments();
                } else {
                    alert('卸载失败: ' + data.error);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert('卸载请求失败');
            });
        }

        function simulateProgress(name, action = '安装') {
            let progress = 0;
            const interval = setInterval(() => {
                progress += Math.random() * 20;
                if (progress >= 100) {
                    progress = 100;
                    clearInterval(interval);
                    setTimeout(() => {
                        closeProgressModal();
                        loadEnvironments();
                    }, 1000);
                }

                document.getElementById('progressFill').style.width = progress + '%';
                document.getElementById('progressMessage').textContent = ` + "`" + `${action}进度: ${Math.round(progress)}%` + "`" + `;
            }, 500);
        }

        function closeProgressModal() {
            document.getElementById('progressModal').style.display = 'none';
        }

        // PHP插件管理
        function showPHPExtensions() {
            document.getElementById('phpExtensionsModal').style.display = 'flex';
            loadPHPExtensions();
        }

        function closePHPExtensionsModal() {
            document.getElementById('phpExtensionsModal').style.display = 'none';
        }

        function loadPHPExtensions() {
            fetch('/api/php/extensions')
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        renderPHPExtensions(data.data);
                    } else {
                        document.getElementById('phpExtensionsList').innerHTML = '<p>加载失败: ' + data.error + '</p>';
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    document.getElementById('phpExtensionsList').innerHTML = '<p>加载失败</p>';
                });
        }

        function renderPHPExtensions(extensions) {
            const container = document.getElementById('phpExtensionsList');
            container.innerHTML = extensions.map(ext => ` + "`" + `
                <div style="border: 1px solid #e2e8f0; border-radius: 8px; padding: 1rem; margin-bottom: 0.5rem; display: flex; justify-content: space-between; align-items: center;">
                    <div>
                        <div style="font-weight: 600; color: #1e293b;">${ext.display_name}</div>
                        <div style="color: #64748b; font-size: 0.875rem; margin-top: 0.25rem;">${ext.description}</div>
                        ${ext.version ? ` + "`" + `<div style="color: #10b981; font-size: 0.75rem; margin-top: 0.25rem;">版本: ${ext.version}</div>` + "`" + ` : ''}
                    </div>
                    <div style="display: flex; align-items: center; gap: 0.5rem;">
                        <span class="status ${ext.status}" style="padding: 0.25rem 0.75rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 500;">
                            ${getExtensionStatusText(ext.status)}
                        </span>
                        ${ext.status === 'enabled' && !ext.required ?
                            ` + "`" + `<button class="btn btn-danger" onclick="togglePHPExtension('${ext.name}', 'disable')" style="padding: 0.25rem 0.75rem; font-size: 0.75rem;">禁用</button>` + "`" + ` :
                            ext.status === 'disabled' || ext.status === 'not_installed' ?
                            ` + "`" + `<button class="btn btn-primary" onclick="togglePHPExtension('${ext.name}', 'enable')" style="padding: 0.25rem 0.75rem; font-size: 0.75rem;">启用</button>` + "`" + ` :
                            ''
                        }
                    </div>
                </div>
            ` + "`" + `).join('');
        }

        function getExtensionStatusText(status) {
            const statusMap = {
                'enabled': '已启用',
                'disabled': '已禁用',
                'not_installed': '未安装'
            };
            return statusMap[status] || status;
        }

        function togglePHPExtension(name, action) {
            const actionText = action === 'enable' ? '启用' : '禁用';

            if (!confirm(` + "`" + `确定要${actionText}扩展 ${name} 吗？` + "`" + `)) {
                return;
            }

            fetch(` + "`" + `/api/php/extensions/${action}` + "`" + `, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ name: name })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert(data.data.message);
                    loadPHPExtensions(); // 重新加载列表
                } else {
                    alert(` + "`" + `${actionText}失败: ` + "`" + ` + data.error);
                }
            })
            .catch(error => {
                console.error('Error:', error);
                alert(` + "`" + `${actionText}请求失败` + "`" + `);
            });
        }

        // 页面加载时获取环境列表
        loadEnvironments();
    </script>
</body>
</html>`

	t, _ := template.New("environment").Parse(tmpl)
	data := map[string]string{
		"Username": username,
	}
	t.Execute(w, data)
}

// EnvironmentStatus 获取环境状态
func (h *Handlers) EnvironmentStatus(w http.ResponseWriter, r *http.Request) {
	environments := h.environment.GetAvailableEnvironments()
	h.jsonSuccess(w, environments)
}

// InstallEnvironment 安装环境
func (h *Handlers) InstallEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
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
		if err := h.environment.InstallEnvironment(req.Name, req.Version, progressChan); err != nil {
			progressChan <- environment.InstallProgress{
				Environment: req.Name,
				Progress:    0,
				Message:     fmt.Sprintf("安装失败: %v", err),
				Status:      "error",
			}
		}
	}()

	h.jsonSuccess(w, map[string]string{
		"message": fmt.Sprintf("开始安装 %s %s", req.Name, req.Version),
		"status":  "started",
	})
}

// UninstallEnvironment 卸载环境
func (h *Handlers) UninstallEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if err := h.environment.UninstallEnvironment(req.Name); err != nil {
		h.jsonError(w, fmt.Sprintf("卸载失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{
		"message": "卸载成功",
	})
}

// UpgradeEnvironment 升级环境
func (h *Handlers) UpgradeEnvironment(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
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
		if err := h.environment.UpgradeEnvironment(req.Name, req.Version, progressChan); err != nil {
			progressChan <- environment.InstallProgress{
				Environment: req.Name,
				Progress:    0,
				Message:     fmt.Sprintf("升级失败: %v", err),
				Status:      "error",
			}
		}
	}()

	h.jsonSuccess(w, map[string]string{
		"message": fmt.Sprintf("开始升级 %s 到 %s", req.Name, req.Version),
		"status":  "started",
	})
}

// SystemInfo 系统信息页面
func (h *Handlers) SystemInfo(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Username")

	// 简化的系统信息页面模板
	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>系统监控 - 服务器管理面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f8fafc;
            color: #334155;
        }
        .header {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 1rem 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .nav {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 0 2rem;
        }
        .nav ul {
            list-style: none;
            display: flex;
            gap: 2rem;
        }
        .nav a {
            display: block;
            padding: 1rem 0;
            text-decoration: none;
            color: #64748b;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }
        .nav a:hover, .nav a.active {
            color: #3b82f6;
            border-bottom-color: #3b82f6;
        }
        .container {
            max-width: 1200px;
            margin: 2rem auto;
            padding: 0 2rem;
        }
        .card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            margin-bottom: 1.5rem;
        }
        .btn {
            background: #3b82f6;
            color: white;
            border: none;
            padding: 0.5rem 1rem;
            border-radius: 6px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
        }
        .btn-danger {
            background: #ef4444;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 服务器管理面板</h1>
        <div>
            <span>欢迎, {{.Username}}</span>
            <form method="post" action="/logout" style="display: inline;">
                <button type="submit" class="btn btn-danger">退出</button>
            </form>
        </div>
    </div>

    <nav class="nav">
        <ul>
            <li><a href="/dashboard">仪表板</a></li>
            <li><a href="/environment">环境管理</a></li>
            <li><a href="/system" class="active">系统监控</a></li>
            <li><a href="/files">文件管理</a></li>
            <li><a href="/logs">日志查看</a></li>
            <li><a href="/settings">设置</a></li>
        </ul>
    </nav>

    <div class="container">
        <div class="card">
            <h3>系统监控</h3>
            <p>系统监控功能正在开发中...</p>
        </div>
    </div>
</body>
</html>`

	t, _ := template.New("system").Parse(tmpl)
	data := map[string]string{
		"Username": username,
	}
	t.Execute(w, data)
}

// ProcessList 获取进程列表
func (h *Handlers) ProcessList(w http.ResponseWriter, r *http.Request) {
	processes, err := h.system.GetProcessList()
	if err != nil {
		h.jsonError(w, "获取进程列表失败", http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, processes)
}

// ServiceList 获取服务列表
func (h *Handlers) ServiceList(w http.ResponseWriter, r *http.Request) {
	services, err := h.system.GetServiceList()
	if err != nil {
		h.jsonError(w, "获取服务列表失败", http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, services)
}

// FileManager 文件管理页面
func (h *Handlers) FileManager(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Username")

	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>文件管理 - 服务器管理面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f8fafc; }
        .header { background: white; border-bottom: 1px solid #e2e8f0; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { background: white; border-bottom: 1px solid #e2e8f0; padding: 0 2rem; }
        .nav ul { list-style: none; display: flex; gap: 2rem; }
        .nav a { display: block; padding: 1rem 0; text-decoration: none; color: #64748b; }
        .nav a.active { color: #3b82f6; border-bottom: 2px solid #3b82f6; }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .card { background: white; border-radius: 8px; padding: 1.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .btn { background: #3b82f6; color: white; border: none; padding: 0.5rem 1rem; border-radius: 6px; cursor: pointer; text-decoration: none; display: inline-block; }
        .btn-danger { background: #ef4444; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 服务器管理面板</h1>
        <div>
            <span>欢迎, {{.Username}}</span>
            <form method="post" action="/logout" style="display: inline;">
                <button type="submit" class="btn btn-danger">退出</button>
            </form>
        </div>
    </div>

    <nav class="nav">
        <ul>
            <li><a href="/dashboard">仪表板</a></li>
            <li><a href="/environment">环境管理</a></li>
            <li><a href="/system">系统监控</a></li>
            <li><a href="/files" class="active">文件管理</a></li>
            <li><a href="/logs">日志查看</a></li>
            <li><a href="/settings">设置</a></li>
        </ul>
    </nav>

    <div class="container">
        <div class="card">
            <h3>文件管理</h3>
            <p>文件管理功能正在开发中...</p>
        </div>
    </div>
</body>
</html>`

	t, _ := template.New("files").Parse(tmpl)
	data := map[string]string{
		"Username": username,
	}
	t.Execute(w, data)
}

// 其他简化的处理器方法
func (h *Handlers) FileList(w http.ResponseWriter, r *http.Request) {
	h.jsonSuccess(w, []string{})
}

func (h *Handlers) FileUpload(w http.ResponseWriter, r *http.Request) {
	h.jsonSuccess(w, map[string]string{"message": "上传功能开发中"})
}

func (h *Handlers) FileDelete(w http.ResponseWriter, r *http.Request) {
	h.jsonSuccess(w, map[string]string{"message": "删除功能开发中"})
}

func (h *Handlers) LogViewer(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Username")

	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>日志查看 - 服务器管理面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f8fafc; }
        .header { background: white; border-bottom: 1px solid #e2e8f0; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { background: white; border-bottom: 1px solid #e2e8f0; padding: 0 2rem; }
        .nav ul { list-style: none; display: flex; gap: 2rem; }
        .nav a { display: block; padding: 1rem 0; text-decoration: none; color: #64748b; }
        .nav a.active { color: #3b82f6; border-bottom: 2px solid #3b82f6; }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .card { background: white; border-radius: 8px; padding: 1.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); }
        .btn { background: #3b82f6; color: white; border: none; padding: 0.5rem 1rem; border-radius: 6px; cursor: pointer; text-decoration: none; display: inline-block; }
        .btn-danger { background: #ef4444; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 服务器管理面板</h1>
        <div>
            <span>欢迎, {{.Username}}</span>
            <form method="post" action="/logout" style="display: inline;">
                <button type="submit" class="btn btn-danger">退出</button>
            </form>
        </div>
    </div>

    <nav class="nav">
        <ul>
            <li><a href="/dashboard">仪表板</a></li>
            <li><a href="/environment">环境管理</a></li>
            <li><a href="/system">系统监控</a></li>
            <li><a href="/files">文件管理</a></li>
            <li><a href="/logs" class="active">日志查看</a></li>
            <li><a href="/settings">设置</a></li>
        </ul>
    </nav>

    <div class="container">
        <div class="card">
            <h3>日志查看</h3>
            <p>日志查看功能正在开发中...</p>
        </div>
    </div>
</body>
</html>`

	t, _ := template.New("logs").Parse(tmpl)
	data := map[string]string{
		"Username": username,
	}
	t.Execute(w, data)
}

func (h *Handlers) ReadLogs(w http.ResponseWriter, r *http.Request) {
	h.jsonSuccess(w, []string{})
}

func (h *Handlers) Settings(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Username")

	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>设置 - 服务器管理面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f8fafc; }
        .header { background: white; border-bottom: 1px solid #e2e8f0; padding: 1rem 2rem; display: flex; justify-content: space-between; align-items: center; }
        .nav { background: white; border-bottom: 1px solid #e2e8f0; padding: 0 2rem; }
        .nav ul { list-style: none; display: flex; gap: 2rem; }
        .nav a { display: block; padding: 1rem 0; text-decoration: none; color: #64748b; }
        .nav a.active { color: #3b82f6; border-bottom: 2px solid #3b82f6; }
        .container { max-width: 1200px; margin: 2rem auto; padding: 0 2rem; }
        .card { background: white; border-radius: 8px; padding: 1.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.1); margin-bottom: 1.5rem; }
        .btn { border: none; padding: 0.5rem 1rem; border-radius: 6px; cursor: pointer; text-decoration: none; display: inline-block; margin-right: 0.5rem; }
        .btn-primary { background: #3b82f6; color: white; }
        .btn-secondary { background: #6b7280; color: white; }
        .btn-warning { background: #f59e0b; color: white; }
        .btn-danger { background: #ef4444; color: white; }
        .btn-success { background: #10b981; color: white; }
        .cert-status { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 6px; padding: 1rem; margin-bottom: 1.5rem; }
        .cert-valid { border-color: #10b981; background: #f0fdf4; }
        .cert-warning { border-color: #f59e0b; background: #fffbeb; }
        .cert-invalid { border-color: #ef4444; background: #fef2f2; }
        .action-group { display: flex; gap: 0.5rem; flex-wrap: wrap; margin-top: 1rem; }
        .form-section { background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 6px; padding: 1.5rem; margin-top: 1.5rem; }
        .form-group { margin-bottom: 1rem; }
        .form-group label { display: block; margin-bottom: 0.5rem; font-weight: 500; }
        .form-group input { width: 100%; padding: 0.5rem; border: 1px solid #d1d5db; border-radius: 4px; box-sizing: border-box; }
        .form-group small { color: #6b7280; font-size: 0.875rem; }
        .status-badge { display: inline-block; padding: 0.25rem 0.5rem; border-radius: 4px; font-size: 0.875rem; font-weight: 500; }
        .status-valid { background: #d1fae5; color: #065f46; }
        .status-warning { background: #fef3c7; color: #92400e; }
        .status-invalid { background: #fee2e2; color: #991b1b; }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 服务器管理面板</h1>
        <div>
            <span>欢迎, {{.Username}}</span>
            <form method="post" action="/logout" style="display: inline;">
                <button type="submit" class="btn btn-danger">退出</button>
            </form>
        </div>
    </div>

    <nav class="nav">
        <ul>
            <li><a href="/dashboard">仪表板</a></li>
            <li><a href="/environment">环境管理</a></li>
            <li><a href="/system">系统监控</a></li>
            <li><a href="/files">文件管理</a></li>
            <li><a href="/logs">日志查看</a></li>
            <li><a href="/settings" class="active">设置</a></li>
        </ul>
    </nav>

    <div class="container">
        <div class="card">
            <h3>🔒 SSL证书管理</h3>
            <div id="cert-status" class="cert-status">
                <p>正在加载证书信息...</p>
            </div>

            <div class="cert-actions">
                <h4>证书操作</h4>
                <div class="action-group">
                    <button onclick="generateSelfSigned()" class="btn btn-secondary">生成自签名证书</button>
                    <button onclick="showLetsEncryptForm()" class="btn btn-primary">申请Let's Encrypt证书</button>
                    <button onclick="renewCertificate()" class="btn btn-warning">续期证书</button>
                    <button onclick="deleteCertificate()" class="btn btn-danger">删除证书</button>
                </div>
            </div>

            <div id="letsencrypt-form" class="form-section" style="display: none;">
                <h4>申请Let's Encrypt证书</h4>
                <form onsubmit="requestLetsEncrypt(event)">
                    <div class="form-group">
                        <label for="domain">域名:</label>
                        <input type="text" id="domain" name="domain" required placeholder="example.com">
                        <small>请确保域名已正确解析到此服务器</small>
                    </div>
                    <div class="form-group">
                        <label for="email">邮箱:</label>
                        <input type="email" id="email" name="email" required placeholder="admin@example.com">
                        <small>用于接收证书到期提醒</small>
                    </div>
                    <div class="form-group">
                        <button type="submit" class="btn btn-primary">申请证书</button>
                        <button type="button" onclick="hideLetsEncryptForm()" class="btn btn-secondary">取消</button>
                    </div>
                </form>
            </div>

            <div id="self-signed-form" class="form-section" style="display: none;">
                <h4>生成自签名证书</h4>
                <form onsubmit="generateSelfSignedCert(event)">
                    <div class="form-group">
                        <label for="self-domain">域名/IP:</label>
                        <input type="text" id="self-domain" name="domain" placeholder="localhost (可选)">
                        <small>留空将使用localhost和服务器IP</small>
                    </div>
                    <div class="form-group">
                        <button type="submit" class="btn btn-primary">生成证书</button>
                        <button type="button" onclick="hideSelfSignedForm()" class="btn btn-secondary">取消</button>
                    </div>
                </form>
            </div>
        </div>

        <div class="card">
            <h3>⚙️ 系统设置</h3>
            <p>其他系统设置功能开发中...</p>
        </div>
    </div>

    <script>
        // 加载证书信息
        function loadCertInfo() {
            fetch('/api/ssl/status')
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        displayCertInfo(data.data);
                    } else {
                        document.getElementById('cert-status').innerHTML =
                            '<p style="color: #ef4444;">加载证书信息失败: ' + data.message + '</p>';
                    }
                })
                .catch(error => {
                    document.getElementById('cert-status').innerHTML =
                        '<p style="color: #ef4444;">加载证书信息失败: ' + error.message + '</p>';
                });
        }

        // 显示证书信息
        function displayCertInfo(cert) {
            const statusDiv = document.getElementById('cert-status');
            let statusClass = 'cert-invalid';
            let statusBadge = 'status-invalid';
            let statusText = '无效';

            if (cert.type === 'none') {
                statusDiv.innerHTML = '<p>未安装SSL证书</p>';
                return;
            }

            if (cert.is_valid) {
                if (cert.days_left > 30) {
                    statusClass = 'cert-valid';
                    statusBadge = 'status-valid';
                    statusText = '有效';
                } else {
                    statusClass = 'cert-warning';
                    statusBadge = 'status-warning';
                    statusText = '即将过期';
                }
            }

            statusDiv.className = 'cert-status ' + statusClass;
            statusDiv.innerHTML =
                '<div style="display: flex; justify-content: space-between; align-items: center;">' +
                    '<div>' +
                        '<h4>证书状态: <span class="status-badge ' + statusBadge + '">' + statusText + '</span></h4>' +
                        '<p><strong>类型:</strong> ' + (cert.type === 'letsencrypt' ? "Let's Encrypt" : '自签名') + '</p>' +
                        '<p><strong>域名:</strong> ' + cert.domain + '</p>' +
                        '<p><strong>过期时间:</strong> ' + new Date(cert.expiry_date).toLocaleString() + '</p>' +
                        '<p><strong>剩余天数:</strong> ' + cert.days_left + ' 天</p>' +
                    '</div>' +
                '</div>';
        }

        // 显示Let's Encrypt申请表单
        function showLetsEncryptForm() {
            document.getElementById('letsencrypt-form').style.display = 'block';
            document.getElementById('self-signed-form').style.display = 'none';
        }

        // 隐藏Let's Encrypt申请表单
        function hideLetsEncryptForm() {
            document.getElementById('letsencrypt-form').style.display = 'none';
        }

        // 显示自签名证书表单
        function generateSelfSigned() {
            document.getElementById('self-signed-form').style.display = 'block';
            document.getElementById('letsencrypt-form').style.display = 'none';
        }

        // 隐藏自签名证书表单
        function hideSelfSignedForm() {
            document.getElementById('self-signed-form').style.display = 'none';
        }

        // 申请Let's Encrypt证书
        function requestLetsEncrypt(event) {
            event.preventDefault();
            const domain = document.getElementById('domain').value;
            const email = document.getElementById('email').value;

            if (!domain || !email) {
                alert('请填写所有必填字段');
                return;
            }

            const button = event.target.querySelector('button[type="submit"]');
            button.disabled = true;
            button.textContent = '申请中...';

            fetch('/api/ssl/letsencrypt', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ domain, email })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('Let\\'s Encrypt证书申请成功！');
                    hideLetsEncryptForm();
                    loadCertInfo();
                } else {
                    alert('申请失败: ' + data.message);
                }
            })
            .catch(error => {
                alert('申请失败: ' + error.message);
            })
            .finally(() => {
                button.disabled = false;
                button.textContent = '申请证书';
            });
        }

        // 生成自签名证书
        function generateSelfSignedCert(event) {
            event.preventDefault();
            const domain = document.getElementById('self-domain').value || 'localhost';

            const button = event.target.querySelector('button[type="submit"]');
            button.disabled = true;
            button.textContent = '生成中...';

            fetch('/api/ssl/self-signed', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ domain })
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    alert('自签名证书生成成功！');
                    hideSelfSignedForm();
                    loadCertInfo();
                } else {
                    alert('生成失败: ' + data.message);
                }
            })
            .catch(error => {
                alert('生成失败: ' + error.message);
            })
            .finally(() => {
                button.disabled = false;
                button.textContent = '生成证书';
            });
        }

        // 续期证书
        function renewCertificate() {
            if (!confirm('确定要续期证书吗？')) return;

            fetch('/api/ssl/renew', { method: 'POST' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert('证书续期成功！');
                        loadCertInfo();
                    } else {
                        alert('续期失败: ' + data.message);
                    }
                })
                .catch(error => {
                    alert('续期失败: ' + error.message);
                });
        }

        // 删除证书
        function deleteCertificate() {
            if (!confirm('确定要删除当前证书吗？删除后将无法使用HTTPS访问。')) return;

            fetch('/api/ssl/delete', { method: 'DELETE' })
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        alert('证书删除成功！');
                        loadCertInfo();
                    } else {
                        alert('删除失败: ' + data.message);
                    }
                })
                .catch(error => {
                    alert('删除失败: ' + error.message);
                });
        }

        // 页面加载时获取证书信息
        loadCertInfo();
    </script>
</body>
</html>`

	t, _ := template.New("settings").Parse(tmpl)
	data := map[string]string{
		"Username": username,
	}
	t.Execute(w, data)
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		h.auth.Logout(cookie.Value)
	}

	// 清除cookie
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/login", http.StatusFound)
}

// PHPExtensions 获取PHP扩展列表
func (h *Handlers) PHPExtensions(w http.ResponseWriter, r *http.Request) {
	extensions, err := h.environment.GetPHPExtensions()
	if err != nil {
		h.jsonError(w, "获取PHP扩展列表失败", http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, extensions)
}

// EnablePHPExtension 启用PHP扩展
func (h *Handlers) EnablePHPExtension(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if err := h.environment.EnablePHPExtension(req.Name); err != nil {
		h.jsonError(w, fmt.Sprintf("启用扩展失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{
		"message": fmt.Sprintf("扩展 %s 已启用", req.Name),
	})
}

// DisablePHPExtension 禁用PHP扩展
func (h *Handlers) DisablePHPExtension(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if err := h.environment.DisablePHPExtension(req.Name); err != nil {
		h.jsonError(w, fmt.Sprintf("禁用扩展失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{
		"message": fmt.Sprintf("扩展 %s 已禁用", req.Name),
	})
}

// Projects 项目管理页面
func (h *Handlers) Projects(w http.ResponseWriter, r *http.Request) {
	username := r.Header.Get("X-Username")

	tmpl := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>项目管理 - 服务器管理面板</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: #f8fafc;
            color: #334155;
        }
        .header {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 1rem 2rem;
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        .header h1 {
            color: #1e293b;
            font-size: 1.5rem;
        }
        .user-info {
            display: flex;
            align-items: center;
            gap: 1rem;
        }
        .nav {
            background: white;
            border-bottom: 1px solid #e2e8f0;
            padding: 0 2rem;
        }
        .nav ul {
            list-style: none;
            display: flex;
            gap: 2rem;
        }
        .nav a {
            display: block;
            padding: 1rem 0;
            text-decoration: none;
            color: #64748b;
            border-bottom: 2px solid transparent;
            transition: all 0.2s;
        }
        .nav a:hover, .nav a.active {
            color: #3b82f6;
            border-bottom-color: #3b82f6;
        }
        .container {
            max-width: 1200px;
            margin: 2rem auto;
            padding: 0 2rem;
        }
        .projects-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2rem;
        }
        .btn {
            background: #3b82f6;
            color: white;
            border: none;
            padding: 0.75rem 1.5rem;
            border-radius: 6px;
            cursor: pointer;
            text-decoration: none;
            display: inline-block;
            font-weight: 500;
            transition: background 0.2s;
        }
        .btn:hover {
            background: #2563eb;
        }
        .btn-danger {
            background: #ef4444;
        }
        .btn-danger:hover {
            background: #dc2626;
        }
        .projects-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
            gap: 1.5rem;
            margin-bottom: 2rem;
        }
        .project-card {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            border: 1px solid #e2e8f0;
        }
        .project-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 1rem;
        }
        .project-title {
            font-size: 1.25rem;
            font-weight: 600;
            color: #1e293b;
        }
        .project-status {
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.875rem;
            font-weight: 500;
        }
        .status-active {
            background: #dcfce7;
            color: #166534;
        }
        .status-inactive {
            background: #fef2f2;
            color: #991b1b;
        }
        .project-info {
            margin-bottom: 1rem;
        }
        .project-info div {
            margin-bottom: 0.5rem;
            font-size: 0.875rem;
            color: #64748b;
        }
        .project-actions {
            display: flex;
            gap: 0.5rem;
            flex-wrap: wrap;
        }
        .btn-sm {
            padding: 0.5rem 1rem;
            font-size: 0.875rem;
        }
        .databases-section {
            background: white;
            border-radius: 8px;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            margin-top: 2rem;
        }
        .databases-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
            gap: 1rem;
            margin-top: 1rem;
        }
        .database-card {
            border: 1px solid #e2e8f0;
            border-radius: 6px;
            padding: 1rem;
        }
        .database-name {
            font-weight: 600;
            color: #1e293b;
            margin-bottom: 0.5rem;
        }
        .database-info {
            font-size: 0.875rem;
            color: #64748b;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>🚀 服务器管理面板</h1>
        <div class="user-info">
            <span>欢迎, {{.Username}}</span>
            <form method="post" action="/logout" style="display: inline;">
                <button type="submit" class="btn btn-danger">退出</button>
            </form>
        </div>
    </div>

    <nav class="nav">
        <ul>
            <li><a href="/dashboard">仪表板</a></li>
            <li><a href="/environment">环境管理</a></li>
            <li><a href="/projects" class="active">项目管理</a></li>
        </ul>
    </nav>

    <div class="container">
        <div class="projects-header">
            <h2>项目管理</h2>
            <div>
                <button class="btn" onclick="scanProjects()">🔍 扫描项目</button>
                <button class="btn" onclick="createProject()">➕ 创建项目</button>
            </div>
        </div>

        <div class="projects-grid" id="projectsGrid">
            <div style="text-align: center; padding: 2rem; color: #64748b;">
                点击"扫描项目"来发现现有项目
            </div>
        </div>

        <div class="databases-section">
            <h3>数据库管理</h3>
            <div class="databases-grid" id="databasesGrid">
                加载中...
            </div>
        </div>
    </div>

    <script>
        let projects = [];
        let databases = [];

        function scanProjects() {
            fetch('/api/projects/scan')
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        projects = data.data;
                        renderProjects();
                    } else {
                        alert('扫描失败: ' + data.error);
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    alert('扫描请求失败');
                });
        }

        function loadDatabases() {
            fetch('/api/projects/databases')
                .then(response => response.json())
                .then(data => {
                    if (data.success) {
                        databases = data.data;
                        renderDatabases();
                    } else {
                        document.getElementById('databasesGrid').innerHTML = '<p>加载失败: ' + data.error + '</p>';
                    }
                })
                .catch(error => {
                    console.error('Error:', error);
                    document.getElementById('databasesGrid').innerHTML = '<p>加载失败</p>';
                });
        }

        function renderProjects() {
            const grid = document.getElementById('projectsGrid');
            if (projects.length === 0) {
                grid.innerHTML = '<div style="text-align: center; padding: 2rem; color: #64748b;">未发现任何项目</div>';
                return;
            }

            grid.innerHTML = projects.map(project => ` + "`" + `
                <div class="project-card">
                    <div class="project-header">
                        <div class="project-title">${project.name}</div>
                        <div class="project-status status-${project.status}">${getStatusText(project.status)}</div>
                    </div>
                    <div class="project-info">
                        <div><strong>类型:</strong> ${getTypeText(project.type)}</div>
                        <div><strong>路径:</strong> ${project.path}</div>
                        ${project.domain ? ` + "`" + `<div><strong>域名:</strong> ${project.domain}</div>` + "`" + ` : ''}
                        ${project.database ? ` + "`" + `<div><strong>数据库:</strong> ${project.database}</div>` + "`" + ` : ''}
                        ${project.php_version ? ` + "`" + `<div><strong>PHP版本:</strong> ${project.php_version}</div>` + "`" + ` : ''}
                        <div><strong>大小:</strong> ${formatBytes(project.size)}</div>
                        <div><strong>更新时间:</strong> ${formatDate(project.updated_at)}</div>
                    </div>
                    <div class="project-actions">
                        <button class="btn btn-sm" onclick="openProject('${project.path}')">📁 打开</button>
                        <button class="btn btn-sm" onclick="editProject('${project.id}')">✏️ 编辑</button>
                        ${project.status === 'inactive' ?
                            ` + "`" + `<button class="btn btn-sm" onclick="activateProject('${project.id}')">🚀 激活</button>` + "`" + ` :
                            ` + "`" + `<button class="btn btn-sm btn-danger" onclick="deactivateProject('${project.id}')">⏸️ 停用</button>` + "`" + `
                        }
                    </div>
                </div>
            ` + "`" + `).join('');
        }

        function renderDatabases() {
            const grid = document.getElementById('databasesGrid');
            if (databases.length === 0) {
                grid.innerHTML = '<div style="text-align: center; padding: 1rem; color: #64748b;">未发现数据库</div>';
                return;
            }

            grid.innerHTML = databases.map(db => ` + "`" + `
                <div class="database-card">
                    <div class="database-name">${db.name}</div>
                    <div class="database-info">
                        <div>类型: ${db.type}</div>
                        <div>大小: ${db.size}</div>
                        <div>表数量: ${db.tables}</div>
                        <div>字符集: ${db.charset}</div>
                    </div>
                </div>
            ` + "`" + `).join('');
        }

        function getStatusText(status) {
            const statusMap = {
                'active': '运行中',
                'inactive': '未激活',
                'error': '错误'
            };
            return statusMap[status] || status;
        }

        function getTypeText(type) {
            const typeMap = {
                'wordpress': 'WordPress',
                'laravel': 'Laravel',
                'nodejs': 'Node.js',
                'php': 'PHP',
                'static': '静态网站',
                'unknown': '未知'
            };
            return typeMap[type] || type;
        }

        function formatBytes(bytes) {
            if (bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'KB', 'MB', 'GB'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
        }

        function formatDate(dateStr) {
            const date = new Date(dateStr);
            return date.toLocaleDateString('zh-CN') + ' ' + date.toLocaleTimeString('zh-CN');
        }

        function createProject() {
            alert('创建项目功能开发中...');
        }

        function openProject(path) {
            alert('打开项目: ' + path);
        }

        function editProject(id) {
            alert('编辑项目: ' + id);
        }

        function activateProject(id) {
            alert('激活项目: ' + id);
        }

        function deactivateProject(id) {
            alert('停用项目: ' + id);
        }

        // 页面加载时获取数据库列表
        loadDatabases();
    </script>
</body>
</html>`

	t, _ := template.New("projects").Parse(tmpl)
	data := map[string]string{
		"Username": username,
	}
	t.Execute(w, data)
}

// ScanProjects 扫描项目
func (h *Handlers) ScanProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := h.projects.ScanProjects()
	if err != nil {
		h.jsonError(w, "扫描项目失败", http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, projects)
}

// ProjectDatabases 获取数据库列表
func (h *Handlers) ProjectDatabases(w http.ResponseWriter, r *http.Request) {
	databases, err := h.projects.GetDatabases()
	if err != nil {
		h.jsonError(w, "获取数据库列表失败", http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, databases)
}

// CreateProject 创建项目
func (h *Handlers) CreateProject(w http.ResponseWriter, r *http.Request) {
	h.jsonSuccess(w, map[string]string{
		"message": "创建项目功能开发中",
	})
}

// SSL证书管理API

// SSLStatus 获取SSL证书状态
func (h *Handlers) SSLStatus(w http.ResponseWriter, r *http.Request) {
	certInfo, err := h.ssl.GetCertInfo()
	if err != nil {
		h.jsonError(w, fmt.Sprintf("获取证书信息失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, certInfo)
}

// GenerateSelfSigned 生成自签名证书
func (h *Handlers) GenerateSelfSigned(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	domain := req.Domain
	if domain == "" {
		domain = "localhost"
	}

	if err := h.ssl.GenerateSelfSigned(domain); err != nil {
		h.jsonError(w, fmt.Sprintf("生成自签名证书失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{
		"message": "自签名证书生成成功",
	})
}

// RequestLetsEncrypt 申请Let's Encrypt证书
func (h *Handlers) RequestLetsEncrypt(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Domain string `json:"domain"`
		Email  string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.jsonError(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	if req.Domain == "" || req.Email == "" {
		h.jsonError(w, "域名和邮箱不能为空", http.StatusBadRequest)
		return
	}

	if err := h.ssl.RequestLetsEncrypt(req.Domain, req.Email); err != nil {
		h.jsonError(w, fmt.Sprintf("申请Let's Encrypt证书失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{
		"message": "Let's Encrypt证书申请成功",
	})
}

// RenewCertificate 续期证书
func (h *Handlers) RenewCertificate(w http.ResponseWriter, r *http.Request) {
	if err := h.ssl.RenewLetsEncrypt(); err != nil {
		h.jsonError(w, fmt.Sprintf("续期证书失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{
		"message": "证书续期成功",
	})
}

// DeleteCertificate 删除证书
func (h *Handlers) DeleteCertificate(w http.ResponseWriter, r *http.Request) {
	if err := h.ssl.DeleteCertificate(); err != nil {
		h.jsonError(w, fmt.Sprintf("删除证书失败: %v", err), http.StatusInternalServerError)
		return
	}

	h.jsonSuccess(w, map[string]string{
		"message": "证书删除成功",
	})
}
