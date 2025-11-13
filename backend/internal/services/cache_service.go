package services

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// CacheService handles Redis caching
type CacheService struct {
	client *redis.Client
	ctx    context.Context
}

// NewCacheService creates a new cache service
func NewCacheService(redisURL string) *CacheService {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		// If Redis is not configured, return nil (caching disabled)
		return nil
	}

	client := redis.NewClient(opt)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		// Redis not available, caching disabled
		return nil
	}

	return &CacheService{
		client: client,
		ctx:    ctx,
	}
}

// Get retrieves a value from cache
func (s *CacheService) Get(key string, dest interface{}) error {
	if s == nil || s.client == nil {
		return redis.Nil
	}

	val, err := s.client.Get(s.ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

// Set stores a value in cache with expiration
func (s *CacheService) Set(key string, value interface{}, expiration time.Duration) error {
	if s == nil || s.client == nil {
		return nil // Silently fail if Redis not available
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return s.client.Set(s.ctx, key, jsonData, expiration).Err()
}

// Delete removes a key from cache
func (s *CacheService) Delete(key string) error {
	if s == nil || s.client == nil {
		return nil
	}

	return s.client.Del(s.ctx, key).Err()
}

// DeletePattern removes all keys matching a pattern
func (s *CacheService) DeletePattern(pattern string) error {
	if s == nil || s.client == nil {
		return nil
	}

	iter := s.client.Scan(s.ctx, 0, pattern, 0).Iterator()
	for iter.Next(s.ctx) {
		if err := s.client.Del(s.ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Exists checks if a key exists
func (s *CacheService) Exists(key string) bool {
	if s == nil || s.client == nil {
		return false
	}

	count, err := s.client.Exists(s.ctx, key).Result()
	return err == nil && count > 0
}

// IsEnabled returns true if caching is available
func (s *CacheService) IsEnabled() bool {
	return s != nil && s.client != nil
}
