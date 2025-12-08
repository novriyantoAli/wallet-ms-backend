package migration

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/entity"
	productEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	purchaseEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/entity"
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

	// Use raw SQL to create tables if they don't exist to avoid GORM conflicts
	// This ensures idempotency and prevents "insufficient arguments" errors

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			level VARCHAR(20) DEFAULT 'user' NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			amount DECIMAL(10,2) NOT NULL,
			currency VARCHAR(3) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			description VARCHAR(500),
			user_id INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS wallets (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			balance DECIMAL(10,2) DEFAULT 0,
			currency VARCHAR(3) DEFAULT 'IDR',
			status VARCHAR(50) DEFAULT 'active',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS wallet_transactions (
			id SERIAL PRIMARY KEY,
			wallet_id INTEGER NOT NULL,
			type VARCHAR(50) NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			status VARCHAR(50) DEFAULT 'pending',
			description VARCHAR(255),
			balance_after DECIMAL(10,2) NOT NULL,
			reference_id VARCHAR(100),
			related_wallet_id INTEGER,
			metadata TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			CONSTRAINT fk_wallet_transaction_wallet_id FOREIGN KEY(wallet_id) REFERENCES wallets(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_wallet_transaction_wallet_id ON wallet_transactions(wallet_id)`,
		`CREATE INDEX IF NOT EXISTS idx_wallet_transaction_reference_id ON wallet_transactions(reference_id)`,
		`CREATE TABLE IF NOT EXISTS products (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description VARCHAR(1000),
			price DECIMAL(10,2) NOT NULL,
			sku VARCHAR(100) UNIQUE NOT NULL,
			stock INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			CONSTRAINT sku_unique UNIQUE(sku)
		)`,
		`CREATE TABLE IF NOT EXISTS wifi_products (
			id SERIAL PRIMARY KEY,
			product_id INTEGER NOT NULL UNIQUE,
			quota DECIMAL(10,2) NOT NULL,
			duration INTEGER NOT NULL,
			speed_limit DECIMAL(10,2) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			CONSTRAINT fk_wifi_product_id FOREIGN KEY(product_id) REFERENCES products(id) ON DELETE CASCADE,
			CONSTRAINT uk_wifi_product_id UNIQUE(product_id)
		)`,
		`CREATE TABLE IF NOT EXISTS purchases (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			product_id INTEGER NOT NULL,
			quantity INTEGER NOT NULL,
			total_price DECIMAL(10,2) NOT NULL,
			status VARCHAR(50) DEFAULT 'completed',
			notes VARCHAR(500),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			deleted_at TIMESTAMP,
			CONSTRAINT fk_purchase_user_id FOREIGN KEY(user_id) REFERENCES users(id),
			CONSTRAINT fk_purchase_product_id FOREIGN KEY(product_id) REFERENCES products(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_purchase_user_id ON purchases(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_purchase_product_id ON purchases(product_id)`,
	}

	for _, migration := range migrations {
		if err := s.db.Exec(migration).Error; err != nil {
			s.logger.Error("Failed to execute migration", zap.Error(err), zap.String("migration", migration[:50]))
			return err
		}
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
		&purchaseEntity.Purchase{},
		&productEntity.WiFiProduct{},
		&productEntity.Product{},
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
