package service

import (
	"context"
	"errors"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type WiFiProductService interface {
	CreateWiFiProduct(ctx context.Context, req *dto.CreateWiFiProductRequest) (*dto.WiFiProductResponse, error)
	GetWiFiProductByID(ctx context.Context, id uint) (*dto.WiFiProductResponse, error)
	GetWiFiProductByProductID(ctx context.Context, productID uint) (*dto.WiFiProductResponse, error)
	GetWiFiProducts(ctx context.Context, filter *dto.WiFiProductFilter) (*dto.WiFiProductListResponse, error)
	UpdateWiFiProduct(ctx context.Context, id uint, req *dto.UpdateWiFiProductRequest) (*dto.WiFiProductResponse, error)
	DeleteWiFiProduct(ctx context.Context, id uint) error
}

type wifiProductService struct {
	repo   repository.WiFiProductRepository
	logger *zap.Logger
}

func NewWiFiProductService(
	repo repository.WiFiProductRepository,
	logger *zap.Logger,
) WiFiProductService {
	return &wifiProductService{
		repo:   repo,
		logger: logger,
	}
}

func (s *wifiProductService) CreateWiFiProduct(ctx context.Context, req *dto.CreateWiFiProductRequest) (*dto.WiFiProductResponse, error) {
	// Validate request
	if req.ProductID == 0 || req.Quota <= 0 || req.Duration <= 0 || req.SpeedLimit <= 0 {
		return nil, errors.New("invalid wifi product data")
	}

	// Check if WiFi product already exists for this product
	existing, _ := s.repo.GetByProductID(ctx, req.ProductID)
	if existing != nil {
		return nil, errors.New("wifi product already exists for this product")
	}

	product := &entity.WiFiProduct{
		ProductID:  req.ProductID,
		Quota:      req.Quota,
		Duration:   req.Duration,
		SpeedLimit: req.SpeedLimit,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := s.repo.Create(ctx, product)
	if err != nil {
		s.logger.Error("Failed to create wifi product", zap.Error(err), zap.Uint("product_id", req.ProductID))
		return nil, errors.New("failed to create wifi product")
	}

	s.logger.Info("WiFi product created", zap.Uint("id", product.ID), zap.Uint("product_id", req.ProductID))
	return s.entityToResponse(product), nil
}

func (s *wifiProductService) GetWiFiProductByID(ctx context.Context, id uint) (*dto.WiFiProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wifi product not found")
		}
		s.logger.Error("Failed to get wifi product", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *wifiProductService) GetWiFiProductByProductID(ctx context.Context, productID uint) (*dto.WiFiProductResponse, error) {
	product, err := s.repo.GetByProductID(ctx, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wifi product not found for this product")
		}
		s.logger.Error("Failed to get wifi product by product_id", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(product), nil
}

func (s *wifiProductService) GetWiFiProducts(ctx context.Context, filter *dto.WiFiProductFilter) (*dto.WiFiProductListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	products, total, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to get wifi products", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.WiFiProductResponse, 0, len(products))
	for _, p := range products {
		responses = append(responses, *s.entityToResponse(&p))
	}

	totalPages := (total + int64(filter.Limit) - 1) / int64(filter.Limit)

	return &dto.WiFiProductListResponse{
		Data:       responses,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *wifiProductService) UpdateWiFiProduct(ctx context.Context, id uint, req *dto.UpdateWiFiProductRequest) (*dto.WiFiProductResponse, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wifi product not found")
		}
		return nil, err
	}

	// Update fields if provided
	if req.Quota > 0 {
		product.Quota = req.Quota
	}
	if req.Duration > 0 {
		product.Duration = req.Duration
	}
	if req.SpeedLimit > 0 {
		product.SpeedLimit = req.SpeedLimit
	}
	product.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, product)
	if err != nil {
		s.logger.Error("Failed to update wifi product", zap.Error(err))
		return nil, errors.New("failed to update wifi product")
	}

	return s.entityToResponse(product), nil
}

func (s *wifiProductService) DeleteWiFiProduct(ctx context.Context, id uint) error {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("wifi product not found")
		}
		return err
	}

	return s.repo.Delete(ctx, product.ID)
}

func (s *wifiProductService) entityToResponse(product *entity.WiFiProduct) *dto.WiFiProductResponse {
	return &dto.WiFiProductResponse{
		ID:         product.ID,
		ProductID:  product.ProductID,
		Quota:      product.Quota,
		Duration:   product.Duration,
		SpeedLimit: product.SpeedLimit,
		CreatedAt:  product.CreatedAt,
		UpdatedAt:  product.UpdatedAt,
	}
}
