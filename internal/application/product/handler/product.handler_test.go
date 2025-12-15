package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupProductHandlerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Product{}, &entity.WiFiProduct{})
	require.NoError(t, err)

	return db
}

func setupProductHandlerForTest(t *testing.T) (*ProductHandler, *gorm.DB) {
	db := setupProductHandlerTestDB(t)
	logger := zap.NewNop()

	repo := repository.NewProductRepository(db, logger)
	wifiRepo := repository.NewWiFiProductRepository(db, logger)

	txManager := database.NewTransactionManager(db)
	svc := service.NewProductService(txManager, repo, wifiRepo, logger)
	wifiSvc := service.NewWiFiProductService(wifiRepo, logger)

	handler := NewProductHandler(svc, wifiSvc, logger)
	return handler, db
}

func TestProductHandler_CreateProduct(t *testing.T) {
	handler, _ := setupProductHandlerForTest(t)

	tests := []struct {
		name           string
		req            interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "create valid wifi product",
			req: dto.CreateProductRequest{
				Name:     "WiFi 10GB",
				Price:    50000,
				Category: "wifi",
				Stock:    100,
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name: "create valid pulsa product",
			req: dto.CreateProductRequest{
				Name:     "Pulsa 50k",
				Price:    50000,
				Category: "pulsa",
				Stock:    50,
			},
			expectedStatus: http.StatusCreated,
			expectedError:  false,
		},
		{
			name:           "create with invalid JSON",
			req:            "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "create product with empty name",
			req: dto.CreateProductRequest{
				Price:    5000,
				Category: "wifi",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "create product without category",
			req: dto.CreateProductRequest{
				Name:  "No Category",
				Price: 5000,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "create product with invalid category",
			req: dto.CreateProductRequest{
				Name:     "Invalid Cat",
				Price:    5000,
				Category: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.req)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/products", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.CreateProduct(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError {
				var errResp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errResp)
				assert.NoError(t, err)
				assert.Contains(t, errResp, "error")
			}
		})
	}
}

func TestProductHandler_GetProduct(t *testing.T) {
	handler, db := setupProductHandlerForTest(t)

	// Create test product
	repo := repository.NewProductRepository(db, zap.NewNop())
	product := &entity.Product{
		Name:  "Test Product",
		Price: 5000,
		SKU:   "GET-TEST-001",
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "get existing product",
			productID:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent product",
			productID:      "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "get with invalid product ID",
			productID:      "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tt.productID})

			handler.GetProduct(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_GetProductBySKU(t *testing.T) {
	handler, db := setupProductHandlerForTest(t)

	// Create test product
	repo := repository.NewProductRepository(db, zap.NewNop())
	product := &entity.Product{
		Name:  "Test Product",
		Price: 5000,
		SKU:   "UNIQUE-SKU-001",
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	tests := []struct {
		name           string
		sku            string
		expectedStatus int
	}{
		{
			name:           "get product by existing SKU",
			sku:            "UNIQUE-SKU-001",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get product by non-existent SKU",
			sku:            "NON-EXISTENT",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "get with empty SKU",
			sku:            "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/products/sku/"+tt.sku, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "sku", Value: tt.sku})

			handler.GetProductBySKU(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_ListProducts(t *testing.T) {
	handler, db := setupProductHandlerForTest(t)

	// Create test products
	repo := repository.NewProductRepository(db, zap.NewNop())
	products := []entity.Product{
		{Name: "Laptop", Price: 10000, SKU: "LAP-001", Stock: 5},
		{Name: "Phone", Price: 5000, SKU: "PHN-001", Stock: 10},
	}
	for i := range products {
		err := repo.Create(context.Background(), &products[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		page           string
		limit          string
		expectedStatus int
	}{
		{
			name:           "list products with valid pagination",
			page:           "1",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list products with invalid page (uses default)",
			page:           "invalid",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/products?page=" + tt.page + "&limit=" + tt.limit
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ListProducts(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_UpdateProduct(t *testing.T) {
	handler, db := setupProductHandlerForTest(t)

	// Create test product
	repo := repository.NewProductRepository(db, zap.NewNop())
	product := &entity.Product{
		Name:     "Original Name",
		Price:    5000,
		SKU:      "UPDATE-001",
		Category: entity.ProductCategoryWiFi,
		Status:   entity.ProductStatusInactive,
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	tests := []struct {
		name           string
		productID      string
		req            interface{}
		expectedStatus int
	}{
		{
			name:      "update existing product",
			productID: "1",
			req: dto.UpdateProductRequest{
				Name:  "Updated Name",
				Price: 6000,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "update product status to inactive",
			productID: "1",
			req: dto.UpdateProductRequest{
				Status: "inactive",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "update non-existent product",
			productID: "999",
			req: dto.UpdateProductRequest{
				Name: "Test",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.req)
			require.NoError(t, err)

			req := httptest.NewRequest("PUT", "/api/v1/products/"+tt.productID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tt.productID})

			handler.UpdateProduct(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_DeleteProduct(t *testing.T) {
	handler, db := setupProductHandlerForTest(t)

	// Create test product
	repo := repository.NewProductRepository(db, zap.NewNop())
	product := &entity.Product{
		Name:  "Delete Test",
		Price: 5000,
		SKU:   "DEL-001",
	}
	err := repo.Create(context.Background(), product)
	require.NoError(t, err)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "delete existing product",
			productID:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "delete non-existent product",
			productID:      "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/v1/products/"+tt.productID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tt.productID})

			handler.DeleteProduct(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestProductHandler_UpdateProductStatus(t *testing.T) {
	handler, db := setupProductHandlerForTest(t)

	// Create a test product
	product := &entity.Product{
		Name:        "WiFi 10GB",
		Description: "WiFi package",
		Price:       50000,
		SKU:         "WIFI-001",
		Category:    entity.ProductCategoryWiFi,
		Status:      entity.ProductStatusInactive,
		Stock:       100,
	}
	err := db.Create(product).Error
	require.NoError(t, err)

	tests := []struct {
		name           string
		productID      string
		requestBody    dto.UpdateProductStatusRequest
		expectedStatus int
		expectError    bool
	}{
		{
			name:      "update product status to active successfully",
			productID: "1",
			requestBody: dto.UpdateProductStatusRequest{
				Status: "active",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:      "update product status to inactive successfully",
			productID: "1",
			requestBody: dto.UpdateProductStatusRequest{
				Status: "inactive",
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "update status for non-existent product",
			productID:      "999",
			requestBody:    dto.UpdateProductStatusRequest{Status: "active"},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/api/v1/products/"+tt.productID+"/status", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tt.productID})

			handler.UpdateProductStatus(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if !tt.expectError {
				var response dto.ProductResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody.Status, response.Status)
			}
		})
	}
}
