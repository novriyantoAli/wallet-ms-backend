package repository

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TransactionRepository defines the interface for wallet transaction data access
type TransactionRepository interface {
	Create(transaction *entity.WalletTransaction) error
	GetByID(id uint) (*entity.WalletTransaction, error)
	GetByWalletID(filter *dto.TransactionFilter) ([]entity.WalletTransaction, int64, error)
	Update(transaction *entity.WalletTransaction) error
	Delete(id uint) error
	GetByReferenceID(referenceID string) (*entity.WalletTransaction, error)
}

type transactionRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTransactionRepository creates a new transaction repository instance
func NewTransactionRepository(db *gorm.DB, logger *zap.Logger) TransactionRepository {
	return &transactionRepository{
		db:     db,
		logger: logger,
	}
}

func (r *transactionRepository) Create(transaction *entity.WalletTransaction) error {
	r.logger.Info("Creating wallet transaction",
		zap.Uint("wallet_id", transaction.WalletID),
		zap.String("type", string(transaction.Type)))
	return r.db.Create(transaction).Error
}

func (r *transactionRepository) GetByID(id uint) (*entity.WalletTransaction, error) {
	var transaction entity.WalletTransaction
	err := r.db.Where("id = ?", id).First(&transaction).Error
	if err != nil {
		r.logger.Error("Failed to get transaction by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) GetByWalletID(filter *dto.TransactionFilter) ([]entity.WalletTransaction, int64, error) {
	var transactions []entity.WalletTransaction
	var totalCount int64

	query := r.db.Where("wallet_id = ?", filter.WalletID)

	// Apply optional filters
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.StartDate != nil {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	// Get total count
	if err := query.Model(&entity.WalletTransaction{}).Count(&totalCount).Error; err != nil {
		r.logger.Error("Failed to count transactions", zap.Error(err))
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
	err := query.Offset(offset).Limit(filter.PageSize).Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		r.logger.Error("Failed to get transactions", zap.Error(err))
		return nil, 0, err
	}

	return transactions, totalCount, nil
}

func (r *transactionRepository) Update(transaction *entity.WalletTransaction) error {
	r.logger.Info("Updating wallet transaction", zap.Uint("id", transaction.ID))
	return r.db.Save(transaction).Error
}

func (r *transactionRepository) Delete(id uint) error {
	r.logger.Info("Deleting wallet transaction", zap.Uint("id", id))
	return r.db.Delete(&entity.WalletTransaction{}, id).Error
}

func (r *transactionRepository) GetByReferenceID(referenceID string) (*entity.WalletTransaction, error) {
	var transaction entity.WalletTransaction
	err := r.db.Where("reference_id = ?", referenceID).First(&transaction).Error
	if err != nil {
		r.logger.Error("Failed to get transaction by reference ID",
			zap.String("reference_id", referenceID),
			zap.Error(err))
		return nil, err
	}
	return &transaction, nil
}
