package system

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"server-panel/internal/cache"
)

// Monitor 系统监控器
type Monitor struct {
	cache      *dataCache
	cacheMutex sync.RWMutex
}

// dataCache 数据缓存
type dataCache struct {
	systemStats   *SystemStats
	processCache  []Process
	diskCache     DiskStats
	lastUpdate    time.Time
	lastDiskUpdate time.Time
	lastProcessUpdate time.Time
	cacheDuration time.Duration
	diskCacheDuration time.Duration
	processCacheDuration time.Duration
}

// NewMonitor 创建新的监控器实例
func NewMonitor() *Monitor {
	return &Monitor{
		cache: &dataCache{
			cacheDuration: 5 * time.Second,        // 基础数据缓存5秒
			diskCacheDuration: 30 * time.Second,   // 磁盘数据缓存30秒
			processCacheDuration: 15 * time.Second, // 进程数据缓存15秒
		},
	}
}

// SystemStats 系统统计信息
type SystemStats struct {
	CPU        CPUStats    `json:"cpu"`
	Memory     MemoryStats `json:"memory"`
	Disk       DiskStats   `json:"disk"`
	Network    NetworkStats `json:"network"`
	LoadAvg    LoadAvg     `json:"load_avg"`
	Uptime     string      `json:"uptime"`
	Hostname   string      `json:"hostname"`
	OS         string      `json:"os"`
	Kernel     string      `json:"kernel"`
	Timestamp  time.Time   `json:"timestamp"`
}

// CPUStats CPU统计
type CPUStats struct {
	Usage     float64 `json:"usage"`
	Cores     int     `json:"cores"`
	Model     string  `json:"model"`
	Frequency string  `json:"frequency"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	Total     uint64  `json:"total"`
	Used      uint64  `json:"used"`
	Free      uint64  `json:"free"`
	Available uint64  `json:"available"`
	Usage     float64 `json:"usage"`
	Swap      SwapStats `json:"swap"`
}

// SwapStats 交换分区统计
type SwapStats struct {
	Total uint64  `json:"total"`
	Used  uint64  `json:"used"`
	Free  uint64  `json:"free"`
	Usage float64 `json:"usage"`
}

// DiskStats 磁盘统计
type DiskStats struct {
	Total uint64  `json:"total"`
	Used  uint64  `json:"used"`
	Free  uint64  `json:"free"`
	Usage float64 `json:"usage"`
}

// NetworkStats 网络统计
type NetworkStats struct {
	BytesReceived uint64 `json:"bytes_received"`
	BytesSent     uint64 `json:"bytes_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	PacketsSent   uint64 `json:"packets_sent"`
}

// LoadAvg 系统负载
type LoadAvg struct {
	Load1  float64 `json:"load1"`
	Load5  float64 `json:"load5"`
	Load15 float64 `json:"load15"`
}

// Process 进程信息
type Process struct {
	PID     int     `json:"pid"`
	Name    string  `json:"name"`
	User    string  `json:"user"`
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Status  string  `json:"status"`
	Command string  `json:"command"`
}

// Service 服务信息
type Service struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Enabled     bool   `json:"enabled"`
	Description string `json:"description"`
}

// 注意：此处有重复的NewMonitor函数声明，已注释掉
// func NewMonitor() *Monitor {
// 	return &Monitor{}
// }

// GetSystemStats 获取系统统计信息（带缓存）
func (m *Monitor) GetSystemStats() (*SystemStats, error) {
	// 先检查全局缓存
	if cached, found := cache.GlobalCache.Get("system_stats"); found {
		if stats, ok := cached.(*SystemStats); ok {
			// 更新时间戳
			statsCopy := *stats
			statsCopy.Timestamp = time.Now()
			return &statsCopy, nil
		}
	}

	// 缓存未命中，获取新数据
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// 双重检查全局缓存
	if cached, found := cache.GlobalCache.Get("system_stats"); found {
		if stats, ok := cached.(*SystemStats); ok {
			statsCopy := *stats
			statsCopy.Timestamp = time.Now()
			return &statsCopy, nil
		}
	}

	stats := &SystemStats{
		Timestamp: time.Now(),
	}

	// 获取主机名（缓存静态信息）
	if m.cache.systemStats == nil {
		hostname, _ := os.Hostname()
		stats.Hostname = hostname
		stats.OS = m.getOSInfo()
		stats.Kernel = m.getKernelVersion()
	} else {
		// 复用静态信息
		stats.Hostname = m.cache.systemStats.Hostname
		stats.OS = m.cache.systemStats.OS
		stats.Kernel = m.cache.systemStats.Kernel
	}

	// 获取动态信息
	stats.CPU = m.getCPUStats()
	stats.Memory = m.getMemoryStats()
	stats.Network = m.getNetworkStats()
	stats.LoadAvg = m.getLoadAvg()
	stats.Uptime = m.getUptime()

	// 获取磁盘信息（使用独立缓存）
	stats.Disk = m.getDiskStatsWithCache()

	// 更新本地缓存
	m.cache.systemStats = stats
	m.cache.lastUpdate = time.Now()

	// 更新全局缓存
	cache.GlobalCache.Set("system_stats", stats, cache.CacheTTL["system_stats"])

	return stats, nil
}

// getCPUStats 获取CPU统计
func (m *Monitor) getCPUStats() CPUStats {
	stats := CPUStats{
		Cores: runtime.NumCPU(),
	}

	// 读取CPU使用率
	if usage := m.getCPUUsage(); usage >= 0 {
		stats.Usage = usage
	}

	// 读取CPU信息
	if info := m.getCPUInfo(); info != "" {
		stats.Model = info
	}

	return stats
}

// getCPUUsage 获取CPU使用率
func (m *Monitor) getCPUUsage() float64 {
	// 简化版本：读取/proc/loadavg
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return -1
	}

	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		if load, err := strconv.ParseFloat(fields[0], 64); err == nil {
			// 将负载转换为百分比（简化计算）
			return (load / float64(runtime.NumCPU())) * 100
		}
	}

	return -1
}

// getCPUInfo 获取CPU信息
func (m *Monitor) getCPUInfo() string {
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

// getMemoryStats 获取内存统计
func (m *Monitor) getMemoryStats() MemoryStats {
	stats := MemoryStats{}

	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return stats
	}

	lines := strings.Split(string(data), "\n")
	memInfo := make(map[string]uint64)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			key := strings.TrimSuffix(fields[0], ":")
			if value, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				memInfo[key] = value * 1024 // 转换为字节
			}
		}
	}

	stats.Total = memInfo["MemTotal"]
	stats.Free = memInfo["MemFree"]
	stats.Available = memInfo["MemAvailable"]
	stats.Used = stats.Total - stats.Available

	if stats.Total > 0 {
		stats.Usage = float64(stats.Used) / float64(stats.Total) * 100
	}

	// 交换分区信息
	stats.Swap.Total = memInfo["SwapTotal"]
	stats.Swap.Free = memInfo["SwapFree"]
	stats.Swap.Used = stats.Swap.Total - stats.Swap.Free

	if stats.Swap.Total > 0 {
		stats.Swap.Usage = float64(stats.Swap.Used) / float64(stats.Swap.Total) * 100
	}

	return stats
}

// getDiskStatsWithCache 获取磁盘统计（带缓存）
func (m *Monitor) getDiskStatsWithCache() DiskStats {
	// 检查磁盘缓存
	if time.Since(m.cache.lastDiskUpdate) < m.cache.diskCacheDuration {
		return m.cache.diskCache
	}

	// 缓存过期，重新获取
	stats := m.getDiskStats()
	m.cache.diskCache = stats
	m.cache.lastDiskUpdate = time.Now()
	return stats
}

// getDiskStats 获取磁盘统计
func (m *Monitor) getDiskStats() DiskStats {
	stats := DiskStats{}

	// 添加超时机制，减少超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "df", "-B1", "/")
	output, err := cmd.Output()
	if err != nil {
		return stats
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) >= 2 {
		fields := strings.Fields(lines[1])
		if len(fields) >= 4 {
			if total, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
				stats.Total = total
			}
			if used, err := strconv.ParseUint(fields[2], 10, 64); err == nil {
				stats.Used = used
			}
			if free, err := strconv.ParseUint(fields[3], 10, 64); err == nil {
				stats.Free = free
			}

			if stats.Total > 0 {
				stats.Usage = float64(stats.Used) / float64(stats.Total) * 100
			}
		}
	}

	return stats
}

// getNetworkStats 获取网络统计
func (m *Monitor) getNetworkStats() NetworkStats {
	stats := NetworkStats{}

	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return stats
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") && !strings.Contains(line, "lo:") {
			parts := strings.Split(line, ":")
			if len(parts) == 2 {
				fields := strings.Fields(parts[1])
				if len(fields) >= 9 {
					if rx, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
						stats.BytesReceived += rx
					}
					if tx, err := strconv.ParseUint(fields[8], 10, 64); err == nil {
						stats.BytesSent += tx
					}
				}
			}
		}
	}

	return stats
}

// getLoadAvg 获取系统负载
func (m *Monitor) getLoadAvg() LoadAvg {
	loadAvg := LoadAvg{}

	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return loadAvg
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		if load1, err := strconv.ParseFloat(fields[0], 64); err == nil {
			loadAvg.Load1 = load1
		}
		if load5, err := strconv.ParseFloat(fields[1], 64); err == nil {
			loadAvg.Load5 = load5
		}
		if load15, err := strconv.ParseFloat(fields[2], 64); err == nil {
			loadAvg.Load15 = load15
		}
	}

	return loadAvg
}

// getUptime 获取系统运行时间
func (m *Monitor) getUptime() string {
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(data))
	if len(fields) > 0 {
		if seconds, err := strconv.ParseFloat(fields[0], 64); err == nil {
			duration := time.Duration(seconds) * time.Second
			days := int(duration.Hours()) / 24
			hours := int(duration.Hours()) % 24
			minutes := int(duration.Minutes()) % 60

			if days > 0 {
				return fmt.Sprintf("%d天 %d小时 %d分钟", days, hours, minutes)
			} else if hours > 0 {
				return fmt.Sprintf("%d小时 %d分钟", hours, minutes)
			} else {
				return fmt.Sprintf("%d分钟", minutes)
			}
		}
	}

	return ""
}

// getKernelVersion 获取内核版本
func (m *Monitor) getKernelVersion() string {
	data, err := os.ReadFile("/proc/version")
	if err != nil {
		return ""
	}

	fields := strings.Fields(string(data))
	if len(fields) >= 3 {
		return fields[2]
	}

	return ""
}

// getOSInfo 获取操作系统信息
func (m *Monitor) getOSInfo() string {
	// 尝试读取 /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		var name, version string

		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				name = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
			if strings.HasPrefix(line, "NAME=") && name == "" {
				name = strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
			}
			if strings.HasPrefix(line, "VERSION=") && version == "" {
				version = strings.Trim(strings.TrimPrefix(line, "VERSION="), "\"")
			}
		}

		if name != "" {
			if version != "" && !strings.Contains(name, version) {
				return fmt.Sprintf("%s %s", name, version)
			}
			return name
		}
	}

	// 备用方案：尝试读取 /etc/issue
	if data, err := os.ReadFile("/etc/issue"); err == nil {
		issue := strings.TrimSpace(string(data))
		if issue != "" && !strings.Contains(issue, "\\") {
			return strings.Fields(issue)[0]
		}
	}

	// 最后备用方案
	return fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
}

// GetSystemDetails 获取详细系统信息
func (m *Monitor) GetSystemDetails() map[string]string {
	details := make(map[string]string)

	// 基本信息
	details["hostname"], _ = os.Hostname()
	details["os"] = m.getOSInfo()
	details["kernel"] = m.getKernelVersion()
	details["architecture"] = runtime.GOARCH
	details["cpu_cores"] = fmt.Sprintf("%d", runtime.NumCPU())

	// 运行时间
	details["uptime"] = m.getUptime()

	// CPU信息
	details["cpu_model"] = m.getCPUInfo()

	// 内存信息
	memStats := m.getMemoryStats()
	details["total_memory"] = m.formatBytes(memStats.Total)
	details["available_memory"] = m.formatBytes(memStats.Available)

	// 磁盘信息
	diskStats := m.getDiskStats()
	details["total_disk"] = m.formatBytes(diskStats.Total)
	details["available_disk"] = m.formatBytes(diskStats.Free)

	return details
}

// formatBytes 格式化字节数
func (m *Monitor) formatBytes(bytes uint64) string {
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

// GetProcessList 获取进程列表（带缓存和限制）
func (m *Monitor) GetProcessList() ([]Process, error) {
	m.cacheMutex.RLock()
	// 检查进程缓存
	if len(m.cache.processCache) > 0 && time.Since(m.cache.lastProcessUpdate) < m.cache.processCacheDuration {
		processes := make([]Process, len(m.cache.processCache))
		copy(processes, m.cache.processCache)
		m.cacheMutex.RUnlock()
		return processes, nil
	}
	m.cacheMutex.RUnlock()

	// 缓存过期，重新获取
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// 双重检查
	if len(m.cache.processCache) > 0 && time.Since(m.cache.lastProcessUpdate) < m.cache.processCacheDuration {
		processes := make([]Process, len(m.cache.processCache))
		copy(processes, m.cache.processCache)
		return processes, nil
	}

	processes, err := m.getProcessListInternal()
	if err != nil {
		return nil, err
	}

	// 更新缓存
	m.cache.processCache = processes
	m.cache.lastProcessUpdate = time.Now()

	return processes, nil
}

// getProcessListInternal 内部获取进程列表
func (m *Monitor) getProcessListInternal() ([]Process, error) {
	// 减少超时时间，优化性能
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用更高效的ps命令，只获取需要的字段，限制输出行数
	cmd := exec.CommandContext(ctx, "ps", "axo", "pid,user,%cpu,%mem,stat,comm", "--sort=-%cpu", "--no-headers")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var processes []Process
	lines := strings.Split(string(output), "\n")

	// 限制进程数量，只取前50个进程以提高性能
	maxProcesses := 50
	count := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || count >= maxProcesses {
			break
		}

		fields := strings.Fields(line)
		if len(fields) >= 6 {
			pid, _ := strconv.Atoi(fields[0])
			cpu, _ := strconv.ParseFloat(fields[2], 64)
			memory, _ := strconv.ParseFloat(fields[3], 64)

			process := Process{
				PID:     pid,
				User:    fields[1],
				CPU:     cpu,
				Memory:  memory,
				Status:  fields[4],
				Name:    fields[5],
				Command: fields[5], // 简化命令显示
			}

			processes = append(processes, process)
			count++
		}
	}

	return processes, nil
}

// GetServiceList 获取服务列表
func (m *Monitor) GetServiceList() ([]Service, error) {
	cmd := exec.Command("systemctl", "list-units", "--type=service", "--no-pager")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var services []Service
	lines := strings.Split(string(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasSuffix(line, ".service") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				service := Service{
					Name:   strings.TrimSuffix(fields[0], ".service"),
					Status: fields[2],
				}

				// 获取服务描述
				if len(fields) > 4 {
					service.Description = strings.Join(fields[4:], " ")
				}

				services = append(services, service)
			}
		}
	}

	return services, nil
}

// RestartService 重启服务
func (m *Monitor) RestartService(serviceName string) error {
	cmd := exec.Command("systemctl", "restart", serviceName)
	return cmd.Run()
}

// StopService 停止服务
func (m *Monitor) StopService(serviceName string) error {
	cmd := exec.Command("systemctl", "stop", serviceName)
	return cmd.Run()
}

// StartService 启动服务
func (m *Monitor) StartService(serviceName string) error {
	cmd := exec.Command("systemctl", "start", serviceName)
	return cmd.Run()
}

// KillProcess 终止进程
func (m *Monitor) KillProcess(pid int) error {
	cmd := exec.Command("kill", "-9", strconv.Itoa(pid))
	return cmd.Run()
}
