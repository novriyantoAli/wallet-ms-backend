package repository

import (
	"testing"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupProductTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Product{})
	require.NoError(t, err)

	return db
}

func TestProductRepository_Create(t *testing.T) {
	db := setupProductTestDB(t)
	logger := zap.NewNop()
	repo := NewProductRepository(db, logger)

	tests := []struct {
		name      string
		product   *entity.Product
		expectErr bool
	}{
		{
			name: "create valid wifi product with active status",
			product: &entity.Product{
				Name:        "WiFi 10GB",
				Description: "WiFi package",
				Price:       50000,
				SKU:         "WIFI-001",
				Category:    entity.ProductCategoryWiFi,
				Status:      entity.ProductStatusActive,
				Stock:       100,
			},
			expectErr: false,
		},
		{
			name: "create valid pulsa product with inactive status",
			product: &entity.Product{
				Name:        "Pulsa 50k",
				Description: "Pulsa package",
				Price:       50000,
				SKU:         "PULSA-001",
				Category:    entity.ProductCategoryPulsa,
				Status:      entity.ProductStatusInactive,
				Stock:       200,
			},
			expectErr: false,
		},
		{
			name: "create product with minimum fields and default status",
			product: &entity.Product{
				Name:     "Phone",
				Price:    5000,
				SKU:      "PHN-001",
				Category: entity.ProductCategoryWiFi,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.product)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.product.ID)

				// Verify it was saved with all fields including status
				saved, err := repo.GetByID(tt.product.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.product.Name, saved.Name)
				assert.Equal(t, tt.product.Price, saved.Price)
				assert.Equal(t, tt.product.Category, saved.Category)
				assert.Equal(t, tt.product.SKU, saved.SKU)
				// Verify status is set (either specified or default active)
				expectedStatus := tt.product.Status
				if expectedStatus == "" {
					expectedStatus = entity.ProductStatusInactive
				}
				assert.Equal(t, expectedStatus, saved.Status)
			}
		})
	}
}

func TestProductRepository_GetByID(t *testing.T) {
	db := setupProductTestDB(t)
	logger := zap.NewNop()
	repo := NewProductRepository(db, logger)

	// Create a test product with category
	product := &entity.Product{
		Name:     "Test Product",
		Price:    5000,
		SKU:      "TEST-001",
		Stock:    10,
		Category: entity.ProductCategoryWiFi,
	}
	err := repo.Create(product)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		expectErr bool
		expectNil bool
	}{
		{
			name:      "get existing product",
			id:        product.ID,
			expectErr: false,
			expectNil: false,
		},
		{
			name:      "get non-existent product",
			id:        999,
			expectErr: true,
			expectNil: true,
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
				assert.Equal(t, "Test Product", result.Name)
				assert.Equal(t, 5000.0, result.Price)
				assert.Equal(t, entity.ProductCategoryWiFi, result.Category)
			}
		})
	}
}

func TestProductRepository_GetBySKU(t *testing.T) {
	db := setupProductTestDB(t)
	logger := zap.NewNop()
	repo := NewProductRepository(db, logger)

	product := &entity.Product{
		Name:     "Test Product",
		Price:    5000,
		SKU:      "UNIQUE-SKU-001",
		Stock:    10,
		Category: entity.ProductCategoryWiFi,
	}
	err := repo.Create(product)
	require.NoError(t, err)

	tests := []struct {
		name      string
		sku       string
		expectErr bool
	}{
		{
			name:      "get product by existing SKU",
			sku:       "UNIQUE-SKU-001",
			expectErr: false,
		},
		{
			name:      "get product by non-existent SKU",
			sku:       "NON-EXISTENT-SKU",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetBySKU(tt.sku)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Test Product", result.Name)
				assert.Equal(t, tt.sku, result.SKU)
				assert.Equal(t, entity.ProductCategoryWiFi, result.Category)
			}
		})
	}
}

func TestProductRepository_GetAll(t *testing.T) {
	db := setupProductTestDB(t)
	logger := zap.NewNop()
	repo := NewProductRepository(db, logger)

	// Create test products
	products := []entity.Product{
		{Name: "Laptop", Price: 10000, SKU: "LAP-001", Stock: 5, Category: entity.ProductCategoryWiFi},
		{Name: "Phone", Price: 5000, SKU: "PHN-001", Stock: 10, Category: entity.ProductCategoryPulsa},
		{Name: "Tablet", Price: 3000, SKU: "TAB-001", Stock: 15, Category: entity.ProductCategoryWiFi},
	}

	for i := range products {
		err := repo.Create(&products[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name             string
		filter           *dto.ProductFilter
		expectedCount    int64
		expectedMinItems int
	}{
		{
			name:             "get all products with default pagination",
			filter:           &dto.ProductFilter{Page: 1, Limit: 10},
			expectedCount:    3,
			expectedMinItems: 3,
		},
		{
			name:             "get products with limit",
			filter:           &dto.ProductFilter{Page: 1, Limit: 2},
			expectedCount:    3,
			expectedMinItems: 2,
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

func TestProductRepository_Update(t *testing.T) {
	db := setupProductTestDB(t)
	logger := zap.NewNop()
	repo := NewProductRepository(db, logger)

	// Create a test product
	product := &entity.Product{
		Name:     "Original Name",
		Price:    5000,
		SKU:      "UPDATE-TEST",
		Stock:    10,
		Category: entity.ProductCategoryWiFi,
	}
	err := repo.Create(product)
	require.NoError(t, err)

	tests := []struct {
		name      string
		product   *entity.Product
		expectErr bool
	}{
		{
			name: "update product fields including status",
			product: &entity.Product{
				ID:       product.ID,
				Name:     "Updated Name",
				Price:    6000,
				SKU:      "UPDATE-TEST",
				Stock:    20,
				Category: entity.ProductCategoryPulsa,
				Status:   entity.ProductStatusInactive,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(tt.product)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the update including status
				saved, err := repo.GetByID(tt.product.ID)
				assert.NoError(t, err)
				assert.Equal(t, "Updated Name", saved.Name)
				assert.Equal(t, 6000.0, saved.Price)
				assert.Equal(t, 20, saved.Stock)
				assert.Equal(t, entity.ProductCategoryPulsa, saved.Category)
				assert.Equal(t, entity.ProductStatusInactive, saved.Status)
			}
		})
	}
}

func TestProductRepository_Delete(t *testing.T) {
	db := setupProductTestDB(t)
	logger := zap.NewNop()
	repo := NewProductRepository(db, logger)

	product := &entity.Product{
		Name:     "Delete Test",
		Price:    5000,
		SKU:      "DEL-TEST",
		Stock:    5,
		Category: entity.ProductCategoryWiFi,
	}
	err := repo.Create(product)
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
