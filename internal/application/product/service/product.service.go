package service

import (
	"context"
	"errors"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/sku"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProductByID(ctx context.Context, id uint) (*dto.ProductResponse, error)
	GetProductBySKU(ctx context.Context, sku string) (*dto.ProductResponse, error)
	GetProducts(ctx context.Context, filter *dto.ProductFilter) (*dto.ProductListResponse, error)
	GetActiveProducts(ctx context.Context, filter *dto.ProductFilter) (*dto.ProductListResponse, error)
	UpdateProduct(ctx context.Context, id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error)
	UpdateProductStatus(ctx context.Context, id uint, status string) (*dto.ProductResponse, error)
	DeleteProduct(ctx context.Context, id uint) error
}

type productService struct {
	txManager       database.TransactionManagerI
	repo            repository.ProductRepository
	wifiProductRepo repository.WiFiProductRepository
	skuGen          *sku.Generator
	logger          *zap.Logger
}

func NewProductService(
	txManager database.TransactionManagerI,
	repo repository.ProductRepository,
	wifiProductRepo repository.WiFiProductRepository,
	logger *zap.Logger,
) ProductService {
	return &productService{
		txManager:       txManager,
		repo:            repo,
		wifiProductRepo: wifiProductRepo,
		skuGen:          sku.NewGenerator("PRODUCT"),
		logger:          logger,
	}
}

func (s *productService) CreateProduct(ctx context.Context, req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Validate input
	if req.Name == "" || req.Price <= 0 || req.Category == "" {
		return nil, errors.New("invalid product data")
	}

	// Auto-generate unique SKU using SKU generator
	skuValue, err := s.skuGen.GenerateWithPattern("PRODUCT-" + req.Category + "-{date}-{random}")
	if err != nil {
		s.logger.Error("Failed to generate SKU", zap.Error(err))
		return nil, errors.New("failed to generate product SKU")
	}

	// Check if SKU already exists (extremely unlikely with timestamp but good practice)
	existing, _ := s.repo.GetBySKU(ctx, skuValue)
	if existing != nil {
		s.logger.Warn("Product with auto-generated SKU already exists", zap.String("sku", skuValue))
		return nil, errors.New("product sku already exists")
	}

	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		SKU:         skuValue,
		Category:    entity.ProductCategory(req.Category),
		Status:      entity.ProductStatusInactive, // Default status is inactive
		Stock:       req.Stock,
	}

	if !product.IsValid() {
		s.logger.Warn("Invalid product data", zap.Any("product", product))
		return nil, errors.New("invalid product data")
	}

	// Use transaction for atomic operations: create product + wifi product (if applicable)
	var result *entity.Product
	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Step 1: Create product
		if err := s.repo.Create(txCtx, product); err != nil {
			s.logger.Error("Failed to create product in transaction", zap.Error(err))
			return err
		}
		result = product

		// Step 2: If category is WiFi, create WiFi product entry
		if product.Category == entity.ProductCategoryWiFi {
			wifiProduct := &entity.WiFiProduct{
				ProductID:  product.ID,
				Quota:      0, // Will be updated through separate endpoint
				Duration:   0, // Will be updated through separate endpoint
				SpeedLimit: 0, // Will be updated through separate endpoint
			}

			if err := s.wifiProductRepo.Create(txCtx, wifiProduct); err != nil {
				s.logger.Error("Failed to create wifi product in transaction", zap.Error(err))
				return err // Rollback transaction on failure
			}
		}

		return nil
	})

	if err != nil {
		s.logger.Error("Transaction failed for product creation", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Product created successfully with transaction", zap.Uint("id", result.ID), zap.String("sku", skuValue), zap.String("category", string(result.Category)))
	return s.entityToResponse(result), nil
}

func (s *productService) GetProductByID(ctx context.Context, id uint) (*dto.ProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *productService) GetProductBySKU(ctx context.Context, sku string) (*dto.ProductResponse, error) {
	product, err := s.repo.GetBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to get product by SKU", zap.Error(err), zap.String("sku", sku))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *productService) GetProducts(ctx context.Context, filter *dto.ProductFilter) (*dto.ProductListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 10
	}

	products, total, err := s.repo.GetAll(ctx, filter)
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

func (s *productService) GetActiveProducts(ctx context.Context, filter *dto.ProductFilter) (*dto.ProductListResponse, error) {
	products, total, err := s.repo.GetActiveProducts(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get active products", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = *s.entityToResponse(&product)
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

func (s *productService) UpdateProduct(ctx context.Context, id uint, req *dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
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
	if req.Status != "" {
		product.Status = entity.ProductStatus(req.Status)
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}

	if err := s.repo.Update(ctx, product); err != nil {
		s.logger.Error("Failed to update product", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *productService) UpdateProductStatus(ctx context.Context, id uint, status string) (*dto.ProductResponse, error) {
	// Validate status value
	if status != "active" && status != "inactive" {
		return nil, errors.New("invalid status: must be 'active' or 'inactive'")
	}

	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	product.Status = entity.ProductStatus(status)

	if err := s.repo.Update(ctx, product); err != nil {
		s.logger.Error("Failed to update product status", zap.Error(err), zap.Uint("id", id))
		return nil, err
	}

	s.logger.Info("Product status updated successfully", zap.Uint("id", id), zap.String("status", status))

	return s.entityToResponse(product), nil
}

func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	// Check if product exists (before cascade delete)
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err), zap.Uint("id", id))
		return err
	}

	// Delete product (cascade deletes wifi_products)
	if err := s.repo.Delete(ctx, id); err != nil {
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
		Status:      string(product.Status),
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
