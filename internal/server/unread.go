package server

import (
	"sync"
	"time"

	"github.com/larsartmann/webphone/internal/domain"
)

// unreadCache memoizes the per-extension unread total for one shell
// render TTL. Every server-side path that can change a thread's unread
// counter (send, inbound webhook, open-thread mark-read) drops the
// extension's entry, so a badge only ever lags by the TTL when a change
// bypasses this process — which cannot happen: the store is in-process.
type unreadCache struct {
	mu    sync.Mutex
	ttl   time.Duration
	items map[domain.Extension]unreadEntry
}

type unreadEntry struct {
	total   int
	expires time.Time
}

func newUnreadCache(ttl time.Duration) *unreadCache {
	return &unreadCache{ttl: ttl, items: map[domain.Extension]unreadEntry{}}
}

func (c *unreadCache) get(ext domain.Extension) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.items[ext]
	if !ok || time.Now().After(entry.expires) {
		return 0, false
	}
	return entry.total, true
}

func (c *unreadCache) put(ext domain.Extension, total int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[ext] = unreadEntry{total: total, expires: time.Now().Add(c.ttl)}
}

func (c *unreadCache) drop(ext domain.Extension) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, ext)
}
