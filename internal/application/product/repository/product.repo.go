package repository

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProductRepository interface {
	Create(product *entity.Product) error
	GetByID(id uint) (*entity.Product, error)
	GetBySKU(sku string) (*entity.Product, error)
	GetAll(filter *dto.ProductFilter) ([]entity.Product, int64, error)
	Update(product *entity.Product) error
	Delete(id uint) error
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

func (r *productRepository) Create(product *entity.Product) error {
	if err := r.db.Create(product).Error; err != nil {
		r.logger.Error("Failed to create product", zap.Error(err), zap.String("sku", product.SKU))
		return err
	}
	return nil
}

func (r *productRepository) GetByID(id uint) (*entity.Product, error) {
	var product entity.Product
	if err := r.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Product not found", zap.Uint("id", id))
			return nil, err
		}
		r.logger.Error("Failed to get product by ID", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetBySKU(sku string) (*entity.Product, error) {
	var product entity.Product
	if err := r.db.Where("sku = ?", sku).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			r.logger.Debug("Product not found by SKU", zap.String("sku", sku))
			return nil, err
		}
		r.logger.Error("Failed to get product by SKU", zap.Error(err), zap.String("sku", sku))
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetAll(filter *dto.ProductFilter) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	query := r.db

	if filter.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if filter.SKU != "" {
		query = query.Where("sku ILIKE ?", "%"+filter.SKU+"%")
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

func (r *productRepository) Update(product *entity.Product) error {
	if err := r.db.Save(product).Error; err != nil {
		r.logger.Error("Failed to update product", zap.Error(err), zap.Uint("id", product.ID))
		return err
	}
	return nil
}

func (r *productRepository) Delete(id uint) error {
	// Use Unscoped to perform hard delete for cascade delete to work
	if err := r.db.Unscoped().Delete(&entity.Product{}, id).Error; err != nil {
		r.logger.Error("Failed to delete product", zap.Error(err), zap.Uint("id", id))
		return err
	}
	return nil
}
