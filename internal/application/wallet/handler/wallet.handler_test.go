package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	userDto "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet(ctx context.Context, req *dto.CreateWalletRequest) (*dto.WalletResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWalletByID(ctx context.Context, id uint) (*dto.WalletResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWalletByUserID(ctx context.Context, userID uint) (*dto.WalletResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWallets(ctx context.Context, filter *dto.WalletFilter) (*dto.WalletListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletListResponse), args.Error(1)
}

func (m *MockWalletService) UpdateWalletBalance(ctx context.Context, walletID uint, req *dto.UpdateWalletBalanceRequest) (*dto.WalletBalanceResponse, error) {
	args := m.Called(ctx, walletID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.WalletBalanceResponse), args.Error(1)
}

func (m *MockWalletService) DeleteWallet(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWalletService) TransferFunds(ctx context.Context, req *dto.TransferRequest) (*dto.TransferResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransferResponse), args.Error(1)
}

func (m *MockWalletService) Transfer(ctx context.Context, senderID uint, req *dto.TransferWalletRequest) (*dto.TransferResponse, error) {
	args := m.Called(ctx, senderID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransferResponse), args.Error(1)
}

func (m *MockWalletService) CreateTransaction(ctx context.Context, req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransactionResponse), args.Error(1)
}

func (m *MockWalletService) GetTransactions(ctx context.Context, filter *dto.TransactionFilter) (*dto.TransactionListResponse, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransactionListResponse), args.Error(1)
}

func (m *MockWalletService) GetTransactionByID(ctx context.Context, id uint) (*dto.TransactionResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.TransactionResponse), args.Error(1)
}

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) Register(req *userDto.RegisterRequest) (*userDto.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) Login(req *userDto.LoginRequest) (*userDto.LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetCurrentUser(token string) (*userDto.UserResponse, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uint) (*userDto.UserResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUsers(filter *userDto.UserFilter) (*userDto.UserListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserListResponse), args.Error(1)
}

func (m *MockUserService) CreateUser(req *userDto.CreateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(id uint, req *userDto.UpdateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockUserService) GetUserByEmail(email string) (*userDto.UserResponse, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUserPassword(id uint, req *userDto.UpdateUserPasswordRequest) error {
	args := m.Called(id, req)
	return args.Error(0)
}

func TestWalletHandler_CreateWallet(t *testing.T) {
	t.Run("should create wallet successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

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

		mockService.On("CreateWallet", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), mock.MatchedBy(func(r *dto.CreateWalletRequest) bool {
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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

		resp := &dto.WalletResponse{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetWalletByID", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(1)).Return(resp, nil)

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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

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

		mockService.On("GetWallets", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), mock.AnythingOfType("*dto.WalletFilter")).Return(respList, nil)

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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

		mockService.On("DeleteWallet", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(1)).Return(nil)

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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

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

		mockService.On("TransferFunds", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), mock.MatchedBy(func(r *dto.TransferRequest) bool {
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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   1, // Same wallet
			Amount:       250000,
			Description:  "Invalid transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), mock.MatchedBy(func(r *dto.TransferRequest) bool {
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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   2,
			Amount:       1000000,
			Description:  "Test transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), mock.MatchedBy(func(r *dto.TransferRequest) bool {
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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 999,
			ToWalletID:   2,
			Amount:       250000,
			Description:  "Test transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), mock.MatchedBy(func(r *dto.TransferRequest) bool {
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
		mockUserService := &MockUserService{}
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   999,
			Amount:       250000,
			Description:  "Test transfer",
		}

		mockService.On("TransferFunds", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), mock.MatchedBy(func(r *dto.TransferRequest) bool {
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

func TestWalletHandler_Transfer(t *testing.T) {
	t.Run("should transfer funds successfully with valid auth", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		token := "valid_token"
		userResp := &userDto.UserResponse{
			ID:    1,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		req := &dto.TransferWalletRequest{
			RecipientUserID: 2,
			Amount:          50000,
			Description:     "Test transfer",
		}

		transferResp := &dto.TransferResponse{
			TransferID:      1,
			FromWalletID:    1,
			ToWalletID:      2,
			Amount:          50000,
			Description:     "Test transfer",
			FromPrevBalance: 100000,
			FromNewBalance:  50000,
			ToPrevBalance:   30000,
			ToNewBalance:    80000,
			Status:          "success",
			TransferredAt:   time.Now(),
		}

		mockUserService.On("GetCurrentUser", token).Return(userResp, nil).Once()
		mockService.On("Transfer", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), userResp.ID, mock.MatchedBy(func(r *dto.TransferWalletRequest) bool {
			return r.RecipientUserID == req.RecipientUserID && r.Amount == req.Amount
		})).Return(transferResp, nil).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockUserService.AssertExpectations(t)
		mockService.AssertExpectations(t)
	})

	t.Run("should reject transfer when user level is regular user", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		token := "valid_token"
		userResp := &userDto.UserResponse{
			ID:    5,
			Name:  "Regular User",
			Email: "user@example.com",
			Level: "user",
		}

		req := &dto.TransferWalletRequest{
			RecipientUserID: 6,
			Amount:          50000,
			Description:     "Should fail",
		}

		mockUserService.On("GetCurrentUser", token).Return(userResp, nil).Once()
		mockService.On("Transfer", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), userResp.ID, mock.MatchedBy(func(r *dto.TransferWalletRequest) bool {
			return r.RecipientUserID == req.RecipientUserID
		})).Return(nil, assert.AnError).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return forbidden when user tries to transfer", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		token := "valid_token"
		userResp := &userDto.UserResponse{
			ID:    5,
			Name:  "Regular User",
			Email: "user@example.com",
			Level: "user",
		}

		req := &dto.TransferWalletRequest{
			RecipientUserID: 6,
			Amount:          50000,
			Description:     "Should fail",
		}

		mockUserService.On("GetCurrentUser", token).Return(userResp, nil).Once()
		mockService.On("Transfer", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), userResp.ID, mock.MatchedBy(func(r *dto.TransferWalletRequest) bool {
			return r.RecipientUserID == req.RecipientUserID
		})).Return(nil, fmt.Errorf("users cannot transfer funds")).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return unauthorized when no auth header", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.TransferWalletRequest{
			RecipientUserID: 2,
			Amount:          50000,
			Description:     "Test transfer",
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		// No Authorization header

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return unauthorized on invalid token", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		token := "invalid_token"

		req := &dto.TransferWalletRequest{
			RecipientUserID: 2,
			Amount:          50000,
			Description:     "Test transfer",
		}

		mockUserService.On("GetCurrentUser", token).Return(nil, fmt.Errorf("invalid or expired token")).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return bad request on invalid request body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		token := "valid_token"
		userResp := &userDto.UserResponse{
			ID:    1,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		mockUserService.On("GetCurrentUser", token).Return(userResp, nil).Once()

		// Invalid JSON body
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return 404 when recipient user not found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		token := "valid_token"
		userResp := &userDto.UserResponse{
			ID:    1,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		req := &dto.TransferWalletRequest{
			RecipientUserID: 999,
			Amount:          50000,
			Description:     "Not found",
		}

		mockUserService.On("GetCurrentUser", token).Return(userResp, nil).Once()
		mockService.On("Transfer", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), userResp.ID, mock.MatchedBy(func(r *dto.TransferWalletRequest) bool {
			return r.RecipientUserID == req.RecipientUserID
		})).Return(nil, fmt.Errorf("recipient user not found")).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return 400 when insufficient balance", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		token := "valid_token"
		userResp := &userDto.UserResponse{
			ID:    1,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		req := &dto.TransferWalletRequest{
			RecipientUserID: 2,
			Amount:          999999,
			Description:     "Insufficient balance",
		}

		mockUserService.On("GetCurrentUser", token).Return(userResp, nil).Once()
		mockService.On("Transfer", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), userResp.ID, mock.MatchedBy(func(r *dto.TransferWalletRequest) bool {
			return r.RecipientUserID == req.RecipientUserID
		})).Return(nil, fmt.Errorf("insufficient balance for transfer")).Once()

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPost, "/api/v1/wallets/transfer-to-user", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq

		handler.Transfer(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockUserService.AssertExpectations(t)
	})
}

func TestWalletHandler_GetUserWallet(t *testing.T) {
	t.Run("should get user wallet successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		resp := &dto.WalletResponse{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetWalletByUserID", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(1)).Return(resp, nil)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/1/wallet", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetUserWallet(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var result dto.WalletResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, uint(1), result.UserID)
		mockService.AssertExpectations(t)
	})

	t.Run("should return 400 for invalid user ID", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/invalid/wallet", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.GetUserWallet(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 404 when wallet not found for user", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		mockService.On("GetWalletByUserID", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(999)).Return(nil, assert.AnError)

		httpReq := httptest.NewRequest(http.MethodGet, "/api/v1/users/999/wallet", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.GetUserWallet(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestWalletHandler_UpdateWalletBalance(t *testing.T) {
	t.Run("should update wallet balance with credit successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.UpdateWalletBalanceRequest{
			Amount:          50000,
			TransactionType: "credit",
			Description:     "Top up",
		}

		resp := &dto.WalletBalanceResponse{
			WalletID:        1,
			PreviousBalance: 100000,
			NewBalance:      150000,
			Amount:          50000,
			TransactionType: "credit",
		}

		mockService.On("UpdateWalletBalance", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(1), mock.MatchedBy(func(r *dto.UpdateWalletBalanceRequest) bool {
			return r.Amount == req.Amount && r.TransactionType == req.TransactionType
		})).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/wallets/1/balance", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateWalletBalance(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var result dto.WalletBalanceResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, float64(150000), result.NewBalance)
		mockService.AssertExpectations(t)
	})

	t.Run("should update wallet balance with debit successfully", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.UpdateWalletBalanceRequest{
			Amount:          30000,
			TransactionType: "debit",
			Description:     "Withdrawal",
		}

		resp := &dto.WalletBalanceResponse{
			WalletID:        1,
			PreviousBalance: 100000,
			NewBalance:      70000,
			Amount:          30000,
			TransactionType: "debit",
		}

		mockService.On("UpdateWalletBalance", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(1), mock.MatchedBy(func(r *dto.UpdateWalletBalanceRequest) bool {
			return r.Amount == req.Amount && r.TransactionType == req.TransactionType
		})).Return(resp, nil)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/wallets/1/balance", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateWalletBalance(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var result dto.WalletBalanceResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Equal(t, float64(70000), result.NewBalance)
		mockService.AssertExpectations(t)
	})

	t.Run("should return 400 for invalid wallet ID", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.UpdateWalletBalanceRequest{
			Amount:          50000,
			TransactionType: "credit",
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/wallets/invalid/balance", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.UpdateWalletBalance(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 for invalid request body", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/wallets/1/balance", bytes.NewReader([]byte("invalid json")))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateWalletBalance(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return 400 when insufficient balance for debit", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.UpdateWalletBalanceRequest{
			Amount:          999999,
			TransactionType: "debit",
			Description:     "Insufficient balance",
		}

		mockService.On("UpdateWalletBalance", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(1), mock.MatchedBy(func(r *dto.UpdateWalletBalanceRequest) bool {
			return r.Amount == req.Amount
		})).Return(nil, assert.AnError)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/wallets/1/balance", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateWalletBalance(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return 404 when wallet not found", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		mockService := &MockWalletService{}
		mockUserService := &MockUserService{}
		logger := testutil.NewSilentLogger()
		handler := NewWalletHandler(mockService, mockUserService, logger)

		req := &dto.UpdateWalletBalanceRequest{
			Amount:          50000,
			TransactionType: "credit",
		}

		mockService.On("UpdateWalletBalance", mock.MatchedBy(func(ctx context.Context) bool {
			return ctx != nil
		}), uint(999), mock.AnythingOfType("*dto.UpdateWalletBalanceRequest")).
			Return(nil, assert.AnError)

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest(http.MethodPut, "/api/v1/wallets/999/balance", bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httpReq
		c.Params = gin.Params{{Key: "id", Value: "999"}}

		handler.UpdateWalletBalance(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}
