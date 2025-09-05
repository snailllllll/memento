package utils

import (
	"errors"
	"sync"
	"time"
)

// cacheLock 本地内存缓存锁实现
type cacheLock struct {
	mu    sync.Mutex
	locks map[string]int64 // 存储锁的过期时间戳
}

var (
	instance *cacheLock
	once     sync.Once
)

// initCacheLock 初始化缓存锁单例
func initCacheLock() {
	once.Do(func() {
		instance = &cacheLock{
			locks: make(map[string]int64),
		}
	})
}

// SetLock 创建并返回一个锁对象
func SetLock(key string, ttl time.Duration) *cacheLock {
	initCacheLock()
	return &cacheLock{}
}

// LockExists 检查锁是否存在且未过期（包级函数）
func LockExists(key string) bool {
	initCacheLock()
	return instance.lockExists(key)
}

// DeleteLock 删除指定锁（包级函数）
func DeleteLock(key string) {
	initCacheLock()
	instance.deleteLock(key)
}

// TryLock 尝试获取锁，如果锁不存在或已过期则成功上锁，否则返回错误
func TryLock(key string, ttl time.Duration) error {
	initCacheLock()
	return instance.tryLock(key, ttl)
}

// setLock 设置锁并指定过期时间（私有方法）
func (c *cacheLock) setLock(key string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 计算过期时间戳（毫秒）
	expireAt := time.Now().Add(ttl).UnixNano() / int64(time.Millisecond)
	c.locks[key] = expireAt
}

// lockExists 检查锁是否存在且未过期（私有方法）
func (c *cacheLock) lockExists(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查键是否存在
	expireAt, exists := c.locks[key]
	if !exists {
		return false
	}

	// 检查是否已过期
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)
	if currentTime > expireAt {
		// 自动清理过期锁
		delete(c.locks, key)
		return false
	}
	return true
}

// deleteLock 删除指定锁（私有方法）
func (c *cacheLock) deleteLock(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.locks, key)
}

// tryLock 原子性的检查并设置锁（私有方法）
func (c *cacheLock) tryLock(key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查锁是否存在且未过期
	expireAt, exists := c.locks[key]
	currentTime := time.Now().UnixNano() / int64(time.Millisecond)

	if exists && currentTime <= expireAt {
		return errors.New("lock already exists and is not expired")
	}

	// 锁不存在或已过期，设置新锁
	expireAt = time.Now().Add(ttl).UnixNano() / int64(time.Millisecond)
	c.locks[key] = expireAt

	return nil
}
