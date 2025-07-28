package environment

import "time"

// ServiceStatus represents the status of a service
type ServiceStatus string

const (
	StatusInstalled    ServiceStatus = "installed"
	StatusNotInstalled ServiceStatus = "not_installed"
	StatusInstalling   ServiceStatus = "installing"
	StatusUninstalling ServiceStatus = "uninstalling"
	StatusError        ServiceStatus = "error"
)

// Service represents a system service/environment
type Service struct {
	Name        string        `json:"name"`
	DisplayName string        `json:"display_name"`
	Status      ServiceStatus `json:"status"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Icon        string        `json:"icon"`
	IsRunning   bool          `json:"is_running"`
	Port        int           `json:"port,omitempty"`
	ConfigPath  string        `json:"config_path,omitempty"`
}

// PHPExtension represents a PHP extension
type PHPExtension struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
	Installed   bool   `json:"installed"`
	Description string `json:"description"`
}

// InstallationProgress represents the progress of an installation
type InstallationProgress struct {
	ServiceName string    `json:"service_name"`
	Progress    int       `json:"progress"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	StartTime   time.Time `json:"start_time"`
	Error       string    `json:"error,omitempty"`
}

// BulkInstallRequest represents a bulk installation request
type BulkInstallRequest struct {
	Services []string `json:"services"`
	Confirm  bool     `json:"confirm"`
}

// EnvironmentOverview represents the overall environment status
type EnvironmentOverview struct {
	Services          []Service             `json:"services"`
	PHPExtensions     []PHPExtension        `json:"php_extensions,omitempty"`
	InstallProgress   *InstallationProgress `json:"install_progress,omitempty"`
	FirstTimeSetup    bool                  `json:"first_time_setup"`
	RecommendedSetup  []string              `json:"recommended_setup"`
}
