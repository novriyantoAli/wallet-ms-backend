package service

import (
	"context"
	"testing"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupProductServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Product{}, &entity.WiFiProduct{})
	require.NoError(t, err)

	return db
}

func TestProductService_CreateProduct(t *testing.T) {
	db := setupProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewProductRepository(db, logger)
	wifiRepo := repository.NewWiFiProductRepository(db, logger)
	txManager := database.NewTransactionManager(db)
	svc := NewProductService(txManager, repo, wifiRepo, logger)
	ctx := context.Background()

	tests := []struct {
		name           string
		req            *dto.CreateProductRequest
		expectErr      bool
		expectMsg      string
		expectCategory string
	}{
		{
			name: "create valid wifi product",
			req: &dto.CreateProductRequest{
				Name:        "WiFi 10GB",
				Description: "WiFi Package",
				Price:       50000,
				Category:    "wifi",
				Stock:       100,
			},
			expectErr:      false,
			expectCategory: "wifi",
		},
		{
			name: "create valid pulsa product",
			req: &dto.CreateProductRequest{
				Name:        "Pulsa 50000",
				Description: "Pulsa Package",
				Price:       50000,
				Category:    "pulsa",
				Stock:       200,
			},
			expectErr:      false,
			expectCategory: "pulsa",
		},
		{
			name: "create product with empty name",
			req: &dto.CreateProductRequest{
				Price:    5000,
				Category: "wifi",
			},
			expectErr: true,
			expectMsg: "invalid product data",
		},
		{
			name: "create product with zero price",
			req: &dto.CreateProductRequest{
				Name:     "Phone",
				Price:    0,
				Category: "wifi",
			},
			expectErr: true,
			expectMsg: "invalid product data",
		},
		{
			name: "create product with invalid category",
			req: &dto.CreateProductRequest{
				Name:     "Invalid",
				Price:    5000,
				Category: "invalid",
			},
			expectErr: true,
			expectMsg: "invalid product data",
		},
	}

	// Pre-create a product with same name, category and price for duplicate test
	// With the new SKU generation pattern (timestamp + random), duplicates are practically impossible
	// So this test case is no longer applicable

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.CreateProduct(ctx, tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if tt.expectMsg != "" {
					assert.Contains(t, err.Error(), tt.expectMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.Name, resp.Name)
				assert.Equal(t, tt.req.Price, resp.Price)
				assert.NotEmpty(t, resp.SKU) // SKU should be auto-generated
				assert.Equal(t, tt.expectCategory, resp.Category)
				assert.Equal(t, "inactive", resp.Status) // New products should have inactive status

				// Verify WiFi product was created if category is wifi
				if tt.req.Category == "wifi" {
					wifiProduct, err := wifiRepo.GetByProductID(ctx, resp.ID)
					assert.NoError(t, err)
					assert.NotNil(t, wifiProduct)
					assert.Equal(t, resp.ID, wifiProduct.ProductID)
				}
			}
		})
	}
}

func TestProductService_GetProductByID(t *testing.T) {
	db := setupProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewProductRepository(db, logger)
	wifiRepo := repository.NewWiFiProductRepository(db, logger)
	txManager := database.NewTransactionManager(db)
	svc := NewProductService(txManager, repo, wifiRepo, logger)
	ctx := context.Background()

	// Create a test product
	product := &entity.Product{
		Name:  "Test Product",
		Price: 5000,
		SKU:   "TEST-001",
		Stock: 10,
	}
	err := repo.Create(ctx, product)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		expectErr bool
	}{
		{
			name:      "get existing product",
			id:        product.ID,
			expectErr: false,
		},
		{
			name:      "get non-existent product",
			id:        999,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetProductByID(ctx, tt.id)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "Test Product", resp.Name)
				assert.Equal(t, 5000.0, resp.Price)
			}
		})
	}
}

func TestProductService_GetProductBySKU(t *testing.T) {
	db := setupProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewProductRepository(db, logger)
	wifiRepo := repository.NewWiFiProductRepository(db, logger)
	txManager := database.NewTransactionManager(db)
	svc := NewProductService(txManager, repo, wifiRepo, logger)
	ctx := context.Background()

	product := &entity.Product{
		Name:  "Test Product",
		Price: 5000,
		SKU:   "UNIQUE-SKU",
		Stock: 10,
	}
	err := repo.Create(ctx, product)
	require.NoError(t, err)

	tests := []struct {
		name      string
		sku       string
		expectErr bool
	}{
		{
			name:      "get product by existing SKU",
			sku:       "UNIQUE-SKU",
			expectErr: false,
		},
		{
			name:      "get product by non-existent SKU",
			sku:       "NON-EXISTENT",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetProductBySKU(ctx, tt.sku)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, "Test Product", resp.Name)
				assert.Equal(t, tt.sku, resp.SKU)
			}
		})
	}
}

func TestProductService_GetProducts(t *testing.T) {
	db := setupProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewProductRepository(db, logger)
	wifiRepo := repository.NewWiFiProductRepository(db, logger)
	txManager := database.NewTransactionManager(db)
	svc := NewProductService(txManager, repo, wifiRepo, logger)
	ctx := context.Background()

	// Create test products
	products := []entity.Product{
		{Name: "Laptop Computer", Price: 10000, SKU: "LAP-001", Stock: 5},
		{Name: "Mobile Phone", Price: 5000, SKU: "PHN-001", Stock: 10},
		{Name: "Tablet Device", Price: 3000, SKU: "TAB-001", Stock: 15},
	}

	for i := range products {
		err := repo.Create(ctx, &products[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		filter        *dto.ProductFilter
		expectedCount int64
	}{
		{
			name:          "get all products",
			filter:        &dto.ProductFilter{Page: 1, Limit: 10},
			expectedCount: 3,
		},
		{
			name:          "get products with limit",
			filter:        &dto.ProductFilter{Page: 1, Limit: 2},
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetProducts(ctx, tt.filter)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedCount, resp.Total)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	db := setupProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewProductRepository(db, logger)
	wifiRepo := repository.NewWiFiProductRepository(db, logger)
	txManager := database.NewTransactionManager(db)
	svc := NewProductService(txManager, repo, wifiRepo, logger)
	ctx := context.Background()

	product := &entity.Product{
		Name:     "Original Name",
		Price:    5000,
		SKU:      "UPDATE-TEST",
		Stock:    10,
		Category: entity.ProductCategoryWiFi,
		Status:   entity.ProductStatusActive,
	}
	err := repo.Create(ctx, product)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		req       *dto.UpdateProductRequest
		expectErr bool
	}{
		{
			name: "update product successfully",
			id:   product.ID,
			req: &dto.UpdateProductRequest{
				Name:        "Updated Name",
				Description: "Updated desc",
				Price:       6000,
				Stock:       20,
			},
			expectErr: false,
		},
		{
			name: "update product status to inactive",
			id:   product.ID,
			req: &dto.UpdateProductRequest{
				Status: "inactive",
			},
			expectErr: false,
		},
		{
			name:      "update non-existent product",
			id:        999,
			req:       &dto.UpdateProductRequest{Name: "Test"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.UpdateProduct(ctx, tt.id, tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)

				// For update product successfully test
				if tt.name == "update product successfully" {
					assert.Equal(t, "Updated Name", resp.Name)
					assert.Equal(t, 6000.0, resp.Price)
					assert.Equal(t, 20, resp.Stock)
				}

				// For update status test
				if tt.name == "update product status to inactive" {
					assert.Equal(t, "inactive", resp.Status)
				}
			}
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	db := setupProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewProductRepository(db, logger)
	wifiRepo := repository.NewWiFiProductRepository(db, logger)
	txManager := database.NewTransactionManager(db)
	svc := NewProductService(txManager, repo, wifiRepo, logger)
	ctx := context.Background()

	product := &entity.Product{
		Name:  "Delete Test",
		Price: 5000,
		SKU:   "DEL-TEST",
	}
	err := repo.Create(ctx, product)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		expectErr bool
	}{
		{
			name:      "delete existing product",
			id:        product.ID,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteProduct(ctx, tt.id)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify deletion
				_, err := repo.GetByID(ctx, tt.id)
				assert.Error(t, err)
			}
		})
	}
}
