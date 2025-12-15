package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupWiFiProductServiceTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Product{}, &entity.WiFiProduct{})
	require.NoError(t, err)

	return db
}

func createTestProductForWiFi(t *testing.T, db *gorm.DB) *entity.Product {
	product := &entity.Product{
		Name:  "WiFi Base Product",
		Price: 50000,
		SKU:   "WIFI-BASE-" + randomSKU(),
		Stock: 100,
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

// Helper for unique SKUs
func randomSKU() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func TestWiFiProductService_CreateWiFiProduct(t *testing.T) {
	ctx := context.Background()
	db := setupWiFiProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewWiFiProductRepository(db, logger)
	svc := NewWiFiProductService(repo, logger)

	product := createTestProductForWiFi(t, db)

	tests := []struct {
		name      string
		req       *dto.CreateWiFiProductRequest
		expectErr bool
		expectMsg string
	}{
		{
			name: "create valid wifi product",
			req: &dto.CreateWiFiProductRequest{
				ProductID:  product.ID,
				Quota:      10,
				Duration:   30,
				SpeedLimit: 10,
			},
			expectErr: false,
		},
		{
			name: "create wifi product with high quota",
			req: &dto.CreateWiFiProductRequest{
				ProductID:  product.ID,
				Quota:      100,
				Duration:   365,
				SpeedLimit: 100,
			},
			expectErr: true, // Second one should fail due to unique constraint
			expectMsg: "wifi product already exists",
		},
		{
			name: "create wifi product with zero quota",
			req: &dto.CreateWiFiProductRequest{
				ProductID:  product.ID,
				Quota:      0,
				Duration:   30,
				SpeedLimit: 10,
			},
			expectErr: true,
			expectMsg: "invalid wifi product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.CreateWiFiProduct(ctx, tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
				if tt.expectMsg != "" {
					assert.Contains(t, err.Error(), tt.expectMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.req.ProductID, resp.ProductID)
				assert.Equal(t, tt.req.Quota, resp.Quota)
				assert.Equal(t, tt.req.Duration, resp.Duration)
			}
		})
	}
}

func TestWiFiProductService_GetWiFiProductByID(t *testing.T) {
	ctx := context.Background()
	db := setupWiFiProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewWiFiProductRepository(db, logger)
	svc := NewWiFiProductService(repo, logger)

	product := createTestProductForWiFi(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      20,
		Duration:   30,
		SpeedLimit: 50,
	}
	err := repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		expectErr bool
	}{
		{
			name:      "get existing wifi product",
			id:        wifiProd.ID,
			expectErr: false,
		},
		{
			name:      "get non-existent wifi product",
			id:        999,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetWiFiProductByID(ctx, tt.id)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, product.ID, resp.ProductID)
				assert.Equal(t, 20.0, resp.Quota)
			}
		})
	}
}

func TestWiFiProductService_GetWiFiProductByProductID(t *testing.T) {
	ctx := context.Background()
	db := setupWiFiProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewWiFiProductRepository(db, logger)
	svc := NewWiFiProductService(repo, logger)

	product := createTestProductForWiFi(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      25,
		Duration:   60,
		SpeedLimit: 75,
	}
	err := repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name      string
		productID uint
		expectErr bool
	}{
		{
			name:      "get wifi product by existing product id",
			productID: product.ID,
			expectErr: false,
		},
		{
			name:      "get wifi product by non-existent product id",
			productID: 999,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetWiFiProductByProductID(ctx, tt.productID)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, product.ID, resp.ProductID)
				assert.Equal(t, 25.0, resp.Quota)
			}
		})
	}
}

func TestWiFiProductService_GetWiFiProducts(t *testing.T) {
	ctx := context.Background()
	db := setupWiFiProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewWiFiProductRepository(db, logger)
	svc := NewWiFiProductService(repo, logger)

	// Create test products and wifi products
	product1 := createTestProductForWiFi(t, db)
	product2 := createTestProductForWiFi(t, db)

	wifiProds := []entity.WiFiProduct{
		{ProductID: product1.ID, Quota: 10, Duration: 30, SpeedLimit: 10},
		{ProductID: product2.ID, Quota: 20, Duration: 60, SpeedLimit: 50},
	}

	for i := range wifiProds {
		err := repo.Create(ctx, &wifiProds[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		filter        *dto.WiFiProductFilter
		expectedCount int64
	}{
		{
			name:          "get all wifi products",
			filter:        &dto.WiFiProductFilter{Page: 1, Limit: 10},
			expectedCount: 2,
		},
		{
			name:          "get wifi products with limit",
			filter:        &dto.WiFiProductFilter{Page: 1, Limit: 1},
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.GetWiFiProducts(ctx, tt.filter)

			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.expectedCount, resp.Total)
		})
	}
}

func TestWiFiProductService_UpdateWiFiProduct(t *testing.T) {
	ctx := context.Background()
	db := setupWiFiProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewWiFiProductRepository(db, logger)
	svc := NewWiFiProductService(repo, logger)

	product := createTestProductForWiFi(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err := repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		req       *dto.UpdateWiFiProductRequest
		expectErr bool
	}{
		{
			name: "update wifi product successfully",
			id:   wifiProd.ID,
			req: &dto.UpdateWiFiProductRequest{
				Quota:      50,
				Duration:   90,
				SpeedLimit: 100,
			},
			expectErr: false,
		},
		{
			name:      "update non-existent wifi product",
			id:        999,
			req:       &dto.UpdateWiFiProductRequest{Quota: 10},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := svc.UpdateWiFiProduct(ctx, tt.id, tt.req)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, 50.0, resp.Quota)
				assert.Equal(t, 90, resp.Duration)
				assert.Equal(t, 100.0, resp.SpeedLimit)
			}
		})
	}
}

func TestWiFiProductService_DeleteWiFiProduct(t *testing.T) {
	ctx := context.Background()
	db := setupWiFiProductServiceTestDB(t)
	logger := zap.NewNop()
	repo := repository.NewWiFiProductRepository(db, logger)
	svc := NewWiFiProductService(repo, logger)

	product := createTestProductForWiFi(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      15,
		Duration:   30,
		SpeedLimit: 25,
	}
	err := repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		expectErr bool
	}{
		{
			name:      "delete existing wifi product",
			id:        wifiProd.ID,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.DeleteWiFiProduct(ctx, tt.id)

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
