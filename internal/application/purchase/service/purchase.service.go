package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/entity"
	purchaseRepo "github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/repository"
	walletDTO "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	walletService "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/queue"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type PurchaseService interface {
	CreatePurchase(req *dto.CreatePurchaseRequest) (*dto.PurchaseResponse, error)
	GetPurchaseByID(id uint) (*dto.PurchaseResponse, error)
	GetUserPurchases(userID uint, filter *dto.PurchaseFilter) (*dto.PurchaseListResponse, error)
	GetUserPurchasesWithDetails(userID uint, filter *dto.PurchaseFilter) (*dto.PurchaseListResponseWithDetails, error)
	GetAllPurchases(filter *dto.PurchaseFilter) (*dto.PurchaseListResponse, error)
	GetAllPurchasesWithDetails(filter *dto.PurchaseFilter) (*dto.PurchaseListResponseWithDetails, error)
	UpdatePurchaseStatus(id uint, status string) (*dto.PurchaseResponse, error)
}

type purchaseService struct {
	db            *gorm.DB
	purchaseRepo  purchaseRepo.PurchaseRepository
	productRepo   repository.ProductRepository
	walletService walletService.WalletService
	queueClient   *queue.Client
	logger        *zap.Logger
}

func NewPurchaseService(
	db *gorm.DB,
	purchaseRepo purchaseRepo.PurchaseRepository,
	productRepo repository.ProductRepository,
	walletService walletService.WalletService,
	logger *zap.Logger,
) PurchaseService {
	return &purchaseService{
		db:            db,
		purchaseRepo:  purchaseRepo,
		productRepo:   productRepo,
		walletService: walletService,
		queueClient:   nil, // Queue client is optional
		logger:        logger,
	}
}

func (s *purchaseService) CreatePurchase(req *dto.CreatePurchaseRequest) (*dto.PurchaseResponse, error) {
	// Validate request
	if req.UserID == 0 || req.ProductID == 0 || req.Quantity <= 0 {
		return nil, errors.New("invalid purchase data")
	}

	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get product with lock to prevent race conditions
	var product struct {
		ID    uint
		Price float64
		Stock int
	}
	if err := tx.Raw("SELECT id, price, stock FROM products WHERE id = ? FOR UPDATE", req.ProductID).Scan(&product).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("product not found")
		}
		s.logger.Error("Failed to get product", zap.Error(err))
		return nil, err
	}

	// Check stock
	if product.Stock < req.Quantity {
		tx.Rollback()
		return nil, errors.New("insufficient stock")
	}

	// Get wallet by user ID using transaction's DB
	var wallet struct {
		ID      uint
		Balance float64
	}
	if err := tx.Raw("SELECT id, balance FROM wallets WHERE user_id = ?", req.UserID).Scan(&wallet).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		s.logger.Error("Failed to get wallet", zap.Error(err))
		return nil, err
	}

	totalPrice := product.Price * float64(req.Quantity)
	if wallet.Balance < totalPrice {
		tx.Rollback()
		return nil, errors.New("insufficient wallet balance")
	}

	// Deduct wallet balance
	if err := tx.Table("wallets").
		Where("user_id = ?", req.UserID).
		Update("balance", gorm.Expr("balance - ?", totalPrice)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to deduct wallet balance", zap.Error(err))
		return nil, errors.New("failed to deduct wallet balance")
	}

	// Reduce product stock
	if err := tx.Table("products").
		Where("id = ?", req.ProductID).
		Update("stock", gorm.Expr("stock - ?", req.Quantity)).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to reduce stock", zap.Error(err))
		return nil, errors.New("failed to reduce stock")
	}

	// Create purchase record
	purchase := &entity.Purchase{
		UserID:     req.UserID,
		ProductID:  req.ProductID,
		Quantity:   req.Quantity,
		TotalPrice: totalPrice,
		Status:     entity.PurchaseStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := tx.Create(purchase).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create purchase", zap.Error(err))
		return nil, errors.New("failed to create purchase")
	}

	// Record wallet transaction within the same transaction
	transactionReq := &walletDTO.CreateTransactionRequest{
		WalletID:    wallet.ID,
		Type:        "payment",
		Amount:      totalPrice,
		Description: fmt.Sprintf("Purchase of product #%d (Qty: %d)", req.ProductID, req.Quantity),
	}

	// Create wallet transaction using the transaction's DB
	walletTransaction := map[string]interface{}{
		"wallet_id":     wallet.ID,
		"type":          "payment",
		"amount":        totalPrice,
		"balance_after": (wallet.Balance - totalPrice),
		"description":   transactionReq.Description,
		"status":        "completed",
		"created_at":    time.Now(),
		"updated_at":    time.Now(),
	}

	if err := tx.Table("wallet_transactions").Create(walletTransaction).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to record wallet transaction", zap.Error(err))
		return nil, errors.New("failed to record wallet transaction")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.Error(err))
		return nil, errors.New("failed to complete purchase")
	}

	s.logger.Info("Purchase created successfully", zap.Uint("id", purchase.ID), zap.Uint("user_id", req.UserID))

	// Send to background job queue for gRPC notification (non-blocking)
	if s.queueClient != nil {
		payload := map[string]interface{}{
			"purchase_id": purchase.ID,
			"user_id":     purchase.UserID,
			"product_id":  purchase.ProductID,
		}
		payloadBytes, _ := json.Marshal(payload)
		task := asynq.NewTask("purchase_notification", payloadBytes)
		s.queueClient.Enqueue(task)
	}

	return s.entityToResponse(purchase), nil
}

func (s *purchaseService) GetPurchaseByID(id uint) (*dto.PurchaseResponse, error) {
	purchase, err := s.purchaseRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("purchase not found")
		}
		s.logger.Error("Failed to get purchase", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(purchase), nil
}

func (s *purchaseService) GetUserPurchases(userID uint, filter *dto.PurchaseFilter) (*dto.PurchaseListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	purchases, count, err := s.purchaseRepo.GetByUserID(userID, filter)
	if err != nil {
		s.logger.Error("Failed to get purchases", zap.Error(err))
		return nil, err
	}

	responses := make([]dto.PurchaseResponse, len(purchases))
	for i, p := range purchases {
		responses[i] = *s.entityToResponse(&p)
	}

	totalPages := (count + int64(filter.Limit) - 1) / int64(filter.Limit)
	return &dto.PurchaseListResponse{
		Data:       responses,
		Total:      count,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *purchaseService) GetUserPurchasesWithDetails(userID uint, filter *dto.PurchaseFilter) (*dto.PurchaseListResponseWithDetails, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	// SQL query with JOIN to get user and product information
	type PurchaseDetail struct {
		ID           uint
		Quantity     int
		TotalPrice   float64
		Status       string
		Notes        string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		UserID       uint
		UserName     string
		UserEmail    string
		UserLevel    string
		ProductID    uint
		ProductName  string
		ProductSKU   string
		ProductPrice float64
		ProductStock int
	}

	var purchaseDetails []PurchaseDetail
	var count int64

	baseQuery := s.db.
		Table("purchases").
		Select(
			"purchases.id, purchases.quantity, purchases.total_price, purchases.status, purchases.notes, purchases.created_at, purchases.updated_at",
			"purchases.user_id, users.name as user_name, users.email as user_email, users.level as user_level",
			"purchases.product_id, products.name as product_name, products.sku as product_sku, products.price as product_price, products.stock as product_stock",
		).
		Joins("LEFT JOIN users ON purchases.user_id = users.id").
		Joins("LEFT JOIN products ON purchases.product_id = products.id").
		Where("purchases.user_id = ?", userID)

	// Apply status filter if provided
	if filter.Status != "" {
		baseQuery = baseQuery.Where("purchases.status = ?", filter.Status)
	}

	// Get total count
	if err := baseQuery.Count(&count).Error; err != nil {
		s.logger.Error("Failed to count user purchases with details", zap.Error(err), zap.Uint("user_id", userID))
		return nil, errors.New("failed to get purchases count")
	}

	// Get paginated results with JOIN
	offset := (filter.Page - 1) * filter.Limit
	if err := baseQuery.
		Offset(offset).
		Limit(filter.Limit).
		Order("purchases.created_at DESC").
		Scan(&purchaseDetails).Error; err != nil {
		s.logger.Error("Failed to get user purchases with details", zap.Error(err), zap.Uint("user_id", userID))
		return nil, errors.New("failed to get purchases")
	}

	// Transform to response format
	responses := make([]dto.PurchaseDetailResponse, len(purchaseDetails))
	for i, pd := range purchaseDetails {
		responses[i] = dto.PurchaseDetailResponse{
			ID:         pd.ID,
			Quantity:   pd.Quantity,
			TotalPrice: pd.TotalPrice,
			Status:     pd.Status,
			Notes:      pd.Notes,
			CreatedAt:  pd.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  pd.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			User: &dto.PurchaseUserInfo{
				ID:    pd.UserID,
				Name:  pd.UserName,
				Email: pd.UserEmail,
				Level: pd.UserLevel,
			},
			Product: &dto.PurchaseProductInfo{
				ID:    pd.ProductID,
				Name:  pd.ProductName,
				SKU:   pd.ProductSKU,
				Price: pd.ProductPrice,
				Stock: pd.ProductStock,
			},
		}
	}

	totalPages := (count + int64(filter.Limit) - 1) / int64(filter.Limit)
	return &dto.PurchaseListResponseWithDetails{
		Data:       responses,
		Total:      count,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *purchaseService) GetAllPurchases(filter *dto.PurchaseFilter) (*dto.PurchaseListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	// Get all purchases with pagination using raw SQL
	var purchases []entity.Purchase
	var count int64

	query := s.db.Model(&entity.Purchase{})

	// Apply status filter if provided
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	// Get total count
	if err := query.Count(&count).Error; err != nil {
		s.logger.Error("Failed to count purchases", zap.Error(err))
		return nil, errors.New("failed to get purchases count")
	}

	// Get paginated results
	offset := (filter.Page - 1) * filter.Limit
	if err := query.
		Offset(offset).
		Limit(filter.Limit).
		Order("created_at DESC").
		Find(&purchases).Error; err != nil {
		s.logger.Error("Failed to get all purchases", zap.Error(err))
		return nil, errors.New("failed to get purchases")
	}

	responses := make([]dto.PurchaseResponse, len(purchases))
	for i, p := range purchases {
		responses[i] = *s.entityToResponse(&p)
	}

	totalPages := (count + int64(filter.Limit) - 1) / int64(filter.Limit)
	return &dto.PurchaseListResponse{
		Data:       responses,
		Total:      count,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *purchaseService) GetAllPurchasesWithDetails(filter *dto.PurchaseFilter) (*dto.PurchaseListResponseWithDetails, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	// SQL query with JOIN to get user and product information
	type PurchaseDetail struct {
		ID           uint
		Quantity     int
		TotalPrice   float64
		Status       string
		Notes        string
		CreatedAt    time.Time
		UpdatedAt    time.Time
		UserID       uint
		UserName     string
		UserEmail    string
		UserLevel    string
		ProductID    uint
		ProductName  string
		ProductSKU   string
		ProductPrice float64
		ProductStock int
	}

	var purchaseDetails []PurchaseDetail
	var count int64

	baseQuery := s.db.
		Table("purchases").
		Select(
			"purchases.id, purchases.quantity, purchases.total_price, purchases.status, purchases.notes, purchases.created_at, purchases.updated_at",
			"purchases.user_id, users.name as user_name, users.email as user_email, users.level as user_level",
			"purchases.product_id, products.name as product_name, products.sku as product_sku, products.price as product_price, products.stock as product_stock",
		).
		Joins("LEFT JOIN users ON purchases.user_id = users.id").
		Joins("LEFT JOIN products ON purchases.product_id = products.id")

	// Apply filters
	if filter.Status != "" {
		baseQuery = baseQuery.Where("purchases.status = ?", filter.Status)
	}

	// Get total count
	if err := baseQuery.Count(&count).Error; err != nil {
		s.logger.Error("Failed to count purchases with details", zap.Error(err))
		return nil, errors.New("failed to get purchases count")
	}

	// Get paginated results with JOIN
	offset := (filter.Page - 1) * filter.Limit
	if err := baseQuery.
		Offset(offset).
		Limit(filter.Limit).
		Order("purchases.created_at DESC").
		Scan(&purchaseDetails).Error; err != nil {
		s.logger.Error("Failed to get purchases with details", zap.Error(err))
		return nil, errors.New("failed to get purchases")
	}

	// Transform to response format
	responses := make([]dto.PurchaseDetailResponse, len(purchaseDetails))
	for i, pd := range purchaseDetails {
		responses[i] = dto.PurchaseDetailResponse{
			ID:         pd.ID,
			Quantity:   pd.Quantity,
			TotalPrice: pd.TotalPrice,
			Status:     pd.Status,
			Notes:      pd.Notes,
			CreatedAt:  pd.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:  pd.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			User: &dto.PurchaseUserInfo{
				ID:    pd.UserID,
				Name:  pd.UserName,
				Email: pd.UserEmail,
				Level: pd.UserLevel,
			},
			Product: &dto.PurchaseProductInfo{
				ID:    pd.ProductID,
				Name:  pd.ProductName,
				SKU:   pd.ProductSKU,
				Price: pd.ProductPrice,
				Stock: pd.ProductStock,
			},
		}
	}

	totalPages := (count + int64(filter.Limit) - 1) / int64(filter.Limit)
	return &dto.PurchaseListResponseWithDetails{
		Data:       responses,
		Total:      count,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *purchaseService) UpdatePurchaseStatus(id uint, status string) (*dto.PurchaseResponse, error) {
	purchase, err := s.purchaseRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("purchase not found")
		}
		return nil, err
	}

	purchase.Status = entity.PurchaseStatus(status)
	purchase.UpdatedAt = time.Now()

	if err := s.purchaseRepo.Update(purchase); err != nil {
		s.logger.Error("Failed to update purchase status", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Purchase status updated", zap.Uint("id", id), zap.String("status", status))
	return s.entityToResponse(purchase), nil
}

func (s *purchaseService) entityToResponse(purchase *entity.Purchase) *dto.PurchaseResponse {
	return &dto.PurchaseResponse{
		ID:         purchase.ID,
		UserID:     purchase.UserID,
		ProductID:  purchase.ProductID,
		Quantity:   purchase.Quantity,
		TotalPrice: purchase.TotalPrice,
		Status:     string(purchase.Status),
		Notes:      purchase.Notes,
		CreatedAt:  purchase.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  purchase.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
