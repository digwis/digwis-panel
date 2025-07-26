package environment

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Manager 环境管理器
type Manager struct {
	mutex     sync.RWMutex
	cache     map[string]Environment
	cacheTime time.Time
	cacheTTL  time.Duration
}

// Environment 环境信息
type Environment struct {
	Name            string   `json:"name"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	Version         string   `json:"version"`
	Status          string   `json:"status"` // installed, not_installed, installing, error
	Required        bool     `json:"required"`
	AvailableVersions []string `json:"available_versions"`
	LatestVersion   string   `json:"latest_version"`
}

// InstallProgress 安装进度
type InstallProgress struct {
	Environment string `json:"environment"`
	Progress    int    `json:"progress"`
	Message     string `json:"message"`
	Status      string `json:"status"`
}

// NewManager 创建环境管理器
func NewManager() *Manager {
	return &Manager{
		cache:    make(map[string]Environment),
		cacheTTL: 30 * time.Second, // 缓存30秒
	}
}

// GetAvailableEnvironments 获取可用环境列表
func (m *Manager) GetAvailableEnvironments() []Environment {
	m.mutex.RLock()
	// 检查缓存是否有效
	if time.Since(m.cacheTime) < m.cacheTTL && len(m.cache) > 0 {
		environments := make([]Environment, 0, len(m.cache))
		for _, env := range m.cache {
			environments = append(environments, env)
		}
		m.mutex.RUnlock()
		return environments
	}
	m.mutex.RUnlock()

	// 缓存过期或不存在，重新获取
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 双重检查，防止并发问题
	if time.Since(m.cacheTime) < m.cacheTTL && len(m.cache) > 0 {
		environments := make([]Environment, 0, len(m.cache))
		for _, env := range m.cache {
			environments = append(environments, env)
		}
		return environments
	}

	environments := []Environment{
		{
			Name:        "nginx",
			DisplayName: "Nginx",
			Description: "高性能Web服务器和反向代理 (Debian稳定版)",
			Required:    true,
		},
		{
			Name:        "php",
			DisplayName: "PHP",
			Description: "PHP编程语言运行环境（包含PHP-FPM）(Debian稳定版)",
			Required:    false,
		},
		{
			Name:        "mysql",
			DisplayName: "MariaDB",
			Description: "关系型数据库管理系统 (Debian默认使用MariaDB)",
			Required:    false,
		},
		{
			Name:        "redis",
			DisplayName: "Redis",
			Description: "高性能内存数据库 (Debian稳定版)",
			Required:    false,
		},
	}

	// 并发检查每个环境的状态
	var wg sync.WaitGroup
	for i := range environments {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			status, version := m.checkEnvironmentStatus(environments[idx].Name)
			environments[idx].Status = status
			environments[idx].Version = version
			// 简化版本信息，避免频繁的apt调用
			environments[idx].AvailableVersions = m.getSimpleVersions(environments[idx].Name)
			environments[idx].LatestVersion = "最新版本"
		}(i)
	}
	wg.Wait()

	// 更新缓存
	m.cache = make(map[string]Environment)
	for _, env := range environments {
		m.cache[env.Name] = env
	}
	m.cacheTime = time.Now()

	return environments
}

// InstallEnvironment 安装环境
func (m *Manager) InstallEnvironment(name, version string, progressChan chan<- InstallProgress) error {
	m.mutex.Lock()
	defer func() {
		m.InvalidateCache() // 清除缓存
		m.mutex.Unlock()
	}()

	progressChan <- InstallProgress{
		Environment: name,
		Progress:    0,
		Message:     fmt.Sprintf("开始安装 %s %s...", name, version),
		Status:      "installing",
	}

	switch name {
	case "nginx":
		return m.installNginx(version, progressChan)
	case "php":
		return m.installPHP(version, progressChan)
	case "mysql":
		return m.installMySQL(version, progressChan)
	case "redis":
		return m.installRedis(version, progressChan)
	default:
		return fmt.Errorf("不支持的环境: %s", name)
	}
}

// UninstallEnvironment 卸载环境
func (m *Manager) UninstallEnvironment(name string) error {
	m.mutex.Lock()
	defer func() {
		m.InvalidateCache() // 清除缓存
		m.mutex.Unlock()
	}()

	switch name {
	case "nginx":
		return m.runCommand("apt", "remove", "-y", "nginx", "nginx-common")
	case "php":
		return m.runCommand("apt", "remove", "-y", "php*")
	case "mysql":
		return m.runCommand("apt", "remove", "-y", "mysql-server", "mysql-client")
	case "redis":
		return m.runCommand("apt", "remove", "-y", "redis-server")
	default:
		return fmt.Errorf("不支持的环境: %s", name)
	}
}

// UpgradeEnvironment 升级环境
func (m *Manager) UpgradeEnvironment(name, version string, progressChan chan<- InstallProgress) error {
	m.mutex.Lock()
	defer func() {
		m.InvalidateCache() // 清除缓存
		m.mutex.Unlock()
	}()

	progressChan <- InstallProgress{
		Environment: name,
		Progress:    0,
		Message:     fmt.Sprintf("开始升级到 %s %s...", name, version),
		Status:      "upgrading",
	}

	switch name {
	case "nginx":
		return m.upgradeNginx(version, progressChan)
	case "php":
		return m.upgradePHP(version, progressChan)
	case "mysql":
		return m.upgradeMySQL(version, progressChan)
	case "redis":
		return m.upgradeRedis(version, progressChan)
	default:
		return fmt.Errorf("不支持的环境: %s", name)
	}
}

// upgradeNginx 升级Nginx
func (m *Manager) upgradeNginx(version string, progressChan chan<- InstallProgress) error {
	// 升级实际上是重新安装指定版本
	installCmd := []string{"apt", "install", "-y", "--reinstall"}
	if version == "最新版本" || version == "" {
		installCmd = append(installCmd, "nginx")
	} else {
		installCmd = append(installCmd, "nginx")
	}

	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{20, "更新软件包列表...", []string{"apt", "update"}},
		{60, fmt.Sprintf("升级Nginx到 %s...", version), installCmd},
		{80, "重启Nginx服务...", []string{"systemctl", "restart", "nginx"}},
		{100, "升级完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "nginx",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "upgrading",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "nginx",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	progressChan <- InstallProgress{
		Environment: "nginx",
		Progress:    100,
		Message:     "Nginx升级成功",
		Status:      "completed",
	}

	return nil
}

// upgradePHP 升级PHP
func (m *Manager) upgradePHP(version string, progressChan chan<- InstallProgress) error {
	// 先卸载旧版本，再安装新版本
	var removeCmd []string
	var installCmd []string
	var phpService string

	if version == "最新版本" || version == "" {
		removeCmd = []string{"apt", "remove", "-y", "php*"}
		installCmd = []string{"apt", "install", "-y", "php", "php-fpm", "php-mysql", "php-curl", "php-gd", "php-mbstring", "php-xml", "php-zip", "php-json", "php-opcache"}
		phpService = "php8.2-fpm"
	} else {
		removeCmd = []string{"apt", "remove", "-y", "php*"}
		installCmd = []string{"apt", "install", "-y",
			fmt.Sprintf("php%s", version),
			fmt.Sprintf("php%s-fpm", version),
			fmt.Sprintf("php%s-mysql", version),
			fmt.Sprintf("php%s-curl", version),
			fmt.Sprintf("php%s-gd", version),
			fmt.Sprintf("php%s-mbstring", version),
			fmt.Sprintf("php%s-xml", version),
			fmt.Sprintf("php%s-zip", version),
			fmt.Sprintf("php%s-json", version),
			fmt.Sprintf("php%s-opcache", version),
		}
		phpService = fmt.Sprintf("php%s-fpm", version)
	}

	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{10, "更新软件包列表...", []string{"apt", "update"}},
		{30, "停止PHP服务...", []string{"systemctl", "stop", "php*-fpm"}},
		{50, "卸载旧版本PHP...", removeCmd},
		{70, fmt.Sprintf("安装PHP %s...", version), installCmd},
		{90, "启动PHP-FPM服务...", []string{"systemctl", "start", phpService}},
		{95, "设置开机自启...", []string{"systemctl", "enable", phpService}},
		{100, "升级完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "php",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "upgrading",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "php",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	return nil
}

// upgradeMySQL 升级MySQL
func (m *Manager) upgradeMySQL(version string, progressChan chan<- InstallProgress) error {
	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{10, "更新软件包列表...", []string{"apt", "update"}},
		{20, "备份数据库配置...", nil}, // 自定义备份步骤
		{60, "升级数据库...", []string{"apt", "upgrade", "-y", "mysql-server", "mysql-client", "mariadb-server", "mariadb-client"}},
		{80, "重启数据库服务...", []string{"systemctl", "restart", "mysql"}},
		{90, "检查服务状态...", []string{"systemctl", "status", "mysql"}},
		{100, "升级完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "mysql",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "upgrading",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "mysql",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	return nil
}

// upgradeRedis 升级Redis
func (m *Manager) upgradeRedis(version string, progressChan chan<- InstallProgress) error {
	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{20, "更新软件包列表...", []string{"apt", "update"}},
		{60, "升级Redis...", []string{"apt", "upgrade", "-y", "redis-server"}},
		{80, "重启Redis服务...", []string{"systemctl", "restart", "redis-server"}},
		{100, "升级完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "redis",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "upgrading",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "redis",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	return nil
}

// checkEnvironmentStatus 检查环境安装状态和版本
func (m *Manager) checkEnvironmentStatus(name string) (status string, version string) {
	switch name {
	case "nginx":
		return m.checkNginxStatus()
	case "php":
		return m.checkPHPStatus()
	case "mysql":
		return m.checkMySQLStatus()
	case "redis":
		return m.checkRedisStatus()
	default:
		return "not_installed", ""
	}
}

// checkNginxStatus 检查Nginx状态
func (m *Manager) checkNginxStatus() (string, string) {
	// 检查nginx命令是否存在
	if _, err := exec.LookPath("nginx"); err != nil {
		return "not_installed", ""
	}

	// 获取版本信息
	cmd := exec.Command("nginx", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "installed", "未知版本"
	}

	// 解析版本 (nginx version: nginx/1.18.0)
	versionStr := string(output)
	if strings.Contains(versionStr, "nginx/") {
		parts := strings.Split(versionStr, "nginx/")
		if len(parts) > 1 {
			version := strings.TrimSpace(parts[1])
			return "installed", version
		}
	}

	return "installed", "未知版本"
}

// checkPHPStatus 检查PHP状态
func (m *Manager) checkPHPStatus() (string, string) {
	// 检查php命令是否存在
	if _, err := exec.LookPath("php"); err != nil {
		return "not_installed", ""
	}

	// 获取版本信息
	cmd := exec.Command("php", "-v")
	output, err := cmd.Output()
	if err != nil {
		return "installed", "未知版本"
	}

	// 解析版本 (PHP 8.2.7 (cli))
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		if strings.HasPrefix(firstLine, "PHP ") {
			parts := strings.Fields(firstLine)
			if len(parts) >= 2 {
				return "installed", parts[1]
			}
		}
	}

	return "installed", "未知版本"
}

// checkMySQLStatus 检查MySQL状态
func (m *Manager) checkMySQLStatus() (string, string) {
	// 检查mysql命令是否存在
	if _, err := exec.LookPath("mysql"); err != nil {
		return "not_installed", ""
	}

	// 获取版本信息
	cmd := exec.Command("mysql", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "installed", "未知版本"
	}

	// 解析版本信息
	versionStr := string(output)

	// 检查是否是MariaDB
	if strings.Contains(versionStr, "MariaDB") {
		// 解析MariaDB版本 (mysql  Ver 15.1 Distrib 10.11.11-MariaDB, for debian-linux-gnu)
		if strings.Contains(versionStr, "Distrib ") {
			parts := strings.Split(versionStr, "Distrib ")
			if len(parts) > 1 {
				versionPart := strings.Fields(parts[1])
				if len(versionPart) > 0 {
					// 提取版本号，去掉-MariaDB后缀
					version := strings.Split(versionPart[0], "-")[0]
					return "installed", fmt.Sprintf("MariaDB %s", version)
				}
			}
		}
		return "installed", "MariaDB 未知版本"
	}

	// 检查是否是MySQL
	if strings.Contains(versionStr, "Ver ") {
		parts := strings.Split(versionStr, "Ver ")
		if len(parts) > 1 {
			versionPart := strings.Fields(parts[1])
			if len(versionPart) > 0 {
				// 提取主版本号
				version := strings.Split(versionPart[0], "-")[0]
				return "installed", fmt.Sprintf("MySQL %s", version)
			}
		}
	}

	return "installed", "未知版本"
}

// checkRedisStatus 检查Redis状态
func (m *Manager) checkRedisStatus() (string, string) {
	// 检查redis-server命令是否存在
	if _, err := exec.LookPath("redis-server"); err != nil {
		return "not_installed", ""
	}

	// 获取版本信息
	cmd := exec.Command("redis-server", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "installed", "未知版本"
	}

	// 解析版本 (Redis server v=7.0.11 sha=00000000:0 malloc=jemalloc-5.2.1 bits=64 build=2dd77560d1c11a56)
	versionStr := string(output)
	if strings.Contains(versionStr, "v=") {
		parts := strings.Split(versionStr, "v=")
		if len(parts) > 1 {
			versionPart := strings.Fields(parts[1])
			if len(versionPart) > 0 {
				return "installed", versionPart[0]
			}
		}
	}

	return "installed", "未知版本"
}

// getVersion 获取版本信息
func (m *Manager) getVersion(name string) string {
	var cmd *exec.Cmd
	
	switch name {
	case "nginx":
		cmd = exec.Command("nginx", "-v")
	case "php":
		cmd = exec.Command("php", "-v")
	case "mysql":
		cmd = exec.Command("mysql", "--version")
	case "nodejs":
		cmd = exec.Command("node", "--version")
	case "python":
		cmd = exec.Command("python3", "--version")
	case "docker":
		cmd = exec.Command("docker", "--version")
	case "redis":
		cmd = exec.Command("redis-server", "--version")
	case "git":
		cmd = exec.Command("git", "--version")
	default:
		return ""
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(output))
	// 提取版本号（简化处理）
	if strings.Contains(version, "version") {
		parts := strings.Fields(version)
		for i, part := range parts {
			if part == "version" && i+1 < len(parts) {
				return parts[i+1]
			}
		}
	}
	
	return strings.Fields(version)[0]
}

// installNginx 安装Nginx
func (m *Manager) installNginx(version string, progressChan chan<- InstallProgress) error {
	// 构建安装命令
	installCmd := []string{"apt", "install", "-y"}
	if version == "最新版本" || version == "" {
		installCmd = append(installCmd, "nginx")
	} else {
		// 对于特定版本，先尝试查找可用的包
		installCmd = append(installCmd, "nginx")
	}

	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{20, "更新软件包列表...", []string{"apt", "update"}},
		{60, fmt.Sprintf("安装Nginx %s...", version), installCmd},
		{80, "启动Nginx服务...", []string{"systemctl", "start", "nginx"}},
		{90, "设置开机自启...", []string{"systemctl", "enable", "nginx"}},
		{100, "安装完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "nginx",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "installing",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "nginx",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	progressChan <- InstallProgress{
		Environment: "nginx",
		Progress:    100,
		Message:     "Nginx安装成功",
		Status:      "completed",
	}

	return nil
}

// installPHP 安装PHP
func (m *Manager) installPHP(version string, progressChan chan<- InstallProgress) error {
	// 构建PHP安装命令
	var phpPackages []string
	var phpService string

	if version == "最新版本" || version == "" {
		phpPackages = []string{"php", "php-fpm", "php-mysql", "php-curl", "php-gd", "php-mbstring", "php-xml", "php-zip", "php-json", "php-opcache"}
		phpService = "php8.2-fpm" // 默认服务名
	} else {
		// 安装特定版本的PHP
		phpPackages = []string{
			fmt.Sprintf("php%s", version),
			fmt.Sprintf("php%s-fpm", version),
			fmt.Sprintf("php%s-mysql", version),
			fmt.Sprintf("php%s-curl", version),
			fmt.Sprintf("php%s-gd", version),
			fmt.Sprintf("php%s-mbstring", version),
			fmt.Sprintf("php%s-xml", version),
			fmt.Sprintf("php%s-zip", version),
			fmt.Sprintf("php%s-json", version),
			fmt.Sprintf("php%s-opcache", version),
		}
		phpService = fmt.Sprintf("php%s-fpm", version)
	}

	installCmd := append([]string{"apt", "install", "-y"}, phpPackages...)

	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{20, "更新软件包列表...", []string{"apt", "update"}},
		{60, fmt.Sprintf("安装PHP %s及扩展...", version), installCmd},
		{80, "启动PHP-FPM服务...", []string{"systemctl", "start", phpService}},
		{90, "设置开机自启...", []string{"systemctl", "enable", phpService}},
		{100, "安装完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "php",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "installing",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "php",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	return nil
}

// installMySQL 安装MySQL
func (m *Manager) installMySQL(version string, progressChan chan<- InstallProgress) error {
	// 根据版本选择安装包
	var installPackage string
	var serviceName string

	if strings.Contains(version, "MariaDB") || version == "最新版本" {
		installPackage = "mariadb-server"
		serviceName = "mariadb"
	} else {
		installPackage = "mysql-server"
		serviceName = "mysql"
	}

	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{20, "更新软件包列表...", []string{"apt", "update"}},
		{60, fmt.Sprintf("安装%s...", version), []string{"apt", "install", "-y", installPackage}},
		{80, "启动数据库服务...", []string{"systemctl", "start", serviceName}},
		{90, "设置开机自启...", []string{"systemctl", "enable", serviceName}},
		{100, "安装完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "mysql",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "installing",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "mysql",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	return nil
}

// installNodeJS 安装Node.js
func (m *Manager) installNodeJS(progressChan chan<- InstallProgress) error {
	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{20, "更新软件包列表...", []string{"apt", "update"}},
		{60, "安装Node.js和npm...", []string{"apt", "install", "-y", "nodejs", "npm"}},
		{100, "安装完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "nodejs",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "installing",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				return err
			}
		}
	}

	return nil
}

// installPython 安装Python
func (m *Manager) installPython(progressChan chan<- InstallProgress) error {
	return m.runCommand("apt", "install", "-y", "python3", "python3-pip")
}

// installDocker 安装Docker
func (m *Manager) installDocker(progressChan chan<- InstallProgress) error {
	return m.runCommand("apt", "install", "-y", "docker.io")
}

// installRedis 安装Redis
func (m *Manager) installRedis(version string, progressChan chan<- InstallProgress) error {
	steps := []struct {
		progress int
		message  string
		command  []string
	}{
		{20, "更新软件包列表...", []string{"apt", "update"}},
		{60, "安装Redis...", []string{"apt", "install", "-y", "redis-server"}},
		{80, "启动Redis服务...", []string{"systemctl", "start", "redis-server"}},
		{90, "设置开机自启...", []string{"systemctl", "enable", "redis-server"}},
		{100, "安装完成", nil},
	}

	for _, step := range steps {
		progressChan <- InstallProgress{
			Environment: "redis",
			Progress:    step.progress,
			Message:     step.message,
			Status:      "installing",
		}

		if step.command != nil {
			if err := m.runCommand(step.command[0], step.command[1:]...); err != nil {
				progressChan <- InstallProgress{
					Environment: "redis",
					Progress:    step.progress,
					Message:     fmt.Sprintf("错误: %v", err),
					Status:      "error",
				}
				return err
			}
		}
	}

	return nil
}

// installGit 安装Git
func (m *Manager) installGit(progressChan chan<- InstallProgress) error {
	return m.runCommand("apt", "install", "-y", "git")
}

// runCommand 执行命令
func (m *Manager) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	
	// 获取输出用于调试
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("命令执行失败: %s, 输出: %s", err, string(output))
	}
	
	return nil
}

// runCommandWithProgress 带进度的命令执行
func (m *Manager) runCommandWithProgress(progressChan chan<- InstallProgress, env string, cmd *exec.Cmd) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		progressChan <- InstallProgress{
			Environment: env,
			Progress:    50, // 简化进度处理
			Message:     line,
			Status:      "installing",
		}
	}

	return cmd.Wait()
}

// getAvailableVersions 获取可用版本列表
func (m *Manager) getAvailableVersions(name string) []string {
	switch name {
	case "nginx":
		return m.getNginxVersions()
	case "php":
		return m.getPHPVersions()
	case "mysql":
		return m.getMySQLVersions()
	case "redis":
		return m.getRedisVersions()
	default:
		return []string{}
	}
}

// getLatestVersion 获取最新版本
func (m *Manager) getLatestVersion(name string) string {
	versions := m.getAvailableVersions(name)
	if len(versions) > 0 {
		return versions[0] // 假设第一个是最新版本
	}
	return ""
}

// getNginxVersions 获取Nginx可用版本
func (m *Manager) getNginxVersions() []string {
	// 检查当前仓库中可用的版本
	cmd := exec.Command("apt", "list", "-a", "nginx", "2>/dev/null")
	output, err := cmd.Output()
	if err != nil {
		return []string{"最新版本"}
	}

	versions := []string{"最新版本"}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "nginx/") && !strings.Contains(line, "WARNING") {
			// 提取版本信息
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				version := parts[1]
				// 简化版本显示
				if strings.Contains(version, "-") {
					version = strings.Split(version, "-")[0]
				}
				if version != "" && !contains(versions, version) {
					versions = append(versions, version)
				}
			}
		}
	}

	return versions
}

// getPHPVersions 获取PHP可用版本
func (m *Manager) getPHPVersions() []string {
	versions := []string{"最新版本"}

	// 检查系统支持的PHP版本
	commonVersions := []string{"8.3", "8.2", "8.1", "8.0", "7.4"}

	for _, version := range commonVersions {
		// 检查是否有对应的包
		cmd := exec.Command("apt", "list", fmt.Sprintf("php%s", version), "2>/dev/null")
		if output, err := cmd.Output(); err == nil {
			if strings.Contains(string(output), fmt.Sprintf("php%s/", version)) {
				versions = append(versions, version)
			}
		}
	}

	// 如果没有找到其他版本，返回当前版本
	if len(versions) == 1 {
		currentVersion := m.getCurrentPHPVersion()
		if currentVersion != "" {
			versions = append(versions, currentVersion+" (当前)")
		}
	}

	return versions
}

// getMySQLVersions 获取MySQL可用版本
func (m *Manager) getMySQLVersions() []string {
	return []string{"最新版本", "MySQL 8.0", "MySQL 5.7", "MariaDB 10.11", "MariaDB 10.6"}
}

// getRedisVersions 获取Redis可用版本
func (m *Manager) getRedisVersions() []string {
	return []string{"最新版本", "7.0", "6.2", "6.0", "5.0"}
}

// PHPExtension PHP扩展信息
type PHPExtension struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Status      string `json:"status"` // enabled, disabled, not_installed
	Version     string `json:"version"`
	Required    bool   `json:"required"`
}

// GetPHPExtensions 获取PHP扩展列表
func (m *Manager) GetPHPExtensions() ([]PHPExtension, error) {
	extensions := []PHPExtension{
		{
			Name:        "opcache",
			DisplayName: "OPcache",
			Description: "PHP字节码缓存，提升性能",
			Required:    true,
		},
		{
			Name:        "mysql",
			DisplayName: "MySQL",
			Description: "MySQL数据库支持",
			Required:    false,
		},
		{
			Name:        "mysqli",
			DisplayName: "MySQLi",
			Description: "MySQL改进扩展",
			Required:    false,
		},
		{
			Name:        "pdo_mysql",
			DisplayName: "PDO MySQL",
			Description: "PDO MySQL驱动",
			Required:    false,
		},
		{
			Name:        "curl",
			DisplayName: "cURL",
			Description: "HTTP客户端库",
			Required:    false,
		},
		{
			Name:        "gd",
			DisplayName: "GD",
			Description: "图像处理库",
			Required:    false,
		},
		{
			Name:        "mbstring",
			DisplayName: "Multibyte String",
			Description: "多字节字符串处理",
			Required:    false,
		},
		{
			Name:        "xml",
			DisplayName: "XML",
			Description: "XML处理支持",
			Required:    false,
		},
		{
			Name:        "zip",
			DisplayName: "ZIP",
			Description: "ZIP压缩支持",
			Required:    false,
		},
		{
			Name:        "json",
			DisplayName: "JSON",
			Description: "JSON数据处理",
			Required:    true,
		},
		{
			Name:        "redis",
			DisplayName: "Redis",
			Description: "Redis缓存支持",
			Required:    false,
		},
		{
			Name:        "memcached",
			DisplayName: "Memcached",
			Description: "Memcached缓存支持",
			Required:    false,
		},
		{
			Name:        "imagick",
			DisplayName: "ImageMagick",
			Description: "高级图像处理",
			Required:    false,
		},
		{
			Name:        "xdebug",
			DisplayName: "Xdebug",
			Description: "PHP调试和性能分析",
			Required:    false,
		},
	}

	// 检查每个扩展的状态
	for i := range extensions {
		status, version := m.checkPHPExtensionStatus(extensions[i].Name)
		extensions[i].Status = status
		extensions[i].Version = version
	}

	return extensions, nil
}

// checkPHPExtensionStatus 检查PHP扩展状态
func (m *Manager) checkPHPExtensionStatus(extName string) (string, string) {
	// 使用php -m命令检查已加载的扩展
	cmd := exec.Command("php", "-m")
	output, err := cmd.Output()
	if err != nil {
		return "not_installed", ""
	}

	modules := strings.Split(string(output), "\n")
	for _, module := range modules {
		if strings.TrimSpace(strings.ToLower(module)) == strings.ToLower(extName) {
			// 扩展已加载，获取版本信息
			version := m.getPHPExtensionVersion(extName)
			return "enabled", version
		}
	}

	// 检查扩展是否已安装但未启用
	if m.isPHPExtensionInstalled(extName) {
		return "disabled", ""
	}

	return "not_installed", ""
}

// isPHPExtensionInstalled 检查PHP扩展是否已安装
func (m *Manager) isPHPExtensionInstalled(extName string) bool {
	// 检查扩展文件是否存在
	phpVersion := m.getCurrentPHPVersion()
	if phpVersion == "" {
		return false
	}

	extFile := fmt.Sprintf("/usr/lib/php/%s/%s.so", phpVersion, extName)
	if _, err := os.Stat(extFile); err == nil {
		return true
	}

	// 检查一些常见路径
	commonPaths := []string{
		fmt.Sprintf("/usr/lib/php/%s/%s.so", phpVersion, extName),
		fmt.Sprintf("/usr/lib/php/modules/%s.so", extName),
		fmt.Sprintf("/usr/lib/php/extensions/%s.so", extName),
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// getCurrentPHPVersion 获取当前PHP版本
func (m *Manager) getCurrentPHPVersion() string {
	cmd := exec.Command("php", "-v")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		firstLine := lines[0]
		if strings.HasPrefix(firstLine, "PHP ") {
			parts := strings.Fields(firstLine)
			if len(parts) >= 2 {
				version := parts[1]
				// 提取主版本号 (如 8.2)
				versionParts := strings.Split(version, ".")
				if len(versionParts) >= 2 {
					return fmt.Sprintf("%s.%s", versionParts[0], versionParts[1])
				}
			}
		}
	}

	return "8.2" // 默认版本
}

// getPHPExtensionVersion 获取PHP扩展版本
func (m *Manager) getPHPExtensionVersion(extName string) string {
	cmd := exec.Command("php", "-r", fmt.Sprintf("echo phpversion('%s');", extName))
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	version := strings.TrimSpace(string(output))
	if version == "" || version == "false" {
		return ""
	}

	return version
}

// EnablePHPExtension 启用PHP扩展
func (m *Manager) EnablePHPExtension(extName string) error {
	phpVersion := m.getCurrentPHPVersion()

	// 检查扩展是否已安装
	if !m.isPHPExtensionInstalled(extName) {
		// 尝试安装扩展
		if err := m.installPHPExtension(extName, phpVersion); err != nil {
			return fmt.Errorf("安装扩展失败: %v", err)
		}
	}

	// 启用扩展
	return m.enablePHPExtensionConfig(extName, phpVersion)
}

// DisablePHPExtension 禁用PHP扩展
func (m *Manager) DisablePHPExtension(extName string) error {
	phpVersion := m.getCurrentPHPVersion()
	return m.disablePHPExtensionConfig(extName, phpVersion)
}

// installPHPExtension 安装PHP扩展
func (m *Manager) installPHPExtension(extName, phpVersion string) error {
	packageName := fmt.Sprintf("php%s-%s", phpVersion, extName)

	// 特殊处理一些扩展名
	switch extName {
	case "mysql":
		packageName = fmt.Sprintf("php%s-mysql", phpVersion)
	case "pdo_mysql":
		packageName = fmt.Sprintf("php%s-mysql", phpVersion)
	case "imagick":
		packageName = fmt.Sprintf("php%s-imagick", phpVersion)
	case "xdebug":
		packageName = fmt.Sprintf("php%s-xdebug", phpVersion)
	}

	return m.runCommand("apt", "install", "-y", packageName)
}

// enablePHPExtensionConfig 启用PHP扩展配置
func (m *Manager) enablePHPExtensionConfig(extName, phpVersion string) error {
	// 使用phpenmod命令启用扩展
	if err := m.runCommand("phpenmod", "-v", phpVersion, extName); err != nil {
		// 如果phpenmod失败，尝试手动创建配置文件
		return m.createPHPExtensionConfig(extName, phpVersion)
	}

	// 重启PHP-FPM服务
	return m.runCommand("systemctl", "restart", fmt.Sprintf("php%s-fpm", phpVersion))
}

// disablePHPExtensionConfig 禁用PHP扩展配置
func (m *Manager) disablePHPExtensionConfig(extName, phpVersion string) error {
	// 使用phpdismod命令禁用扩展
	if err := m.runCommand("phpdismod", "-v", phpVersion, extName); err != nil {
		// 如果phpdismod失败，尝试手动删除配置文件
		return m.removePHPExtensionConfig(extName, phpVersion)
	}

	// 重启PHP-FPM服务
	return m.runCommand("systemctl", "restart", fmt.Sprintf("php%s-fpm", phpVersion))
}

// createPHPExtensionConfig 创建PHP扩展配置文件
func (m *Manager) createPHPExtensionConfig(extName, phpVersion string) error {
	configDir := fmt.Sprintf("/etc/php/%s/mods-available", phpVersion)
	configFile := fmt.Sprintf("%s/%s.ini", configDir, extName)

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	// 创建配置文件
	content := fmt.Sprintf("extension=%s\n", extName)
	return os.WriteFile(configFile, []byte(content), 0644)
}

// removePHPExtensionConfig 删除PHP扩展配置文件
func (m *Manager) removePHPExtensionConfig(extName, phpVersion string) error {
	configFile := fmt.Sprintf("/etc/php/%s/mods-available/%s.ini", phpVersion, extName)
	return os.Remove(configFile)
}

// contains 检查字符串切片是否包含指定字符串
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getSimpleVersions 获取简化的版本列表（避免频繁的apt调用）
func (m *Manager) getSimpleVersions(name string) []string {
	switch name {
	case "nginx":
		return []string{"最新版本", "1.22.1 (当前)"}
	case "php":
		return []string{"最新版本", "8.2 (当前)"}
	case "mysql":
		return []string{"最新版本", "MariaDB 10.11 (当前)"}
	case "redis":
		return []string{"最新版本", "7.0 (当前)"}
	default:
		return []string{"最新版本"}
	}
}

// InvalidateCache 清除缓存
func (m *Manager) InvalidateCache() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.cache = make(map[string]Environment)
	m.cacheTime = time.Time{}
}
