package repository

import (
	"errors"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/entity"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PurchaseRepository interface {
	Create(purchase *entity.Purchase) error
	GetByID(id uint) (*entity.Purchase, error)
	GetByUserID(userID uint, filter *dto.PurchaseFilter) ([]entity.Purchase, int64, error)
	Update(purchase *entity.Purchase) error
	Delete(id uint) error
}

type purchaseRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewPurchaseRepository(db *gorm.DB, logger *zap.Logger) PurchaseRepository {
	return &purchaseRepository{db: db, logger: logger}
}

func (r *purchaseRepository) Create(purchase *entity.Purchase) error {
	if err := r.db.Create(purchase).Error; err != nil {
		r.logger.Error("Failed to create purchase", zap.Error(err))
		return err
	}
	return nil
}

func (r *purchaseRepository) GetByID(id uint) (*entity.Purchase, error) {
	var purchase entity.Purchase
	if err := r.db.Where("id = ?", id).First(&purchase).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		r.logger.Error("Failed to get purchase", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}
	return &purchase, nil
}

func (r *purchaseRepository) GetByUserID(userID uint, filter *dto.PurchaseFilter) ([]entity.Purchase, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	var purchases []entity.Purchase
	var count int64

	query := r.db.Where("user_id = ?", userID)
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ProductID > 0 {
		query = query.Where("product_id = ?", filter.ProductID)
	}

	if err := query.Model(&entity.Purchase{}).Count(&count).Error; err != nil {
		r.logger.Error("Failed to count purchases", zap.Error(err))
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&purchases).Error; err != nil {
		r.logger.Error("Failed to get purchases", zap.Error(err), zap.Uint("user_id", userID))
		return nil, 0, err
	}

	return purchases, count, nil
}

func (r *purchaseRepository) Update(purchase *entity.Purchase) error {
	if err := r.db.Save(purchase).Error; err != nil {
		r.logger.Error("Failed to update purchase", zap.Error(err), zap.Uint("id", purchase.ID))
		return err
	}
	return nil
}

func (r *purchaseRepository) Delete(id uint) error {
	if err := r.db.Unscoped().Delete(&entity.Purchase{}, id).Error; err != nil {
		r.logger.Error("Failed to delete purchase", zap.Error(err), zap.Uint("id", id))
		return err
	}
	return nil
}
