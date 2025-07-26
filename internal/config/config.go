package config

import (
	"os"
	"time"
)

// Config 应用配置
type Config struct {
	Debug bool      `yaml:"debug"`
	Auth  AuthConfig `yaml:"auth"`
	Server ServerConfig `yaml:"server"`
	Paths  PathsConfig  `yaml:"paths"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	SessionTimeout time.Duration `yaml:"session_timeout"`
	MaxLoginAttempts int         `yaml:"max_login_attempts"`
	LockoutDuration  time.Duration `yaml:"lockout_duration"`
	SecretKey       string        `yaml:"secret_key"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         string `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

// PathsConfig 路径配置
type PathsConfig struct {
	DataDir    string `yaml:"data_dir"`
	LogDir     string `yaml:"log_dir"`
	TempDir    string `yaml:"temp_dir"`
	BackupDir  string `yaml:"backup_dir"`
}

// Load 加载配置文件
func Load(path string) (*Config, error) {
	// 简化版本：直接返回默认配置
	// 在实际项目中，这里会解析YAML文件
	return Default(), nil
}

// Default 返回默认配置
func Default() *Config {
	return &Config{
		Debug: false,
		Auth: AuthConfig{
			SessionTimeout:   time.Hour,
			MaxLoginAttempts: 5,
			LockoutDuration:  15 * time.Minute,
			SecretKey:       generateSecretKey(),
		},
		Server: ServerConfig{
			Port:         "8443",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		Paths: PathsConfig{
			DataDir:   "/var/lib/server-panel",
			LogDir:    "/var/log/server-panel",
			TempDir:   "/tmp/server-panel",
			BackupDir: "/var/backups/server-panel",
		},
	}
}

// generateSecretKey 生成随机密钥
func generateSecretKey() string {
	// 检查环境变量
	if key := os.Getenv("SERVER_PANEL_SECRET"); key != "" {
		return key
	}
	
	// 默认密钥（生产环境应该使用随机生成的密钥）
	return "your-secret-key-change-in-production"
}

// EnsureDirectories 确保必要的目录存在
func (c *Config) EnsureDirectories() error {
	dirs := []string{
		c.Paths.DataDir,
		c.Paths.LogDir,
		c.Paths.TempDir,
		c.Paths.BackupDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
