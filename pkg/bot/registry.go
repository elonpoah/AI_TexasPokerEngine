package bot

import "sync"

var (
	mu       sync.RWMutex
	registry = make(map[string]Bot)

	defaultBot Bot
)

// 注册 bot
func Register(b Bot) {
	mu.Lock()
	defer mu.Unlock()

	registry[b.Name()] = b
}

// 获取 bot（安全版）
func Get(name string) Bot {
	mu.RLock()
	defer mu.RUnlock()

	b := registry[name]
	if b == nil {
		return defaultBot
	}
	return b
}

// 设置默认 bot（fallback）
func SetDefault(b Bot) {
	mu.Lock()
	defer mu.Unlock()

	defaultBot = b
}

// 列出所有 bot（调试/管理用）
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	var res []string
	for k := range registry {
		res = append(res, k)
	}
	return res
}
