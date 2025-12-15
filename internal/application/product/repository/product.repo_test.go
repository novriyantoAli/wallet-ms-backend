package repository

import (
	"context"
	"fmt"
	"testing"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProductRepository_Create(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewProductRepository(db, logger)
	ctx := context.Background()

	t.Run("should create product successfully", func(t *testing.T) {
		// Given
		product := &entity.Product{
			Name:        "WiFi 10GB",
			Description: "WiFi package",
			Price:       50000,
			SKU:         "PRODUCT-wifi-20251214-ABC123",
			Category:    entity.ProductCategoryWiFi,
			Status:      entity.ProductStatusActive,
			Stock:       100,
		}

		// When
		err := repo.Create(ctx, product)

		// Then
		assert.NoError(t, err)
		assert.NotZero(t, product.ID)

		// Verify product was created in database
		var dbProduct entity.Product
		err = db.First(&dbProduct, product.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, product.SKU, dbProduct.SKU)
		assert.Equal(t, product.Name, dbProduct.Name)
		assert.Equal(t, product.Price, dbProduct.Price)
	})

	t.Run("should fail to create product with duplicate SKU", func(t *testing.T) {
		// Given
		product1 := &entity.Product{
			Name:     "Product 1",
			Price:    50000,
			SKU:      "DUPLICATE-SKU",
			Category: entity.ProductCategoryWiFi,
		}

		product2 := &entity.Product{
			Name:     "Product 2",
			Price:    50000,
			SKU:      "DUPLICATE-SKU",
			Category: entity.ProductCategoryWiFi,
		}

		// When
		err1 := repo.Create(ctx, product1)
		err2 := repo.Create(ctx, product2)

		// Then
		assert.NoError(t, err1)
		assert.Error(t, err2) // Should fail due to unique constraint
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestProductRepository_GetByID(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewProductRepository(db, logger)
	ctx := context.Background()

	t.Run("should get product by ID successfully", func(t *testing.T) {
		// Given
		product := &entity.Product{
			Name:     "Test Product",
			Price:    5000,
			SKU:      "PRODUCT-test-20251214-TEST01",
			Stock:    10,
			Category: entity.ProductCategoryWiFi,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		// When
		foundProduct, err := repo.GetByID(ctx, product.ID)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, product.ID, foundProduct.ID)
		assert.Equal(t, product.SKU, foundProduct.SKU)
		assert.Equal(t, product.Name, foundProduct.Name)
		assert.Equal(t, product.Price, foundProduct.Price)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		// When
		_, err := repo.GetByID(ctx, 999)

		// Then
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestProductRepository_GetBySKU(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewProductRepository(db, logger)
	ctx := context.Background()

	t.Run("should get product by SKU successfully", func(t *testing.T) {
		// Given
		product := &entity.Product{
			Name:     "Test Product",
			Price:    5000,
			SKU:      "PRODUCT-unique-20251214-ABC123",
			Stock:    10,
			Category: entity.ProductCategoryWiFi,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		// When
		foundProduct, err := repo.GetBySKU(ctx, product.SKU)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, product.ID, foundProduct.ID)
		assert.Equal(t, product.SKU, foundProduct.SKU)
		assert.Equal(t, product.Name, foundProduct.Name)
		assert.Equal(t, product.Category, foundProduct.Category)
	})

	t.Run("should return error when SKU not found", func(t *testing.T) {
		// When
		_, err := repo.GetBySKU(ctx, "NON-EXISTENT-SKU")

		// Then
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestProductRepository_GetAll(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewProductRepository(db, logger)
	ctx := context.Background()

	t.Run("should get all products with pagination", func(t *testing.T) {
		// Given - Create multiple products
		for i := 0; i < 5; i++ {
			product := &entity.Product{
				Name:     fmt.Sprintf("Product %d", i),
				Price:    float64(5000 * (i + 1)),
				SKU:      fmt.Sprintf("PRODUCT-test-%d-SKU%02d", 20251214, i),
				Stock:    i * 10,
				Category: entity.ProductCategoryWiFi,
			}
			err := repo.Create(ctx, product)
			require.NoError(t, err)
		}

		filter := &dto.ProductFilter{
			Page:  1,
			Limit: 3,
		}

		// When
		products, totalCount, err := repo.GetAll(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Len(t, products, 3)            // Should return 3 products due to limit
		assert.Equal(t, int64(5), totalCount) // Total count should be 5
	})

	// Cleanup
	testutil.CleanDB(db)

	// New setup for second test
	db2, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger2 := testutil.NewTestLogger(t)
	repo2 := NewProductRepository(db2, logger2)

	t.Run("should filter products by search", func(t *testing.T) {
		// Given
		product1 := &entity.Product{
			Name:     "Laptop Computer",
			Price:    15000,
			SKU:      "PRODUCT-laptop-20251214-PC001",
			Stock:    5,
			Category: entity.ProductCategoryWiFi,
		}
		err := repo2.Create(ctx, product1)
		require.NoError(t, err)

		product2 := &entity.Product{
			Name:     "WiFi Router",
			Price:    3000,
			SKU:      "PRODUCT-wifi-20251214-RT001",
			Stock:    20,
			Category: entity.ProductCategoryWiFi,
		}
		err = repo2.Create(ctx, product2)
		require.NoError(t, err)

		filter := &dto.ProductFilter{
			Search: "Laptop",
			Page:   1,
			Limit:  10,
		}

		// When
		products, totalCount, err := repo2.GetAll(ctx, filter)

		// Then
		assert.NoError(t, err)
		if assert.Greater(t, len(products), 0, "expected at least one product matching 'Laptop'") {
			assert.Equal(t, "Laptop Computer", products[0].Name)
		}
		assert.GreaterOrEqual(t, totalCount, int64(1))
	})

	// Cleanup
	testutil.CleanDB(db2)
}

func TestProductRepository_Update(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewProductRepository(db, logger)
	ctx := context.Background()

	t.Run("should update product successfully", func(t *testing.T) {
		// Given
		product := &entity.Product{
			Name:     "Original Name",
			Price:    5000,
			SKU:      "PRODUCT-update-20251214-UPD01",
			Stock:    10,
			Category: entity.ProductCategoryWiFi,
			Status:   entity.ProductStatusActive,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		// When
		product.Name = "Updated Name"
		product.Price = 6000
		product.Stock = 20
		product.Status = entity.ProductStatusInactive
		err = repo.Update(ctx, product)

		// Then
		assert.NoError(t, err)

		// Verify the update
		savedProduct, err := repo.GetByID(ctx, product.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", savedProduct.Name)
		assert.Equal(t, 6000.0, savedProduct.Price)
		assert.Equal(t, 20, savedProduct.Stock)
		assert.Equal(t, entity.ProductStatusInactive, savedProduct.Status)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestProductRepository_Delete(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewProductRepository(db, logger)
	ctx := context.Background()

	t.Run("should delete product successfully", func(t *testing.T) {
		// Given
		product := &entity.Product{
			Name:     "Delete Test",
			Price:    5000,
			SKU:      "PRODUCT-delete-20251214-DEL01",
			Stock:    5,
			Category: entity.ProductCategoryWiFi,
		}
		err := repo.Create(ctx, product)
		require.NoError(t, err)

		// When
		err = repo.Delete(ctx, product.ID)

		// Then
		assert.NoError(t, err)

		// Verify product was deleted
		_, err = repo.GetByID(ctx, product.ID)
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}
