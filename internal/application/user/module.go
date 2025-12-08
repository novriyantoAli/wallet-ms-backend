package user

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/config"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/jwt"
	redisutil "github.com/novriyantoAli/wallet-ms-backend/internal/pkg/redis"

	redisclient "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func provideJWTManager(cfg *config.Config, redisClient *redisclient.Client) *jwt.JWTManager {
	return jwt.NewJWTManagerWithRedis(jwt.JWTConfig{
		SecretKey: cfg.JWT.SecretKey,
		Expiry:    cfg.JWT.Expiry,
	}, redisClient)
}

func provideRedisClient(cfg *config.Config) *redisclient.Client {
	return redisutil.NewRedisClient(redisutil.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

// Module provides all user domain dependencies
var Module = fx.Options(
	fx.Provide(
		provideRedisClient,
		repository.NewUserRepository,
		provideJWTManager,
		service.NewUserService,
		handler.NewUserHandler,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		provideRedisClient,
		repository.NewUserRepository,
		provideJWTManager,
		service.NewUserService,
	),
)
