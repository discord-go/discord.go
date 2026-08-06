package cache

import (
	"sync"
	"time"
)

type item struct {
	value      any
	expiration int64 // unix nano
}

// MemoryCache is a thread-safe in-memory cache implementation.
type MemoryCache struct {
	mu      sync.RWMutex
	items   map[string]item
	options *Options
}

// NewMemoryCache creates a new MemoryCache with the given options.
func NewMemoryCache(opts ...Option) *MemoryCache {
	options := DefaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &MemoryCache{
		items:   make(map[string]item),
		options: options,
	}
}

// Get retrieves a value from the cache.
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	i, found := c.items[key]
	if !found {
		return nil, false
	}

	if i.expiration > 0 && time.Now().UnixNano() > i.expiration {
		delete(c.items, key)
		return nil, false
	}

	return i.value, true
}

// Set adds or updates a value in the cache.
func (c *MemoryCache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp int64
	if c.options.TTL > 0 {
		exp = time.Now().Add(c.options.TTL).UnixNano()
	}

	// Simple eviction policy: evict a random element if full
	if c.options.MaxSize > 0 && len(c.items) >= c.options.MaxSize {
		// Only delete if the key is not already present, or if it is present it would just replace it.
		// So if the key isn't in there, we need to free 1 space.
		if _, ok := c.items[key]; !ok {
			for k := range c.items {
				delete(c.items, k)
				break
			}
		}
	}

	c.items[key] = item{
		value:      value,
		expiration: exp,
	}
}

// Delete removes a value from the cache.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, key)
}

// Clear empties the cache.
func (c *MemoryCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]item)
}

// CleanUp removes expired items from the cache.
func (c *MemoryCache) CleanUp() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.expiration > 0 && now > v.expiration {
			delete(c.items, k)
		}
	}
}

func (c *MemoryCache) GetGuild(id string) (any, bool) { return c.Get("guild:" + id) }
func (c *MemoryCache) SetGuild(id string, guild any)  { c.Set("guild:"+id, guild) }
func (c *MemoryCache) DeleteGuild(id string)          { c.Delete("guild:" + id) }

func (c *MemoryCache) GetChannel(id string) (any, bool)  { return c.Get("channel:" + id) }
func (c *MemoryCache) SetChannel(id string, channel any) { c.Set("channel:"+id, channel) }
func (c *MemoryCache) DeleteChannel(id string)           { c.Delete("channel:" + id) }

func (c *MemoryCache) GetUser(id string) (any, bool) { return c.Get("user:" + id) }
func (c *MemoryCache) SetUser(id string, user any)   { c.Set("user:"+id, user) }
func (c *MemoryCache) DeleteUser(id string)          { c.Delete("user:" + id) }

func (c *MemoryCache) GetRole(id string) (any, bool) { return c.Get("role:" + id) }
func (c *MemoryCache) SetRole(id string, role any)   { c.Set("role:"+id, role) }
func (c *MemoryCache) DeleteRole(id string)          { c.Delete("role:" + id) }

func (c *MemoryCache) GetMessage(id string) (any, bool)  { return c.Get("message:" + id) }
func (c *MemoryCache) SetMessage(id string, message any) { c.Set("message:"+id, message) }
func (c *MemoryCache) DeleteMessage(id string)           { c.Delete("message:" + id) }

func (c *MemoryCache) GetMember(guildID, userID string) (any, bool) {
	return c.Get("member:" + guildID + ":" + userID)
}
func (c *MemoryCache) SetMember(guildID, userID string, member any) {
	c.Set("member:"+guildID+":"+userID, member)
}
func (c *MemoryCache) DeleteMember(guildID, userID string) {
	c.Delete("member:" + guildID + ":" + userID)
}
