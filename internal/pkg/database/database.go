package database

import (
	"fmt"

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

	log.Info("Database connected successfully")

	// Create tables using raw SQL (CREATE TABLE IF NOT EXISTS to avoid duplicates)
	migrations := []string{
		`DROP TABLE IF EXISTS purchases CASCADE`,
		`DROP TABLE IF EXISTS wifi_products CASCADE`,
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
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
		if err := db.Exec(migration).Error; err != nil {
			log.Error("Failed to execute migration", zap.Error(err), zap.String("migration", migration[:50]))
			return nil, err
		}
	}

	log.Info("Database connected and migrated successfully")
	return db, nil
}
