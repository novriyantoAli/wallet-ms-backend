package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func NewRedisClient(config RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	return client
}

// Ping checks the Redis connection
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
