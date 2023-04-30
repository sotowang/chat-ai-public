package utils

import (
	"sync"
	"time"
)

var AppKeyLock *Lock

type Lock struct {
	mu      sync.Mutex
	locks   map[string]*lockInfo
	timeout time.Duration
}

type lockInfo struct {
	value string
	exp   time.Time
}

func NewLock(timeout time.Duration) *Lock {
	return &Lock{
		locks:   make(map[string]*lockInfo),
		timeout: timeout,
	}
}

func (l *Lock) Lock(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if the lock already exists and hasn't timed out
	if lock, ok := l.locks[key]; ok {
		if lock.exp.After(time.Now()) {
			return false
		}
	}

	// Create a new lock and set the expiration time
	l.locks[key] = &lockInfo{
		value: "",
		exp:   time.Now().Add(l.timeout),
	}
	return true
}

func (l *Lock) Unlock(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.locks, key)
}
