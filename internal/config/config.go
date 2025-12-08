package config

import (
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultStartTimeout = 15 * time.Second
	DefaultStopTimeout  = 10 * time.Second
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Logger     LoggerConfig     `mapstructure:"logger"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Worker     WorkerConfig     `mapstructure:"worker"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	DanaGapura DanaGapuraConfig `mapstructure:"dana_gapura"`
}

type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
	SSLMode  string `mapstructure:"ssl_mode"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type WorkerConfig struct {
	Concurrency          int           `mapstructure:"concurrency"`
	PaymentCheckInterval time.Duration `mapstructure:"payment_check_interval"`
	RetryMaxAttempts     int           `mapstructure:"retry_max_attempts"`
	RetryDelay           time.Duration `mapstructure:"retry_delay"`
}

type JWTConfig struct {
	SecretKey string        `mapstructure:"secret_key"`
	Expiry    time.Duration `mapstructure:"expiry"`
}

type DanaGapuraConfig struct {
	BaseURL        string `mapstructure:"base_url"`
	PartnerID      string `mapstructure:"partner_id"`
	PrivateKey     string `mapstructure:"private_key"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
	PublicKeyPath  string `mapstructure:"public_key_path"`
}

func NewConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.idle_timeout", "60s")

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.db_name", "vibe_db")
	viper.SetDefault("database.ssl_mode", "disable")

	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.format", "json")
	viper.SetDefault("logger.output_path", "stdout")

	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("worker.concurrency", 10)
	viper.SetDefault("worker.payment_check_interval", "5m")
	viper.SetDefault("worker.retry_max_attempts", 3)
	viper.SetDefault("worker.retry_delay", "30s")

	viper.SetDefault("jwt.secret_key", "your-secret-key-change-in-production")
	viper.SetDefault("jwt.expiry", "24h")

	viper.SetDefault("dana_gapura.base_url", "https://api.sandbox.dana.id")
	viper.SetDefault("dana_gapura.partner_id", "")
	viper.SetDefault("dana_gapura.private_key", "")
	viper.SetDefault("dana_gapura.private_key_path", "")
	viper.SetDefault("dana_gapura.public_key_path", "")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
