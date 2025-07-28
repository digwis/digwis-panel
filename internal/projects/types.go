package projects

import "time"

// ProjectStatus represents the status of a project
type ProjectStatus string

const (
	StatusActive   ProjectStatus = "active"
	StatusInactive ProjectStatus = "inactive"
	StatusError    ProjectStatus = "error"
)

// Project represents a web project
type Project struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Domain      string        `json:"domain"`
	Path        string        `json:"path"`
	Status      ProjectStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	DatabaseName string       `json:"database_name"`
	SSLEnabled  bool          `json:"ssl_enabled"`
	SSLCertPath string        `json:"ssl_cert_path,omitempty"`
	BackupEnabled bool        `json:"backup_enabled"`
	BackupSchedule string     `json:"backup_schedule,omitempty"`
	Size        int64         `json:"size"` // in bytes
}

// NginxConfig represents Nginx configuration for a project
type NginxConfig struct {
	ProjectID    string `json:"project_id"`
	ServerName   string `json:"server_name"`
	DocumentRoot string `json:"document_root"`
	Port         int    `json:"port"`
	SSLEnabled   bool   `json:"ssl_enabled"`
	SSLCertPath  string `json:"ssl_cert_path,omitempty"`
	SSLKeyPath   string `json:"ssl_key_path,omitempty"`
	CustomConfig string `json:"custom_config,omitempty"`
}

// FileInfo represents file information
type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	IsDir     bool      `json:"is_dir"`
	ModTime   time.Time `json:"mod_time"`
	Extension string    `json:"extension,omitempty"`
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	ProjectID     string    `json:"project_id"`
	Enabled       bool      `json:"enabled"`
	Schedule      string    `json:"schedule"` // cron format
	Destinations  []string  `json:"destinations"` // local, gdrive
	RetentionDays int       `json:"retention_days"`
	LastBackup    time.Time `json:"last_backup,omitempty"`
}

// BackupInfo represents backup information
type BackupInfo struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	Filename  string    `json:"filename"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	Type      string    `json:"type"` // full, files, database
}

// ProjectOverview represents the overall project status
type ProjectOverview struct {
	Projects      []Project `json:"projects"`
	TotalProjects int       `json:"total_projects"`
	ActiveProjects int      `json:"active_projects"`
	TotalSize     int64     `json:"total_size"`
	FirstTimeSetup bool     `json:"first_time_setup"`
}

// CreateProjectRequest represents a project creation request
type CreateProjectRequest struct {
	Name         string `json:"name"`
	Domain       string `json:"domain"`
	CreateDB     bool   `json:"create_db"`
	EnableSSL    bool   `json:"enable_ssl"`
	EnableBackup bool   `json:"enable_backup"`
}

// UploadProgress represents file upload progress
type UploadProgress struct {
	ProjectID    string `json:"project_id"`
	Filename     string `json:"filename"`
	Progress     int    `json:"progress"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	Error        string `json:"error,omitempty"`
}
