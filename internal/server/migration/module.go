package migration

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/entity"
	userEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	walletEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewServer(db *gorm.DB, logger *zap.Logger) *Server {
	return &Server{
		db:     db,
		logger: logger,
	}
}

func (s *Server) RunMigrations() error {
	s.logger.Info("Starting database migrations")

	// Run auto migrations for all entities
	err := s.db.AutoMigrate(
		&userEntity.User{},
		&walletEntity.Wallet{},
		&walletEntity.WalletTransaction{},
		&entity.Payment{},
	)
	if err != nil {
		s.logger.Error("Failed to run database migrations", zap.Error(err))
		return err
	}

	s.logger.Info("Database migrations completed successfully")
	return nil
}

func (s *Server) SeedData() error {
	s.logger.Info("Starting data seeding")

	// Add any initial data seeding here
	// Example: Create default admin user, initial payment statuses, etc.

	s.logger.Info("Data seeding completed successfully")
	return nil
}

func (s *Server) DropTables() error {
	s.logger.Warn("Dropping all database tables")

	err := s.db.Migrator().DropTable(
		&walletEntity.WalletTransaction{},
		&walletEntity.Wallet{},
		&entity.Payment{},
		&userEntity.User{},
	)
	if err != nil {
		s.logger.Error("Failed to drop database tables", zap.Error(err))
		return err
	}

	s.logger.Info("Database tables dropped successfully")
	return nil
}

func (s *Server) RecreateUserTable() error {
	s.logger.Warn("Recreating user table with new schema")

	// Drop the table if it exists
	if err := s.db.Migrator().DropTable(&userEntity.User{}); err != nil {
		s.logger.Error("Failed to drop user table", zap.Error(err))
		return err
	}

	// Create the table fresh with new schema
	if err := s.db.AutoMigrate(&userEntity.User{}); err != nil {
		s.logger.Error("Failed to create user table", zap.Error(err))
		return err
	}

	s.logger.Info("User table recreated successfully with level column")
	return nil
}
