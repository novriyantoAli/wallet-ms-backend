package repository

import (
	"context"
	"strings"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	GetByID(ctx context.Context, id uint) (*entity.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entity.Product, error)
	GetAll(ctx context.Context, filter *dto.ProductFilter) ([]entity.Product, int64, error)
	GetActiveProducts(ctx context.Context, filter *dto.ProductFilter) ([]entity.Product, int64, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uint) error
}

type productRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewProductRepository(db *gorm.DB, logger *zap.Logger) ProductRepository {
	return &productRepository{
		db:     db,
		logger: logger,
	}
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	r.logger.Info("Creating product", zap.String("sku", product.SKU))
	db := database.GetDB(ctx, r.db)
	return db.Create(product).Error
}

func (r *productRepository) GetByID(ctx context.Context, id uint) (*entity.Product, error) {
	var product entity.Product
	db := database.GetDB(ctx, r.db)
	err := db.First(&product, id).Error
	if err != nil {
		r.logger.Error("Failed to get product by ID", zap.Uint("id", id), zap.Error(err))
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*entity.Product, error) {
	var product entity.Product
	db := database.GetDB(ctx, r.db)
	err := db.Where("sku = ?", sku).First(&product).Error
	if err != nil {
		r.logger.Error("Failed to get product by SKU", zap.String("sku", sku), zap.Error(err))
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetAll(ctx context.Context, filter *dto.ProductFilter) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	db := database.GetDB(ctx, r.db)
	query := db

	if filter.Search != "" {
		query = query.Where("UPPER(name) LIKE ? OR UPPER(description) LIKE ?", "%"+strings.ToUpper(filter.Search)+"%", "%"+strings.ToUpper(filter.Search)+"%")
	}

	if filter.SKU != "" {
		query = query.Where("UPPER(sku) LIKE ?", "%"+strings.ToUpper(filter.SKU)+"%")
	}

	if err := query.Model(&entity.Product{}).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count products", zap.Error(err))
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Find(&products).Error; err != nil {
		r.logger.Error("Failed to get products", zap.Error(err))
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) GetActiveProducts(ctx context.Context, filter *dto.ProductFilter) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	db := database.GetDB(ctx, r.db)
	query := db.Where("status = ?", entity.ProductStatusActive)

	if filter.SKU != "" {
		query = query.Where("sku = ?", filter.SKU)
	}

	if err := query.Model(&entity.Product{}).Count(&total).Error; err != nil {
		r.logger.Error("Failed to count active products", zap.Error(err))
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	if err := query.Offset(offset).Limit(filter.Limit).Find(&products).Error; err != nil {
		r.logger.Error("Failed to get active products", zap.Error(err))
		return nil, 0, err
	}

	return products, total, nil
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	r.logger.Info("Updating product", zap.Uint("id", product.ID))
	db := database.GetDB(ctx, r.db)
	return db.Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id uint) error {
	r.logger.Info("Deleting product", zap.Uint("id", id))
	db := database.GetDB(ctx, r.db)
	return db.Unscoped().Delete(&entity.Product{}, id).Error
}
