package dns

import (
	"net"
	"sync"
	"time"
)

// defaultTimeToLive is the default duration a dns result is cached.
const defaultTimeToLive = 5 * time.Minute

// cache is a ttl-based cache for dns entries
type cache struct {
	sync.Mutex
	ttl      time.Duration // time-to-live for cache entry
	ipCache  map[string]*ipCacheEntry
	txtCache map[string]*txtCacheEntry
}

func newCache() *cache {
	return &cache{
		ttl:      defaultTimeToLive,
		ipCache:  make(map[string]*ipCacheEntry),
		txtCache: make(map[string]*txtCacheEntry),
	}
}

// resolveIPCache resolves the domain through the cache if it is available.
func (c *cache) resolveIPCache(domain string) ([]net.IPAddr, bool) {
	c.Lock()
	defer c.Unlock()

	entry, ok := c.ipCache[domain]

	if !ok {
		return nil, false
	}

	if time.Duration(runtimeNano()-entry.timestamp) > c.ttl {
		// invalidates cache entry
		delete(c.ipCache, domain)
		return nil, false
	}

	return entry.addresses, true
}

// resolveIPCache resolves the txt through the cache if it is available.
func (c *cache) resolveTXTCache(txt string) ([]string, bool) {
	c.Lock()
	defer c.Unlock()

	entry, ok := c.txtCache[txt]

	if !ok {
		return nil, false
	}

	if time.Duration(runtimeNano()-entry.timestamp) > c.ttl {
		// invalidates cache entry
		delete(c.txtCache, txt)
		return nil, false
	}

	return entry.addresses, true
}

// updateIPCache updates the cache entry for the domain.
func (c *cache) updateIPCache(domain string, addr []net.IPAddr) {
	c.Lock()
	defer c.Unlock()

	c.ipCache[domain] = &ipCacheEntry{
		addresses: addr,
		timestamp: runtimeNano(),
	}
}

// updateTXTCache updates the cache entry for the txt.
func (c *cache) updateTXTCache(txt string, addr []string) {
	c.Lock()
	defer c.Unlock()

	c.txtCache[txt] = &txtCacheEntry{
		addresses: addr,
		timestamp: runtimeNano(),
	}
}
