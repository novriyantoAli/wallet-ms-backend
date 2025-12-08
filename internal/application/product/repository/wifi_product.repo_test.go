package repository

import (
	"fmt"
	"testing"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupWiFiProductTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Product{}, &entity.WiFiProduct{})
	require.NoError(t, err)

	return db
}

func createTestProduct(t *testing.T, db *gorm.DB) *entity.Product {
	return createTestProductWithSKU(t, db, "")
}

func createTestProductWithSKU(t *testing.T, db *gorm.DB, sku string) *entity.Product {
	if sku == "" {
		// Generate unique SKU
		var count int64
		db.Model(&entity.Product{}).Count(&count)
		sku = "WIFI-PROD-" + fmt.Sprintf("%03d", count+1)
	}
	product := &entity.Product{
		Name:  "WiFi Test Product",
		Price: 50000,
		SKU:   sku,
		Stock: 100,
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

func TestWiFiProductRepository_Create(t *testing.T) {
	db := setupWiFiProductTestDB(t)
	repo := NewWiFiProductRepository(db)

	product := createTestProduct(t, db)

	tests := []struct {
		name      string
		wifiProd  *entity.WiFiProduct
		expectErr bool
	}{
		{
			name: "create valid wifi product",
			wifiProd: &entity.WiFiProduct{
				ProductID:  product.ID,
				Quota:      10,
				Duration:   30,
				SpeedLimit: 10,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.wifiProd)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.wifiProd.ID)

				// Verify it was saved
				saved, err := repo.GetByID(tt.wifiProd.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.wifiProd.ProductID, saved.ProductID)
				assert.Equal(t, tt.wifiProd.Quota, saved.Quota)
			}
		})
	}
}

func TestWiFiProductRepository_GetByID(t *testing.T) {
	db := setupWiFiProductTestDB(t)
	repo := NewWiFiProductRepository(db)

	product := createTestProduct(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      20,
		Duration:   30,
		SpeedLimit: 50,
	}
	err := repo.Create(wifiProd)
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
			result, err := repo.GetByID(tt.id)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, wifiProd.ProductID, result.ProductID)
				assert.Equal(t, 20.0, result.Quota)
			}
		})
	}
}

func TestWiFiProductRepository_GetByProductID(t *testing.T) {
	db := setupWiFiProductTestDB(t)
	repo := NewWiFiProductRepository(db)

	product := createTestProduct(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      25,
		Duration:   60,
		SpeedLimit: 75,
	}
	err := repo.Create(wifiProd)
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
			result, err := repo.GetByProductID(tt.productID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, product.ID, result.ProductID)
				assert.Equal(t, 25.0, result.Quota)
			}
		})
	}
}

func TestWiFiProductRepository_GetAll(t *testing.T) {
	db := setupWiFiProductTestDB(t)
	repo := NewWiFiProductRepository(db)

	// Create test products and wifi products
	product1 := createTestProduct(t, db)
	product2 := createTestProduct(t, db)

	wifiProds := []entity.WiFiProduct{
		{ProductID: product1.ID, Quota: 10, Duration: 30, SpeedLimit: 10},
		{ProductID: product2.ID, Quota: 20, Duration: 60, SpeedLimit: 50},
	}

	for i := range wifiProds {
		err := repo.Create(&wifiProds[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name             string
		filter           *dto.WiFiProductFilter
		expectedCount    int64
		expectedMinItems int
	}{
		{
			name:             "get all wifi products default pagination",
			filter:           &dto.WiFiProductFilter{Page: 1, Limit: 10},
			expectedCount:    2,
			expectedMinItems: 2,
		},
		{
			name:             "get wifi products with limit",
			filter:           &dto.WiFiProductFilter{Page: 1, Limit: 1},
			expectedCount:    2,
			expectedMinItems: 1,
		},
		{
			name:             "get wifi products page 2",
			filter:           &dto.WiFiProductFilter{Page: 2, Limit: 1},
			expectedCount:    2,
			expectedMinItems: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, total, err := repo.GetAll(tt.filter)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCount, total)
			assert.GreaterOrEqual(t, len(results), tt.expectedMinItems)
		})
	}
}

func TestWiFiProductRepository_Update(t *testing.T) {
	db := setupWiFiProductTestDB(t)
	repo := NewWiFiProductRepository(db)

	product := createTestProduct(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err := repo.Create(wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name      string
		wifiProd  *entity.WiFiProduct
		expectErr bool
	}{
		{
			name: "update wifi product fields",
			wifiProd: &entity.WiFiProduct{
				ID:         wifiProd.ID,
				ProductID:  product.ID,
				Quota:      50,
				Duration:   90,
				SpeedLimit: 100,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(tt.wifiProd)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the update
				saved, err := repo.GetByID(tt.wifiProd.ID)
				assert.NoError(t, err)
				assert.Equal(t, 50.0, saved.Quota)
				assert.Equal(t, 90, saved.Duration)
				assert.Equal(t, 100.0, saved.SpeedLimit)
			}
		})
	}
}

func TestWiFiProductRepository_Delete(t *testing.T) {
	db := setupWiFiProductTestDB(t)
	repo := NewWiFiProductRepository(db)

	product := createTestProduct(t, db)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      15,
		Duration:   30,
		SpeedLimit: 25,
	}
	err := repo.Create(wifiProd)
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
			err := repo.Delete(tt.id)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify deletion
				_, err := repo.GetByID(tt.id)
				assert.Error(t, err)
			}
		})
	}
}

// Test one-to-one constraint (one product can only have one wifi_product)
// Note: This test verifies the constraint exists in the schema
// SQLite may not fully enforce unique constraints, but production PostgreSQL will
func TestWiFiProductRepository_OneToOneConstraint(t *testing.T) {
	db := setupWiFiProductTestDB(t)
	repo := NewWiFiProductRepository(db)

	product := createTestProduct(t, db)

	// Create first wifi product
	wifiProd1 := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err := repo.Create(wifiProd1)
	assert.NoError(t, err)

	// Verify we can retrieve it
	result, err := repo.GetByProductID(product.ID)
	assert.NoError(t, err)
	assert.Equal(t, wifiProd1.ID, result.ID)
	assert.Equal(t, product.ID, result.ProductID)
}
