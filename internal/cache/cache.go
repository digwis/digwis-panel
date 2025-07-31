package cache

import (
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Data      interface{}
	ExpiresAt time.Time
}

// Cache 通用缓存系统
type Cache struct {
	items map[string]*CacheItem
	mutex sync.RWMutex
}

// NewCache 创建新的缓存实例
func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]*CacheItem),
	}
	
	// 启动清理协程
	go cache.cleanup()
	
	return cache
}

// Set 设置缓存项
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.items[key] = &CacheItem{
		Data:      value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get 获取缓存项
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	
	item, exists := c.items[key]
	if !exists {
		return nil, false
	}
	
	// 检查是否过期
	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	
	return item.Data, true
}

// Delete 删除缓存项
func (c *Cache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	delete(c.items, key)
}

// Clear 清空所有缓存
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	
	c.items = make(map[string]*CacheItem)
}

// cleanup 定期清理过期缓存
func (c *Cache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mutex.Lock()
		now := time.Now()
		for key, item := range c.items {
			if now.After(item.ExpiresAt) {
				delete(c.items, key)
			}
		}
		c.mutex.Unlock()
	}
}

// 缓存TTL配置
var CacheTTL = map[string]time.Duration{
	"system_stats":     5 * time.Second,   // 系统统计
	"system_overview":  10 * time.Second,  // 系统概览
	"system_details":   15 * time.Second,  // 系统详情
	"environment":      30 * time.Second,  // 环境信息
	"projects":         10 * time.Second,  // 项目列表
	"processes":        5 * time.Second,   // 进程列表
}

// 全局缓存实例
var GlobalCache = NewCache()
