package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet(req *dto.CreateWalletRequest) (*dto.WalletResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWalletByID(id uint) (*dto.WalletResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWalletByUserID(userID uint) (*dto.WalletResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWallets(filter *dto.WalletFilter) (*dto.WalletListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletListResponse), args.Error(1)
}

func (m *MockWalletService) UpdateWalletBalance(walletID uint, req *dto.UpdateWalletBalanceRequest) (*dto.WalletBalanceResponse, error) {
	args := m.Called(walletID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletBalanceResponse), args.Error(1)
}

func (m *MockWalletService) DeleteWallet(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockWalletService) TransferFunds(req *dto.TransferRequest) (*dto.TransferResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransferResponse), args.Error(1)
}

func (m *MockWalletService) CreateTransaction(req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransactionResponse), args.Error(1)
}

func (m *MockWalletService) GetTransactions(filter *dto.TransactionFilter) (*dto.TransactionListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransactionListResponse), args.Error(1)
}

func (m *MockWalletService) GetTransactionByID(id uint) (*dto.TransactionResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransactionResponse), args.Error(1)
}

func TestWalletHandler_CreateWallet(t *testing.T) {
	t.Run("should create wallet successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		req := &dto.CreateWalletRequest{
			UserID:         1,
			Currency:       "IDR",
			InitialBalance: 100000,
		}
		resp := &dto.WalletResponse{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("CreateWallet", mock.MatchedBy(func(r *dto.CreateWalletRequest) bool {
			return r.UserID == req.UserID
		})).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.CreateWallet(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestWalletHandler_GetWallet(t *testing.T) {
	t.Run("should get wallet by id successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		resp := &dto.WalletResponse{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetWalletByID", uint(1)).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallets/1", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetWallet(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestWalletHandler_ListWallets(t *testing.T) {
	t.Run("should list wallets with pagination", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		respList := &dto.WalletListResponse{
			Data: []dto.WalletResponse{
				{
					ID:        1,
					UserID:    1,
					Balance:   100000,
					Currency:  "IDR",
					Status:    "active",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			TotalCount: 1,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetWallets", mock.AnythingOfType("*dto.WalletFilter")).Return(respList, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/wallets", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.ListWallets(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestWalletHandler_DeleteWallet(t *testing.T) {
	t.Run("should delete wallet successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		mockService.On("DeleteWallet", uint(1)).Return(nil)

		httpReq := httptest.NewRequest(http.MethodDelete, "/api/v1/wallets/1", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.DeleteWallet(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestWalletHandler_TransferFunds(t *testing.T) {
	t.Run("should transfer funds successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   2,
			Amount:       250000,
			Description:  "Test transfer",
		}
		resp := &dto.TransferResponse{
			TransferID:      1,
			FromWalletID:    1,
			ToWalletID:      2,
			Amount:          250000,
			Description:     "Test transfer",
			FromPrevBalance: 1000000,
			FromNewBalance:  750000,
			ToPrevBalance:   500000,
			ToNewBalance:    750000,
			Status:          "success",
			TransferredAt:   time.Now(),
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(r *dto.TransferRequest) bool {
			return r.FromWalletID == req.FromWalletID && r.ToWalletID == req.ToWalletID
		})).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.TransferFunds(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.TransferResponse
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, req.FromWalletID, result.FromWalletID)
		assert.Equal(t, req.ToWalletID, result.ToWalletID)
		assert.Equal(t, req.Amount, result.Amount)
		assert.Equal(t, "success", result.Status)
	})

	t.Run("should return error if invalid request body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.TransferFunds(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error if transfer to same wallet", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   1, // Same wallet
			Amount:       250000,
			Description:  "Invalid transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(r *dto.TransferRequest) bool {
			return r.FromWalletID == r.ToWalletID
		})).Return(nil, assert.AnError)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.TransferFunds(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error if insufficient balance", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   2,
			Amount:       1000000,
			Description:  "Test transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(r *dto.TransferRequest) bool {
			return r.FromWalletID == req.FromWalletID && r.ToWalletID == req.ToWalletID
		})).Return(nil, assert.AnError)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.TransferFunds(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error if source wallet not found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 999,
			ToWalletID:   2,
			Amount:       250000,
			Description:  "Test transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(r *dto.TransferRequest) bool {
			return r.FromWalletID == req.FromWalletID
		})).Return(nil, assert.AnError)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.TransferFunds(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error if destination wallet not found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   999,
			Amount:       250000,
			Description:  "Test transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(r *dto.TransferRequest) bool {
			return r.ToWalletID == req.ToWalletID
		})).Return(nil, assert.AnError)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.TransferFunds(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
