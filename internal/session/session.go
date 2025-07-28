package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"server-panel/internal/database"
)

// Store 会话存储
type Store struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
	dbManager *database.Manager
}

// Session 会话
type Session struct {
	ID       string                 `json:"id"`
	Data     map[string]interface{} `json:"data"`
	Expires  time.Time              `json:"expires"`
	modified bool                   `json:"-"` // 标记是否已修改，不序列化
}

// NewStore 创建新的会话存储
func NewStore() *Store {
	// 使用统一的数据库管理器
	dbManager := database.NewManager()

	store := &Store{
		sessions:  make(map[string]*Session),
		dbManager: dbManager,
	}

	// 加载现有会话到内存
	store.loadSessions()

	// 启动清理 goroutine
	go store.cleanup()

	return store
}



// Get 获取会话
func (s *Store) Get(r *http.Request) (*Session, error) {
	// 从 Cookie 中获取会话 ID
	cookie, err := r.Cookie("session_id")
	if err != nil {
		// 创建新会话
		return s.createSession(), nil
	}
	
	s.mutex.RLock()
	session, exists := s.sessions[cookie.Value]
	s.mutex.RUnlock()
	
	if !exists || session.Expires.Before(time.Now()) {
		// 会话不存在或已过期，使用现有 cookie ID 创建新会话
		return s.createSessionWithID(cookie.Value), nil
	}
	
	// 延长会话过期时间
	session.Expires = time.Now().Add(24 * time.Hour)
	
	return session, nil
}

// createSession 创建新会话
func (s *Store) createSession() *Session {
	sessionID := generateSessionID()

	session := &Session{
		ID:      sessionID,
		Data:    make(map[string]interface{}),
		Expires: time.Now().Add(24 * time.Hour),
	}

	s.mutex.Lock()
	s.sessions[sessionID] = session
	s.saveSession(session) // 保存到数据库
	s.mutex.Unlock()

	return session
}

// createSessionWithID 使用指定 ID 创建新会话
func (s *Store) createSessionWithID(sessionID string) *Session {
	session := &Session{
		ID:      sessionID,
		Data:    make(map[string]interface{}),
		Expires: time.Now().Add(24 * time.Hour),
	}

	s.mutex.Lock()
	s.sessions[sessionID] = session
	s.saveSession(session) // 保存到数据库
	s.mutex.Unlock()

	return session
}

// Save 保存会话
func (s *Session) Save(w http.ResponseWriter) error {
	// 设置 Cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    s.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // 在生产环境中应该设置为 true
		SameSite: http.SameSiteLaxMode, // 改为Lax模式以支持SSE
		Expires:  s.Expires,
	}

	http.SetCookie(w, cookie)
	return nil
}

// SaveWithStore 保存会话到数据库和Cookie
func (s *Session) SaveWithStore(w http.ResponseWriter, store *Store) error {
	// 只有在数据被修改时才保存到数据库
	if s.modified {
		if err := store.saveSession(s); err != nil {
			return err
		}
		s.modified = false // 重置修改标记
	}

	// 设置 Cookie
	return s.Save(w)
}

// Get 获取会话数据
func (s *Session) Get(key string) interface{} {
	return s.Data[key]
}

// Set 设置会话数据
func (s *Session) Set(key string, value interface{}) {
	s.Data[key] = value
	s.modified = true // 标记为已修改
}

// SetWithStore 设置会话数据并立即保存到数据库
func (s *Session) SetWithStore(key string, value interface{}, store *Store) error {
	s.Data[key] = value
	return store.saveSession(s)
}

// Delete 删除会话数据
func (s *Session) Delete(key string) {
	delete(s.Data, key)
}

// cleanup 清理过期会话
func (s *Store) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()

		// 清理内存中的过期会话
		s.mutex.Lock()
		for id, session := range s.sessions {
			if session.Expires.Before(now) {
				delete(s.sessions, id)
			}
		}
		s.mutex.Unlock()

		// 清理数据库中的过期会话
		db := s.dbManager.GetDB()
		db.Exec("DELETE FROM sessions WHERE expires < ?", now)
	}
}

// generateSessionID 生成会话 ID
func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}



// loadSessions 从数据库加载会话到内存
func (s *Store) loadSessions() {
	db := s.dbManager.GetDB()
	rows, err := db.Query("SELECT id, data, expires FROM sessions WHERE expires > ?", time.Now())
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, dataStr string
		var expires time.Time

		if err := rows.Scan(&id, &dataStr, &expires); err != nil {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			continue
		}

		session := &Session{
			ID:      id,
			Data:    data,
			Expires: expires,
		}

		s.sessions[id] = session
	}
}

// saveSession 保存会话到数据库
func (s *Store) saveSession(session *Session) error {
	dataBytes, err := json.Marshal(session.Data)
	if err != nil {
		return err
	}

	db := s.dbManager.GetDB()
	_, err = db.Exec(`
		INSERT OR REPLACE INTO sessions (id, data, expires, updated_at)
		VALUES (?, ?, ?, CURRENT_TIMESTAMP)
	`, session.ID, string(dataBytes), session.Expires)

	return err
}

// deleteSession 从数据库删除会话
func (s *Store) deleteSession(sessionID string) error {
	db := s.dbManager.GetDB()
	_, err := db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

// DestroySession 销毁会话（公开方法）
func (s *Store) DestroySession(sessionID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 从内存中删除
	delete(s.sessions, sessionID)

	// 从数据库中删除
	return s.deleteSession(sessionID)
}
