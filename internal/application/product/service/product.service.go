package service

import (
	"errors"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"

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
	repo   repository.ProductRepository
	logger *zap.Logger
}

func NewProductService(repo repository.ProductRepository, logger *zap.Logger) ProductService {
	return &productService{
		repo:   repo,
		logger: logger,
	}
}

func (s *productService) CreateProduct(req *dto.CreateProductRequest) (*dto.ProductResponse, error) {
	// Check if SKU already exists
	existing, _ := s.repo.GetBySKU(req.SKU)
	if existing != nil {
		s.logger.Warn("Product with SKU already exists", zap.String("sku", req.SKU))
		return nil, errors.New("product with this SKU already exists")
	}

	product := &entity.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		SKU:         req.SKU,
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

	// Check if new SKU is already used by another product
	if req.SKU != "" && req.SKU != product.SKU {
		existing, _ := s.repo.GetBySKU(req.SKU)
		if existing != nil {
			return nil, errors.New("product with this SKU already exists")
		}
		product.SKU = req.SKU
	}

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
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
