package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/eveeze/warung-backend/internal/config"
	"github.com/eveeze/warung-backend/internal/pkg/logger"
)

// RedisClient wraps the redis.Client with additional functionality
type RedisClient struct {
	*redis.Client
	config *config.RedisConfig
}

// NewRedis creates a new Redis connection
func NewRedis(cfg *config.RedisConfig) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	logger.Info("Redis connection established successfully")

	return &RedisClient{
		Client: client,
		config: cfg,
	}, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	logger.Info("Closing Redis connection")
	return r.Client.Close()
}

// Health checks the Redis health
func (r *RedisClient) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	return r.Client.Ping(ctx).Err()
}

// Cache helpers

// SetJSON sets a JSON value with expiration
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Set(ctx, key, value, expiration).Err()
}

// GetJSON gets a JSON value
func (r *RedisClient) GetJSON(ctx context.Context, key string) (string, error) {
	return r.Get(ctx, key).Result()
}

// DeleteKeys deletes multiple keys by pattern
func (r *RedisClient) DeleteKeys(ctx context.Context, pattern string) error {
	iter := r.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := r.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	return iter.Err()
}

// Cache key generators

// ProductCacheKey generates a cache key for a product
func ProductCacheKey(id string) string {
	return fmt.Sprintf("product:%s", id)
}

// ProductListCacheKey generates a cache key for product list
func ProductListCacheKey(page, perPage int, filters string) string {
	return fmt.Sprintf("products:list:%d:%d:%s", page, perPage, filters)
}

// CustomerCacheKey generates a cache key for a customer
func CustomerCacheKey(id string) string {
	return fmt.Sprintf("customer:%s", id)
}

// TransactionCacheKey generates a cache key for a transaction
func TransactionCacheKey(id string) string {
	return fmt.Sprintf("transaction:%s", id)
}

// ReportCacheKey generates a cache key for reports
func ReportCacheKey(reportType, date string) string {
	return fmt.Sprintf("report:%s:%s", reportType, date)
}
