package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWalletHandler_CreateTransaction(t *testing.T) {
	t.Run("should create transaction successfully", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := dto.CreateTransactionRequest{
			WalletID:    1,
			Type:        "deposit",
			Amount:      100.0,
			Description: "Test deposit",
		}

		resp := &dto.TransactionResponse{
			ID:           1,
			WalletID:     1,
			Type:         "deposit",
			Amount:       100.0,
			Status:       "completed",
			BalanceAfter: 200.0,
			CreatedAt:    time.Now(),
		}

		mockService.On("CreateTransaction", &req).Return(resp, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(req)
		ctx.Request = httptest.NewRequest("POST", "/api/v1/wallets/transactions", strings.NewReader(string(body)))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateTransaction(ctx)

		// Then
		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/api/v1/wallets/transactions", strings.NewReader("invalid json"))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateTransaction(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return not found when wallet not found", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := dto.CreateTransactionRequest{
			WalletID:    999,
			Type:        "deposit",
			Amount:      100.0,
			Description: "Test deposit",
		}

		mockService.On("CreateTransaction", &req).Return(nil, errors.New("wallet not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		body, _ := json.Marshal(req)
		ctx.Request = httptest.NewRequest("POST", "/api/v1/wallets/transactions", strings.NewReader(string(body)))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateTransaction(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestWalletHandler_GetTransactions(t *testing.T) {
	t.Run("should get transactions successfully", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		resp := &dto.TransactionListResponse{
			Data: []dto.TransactionResponse{
				{
					ID:           1,
					WalletID:     1,
					Type:         "deposit",
					Amount:       100.0,
					Status:       "completed",
					BalanceAfter: 200.0,
				},
			},
			TotalCount: 1,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetTransactions", mock.MatchedBy(func(f *dto.TransactionFilter) bool {
			return f.WalletID == 1
		})).Return(resp, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/api/v1/wallets/transactions?wallet_id=1", nil)

		// When
		handler.GetTransactions(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request when wallet_id is missing", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/api/v1/wallets/transactions", nil)

		// When
		handler.GetTransactions(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should apply default pagination", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		filter := &dto.TransactionFilter{
			WalletID: 1,
			Page:     1,
			PageSize: 10,
		}

		resp := &dto.TransactionListResponse{
			Data:       []dto.TransactionResponse{},
			TotalCount: 0,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetTransactions", filter).Return(resp, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/api/v1/wallets/transactions?wallet_id=1", nil)

		// When
		handler.GetTransactions(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestWalletHandler_GetTransaction(t *testing.T) {
	t.Run("should get transaction by ID successfully", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		resp := &dto.TransactionResponse{
			ID:           1,
			WalletID:     1,
			Type:         "deposit",
			Amount:       100.0,
			Status:       "completed",
			BalanceAfter: 200.0,
			CreatedAt:    time.Now(),
		}

		mockService.On("GetTransactionByID", uint(1)).Return(resp, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/api/v1/wallets/transactions/1", nil)
		ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: "1"})

		// When
		handler.GetTransaction(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when transaction not found", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		mockService.On("GetTransactionByID", uint(999)).Return(nil, errors.New("transaction not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/api/v1/wallets/transactions/999", nil)
		ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: "999"})

		// When
		handler.GetTransaction(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("should return bad request for invalid transaction ID", func(t *testing.T) {
		// Setup
		mockService := &testutil.MockWalletService{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/api/v1/wallets/transactions/invalid", nil)
		ctx.Params = append(ctx.Params, gin.Param{Key: "id", Value: "invalid"})

		// When
		handler.GetTransaction(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
