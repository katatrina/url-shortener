package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	urlKeyPrefix = "url:"
	urlCacheTTL  = 1 * time.Hour
)

// CachedURL contains only the fields needed for redirect.
// This is what gets stored in Redis - keep it minimal.
type CachedURL struct {
	ID          string `json:"id"`
	OriginalURL string `json:"originalUrl"`

	// Pointer (*int64): nil means "no expiry", distinguishing it from zero value.
	// Unix timestamp: compact and timezone-safe compared to time.Time string.
	// omitempty: when nil, the field is omitted from JSON entirely to save Redis memory.
	ExpiresAt *int64 `json:"expiresAt,omitempty"`
}

type URLCache struct {
	rdb *redis.Client
}

func NewURLCache(rdb *redis.Client) *URLCache {
	return &URLCache{rdb}
}

func urlKey(shortCode string) string {
	return fmt.Sprintf("%s%s", urlKeyPrefix, shortCode)
}

// Set stores a URL in cache.
func (c *URLCache) Set(ctx context.Context, shortCode string, cachedURL *CachedURL) error {
	data, err := json.Marshal(cachedURL)
	if err != nil {
		return fmt.Errorf("cache marshal failed: %w", err)
	}

	return c.rdb.Set(ctx, urlKey(shortCode), data, urlCacheTTL).Err()
}

// Get retrieves a cached URL by short code.
// Returns nil, nil on cache miss (not an error).
func (c *URLCache) Get(ctx context.Context, shortCode string) (*CachedURL, error) {
	data, err := c.rdb.Get(ctx, urlKey(shortCode)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss - not an error.
		}
		return nil, fmt.Errorf("cache get failed: %w", err) // Infrastructure error (Redis down, timeout, network)
	}

	var cached CachedURL
	if err = json.Unmarshal(data, &cached); err != nil {
		// Cache data is corrupted (e.g., schema change after deployment, manual edit via redis-cli, bug elsewhere marshal wrong data format).
		//
		// Strategy: self-healing cache
		// 1. Delete the bad entry so the next request gets a clean cache miss
		//    instead of repeatedly hitting unmarshal failure until TTL expires.
		// 2. Return nil, nil (treat as cache miss) so the caller falls through
		//    to the database. Cache is not the source of truth - a corrupted entry
		//    should never surface as an error to the user.
		// 3. Del error is intentionally discarded (best-effort cleanup).
		//    If Del also fails, TTL will eventually evict the bad entry anyway.
		_ = c.rdb.Del(ctx, urlKey(shortCode)).Err()
		return nil, nil
	}

	return &cached, nil
}

// Delete removes a URL from cache by short code.
//
// Note:
//  - Deleting (DEL) a non-existent key in Redis will not return any error - it returns 0 (no keys were deleted).
//  - Cache invalidation is an idempotent operation - calling it once or ten times should return the same result: the key disappears.
//  - We don't care whether it "existed and was deleted" or "never existed in the first place".
//  - There are many reasons why a key might not exist in the cache at the time you call Del: the TTL has expired, the cache was evicted due to memory shortage, Redis just restarted, or simply the key was never cached. All of these are normal occurrences, not errors, and do not require special handling.
//  - Non-nil errors here are infrastructure-level only (network, timeout, etc.).
func (c *URLCache) Delete(ctx context.Context, shortCode string) error {
	return c.rdb.Del(ctx, urlKey(shortCode)).Err()
}
