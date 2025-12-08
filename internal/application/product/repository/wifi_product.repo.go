package repository

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"

	"gorm.io/gorm"
)

type WiFiProductRepository interface {
	Create(product *entity.WiFiProduct) error
	GetByID(id uint) (*entity.WiFiProduct, error)
	GetByProductID(productID uint) (*entity.WiFiProduct, error)
	GetAll(filter *dto.WiFiProductFilter) ([]entity.WiFiProduct, int64, error)
	Update(product *entity.WiFiProduct) error
	Delete(id uint) error
}

type wifiProductRepository struct {
	db *gorm.DB
}

func NewWiFiProductRepository(db *gorm.DB) WiFiProductRepository {
	return &wifiProductRepository{db: db}
}

func (r *wifiProductRepository) Create(product *entity.WiFiProduct) error {
	return r.db.Create(product).Error
}

func (r *wifiProductRepository) GetByID(id uint) (*entity.WiFiProduct, error) {
	var product entity.WiFiProduct
	err := r.db.Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *wifiProductRepository) GetByProductID(productID uint) (*entity.WiFiProduct, error) {
	var product entity.WiFiProduct
	err := r.db.Where("product_id = ?", productID).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *wifiProductRepository) GetAll(filter *dto.WiFiProductFilter) ([]entity.WiFiProduct, int64, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	var products []entity.WiFiProduct
	var count int64

	offset := (filter.Page - 1) * filter.Limit
	query := r.db

	err := query.Offset(offset).
		Limit(filter.Limit).
		Order("created_at DESC").
		Find(&products).
		Offset(-1).
		Limit(-1).
		Count(&count).Error

	if err != nil {
		return nil, 0, err
	}

	return products, count, nil
}

func (r *wifiProductRepository) Update(product *entity.WiFiProduct) error {
	return r.db.Save(product).Error
}

func (r *wifiProductRepository) Delete(id uint) error {
	return r.db.Delete(&entity.WiFiProduct{}, id).Error
}
