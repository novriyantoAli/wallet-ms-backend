package service

import (
	"errors"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/util"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProductService interface {
	CreateProduct(req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProductByID(id uint) (*dto.ProductResponse, error)
	GetProductBySKU(sku string) (*dto.ProductResponse, error)
	GetProducts(filter *dto.ProductFilter) (*dto.ProductListResponse, error)
	UpdateProduct(id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(id uint) error
}

type productService struct {
	repo            repository.ProductRepository
	wifiProductRepo repository.WiFiProductRepository
	logger          *zap.Logger
}

func NewProductService(repo repository.ProductRepository, wifiProductRepo repository.WiFiProductRepository, logger *zap.Logger) ProductService {
	return &productService{
		repo:            repo,
		wifiProductRepo: wifiProductRepo,
		logger:          logger,
	}
}

func (s *productService) CreateProduct(req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Auto-generate SKU from category, name, and price
	sku := util.GenerateSKU(req.Category, req.Name, req.Price)

	// Check if SKU already exists
	existing, _ := s.repo.GetBySKU(sku)
	if existing != nil {
		s.logger.Warn("Product with auto-generated SKU already exists", zap.String("sku", sku))
		return nil, errors.New("product with this name, category and price already exists")
	}

	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		SKU:         sku,
		Category:    entity.ProductCategory(req.Category),
		Stock:       req.Stock,
	}

	if !product.IsValid() {
		s.logger.Warn("Invalid product data", zap.Any("product", product))
		return nil, errors.New("invalid product data")
	}

	if err := s.repo.Create(product); err != nil {
		s.logger.Error("Failed to create product", zap.Error(err))
		return nil, err
	}

	// If category is WiFi, WiFi product details must be provided
	if product.Category == entity.ProductCategoryWiFi {
		// For WiFi products, create a default WiFi product entry
		// The WiFi product details (quota, duration, speed_limit) should be set by the client through a separate endpoint
		wifiProduct := &entity.WiFiProduct{
			ProductID:  product.ID,
			Quota:      0, // Will be updated through separate endpoint
			Duration:   0, // Will be updated through separate endpoint
			SpeedLimit: 0, // Will be updated through separate endpoint
		}

		if err := s.wifiProductRepo.Create(wifiProduct); err != nil {
			s.logger.Warn("Failed to create wifi product details, but product was created", zap.Error(err), zap.Uint("product_id", product.ID))
			// Don't fail the product creation if wifi product creation fails
		}
	}

	s.logger.Info("Product created successfully", zap.Uint("id", product.ID), zap.String("sku", sku), zap.String("category", string(product.Category)))
	return s.entityToResponse(product), nil
}

func (s *productService) GetProductByID(id uint) (*dto.ProductResponse, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *productService) GetProductBySKU(sku string) (*dto.ProductResponse, error) {
	product, err := s.repo.GetBySKU(sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to get product by SKU", zap.Error(err), zap.String("sku", sku))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *productService) GetProducts(filter *dto.ProductFilter) (*dto.ProductListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 10
	}

	products, total, err := s.repo.GetAll(filter)
	if err != nil {
		s.logger.Error("Failed to get products", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		resp := s.entityToResponse(&product)
		responses[i] = *resp
	}

	totalPages := (int(total) + filter.Limit - 1) / filter.Limit

	return &dto.ProductListResponse{
		Data:       responses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *productService) UpdateProduct(id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	// Note: SKU and Category cannot be updated after product creation
	// SKU is auto-generated from category, name, and price at creation time
	// Category is immutable

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}

	if err := s.repo.Update(product); err != nil {
		s.logger.Error("Failed to update product", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *productService) DeleteProduct(id uint) error {
	// Check if product exists (before cascade delete)
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err), zap.Uint("id", id))
		return err
	}

	// Delete product (cascade deletes wifi_products)
	if err := s.repo.Delete(id); err != nil {
		s.logger.Error("Failed to delete product", zap.Error(err), zap.Uint("id", id))
		return err
	}

	s.logger.Info("Product deleted successfully", zap.Uint("id", id))

	return nil
}

func (s *productService) entityToResponse(product *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		SKU:         product.SKU,
		Category:    string(product.Category),
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
