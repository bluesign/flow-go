package dns

import (
	"context"
	"net"
	"time"
	_ "unsafe" // for linking runtimeNano

	madns "github.com/multiformats/go-multiaddr-dns"

	"github.com/onflow/flow-go/module"
)

//go:linkname runtimeNano runtime.nanotime
func runtimeNano() int64

// Resolver is a cache-based dns resolver for libp2p.
type Resolver struct {
	c         *cache
	res       madns.BasicResolver // underlying resolver
	collector module.ResolverMetrics
}

// optFunc is the option function for Resolver.
type optFunc func(resolver *Resolver)

// WithBasicResolver is an option function for setting the basic resolver of this Resolver.
func WithBasicResolver(basic madns.BasicResolver) func(resolver *Resolver) {
	return func(resolver *Resolver) {
		resolver.res = basic
	}
}

// WithTTL is an option function for setting the time to live for cache entries.
func WithTTL(ttl time.Duration) func(resolver *Resolver) {
	return func(resolver *Resolver) {
		resolver.c.ttl = ttl
	}
}

// NewResolver is the factory function for creating an instance of this resolver.
func NewResolver(collector module.ResolverMetrics, opts ...optFunc) (*madns.Resolver, error) {
	resolver := &Resolver{
		res:       madns.DefaultResolver,
		c:         newCache(),
		collector: collector,
	}

	for _, opt := range opts {
		opt(resolver)
	}

	return madns.NewResolver(madns.WithDefaultResolver(resolver))
}

// LookupIPAddr implements BasicResolver interface for libp2p for looking up ip addresses through resolver.
func (r *Resolver) LookupIPAddr(ctx context.Context, domain string) ([]net.IPAddr, error) {
	started := runtimeNano()

	addr, err := r.lookupIPAddr(ctx, domain)

	r.collector.DNSLookupDuration(
		time.Duration(runtimeNano() - started))
	return addr, err
}

// lookupIPAddr encapsulates the logic of resolving an ip address through cache.
func (r *Resolver) lookupIPAddr(ctx context.Context, domain string) ([]net.IPAddr, error) {
	if addr, ok := r.c.resolveIPCache(domain); ok {
		// resolving address from cache
		r.collector.OnDNSCacheHit()
		return addr, nil
	}

	// resolves domain through underlying resolver
	r.collector.OnDNSCacheMiss()
	addr, err := r.res.LookupIPAddr(ctx, domain)
	if err != nil {
		return nil, err
	}

	r.c.updateIPCache(domain, addr) // updates cache

	return addr, nil
}

// LookupTXT implements BasicResolver interface for libp2p.
func (r *Resolver) LookupTXT(ctx context.Context, txt string) ([]string, error) {

	started := runtimeNano()

	addr, err := r.lookupTXT(ctx, txt)

	r.collector.DNSLookupDuration(
		time.Duration(runtimeNano() - started))
	return addr, err
}

func (r *Resolver) lookupTXT(ctx context.Context, txt string) ([]string, error) {
	if addr, ok := r.c.resolveTXTCache(txt); ok {
		// resolving address from cache
		r.collector.OnDNSCacheHit()
		return addr, nil
	}

	// resolves txt through underlying resolver
	r.collector.OnDNSCacheMiss()
	addr, err := r.res.LookupTXT(ctx, txt)
	if err != nil {
		return nil, err
	}

	r.c.updateTXTCache(txt, addr) // updates cache

	return addr, err
}
