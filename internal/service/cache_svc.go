package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/eveeze/warung-backend/internal/database"
)

// CacheService wraps Redis operations for caching
type CacheService struct {
	redis *database.RedisClient
}

// NewCacheService creates a new CacheService
func NewCacheService(redis *database.RedisClient) *CacheService {
	return &CacheService{redis: redis}
}

// Get retrieves a value from cache and unmarshals it
func (s *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, dest)
}

// Set stores a value in cache with TTL
func (s *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}
	
	return s.redis.Set(ctx, key, data, ttl).Err()
}

// Delete removes a key from cache
func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.redis.Del(ctx, key).Err()
}

// InvalidatePattern deletes all keys matching a pattern
func (s *CacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := s.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	
	if len(keys) == 0 {
		return nil
	}
	
	return s.redis.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in cache
func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := s.redis.Exists(ctx, key).Result()
	return count > 0, err
}
