package repository

import (
	"context"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WalletWithUserData represents wallet joined with user information
type WalletWithUserData struct {
	ID        uint      `gorm:"column:id"`
	UserID    uint      `gorm:"column:user_id"`
	Balance   float64   `gorm:"column:balance"`
	Currency  string    `gorm:"column:currency"`
	Status    string    `gorm:"column:status"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
	// User fields
	UserName  string `gorm:"column:user_name"`
	UserEmail string `gorm:"column:user_email"`
	UserLevel string `gorm:"column:user_level"`
}

// WalletRepository defines the interface for wallet data access
type WalletRepository interface {
	Create(ctx context.Context, wallet *entity.Wallet) error
	GetByID(ctx context.Context, id uint) (*entity.Wallet, error)
	GetByUserID(ctx context.Context, userID uint) (*entity.Wallet, error)
	GetAll(ctx context.Context, filter *dto.WalletFilter) ([]entity.Wallet, int64, error)
	Update(ctx context.Context, wallet *entity.Wallet) error
	Delete(ctx context.Context, id uint) error
	UpdateBalance(ctx context.Context, walletID uint, amount float64) error
	GetByIDWithUser(ctx context.Context, id uint) (*WalletWithUserData, error)
	GetByUserIDWithUser(ctx context.Context, userID uint) (*WalletWithUserData, error)
	GetAllWithUser(ctx context.Context, filter *dto.WalletFilter) ([]WalletWithUserData, int64, error)
}

type walletRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewWalletRepository creates a new wallet repository instance
func NewWalletRepository(db *gorm.DB, logger *zap.Logger) WalletRepository {
	return &walletRepository{
		db:     db,
		logger: logger,
	}
}

func (r *walletRepository) Create(ctx context.Context, wallet *entity.Wallet) error {
	r.logger.Info("Creating wallet", zap.Uint("user_id", wallet.UserID))
	db := database.GetDB(ctx, r.db)
	return db.Create(wallet).Error
}

func (r *walletRepository) GetByID(ctx context.Context, id uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	db := database.GetDB(ctx, r.db)
	err := db.Where("id = ?", id).First(&wallet).Error
	if err != nil {
		r.logger.Error("Failed to get wallet by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	db := database.GetDB(ctx, r.db)
	err := db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		r.logger.Error("Failed to get wallet by user ID", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetAll(ctx context.Context, filter *dto.WalletFilter) ([]entity.Wallet, int64, error) {
	var wallets []entity.Wallet
	var totalCount int64

	db := database.GetDB(ctx, r.db)
	query := db

	// Apply filters
	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Count total records
	if err := query.Model(&entity.Wallet{}).Count(&totalCount).Error; err != nil {
		r.logger.Error("Failed to count wallets", zap.Error(err))
		return nil, 0, err
	}

	// Apply pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Find(&wallets).Error
	if err != nil {
		r.logger.Error("Failed to get wallets", zap.Error(err))
		return nil, 0, err
	}

	return wallets, totalCount, nil
}

func (r *walletRepository) Update(ctx context.Context, wallet *entity.Wallet) error {
	r.logger.Info("Updating wallet", zap.Uint("id", wallet.ID))
	db := database.GetDB(ctx, r.db)
	return db.Save(wallet).Error
}

func (r *walletRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting wallet", zap.Uint("id", id))
	db := database.GetDB(ctx, r.db)
	return db.Delete(&entity.Wallet{}, id).Error
}

func (r *walletRepository) UpdateBalance(ctx context.Context, walletID uint, amount float64) error {
	r.logger.Info("Updating wallet balance", zap.Uint("wallet_id", walletID), zap.Float64("amount", amount))
	db := database.GetDB(ctx, r.db)
	return db.Model(&entity.Wallet{}).Where("id = ?", walletID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func (r *walletRepository) GetByIDWithUser(ctx context.Context, id uint) (*WalletWithUserData, error) {
	var walletWithUser WalletWithUserData
	db := database.GetDB(ctx, r.db)
	err := db.Table("wallets").
		Select(
			"wallets.id",
			"wallets.user_id",
			"wallets.balance",
			"wallets.currency",
			"wallets.status",
			"wallets.created_at",
			"wallets.updated_at",
			"users.name AS user_name",
			"users.email AS user_email",
			"users.level AS user_level",
		).
		Joins("LEFT JOIN users ON wallets.user_id = users.id").
		Where("wallets.id = ?", id).
		First(&walletWithUser).Error

	if err != nil {
		r.logger.Error("Failed to get wallet with user by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &walletWithUser, nil
}

func (r *walletRepository) GetByUserIDWithUser(ctx context.Context, userID uint) (*WalletWithUserData, error) {
	var walletWithUser WalletWithUserData
	db := database.GetDB(ctx, r.db)
	err := db.Table("wallets").
		Select(
			"wallets.id",
			"wallets.user_id",
			"wallets.balance",
			"wallets.currency",
			"wallets.status",
			"wallets.created_at",
			"wallets.updated_at",
			"users.name AS user_name",
			"users.email AS user_email",
			"users.level AS user_level",
		).
		Joins("LEFT JOIN users ON wallets.user_id = users.id").
		Where("wallets.user_id = ?", userID).
		First(&walletWithUser).Error

	if err != nil {
		r.logger.Error("Failed to get wallet with user by user ID", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}
	return &walletWithUser, nil
}

func (r *walletRepository) GetAllWithUser(ctx context.Context, filter *dto.WalletFilter) ([]WalletWithUserData, int64, error) {
	var wallets []WalletWithUserData
	var totalCount int64

	db := database.GetDB(ctx, r.db)
	query := db.Table("wallets").
		Select(
			"wallets.id",
			"wallets.user_id",
			"wallets.balance",
			"wallets.currency",
			"wallets.status",
			"wallets.created_at",
			"wallets.updated_at",
			"users.name AS user_name",
			"users.email AS user_email",
			"users.level AS user_level",
		).
		Joins("LEFT JOIN users ON wallets.user_id = users.id")

	// Apply filters
	if filter.UserID != 0 {
		query = query.Where("wallets.user_id = ?", filter.UserID)
	}
	if filter.Status != "" {
		query = query.Where("wallets.status = ?", filter.Status)
	}

	// Count total records
	if err := query.Model(&entity.Wallet{}).Count(&totalCount).Error; err != nil {
		r.logger.Error("Failed to count wallets with user", zap.Error(err))
		return nil, 0, err
	}

	// Apply pagination
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Find(&wallets).Error
	if err != nil {
		r.logger.Error("Failed to get wallets with user", zap.Error(err))
		return nil, 0, err
	}

	return wallets, totalCount, nil
}
