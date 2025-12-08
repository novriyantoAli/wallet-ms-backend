package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupWiFiProductHandlerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.Product{}, &entity.WiFiProduct{})
	require.NoError(t, err)

	return db
}

func setupWiFiProductHandlerForTest(t *testing.T) (*WiFiProductHandler, *gorm.DB) {
	db := setupWiFiProductHandlerTestDB(t)
	logger := zap.NewNop()

	wifiRepo := repository.NewWiFiProductRepository(db)
	wifiSvc := service.NewWiFiProductService(wifiRepo, logger)
	handler := NewWiFiProductHandler(wifiSvc, logger)

	return handler, db
}

func createTestBaseProduct(t *testing.T, db *gorm.DB) *entity.Product {
	product := &entity.Product{
		Name:  "WiFi Base Product",
		Price: 50000,
		SKU:   "WIFI-BASE-001",
		Stock: 100,
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

func TestWiFiProductHandler_CreateWiFiProduct(t *testing.T) {
	handler, db := setupWiFiProductHandlerForTest(t)

	product := createTestBaseProduct(t, db)

	tests := []struct {
		name           string
		req            interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "create valid wifi product",
			req: dto.CreateWiFiProductRequest{
				ProductID:  product.ID,
				Quota:      10,
				Duration:   30,
				SpeedLimit: 10,
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.req)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/v1/wifi-products", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.CreateWiFiProduct(c)

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

func TestWiFiProductHandler_GetWiFiProduct(t *testing.T) {
	handler, db := setupWiFiProductHandlerForTest(t)

	product := createTestBaseProduct(t, db)
	wifiRepo := repository.NewWiFiProductRepository(db)
	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err := wifiRepo.Create(wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name           string
		wifiProductID  string
		expectedStatus int
	}{
		{
			name:           "get existing wifi product",
			wifiProductID:  "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get non-existent wifi product",
			wifiProductID:  "999",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "get with invalid wifi product ID",
			wifiProductID:  "invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/wifi-products/"+tt.wifiProductID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tt.wifiProductID})

			handler.GetWiFiProduct(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestWiFiProductHandler_GetWiFiProductByProductID(t *testing.T) {
	handler, db := setupWiFiProductHandlerForTest(t)

	product := createTestBaseProduct(t, db)
	wifiRepo := repository.NewWiFiProductRepository(db)
	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err := wifiRepo.Create(wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name           string
		productID      string
		expectedStatus int
	}{
		{
			name:           "get wifi product by existing product id",
			productID:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "get wifi product by non-existent product id",
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
			req := httptest.NewRequest("GET", "/api/v1/products/wifi/product/"+tt.productID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "product_id", Value: tt.productID})

			handler.GetWiFiProductByProductID(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestWiFiProductHandler_ListWiFiProducts(t *testing.T) {
	handler, db := setupWiFiProductHandlerForTest(t)

	product1 := createTestBaseProduct(t, db)
	product2 := &entity.Product{
		Name:  "WiFi Product 2",
		Price: 40000,
		SKU:   "WIFI-BASE-002",
		Stock: 50,
	}
	err := db.Create(product2).Error
	require.NoError(t, err)

	wifiRepo := repository.NewWiFiProductRepository(db)
	wifiProds := []entity.WiFiProduct{
		{ProductID: product1.ID, Quota: 10, Duration: 30, SpeedLimit: 10},
		{ProductID: product2.ID, Quota: 20, Duration: 60, SpeedLimit: 50},
	}
	for i := range wifiProds {
		err := wifiRepo.Create(&wifiProds[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name           string
		page           string
		limit          string
		expectedStatus int
	}{
		{
			name:           "list wifi products with valid pagination",
			page:           "1",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list wifi products with invalid page (uses default)",
			page:           "invalid",
			limit:          "10",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/api/v1/wifi-products?page=" + tt.page + "&limit=" + tt.limit
			req := httptest.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			handler.ListWiFiProducts(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestWiFiProductHandler_UpdateWiFiProduct(t *testing.T) {
	handler, db := setupWiFiProductHandlerForTest(t)

	product := createTestBaseProduct(t, db)
	wifiRepo := repository.NewWiFiProductRepository(db)
	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err := wifiRepo.Create(wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name           string
		wifiProductID  string
		req            interface{}
		expectedStatus int
	}{
		{
			name:          "update existing wifi product",
			wifiProductID: "1",
			req: dto.UpdateWiFiProductRequest{
				Quota:      20,
				Duration:   60,
				SpeedLimit: 50,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "update non-existent wifi product",
			wifiProductID: "999",
			req: dto.UpdateWiFiProductRequest{
				Quota: 10,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.req)
			require.NoError(t, err)

			req := httptest.NewRequest("PUT", "/api/v1/wifi-products/"+tt.wifiProductID, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tt.wifiProductID})

			handler.UpdateWiFiProduct(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestWiFiProductHandler_DeleteWiFiProduct(t *testing.T) {
	handler, db := setupWiFiProductHandlerForTest(t)

	product := createTestBaseProduct(t, db)
	wifiRepo := repository.NewWiFiProductRepository(db)
	wifiProd := &entity.WiFiProduct{
		ProductID:  product.ID,
		Quota:      10,
		Duration:   30,
		SpeedLimit: 10,
	}
	err := wifiRepo.Create(wifiProd)
	require.NoError(t, err)

	tests := []struct {
		name           string
		wifiProductID  string
		expectedStatus int
	}{
		{
			name:           "delete existing wifi product",
			wifiProductID:  "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "delete non-existent wifi product",
			wifiProductID:  "999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/api/v1/wifi-products/"+tt.wifiProductID, nil)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = append(c.Params, gin.Param{Key: "id", Value: tt.wifiProductID})

			handler.DeleteWiFiProduct(c)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
