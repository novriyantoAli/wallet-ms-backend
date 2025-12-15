package redis

import (
	"context"
	"fmt"

	"github.com/novriyantoAli/wallet-ms-backend/internal/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	return client
}

// Ping checks the Redis connection
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
