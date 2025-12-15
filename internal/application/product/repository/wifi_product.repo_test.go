package repository

import (
	"context"
	"testing"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWiFiProductRepository_Create(t *testing.T) {
	// Setup
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewWiFiProductRepository(db, logger)
	ctx := context.Background()

	t.Run("should create WiFi product successfully", func(t *testing.T) {
		// Given
		product := &entity.Product{
			Name:     "WiFi Product",
			Price:    50000,
			SKU:      "PRODUCT-wifi-20251214-WF001",
			Category: entity.ProductCategoryWiFi,
			Stock:    100,
		}
		err := db.Create(product).Error
		require.NoError(t, err)

		wifiProd := &entity.WiFiProduct{
			ProductID:  product.ID,
			Quota:      10,
			Duration:   30,
			SpeedLimit: 10,
		}

		// When
		err = repo.Create(ctx, wifiProd)

		// Then
		assert.NoError(t, err)
		assert.NotZero(t, wifiProd.ID)

		// Verify product was created in database
		var dbWiFiProd entity.WiFiProduct
		err = db.First(&dbWiFiProd, wifiProd.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, wifiProd.ProductID, dbWiFiProd.ProductID)
		assert.Equal(t, wifiProd.Quota, dbWiFiProd.Quota)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestWiFiProductRepository_GetByID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewWiFiProductRepository(db, logger)
	ctx := context.Background()

	// Given - create test product and wifi product
	product := &entity.Product{
		Name:     "Test WiFi",
		Price:    50000,
		SKU:      "PRODUCT-wifi-20251214-WF002",
		Category: entity.ProductCategoryWiFi,
		Stock:    100,
	}
	err = db.Create(product).Error
	require.NoError(t, err)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      20,
		Duration:   30,
		SpeedLimit: 50,
	}
	err = repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	t.Run("should get WiFi product by ID successfully", func(t *testing.T) {
		// When
		result, err := repo.GetByID(ctx, wifiProd.ID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, wifiProd.ProductID, result.ProductID)
		assert.Equal(t, 20.0, result.Quota)
	})

	t.Run("should return error when WiFi product not found", func(t *testing.T) {
		// When
		result, err := repo.GetByID(ctx, 999)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestWiFiProductRepository_GetByProductID(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewWiFiProductRepository(db, logger)
	ctx := context.Background()

	// Given - create test product and wifi product
	product := &entity.Product{
		Name:     "Test WiFi 2",
		Price:    50000,
		SKU:      "PRODUCT-wifi-20251214-WF003",
		Category: entity.ProductCategoryWiFi,
		Stock:    100,
	}
	err = db.Create(product).Error
	require.NoError(t, err)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      25,
		Duration:   60,
		SpeedLimit: 75,
	}
	err = repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	t.Run("should get WiFi product by existing product ID", func(t *testing.T) {
		// When
		result, err := repo.GetByProductID(ctx, product.ID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, product.ID, result.ProductID)
		assert.Equal(t, 25.0, result.Quota)
	})

	t.Run("should return error when product ID not found", func(t *testing.T) {
		// When
		result, err := repo.GetByProductID(ctx, 999)

		// Then
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestWiFiProductRepository_GetAll(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewWiFiProductRepository(db, logger)
	ctx := context.Background()

	// Given - create test products and wifi products
	product1 := &entity.Product{
		Name:     "WiFi Product 1",
		Price:    50000,
		SKU:      "PRODUCT-wifi-20251214-WF004",
		Category: entity.ProductCategoryWiFi,
		Stock:    100,
	}
	err = db.Create(product1).Error
	require.NoError(t, err)

	product2 := &entity.Product{
		Name:     "WiFi Product 2",
		Price:    75000,
		SKU:      "PRODUCT-wifi-20251214-WF005",
		Category: entity.ProductCategoryWiFi,
		Stock:    200,
	}
	err = db.Create(product2).Error
	require.NoError(t, err)

	wifiProds := []entity.WiFiProduct{
		{ProductID: product1.ID, Quota: 10, Duration: 30, SpeedLimit: 10},
		{ProductID: product2.ID, Quota: 20, Duration: 60, SpeedLimit: 50},
	}

	for i := range wifiProds {
		err = repo.Create(ctx, &wifiProds[i])
		require.NoError(t, err)
	}

	t.Run("should get all WiFi products with default pagination", func(t *testing.T) {
		// Given
		filter := &dto.WiFiProductFilter{Page: 1, Limit: 10}

		// When
		results, total, err := repo.GetAll(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.GreaterOrEqual(t, len(results), 2)
	})

	t.Run("should get WiFi products with limit", func(t *testing.T) {
		// Given
		filter := &dto.WiFiProductFilter{Page: 1, Limit: 1}

		// When
		results, total, err := repo.GetAll(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, 1, len(results))
	})

	t.Run("should get WiFi products page 2", func(t *testing.T) {
		// Given
		filter := &dto.WiFiProductFilter{Page: 2, Limit: 1}

		// When
		results, total, err := repo.GetAll(ctx, filter)

		// Then
		assert.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, 1, len(results))
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestWiFiProductRepository_Update(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewWiFiProductRepository(db, logger)
	ctx := context.Background()

	// Given - create test product and wifi product
	product := &entity.Product{
		Name:     "WiFi Update Test",
		Price:    50000,
		SKU:      "PRODUCT-wifi-20251214-WF006",
		Category: entity.ProductCategoryWiFi,
		Stock:    100,
	}
	err = db.Create(product).Error
	require.NoError(t, err)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err = repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	t.Run("should update WiFi product successfully", func(t *testing.T) {
		// Given
		wifiProd.Quota = 50
		wifiProd.Duration = 90
		wifiProd.SpeedLimit = 100

		// When
		err = repo.Update(ctx, wifiProd)

		// Then
		assert.NoError(t, err)

		// Verify the update
		saved, err := repo.GetByID(ctx, wifiProd.ID)
		assert.NoError(t, err)
		assert.Equal(t, 50.0, saved.Quota)
		assert.Equal(t, 90, saved.Duration)
		assert.Equal(t, 100.0, saved.SpeedLimit)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestWiFiProductRepository_Delete(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewWiFiProductRepository(db, logger)
	ctx := context.Background()

	// Given - create test product and wifi product
	product := &entity.Product{
		Name:     "WiFi Delete Test",
		Price:    50000,
		SKU:      "PRODUCT-wifi-20251214-WF007",
		Category: entity.ProductCategoryWiFi,
		Stock:    100,
	}
	err = db.Create(product).Error
	require.NoError(t, err)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      15,
		Duration:   30,
		SpeedLimit: 25,
	}
	err = repo.Create(ctx, wifiProd)
	require.NoError(t, err)

	t.Run("should delete existing WiFi product", func(t *testing.T) {
		// When
		err = repo.Delete(ctx, wifiProd.ID)

		// Then
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, wifiProd.ID)
		assert.Error(t, err)
	})

	// Cleanup
	testutil.CleanDB(db)
}

func TestWiFiProductRepository_OneToOneConstraint(t *testing.T) {
	db, err := testutil.SetupTestDB()
	require.NoError(t, err)
	logger := testutil.NewTestLogger(t)
	repo := NewWiFiProductRepository(db, logger)
	ctx := context.Background()

	// Given
	product := &entity.Product{
		Name:     "Constraint Test",
		Price:    50000,
		SKU:      "PRODUCT-wifi-20251214-WF008",
		Category: entity.ProductCategoryWiFi,
		Stock:    100,
	}
	err = db.Create(product).Error
	require.NoError(t, err)

	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}

	// When
	err = repo.Create(ctx, wifiProd)

	// Then
	assert.NoError(t, err)

	// Verify we can retrieve it
	result, err := repo.GetByProductID(ctx, product.ID)
	assert.NoError(t, err)
	assert.Equal(t, wifiProd.ID, result.ID)
	assert.Equal(t, product.ID, result.ProductID)

	// Cleanup
	testutil.CleanDB(db)
}
