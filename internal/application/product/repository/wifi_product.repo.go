package repository

import (
	"context"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WiFiProductRepository interface {
	Create(ctx context.Context, product *entity.WiFiProduct) error
	GetByID(ctx context.Context, id uint) (*entity.WiFiProduct, error)
	GetByProductID(ctx context.Context, productID uint) (*entity.WiFiProduct, error)
	GetAll(ctx context.Context, filter *dto.WiFiProductFilter) ([]entity.WiFiProduct, int64, error)
	Update(ctx context.Context, product *entity.WiFiProduct) error
	Delete(ctx context.Context, id uint) error
}

type wifiProductRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewWiFiProductRepository(db *gorm.DB, logger *zap.Logger) WiFiProductRepository {
	return &wifiProductRepository{
		db:     db,
		logger: logger,
	}
}

func (r *wifiProductRepository) Create(ctx context.Context, product *entity.WiFiProduct) error {
	r.logger.Info("Creating WiFi product", zap.Uint("product_id", product.ProductID))
	db := database.GetDB(ctx, r.db)
	return db.Create(product).Error
}

func (r *wifiProductRepository) GetByID(ctx context.Context, id uint) (*entity.WiFiProduct, error) {
	var product entity.WiFiProduct
	db := database.GetDB(ctx, r.db)
	err := db.First(&product, id).Error
	if err != nil {
		r.logger.Error("Failed to get WiFi product by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &product, nil
}

func (r *wifiProductRepository) GetByProductID(ctx context.Context, productID uint) (*entity.WiFiProduct, error) {
	var product entity.WiFiProduct
	db := database.GetDB(ctx, r.db)
	err := db.Where("product_id = ?", productID).First(&product).Error
	if err != nil {
		r.logger.Error("Failed to get WiFi product by product ID", zap.Uint("product_id", productID), zap.Error(err))
		return nil, err
	}
	return &product, nil
}

func (r *wifiProductRepository) GetAll(ctx context.Context, filter *dto.WiFiProductFilter) ([]entity.WiFiProduct, int64, error) {
	var products []entity.WiFiProduct
	var total int64

	db := database.GetDB(ctx, r.db)
	query := db

	if err := query.Model(&entity.WiFiProduct{}).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count WiFi products", zap.Error(err))
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Order("created_at DESC").Find(&products).Error; err != nil {
		r.logger.Error("Failed to get WiFi products", zap.Error(err))
		return nil, 0, err
	}

	return products, total, nil
}

func (r *wifiProductRepository) Update(ctx context.Context, product *entity.WiFiProduct) error {
	r.logger.Info("Updating WiFi product", zap.Uint("id", product.ID))
	db := database.GetDB(ctx, r.db)
	return db.Save(product).Error
}

func (r *wifiProductRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting WiFi product", zap.Uint("id", id))
	db := database.GetDB(ctx, r.db)
	return db.Delete(&entity.WiFiProduct{}, id).Error
}
