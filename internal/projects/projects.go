package projects

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Manager 项目管理器
type Manager struct{}

// LegacyProject 旧的项目信息结构（保持兼容性）
type LegacyProject struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	Type        string    `json:"type"`        // website, api, static
	Status      string    `json:"status"`      // active, inactive, error
	Domain      string    `json:"domain"`
	Database    string    `json:"database"`
	PHPVersion  string    `json:"php_version"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Description string    `json:"description"`
}

// Database 数据库信息
type Database struct {
	Name        string `json:"name"`
	Type        string `json:"type"`        // mysql, mariadb, sqlite
	Size        string `json:"size"`
	Tables      int    `json:"tables"`
	Charset     string `json:"charset"`
	Project     string `json:"project"`
	ProjectPath string `json:"project_path"` // 关联的项目路径
}

// NewManager 创建项目管理器
func NewManager() *Manager {
	return &Manager{}
}

// getProjectsDir 获取项目目录路径
func (m *Manager) getProjectsDir() string {
	// 在 Linux 系统中，强制使用 /var/www
	wwwDir := "/var/www"

	// 确保目录存在，如果不存在则创建
	if _, err := os.Stat(wwwDir); os.IsNotExist(err) {
		// 尝试创建目录
		if err := os.MkdirAll(wwwDir, 0755); err != nil {
			// 如果创建失败，尝试使用 sudo 创建
			fmt.Printf("Warning: Cannot create %s directly, trying with elevated permissions...\n", wwwDir)
			// 在开发环境中，我们可以使用用户目录作为备选
			homeDir, _ := os.UserHomeDir()
			userWwwDir := filepath.Join(homeDir, "www")
			os.MkdirAll(userWwwDir, 0755)
			fmt.Printf("Using fallback directory: %s\n", userWwwDir)
			return userWwwDir
		}
	}

	return wwwDir
}

// isProductionEnvironment 检查是否为生产环境
func (m *Manager) isProductionEnvironment() bool {
	// 检查是否存在 Nginx 配置目录
	nginxDirs := []string{"/etc/nginx/sites-available", "/etc/nginx/conf.d"}
	for _, dir := range nginxDirs {
		if _, err := os.Stat(dir); err == nil {
			return true
		}
	}
	return false
}

// GetOverview 获取项目概览
func (m *Manager) GetOverview() (*ProjectOverview, error) {
	projects, err := m.GetProjects()
	if err != nil {
		return nil, err
	}

	activeCount := 0
	var totalSize int64

	for _, project := range projects {
		if project.Status == StatusActive {
			activeCount++
		}
		totalSize += project.Size
	}

	firstTime := len(projects) == 0

	return &ProjectOverview{
		Projects:       projects,
		TotalProjects:  len(projects),
		ActiveProjects: activeCount,
		TotalSize:      totalSize,
		FirstTimeSetup: firstTime,
	}, nil
}

// GetProjects 获取所有项目
func (m *Manager) GetProjects() ([]Project, error) {
	var projects []Project

	// Scan projects directory
	wwwDir := m.getProjectsDir()
	if _, err := os.Stat(wwwDir); os.IsNotExist(err) {
		return projects, nil // No projects directory
	}

	entries, err := os.ReadDir(wwwDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "html" { // Skip default apache html dir
			projectPath := filepath.Join(wwwDir, entry.Name())
			project, err := m.loadProject(entry.Name(), projectPath)
			if err != nil {
				continue // Skip invalid projects
			}
			projects = append(projects, *project)
		}
	}

	return projects, nil
}

// loadProject 加载单个项目信息
func (m *Manager) loadProject(name, path string) (*Project, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	// Calculate directory size
	size, err := m.calculateDirSize(path)
	if err != nil {
		size = 0
	}

	// Check if project has Nginx config
	nginxConfigPath := fmt.Sprintf("/etc/nginx/sites-available/%s", name)
	nginxEnabled := m.isNginxSiteEnabled(name)

	status := StatusInactive
	if nginxEnabled {
		status = StatusActive
	}

	// Generate domain from project name
	domain := fmt.Sprintf("%s.local", name)
	if nginxConfig, err := m.readNginxConfig(nginxConfigPath); err == nil {
		if serverName := m.extractServerName(nginxConfig); serverName != "" {
			domain = serverName
		}
	}

	project := &Project{
		ID:           name,
		Name:         name,
		Domain:       domain,
		Path:         path,
		Status:       status,
		CreatedAt:    info.ModTime(),
		UpdatedAt:    info.ModTime(),
		DatabaseName: name, // Assume database name matches project name
		Size:         size,
	}

	return project, nil
}

// CreateProject 创建新项目
func (m *Manager) CreateProject(req CreateProjectRequest) (*Project, error) {
	// Validate project name
	if req.Name == "" {
		return nil, fmt.Errorf("项目名称不能为空")
	}

	// Get projects directory
	wwwDir := m.getProjectsDir()

	// Create project directory
	projectPath := filepath.Join(wwwDir, req.Name)
	if _, err := os.Stat(projectPath); !os.IsNotExist(err) {
		return nil, fmt.Errorf("项目 %s 已存在", req.Name)
	}

	if err := os.MkdirAll(projectPath, 0755); err != nil {
		return nil, fmt.Errorf("创建项目目录失败: %v", err)
	}

	// Create index.html
	indexContent := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s</title>
</head>
<body>
    <h1>Welcome to %s</h1>
    <p>Your project is ready!</p>
</body>
</html>`, req.Name, req.Name)

	if err := os.WriteFile(filepath.Join(projectPath, "index.html"), []byte(indexContent), 0644); err != nil {
		return nil, fmt.Errorf("创建索引文件失败: %v", err)
	}

	// Create database if requested
	if req.CreateDB {
		if err := m.createDatabase(req.Name); err != nil {
			// Don't fail project creation if database creation fails
			fmt.Printf("Warning: Failed to create database for %s: %v\n", req.Name, err)
		}
	}

	// Create Nginx configuration (skip in development environment)
	if m.isProductionEnvironment() {
		if err := m.createNginxConfig(req.Name, req.Domain, projectPath, req.EnableSSL); err != nil {
			fmt.Printf("Warning: Failed to create Nginx config: %v\n", err)
		} else {
			// Enable the site
			if err := m.enableNginxSite(req.Name); err != nil {
				fmt.Printf("Warning: Failed to enable site: %v\n", err)
			}

			// Reload Nginx
			if err := m.reloadNginx(); err != nil {
				fmt.Printf("Warning: Failed to reload Nginx: %v\n", err)
			}
		}
	}

	// Load and return the created project
	return m.loadProject(req.Name, projectPath)
}

// DeleteProject 删除项目
func (m *Manager) DeleteProject(projectID string) error {
	// Use the same directory logic as creation
	wwwDir := m.getProjectsDir()
	projectPath := filepath.Join(wwwDir, projectID)

	// Remove project directory
	if err := os.RemoveAll(projectPath); err != nil {
		return fmt.Errorf("删除项目目录失败: %v", err)
	}

	// Remove Nginx configuration (only in production environment)
	if m.isProductionEnvironment() {
		nginxConfigPath := fmt.Sprintf("/etc/nginx/sites-available/%s", projectID)
		os.Remove(nginxConfigPath)

		nginxEnabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", projectID)
		os.Remove(nginxEnabledPath)

		// Reload Nginx
		if err := m.reloadNginx(); err != nil {
			fmt.Printf("Warning: Failed to reload Nginx: %v\n", err)
		}
	}

	// Drop database (only if database exists)
	if err := m.dropDatabase(projectID); err != nil {
		fmt.Printf("Warning: Failed to drop database for %s: %v\n", projectID, err)
	}

	return nil
}

// GetProjectFiles 获取项目文件列表
func (m *Manager) GetProjectFiles(projectID, subPath string) ([]FileInfo, error) {
	projectPath := filepath.Join("/var/www", projectID)
	targetPath := filepath.Join(projectPath, subPath)

	// Security check: ensure path is within project directory
	if !strings.HasPrefix(targetPath, projectPath) {
		return nil, fmt.Errorf("无效的路径")
	}

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := FileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(subPath, entry.Name()),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
		}

		if !entry.IsDir() {
			fileInfo.Extension = filepath.Ext(entry.Name())
		}

		files = append(files, fileInfo)
	}

	return files, nil
}

// Helper methods

// calculateDirSize 计算目录大小
func (m *Manager) calculateDirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

// isNginxSiteEnabled 检查Nginx站点是否启用
func (m *Manager) isNginxSiteEnabled(siteName string) bool {
	enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", siteName)
	_, err := os.Stat(enabledPath)
	return err == nil
}

// readNginxConfig 读取Nginx配置文件
func (m *Manager) readNginxConfig(configPath string) (string, error) {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// extractServerName 从Nginx配置中提取server_name
func (m *Manager) extractServerName(config string) string {
	lines := strings.Split(config, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "server_name") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				serverName := parts[1]
				return strings.TrimSuffix(serverName, ";")
			}
		}
	}
	return ""
}

// createDatabase 创建数据库
func (m *Manager) createDatabase(dbName string) error {
	cmd := exec.Command("mysql", "-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName))
	return cmd.Run()
}

// dropDatabase 删除数据库
func (m *Manager) dropDatabase(dbName string) error {
	cmd := exec.Command("mysql", "-e", fmt.Sprintf("DROP DATABASE IF EXISTS `%s`", dbName))
	return cmd.Run()
}

// createNginxConfig 创建Nginx配置
func (m *Manager) createNginxConfig(siteName, domain, projectPath string, enableSSL bool) error {
	if domain == "" {
		domain = fmt.Sprintf("%s.local", siteName)
	}

	config := fmt.Sprintf(`server {
    listen 80;
    server_name %s;
    root %s;
    index index.html index.htm index.php;

    location / {
        try_files $uri $uri/ =404;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/var/run/php/php8.2-fpm.sock;
    }

    location ~ /\.ht {
        deny all;
    }
}`, domain, projectPath)

	configPath := fmt.Sprintf("/etc/nginx/sites-available/%s", siteName)
	return os.WriteFile(configPath, []byte(config), 0644)
}

// enableNginxSite 启用Nginx站点
func (m *Manager) enableNginxSite(siteName string) error {
	availablePath := fmt.Sprintf("/etc/nginx/sites-available/%s", siteName)
	enabledPath := fmt.Sprintf("/etc/nginx/sites-enabled/%s", siteName)

	// Create symlink
	return os.Symlink(availablePath, enabledPath)
}

// reloadNginx 重新加载Nginx
func (m *Manager) reloadNginx() error {
	cmd := exec.Command("systemctl", "reload", "nginx")
	return cmd.Run()
}

// ScanProjects 扫描项目（使用约定的项目路径）- 保持兼容性
func (m *Manager) ScanProjects() ([]LegacyProject, error) {
	var projects []LegacyProject

	// 约定的项目安装路径
	projectsDir := m.getProjectsDir()

	// 直接扫描约定目录
	if projectList, err := m.scanDirectory(projectsDir); err == nil {
		projects = append(projects, projectList...)
	}

	return projects, nil
}

// scanDirectory 扫描目录中的项目
func (m *Manager) scanDirectory(baseDir string) ([]LegacyProject, error) {
	var projects []LegacyProject

	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return projects, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			projectPath := filepath.Join(baseDir, entry.Name())

			// 检查是否是有效的Web项目
			if m.isWebProject(projectPath) {
				project := LegacyProject{
					ID:          m.generateProjectID(projectPath),
					Name:        entry.Name(),
					Path:        projectPath,
					Type:        m.detectProjectType(projectPath),
					Status:      m.getProjectStatus(projectPath),
					Domain:      m.detectDomain(projectPath),
					Database:    m.detectDatabase(projectPath),
					PHPVersion:  m.detectPHPVersion(projectPath),
					Size:        m.getDirectorySize(projectPath),
					CreatedAt:   m.getCreationTime(projectPath),
					UpdatedAt:   m.getModificationTime(projectPath),
					Description: m.generateDescription(projectPath),
				}
				projects = append(projects, project)
			}
		}
	}
	
	return projects, nil
}

// isWebProject 检查是否是Web项目
func (m *Manager) isWebProject(path string) bool {
	// 检查常见的Web项目文件
	indicators := []string{
		"index.html", "index.php", "index.htm",
		"composer.json", "package.json",
		"wp-config.php", // WordPress
		"config.php",    // 通用配置文件
		".htaccess",     // Apache配置
	}
	
	for _, indicator := range indicators {
		if _, err := os.Stat(filepath.Join(path, indicator)); err == nil {
			return true
		}
	}
	
	// 检查是否有子目录包含Web文件
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	
	webFileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if ext == ".php" || ext == ".html" || ext == ".htm" || ext == ".js" || ext == ".css" {
				webFileCount++
			}
		}
	}
	
	return webFileCount >= 3 // 至少有3个Web相关文件
}

// detectProjectType 检测项目类型
func (m *Manager) detectProjectType(path string) string {
	// WordPress
	if _, err := os.Stat(filepath.Join(path, "wp-config.php")); err == nil {
		return "wordpress"
	}
	
	// Laravel
	if _, err := os.Stat(filepath.Join(path, "artisan")); err == nil {
		return "laravel"
	}
	
	// Node.js
	if _, err := os.Stat(filepath.Join(path, "package.json")); err == nil {
		return "nodejs"
	}
	
	// PHP项目
	if m.hasFileWithExtension(path, ".php") {
		return "php"
	}
	
	// 静态网站
	if m.hasFileWithExtension(path, ".html") {
		return "static"
	}
	
	return "unknown"
}

// hasFileWithExtension 检查目录是否包含指定扩展名的文件
func (m *Manager) hasFileWithExtension(path, ext string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	
	for _, entry := range entries {
		if !entry.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ext {
			return true
		}
	}
	
	return false
}

// getProjectStatus 获取项目状态
func (m *Manager) getProjectStatus(path string) string {
	// 检查Nginx配置中是否有此项目
	if m.isProjectInNginxConfig(path) {
		return "active"
	}
	
	return "inactive"
}

// isProjectInNginxConfig 检查项目是否在Nginx配置中
func (m *Manager) isProjectInNginxConfig(path string) bool {
	// 检查Nginx配置文件
	configDirs := []string{
		"/etc/nginx/sites-enabled",
		"/etc/nginx/conf.d",
	}
	
	for _, configDir := range configDirs {
		if entries, err := os.ReadDir(configDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					configFile := filepath.Join(configDir, entry.Name())
					if content, err := os.ReadFile(configFile); err == nil {
						if strings.Contains(string(content), path) {
							return true
						}
					}
				}
			}
		}
	}
	
	return false
}

// detectDomain 检测域名
func (m *Manager) detectDomain(path string) string {
	// 从Nginx配置中提取域名
	configDirs := []string{
		"/etc/nginx/sites-enabled",
		"/etc/nginx/conf.d",
	}
	
	for _, configDir := range configDirs {
		if entries, err := os.ReadDir(configDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					configFile := filepath.Join(configDir, entry.Name())
					if content, err := os.ReadFile(configFile); err == nil {
						if strings.Contains(string(content), path) {
							// 提取server_name
							lines := strings.Split(string(content), "\n")
							for _, line := range lines {
								line = strings.TrimSpace(line)
								if strings.HasPrefix(line, "server_name") {
									parts := strings.Fields(line)
									if len(parts) >= 2 {
										domain := strings.TrimSuffix(parts[1], ";")
										if domain != "_" && domain != "localhost" {
											return domain
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	
	return ""
}

// detectDatabase 检测关联的数据库
func (m *Manager) detectDatabase(path string) string {
	// 检查WordPress配置
	if configFile := filepath.Join(path, "wp-config.php"); m.fileExists(configFile) {
		if content, err := os.ReadFile(configFile); err == nil {
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				if strings.Contains(line, "DB_NAME") {
					// 提取数据库名
					if start := strings.Index(line, "'"); start != -1 {
						if end := strings.Index(line[start+1:], "'"); end != -1 {
							return line[start+1 : start+1+end]
						}
					}
				}
			}
		}
	}
	
	// 检查其他配置文件
	configFiles := []string{"config.php", ".env", "database.php"}
	for _, configFile := range configFiles {
		fullPath := filepath.Join(path, configFile)
		if m.fileExists(fullPath) {
			if dbName := m.extractDatabaseFromConfig(fullPath); dbName != "" {
				return dbName
			}
		}
	}
	
	return ""
}

// extractDatabaseFromConfig 从配置文件中提取数据库名
func (m *Manager) extractDatabaseFromConfig(configFile string) string {
	content, err := os.ReadFile(configFile)
	if err != nil {
		return ""
	}
	
	contentStr := string(content)
	
	// 常见的数据库配置模式
	patterns := []string{
		"database", "db_name", "dbname", "DB_NAME",
	}
	
	for _, pattern := range patterns {
		if strings.Contains(strings.ToLower(contentStr), strings.ToLower(pattern)) {
			// 简化的提取逻辑
			lines := strings.Split(contentStr, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToLower(line), strings.ToLower(pattern)) {
					// 尝试提取值
					if start := strings.Index(line, "'"); start != -1 {
						if end := strings.Index(line[start+1:], "'"); end != -1 {
							value := line[start+1 : start+1+end]
							if value != "" && value != "localhost" && value != "127.0.0.1" {
								return value
							}
						}
					}
				}
			}
		}
	}
	
	return ""
}

// detectPHPVersion 检测PHP版本
func (m *Manager) detectPHPVersion(path string) string {
	// 检查composer.json中的PHP版本要求
	composerFile := filepath.Join(path, "composer.json")
	if m.fileExists(composerFile) {
		// 简化的版本检测
		return "8.2" // 默认返回当前系统PHP版本
	}
	
	return ""
}

// getDirectorySize 获取目录大小
func (m *Manager) getDirectorySize(path string) int64 {
	var size int64
	
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	
	if err != nil {
		return 0
	}
	
	return size
}

// getCreationTime 获取创建时间
func (m *Manager) getCreationTime(path string) time.Time {
	if info, err := os.Stat(path); err == nil {
		return info.ModTime() // Unix系统通常没有创建时间，使用修改时间
	}
	return time.Now()
}

// getModificationTime 获取修改时间
func (m *Manager) getModificationTime(path string) time.Time {
	if info, err := os.Stat(path); err == nil {
		return info.ModTime()
	}
	return time.Now()
}

// generateDescription 生成项目描述
func (m *Manager) generateDescription(path string) string {
	projectType := m.detectProjectType(path)
	
	switch projectType {
	case "wordpress":
		return "WordPress网站"
	case "laravel":
		return "Laravel PHP框架项目"
	case "nodejs":
		return "Node.js应用"
	case "php":
		return "PHP网站项目"
	case "static":
		return "静态网站"
	default:
		return "Web项目"
	}
}

// generateProjectID 生成项目ID
func (m *Manager) generateProjectID(path string) string {
	// 使用路径的哈希作为ID
	return fmt.Sprintf("proj_%x", strings.Replace(path, "/", "_", -1))
}

// fileExists 检查文件是否存在
func (m *Manager) fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// GetDatabases 获取数据库列表
func (m *Manager) GetDatabases() ([]Database, error) {
	var databases []Database
	
	// 连接MySQL/MariaDB获取数据库列表
	cmd := exec.Command("mysql", "-e", "SHOW DATABASES;")
	output, err := cmd.Output()
	if err != nil {
		return databases, err
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] { // 跳过标题行
		line = strings.TrimSpace(line)
		if line != "" && line != "information_schema" && line != "performance_schema" && line != "mysql" && line != "sys" {
			database := Database{
				Name:    line,
				Type:    "mariadb", // 根据实际情况调整
				Size:    m.getDatabaseSize(line),
				Tables:  m.getDatabaseTables(line),
				Charset: "utf8mb4",
			}
			databases = append(databases, database)
		}
	}
	
	return databases, nil
}

// getDatabaseSize 获取数据库大小
func (m *Manager) getDatabaseSize(dbName string) string {
	cmd := exec.Command("mysql", "-e", fmt.Sprintf("SELECT ROUND(SUM(data_length + index_length) / 1024 / 1024, 1) AS 'DB Size in MB' FROM information_schema.tables WHERE table_schema='%s';", dbName))
	output, err := cmd.Output()
	if err != nil {
		return "未知"
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		size := strings.TrimSpace(lines[1])
		if size != "" && size != "NULL" {
			return size + " MB"
		}
	}
	
	return "< 1 MB"
}

// getDatabaseTables 获取数据库表数量
func (m *Manager) getDatabaseTables(dbName string) int {
	cmd := exec.Command("mysql", "-e", fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='%s';", dbName))
	output, err := cmd.Output()
	if err != nil {
		return 0
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		countStr := strings.TrimSpace(lines[1])
		if count, err := strconv.Atoi(countStr); err == nil {
			return count
		}
	}
	
	return 0
}
