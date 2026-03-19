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
		return nil, fmt.Errorf("cache get failed: %w", err)
	}

	var cached CachedURL
	if err = json.Unmarshal(data, &cached); err != nil {
		// Corrupted cache entry - delete it and treat as miss.
		_ = c.rdb.Del(ctx, urlKey(shortCode)).Err()
		return nil, nil
	}

	return &cached, nil
}

// Delete removes a URL from cache.
// Used when a URL is deleted or updated.
func (c *URLCache) Delete(ctx context.Context, shortCode string) error {
	return c.rdb.Del(ctx, urlKey(shortCode)).Err()
}
