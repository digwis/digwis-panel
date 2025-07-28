package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Manager 数据库管理器
type Manager struct {
	db       *sql.DB
	dataDir  string
	dbPath   string
}

// NewManager 创建数据库管理器
func NewManager() *Manager {
	dataDir := getDataDirectory()
	dbPath := filepath.Join(dataDir, "digwis-panel.db")
	
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		log.Fatalf("无法创建数据目录 %s: %v", dataDir, err)
	}
	
	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_timeout=5000")
	if err != nil {
		log.Fatalf("无法打开数据库 %s: %v", dbPath, err)
	}
	
	// 设置连接池参数
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	
	manager := &Manager{
		db:      db,
		dataDir: dataDir,
		dbPath:  dbPath,
	}
	
	// 初始化数据库表
	if err := manager.initTables(); err != nil {
		log.Fatalf("无法初始化数据库表: %v", err)
	}
	
	log.Printf("📊 数据库初始化完成: %s", dbPath)
	return manager
}

// GetDB 获取数据库连接
func (m *Manager) GetDB() *sql.DB {
	return m.db
}

// GetDataDir 获取数据目录
func (m *Manager) GetDataDir() string {
	return m.dataDir
}

// GetDBPath 获取数据库文件路径
func (m *Manager) GetDBPath() string {
	return m.dbPath
}

// Close 关闭数据库连接
func (m *Manager) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// initTables 初始化数据库表
func (m *Manager) initTables() error {
	// 会话表
	sessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		data TEXT NOT NULL,
		expires DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires);
	`
	
	// 系统配置表
	configTable := `
	CREATE TABLE IF NOT EXISTS system_config (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		description TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	
	// 操作日志表
	logsTable := `
	CREATE TABLE IF NOT EXISTS operation_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		action TEXT NOT NULL,
		resource TEXT,
		details TEXT,
		ip_address TEXT,
		user_agent TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_logs_username ON operation_logs(username);
	CREATE INDEX IF NOT EXISTS idx_logs_created_at ON operation_logs(created_at);
	`
	
	// 执行建表语句
	tables := []string{sessionsTable, configTable, logsTable}
	for _, table := range tables {
		if _, err := m.db.Exec(table); err != nil {
			return fmt.Errorf("创建表失败: %v", err)
		}
	}
	
	// 插入默认配置
	return m.insertDefaultConfig()
}

// insertDefaultConfig 插入默认配置
func (m *Manager) insertDefaultConfig() error {
	defaultConfigs := map[string]string{
		"panel_name":        "DigWis Panel",
		"panel_version":     "1.0.0",
		"session_timeout":   "24h",
		"max_login_attempts": "5",
		"lockout_duration":  "15m",
		"auto_backup":       "true",
		"backup_retention":  "7",
	}
	
	for key, value := range defaultConfigs {
		_, err := m.db.Exec(`
			INSERT OR IGNORE INTO system_config (key, value, description) 
			VALUES (?, ?, ?)
		`, key, value, "系统默认配置")
		
		if err != nil {
			return fmt.Errorf("插入默认配置失败: %v", err)
		}
	}
	
	return nil
}

// getDataDirectory 获取数据目录
func getDataDirectory() string {
	// 检查环境变量
	if dataDir := os.Getenv("DIGWIS_DATA_DIR"); dataDir != "" {
		return dataDir
	}
	
	// 检查是否在开发模式
	if isDevMode() {
		if wd, err := os.Getwd(); err == nil {
			return filepath.Join(wd, "data")
		}
		return filepath.Join(os.TempDir(), "digwis-panel-dev")
	}
	
	// 生产模式
	return getProductionDataDir()
}

// isDevMode 检查是否在开发模式
func isDevMode() bool {
	if mode := os.Getenv("DIGWIS_MODE"); mode == "production" {
		return false
	}
	if mode := os.Getenv("DIGWIS_MODE"); mode == "development" {
		return true
	}
	
	// 检查开发文件标识
	devFiles := []string{".air.toml", "go.mod", "package.json"}
	for _, file := range devFiles {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}
	
	return false
}

// getProductionDataDir 获取生产环境数据目录
func getProductionDataDir() string {
	possibleDirs := []string{
		"/opt/digwis-panel/data",      // 首选：应用程序数据目录
		"/var/lib/digwis-panel",       // 备选：系统服务数据目录
		"/usr/local/var/digwis-panel", // 备选：本地变量目录
		"/etc/digwis-panel/data",      // 备选：配置目录
	}
	
	for _, dir := range possibleDirs {
		if err := os.MkdirAll(dir, 0750); err == nil {
			testFile := filepath.Join(dir, ".write_test")
			if file, err := os.Create(testFile); err == nil {
				file.Close()
				os.Remove(testFile)
				return dir
			}
		}
	}
	
	// 用户主目录备选
	if homeDir, err := os.UserHomeDir(); err == nil {
		dataDir := filepath.Join(homeDir, ".digwis-panel")
		os.MkdirAll(dataDir, 0750)
		return dataDir
	}
	
	// 最后备选
	fallbackDir := filepath.Join(os.TempDir(), "digwis-panel")
	os.MkdirAll(fallbackDir, 0750)
	return fallbackDir
}

// Backup 备份数据库
func (m *Manager) Backup() error {
	backupDir := filepath.Join(m.dataDir, "backups")
	if err := os.MkdirAll(backupDir, 0750); err != nil {
		return err
	}
	
	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("digwis-panel_%s.db", timestamp))
	
	// 使用 SQLite 的 VACUUM INTO 命令进行备份
	_, err := m.db.Exec("VACUUM INTO ?", backupPath)
	if err != nil {
		return fmt.Errorf("备份数据库失败: %v", err)
	}
	
	log.Printf("📦 数据库备份完成: %s", backupPath)
	return nil
}

// GetDatabaseInfo 获取数据库信息
func (m *Manager) GetDatabaseInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	// 数据库文件大小
	if stat, err := os.Stat(m.dbPath); err == nil {
		info["size"] = stat.Size()
		info["modified"] = stat.ModTime()
	}
	
	// 表信息
	rows, err := m.db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err == nil {
		defer rows.Close()
		var tables []string
		for rows.Next() {
			var name string
			if rows.Scan(&name) == nil {
				tables = append(tables, name)
			}
		}
		info["tables"] = tables
	}
	
	info["path"] = m.dbPath
	info["data_dir"] = m.dataDir
	
	return info
}
