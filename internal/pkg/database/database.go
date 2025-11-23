package database

import (
	"fmt"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/entity"
	userEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewDatabase(cfg *config.Config, log *zap.Logger) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	err = db.AutoMigrate(
		&userEntity.User{},
		&entity.Payment{},
	)
	if err != nil {
		log.Error("Failed to migrate database", zap.Error(err))
		return nil, err
	}

	log.Info("Database connected and migrated successfully")
	return db, nil
}
