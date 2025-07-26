package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"time"

	"server-panel/internal/config"
)

// Manager 认证管理器
type Manager struct {
	config   config.AuthConfig
	sessions map[string]*Session
	attempts map[string]*LoginAttempts
	mutex    sync.RWMutex
}

// Session 用户会话
type Session struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	UserID    string    `json:"user_id"`
	LoginTime time.Time `json:"login_time"`
	LastSeen  time.Time `json:"last_seen"`
	IPAddress string    `json:"ip_address"`
}

// LoginAttempts 登录尝试记录
type LoginAttempts struct {
	Count     int       `json:"count"`
	LastTry   time.Time `json:"last_try"`
	LockedUntil time.Time `json:"locked_until"`
}

// User 系统用户信息
type User struct {
	Username string   `json:"username"`
	UID      string   `json:"uid"`
	GID      string   `json:"gid"`
	Home     string   `json:"home"`
	Shell    string   `json:"shell"`
	Groups   []string `json:"groups"`
	IsAdmin  bool     `json:"is_admin"`
}

// NewManager 创建认证管理器
func NewManager(cfg config.AuthConfig) *Manager {
	return &Manager{
		config:   cfg,
		sessions: make(map[string]*Session),
		attempts: make(map[string]*LoginAttempts),
	}
}

// Authenticate 验证用户凭据
func (m *Manager) Authenticate(username, password, ipAddress string) (*Session, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查是否被锁定
	if m.isLocked(ipAddress) {
		return nil, fmt.Errorf("IP地址已被锁定，请稍后再试")
	}

	// 验证系统用户
	user, err := m.validateSystemUser(username, password)
	if err != nil {
		m.recordFailedAttempt(ipAddress)
		return nil, err
	}

	// 清除失败尝试记录
	delete(m.attempts, ipAddress)

	// 创建会话
	session := &Session{
		ID:        m.generateSessionID(),
		Username:  user.Username,
		UserID:    user.UID,
		LoginTime: time.Now(),
		LastSeen:  time.Now(),
		IPAddress: ipAddress,
	}

	m.sessions[session.ID] = session

	log.Printf("用户 %s 从 %s 登录成功", username, ipAddress)
	return session, nil
}

// ValidateSession 验证会话
func (m *Manager) ValidateSession(sessionID string) (*Session, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("会话不存在")
	}

	// 检查会话是否过期
	if time.Since(session.LastSeen) > m.config.SessionTimeout {
		delete(m.sessions, sessionID)
		return nil, fmt.Errorf("会话已过期")
	}

	// 更新最后访问时间
	session.LastSeen = time.Now()
	return session, nil
}

// Logout 退出登录
func (m *Manager) Logout(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if session, exists := m.sessions[sessionID]; exists {
		log.Printf("用户 %s 退出登录", session.Username)
		delete(m.sessions, sessionID)
	}

	return nil
}

// validateSystemUser 验证系统用户
func (m *Manager) validateSystemUser(username, password string) (*User, error) {
	// 获取用户信息
	u, err := user.Lookup(username)
	if err != nil {
		return nil, fmt.Errorf("用户不存在")
	}

	// 检查用户ID（排除系统用户）
	uid, _ := strconv.Atoi(u.Uid)
	if uid < 1000 && uid != 0 {
		return nil, fmt.Errorf("不允许系统用户登录")
	}

	// 检查管理员权限
	isAdmin, groups := m.checkAdminPrivileges(username)
	if !isAdmin {
		return nil, fmt.Errorf("用户没有管理员权限")
	}

	// 验证密码
	if !m.validatePassword(username, password) {
		return nil, fmt.Errorf("密码错误")
	}

	return &User{
		Username: u.Username,
		UID:      u.Uid,
		GID:      u.Gid,
		Home:     u.HomeDir,
		Shell:    u.Name,
		Groups:   groups,
		IsAdmin:  isAdmin,
	}, nil
}

// checkAdminPrivileges 检查管理员权限
func (m *Manager) checkAdminPrivileges(username string) (bool, []string) {
	// root用户直接允许
	if username == "root" {
		return true, []string{"root"}
	}

	// 获取用户组
	cmd := exec.Command("groups", username)
	output, err := cmd.Output()
	if err != nil {
		return false, nil
	}

	groupsStr := strings.TrimSpace(string(output))
	// 格式: "username : group1 group2 group3"
	parts := strings.Split(groupsStr, ":")
	if len(parts) < 2 {
		return false, nil
	}

	groups := strings.Fields(strings.TrimSpace(parts[1]))
	
	// 检查管理员组
	adminGroups := []string{"sudo", "wheel", "admin"}
	for _, group := range groups {
		for _, adminGroup := range adminGroups {
			if group == adminGroup {
				return true, groups
			}
		}
	}

	return false, groups
}

// validatePassword 验证密码
func (m *Manager) validatePassword(username, password string) bool {
	// 使用su命令验证密码（更安全的方式）
	cmd := exec.Command("su", "-c", "exit 0", username)
	cmd.Stdin = strings.NewReader(password + "\n")
	
	// 设置环境变量以避免交互
	cmd.Env = append(os.Environ(), "DEBIAN_FRONTEND=noninteractive")
	
	err := cmd.Run()
	return err == nil
}

// isLocked 检查IP是否被锁定
func (m *Manager) isLocked(ipAddress string) bool {
	attempt, exists := m.attempts[ipAddress]
	if !exists {
		return false
	}

	if attempt.Count >= m.config.MaxLoginAttempts {
		if time.Now().Before(attempt.LockedUntil) {
			return true
		}
		// 锁定时间已过，重置计数
		delete(m.attempts, ipAddress)
	}

	return false
}

// recordFailedAttempt 记录失败尝试
func (m *Manager) recordFailedAttempt(ipAddress string) {
	now := time.Now()
	
	attempt, exists := m.attempts[ipAddress]
	if !exists {
		attempt = &LoginAttempts{}
		m.attempts[ipAddress] = attempt
	}

	attempt.Count++
	attempt.LastTry = now

	if attempt.Count >= m.config.MaxLoginAttempts {
		attempt.LockedUntil = now.Add(m.config.LockoutDuration)
		log.Printf("IP %s 因多次登录失败被锁定至 %v", ipAddress, attempt.LockedUntil)
	}
}

// generateSessionID 生成会话ID
func (m *Manager) generateSessionID() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	hash := sha256.Sum256(bytes)
	return hex.EncodeToString(hash[:])
}

// GetActiveSessions 获取活跃会话列表
func (m *Manager) GetActiveSessions() []*Session {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var sessions []*Session
	for _, session := range m.sessions {
		// 只返回未过期的会话
		if time.Since(session.LastSeen) <= m.config.SessionTimeout {
			sessions = append(sessions, session)
		}
	}

	return sessions
}

// CleanupExpiredSessions 清理过期会话
func (m *Manager) CleanupExpiredSessions() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for id, session := range m.sessions {
		if time.Since(session.LastSeen) > m.config.SessionTimeout {
			delete(m.sessions, id)
		}
	}
}
