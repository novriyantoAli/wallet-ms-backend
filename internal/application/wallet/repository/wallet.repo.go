package repository

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// WalletRepository defines the interface for wallet data access
type WalletRepository interface {
	Create(wallet *entity.Wallet) error
	GetByID(id uint) (*entity.Wallet, error)
	GetByUserID(userID uint) (*entity.Wallet, error)
	GetAll(filter *dto.WalletFilter) ([]entity.Wallet, int64, error)
	Update(wallet *entity.Wallet) error
	Delete(id uint) error
	UpdateBalance(walletID uint, amount float64) error
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

func (r *walletRepository) Create(wallet *entity.Wallet) error {
	r.logger.Info("Creating wallet", zap.Uint("user_id", wallet.UserID))
	return r.db.Create(wallet).Error
}

func (r *walletRepository) GetByID(id uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.Where("id = ?", id).First(&wallet).Error
	if err != nil {
		r.logger.Error("Failed to get wallet by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetByUserID(userID uint) (*entity.Wallet, error) {
	var wallet entity.Wallet
	err := r.db.Where("user_id = ?", userID).First(&wallet).Error
	if err != nil {
		r.logger.Error("Failed to get wallet by user ID", zap.Uint("user_id", userID), zap.Error(err))
		return nil, err
	}
	return &wallet, nil
}

func (r *walletRepository) GetAll(filter *dto.WalletFilter) ([]entity.Wallet, int64, error) {
	var wallets []entity.Wallet
	var totalCount int64

	query := r.db

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

func (r *walletRepository) Update(wallet *entity.Wallet) error {
	r.logger.Info("Updating wallet", zap.Uint("id", wallet.ID))
	return r.db.Save(wallet).Error
}

func (r *walletRepository) Delete(id uint) error {
	r.logger.Info("Deleting wallet", zap.Uint("id", id))
	return r.db.Delete(&entity.Wallet{}, id).Error
}

func (r *walletRepository) UpdateBalance(walletID uint, amount float64) error {
	r.logger.Info("Updating wallet balance", zap.Uint("wallet_id", walletID), zap.Float64("amount", amount))
	return r.db.Model(&entity.Wallet{}).Where("id = ?", walletID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}
