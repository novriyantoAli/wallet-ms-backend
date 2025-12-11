package service

import (
	"errors"
	"testing"
	"time"

	userDto "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestWalletService_CreateWallet(t *testing.T) {
	t.Run("should create wallet successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		req := &dto.CreateWalletRequest{
			UserID:         1,
			Currency:       "IDR",
			InitialBalance: 100000,
		}
		userResponse := &userDto.UserResponse{
			ID:    req.UserID,
			Name:  "John Doe",
			Email: "john@example.com",
		}

		walletWithUser := &repository.WalletWithUserData{
			ID:        1,
			UserID:    req.UserID,
			Balance:   req.InitialBalance,
			Currency:  req.Currency,
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserName:  "John Doe",
			UserEmail: "john@example.com",
			UserLevel: "user",
		}

		// Mock expectations
		mockUserService.On("GetUserByID", req.UserID).Return(userResponse, nil)
		mockRepo.On("GetByUserID", req.UserID).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(nil).Run(func(args mock.Arguments) {
			wallet := args.Get(0).(*entity.Wallet)
			wallet.ID = 1
		})
		mockRepo.On("GetByUserIDWithUser", req.UserID).Return(walletWithUser, nil)

		// Execute
		result, err := service.CreateWallet(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.ID)
		assert.Equal(t, req.UserID, result.UserID)
		assert.Equal(t, req.Currency, result.Currency)
		assert.Equal(t, req.InitialBalance, result.Balance)
		assert.Equal(t, "active", result.Status)
		mockRepo.AssertExpectations(t)
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return error if user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		req := &dto.CreateWalletRequest{
			UserID:         999,
			Currency:       "IDR",
			InitialBalance: 100000,
		}

		// Mock expectations
		mockUserService.On("GetUserByID", req.UserID).Return(nil, errors.New("user not found"))

		// Execute
		_, err := service.CreateWallet(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "user not found", err.Error())
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return error if user already has wallet", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		req := &dto.CreateWalletRequest{
			UserID:         1,
			Currency:       "IDR",
			InitialBalance: 100000,
		}
		userResponse := &userDto.UserResponse{
			ID:    req.UserID,
			Name:  "John Doe",
			Email: "john@example.com",
		}
		existingWallet := &entity.Wallet{
			ID:     1,
			UserID: req.UserID,
		}

		// Mock expectations
		mockUserService.On("GetUserByID", req.UserID).Return(userResponse, nil)
		mockRepo.On("GetByUserID", req.UserID).Return(existingWallet, nil)

		// Execute
		_, err := service.CreateWallet(req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "user already has a wallet", err.Error())
		mockUserService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_GetWalletByID(t *testing.T) {
	t.Run("should get wallet by id successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(wallet, nil)

		// Execute
		result, err := service.GetWalletByID(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, wallet.ID, result.ID)
		assert.Equal(t, wallet.UserID, result.UserID)
		assert.Equal(t, wallet.Balance, result.Balance)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if wallet not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		// Mock expectations
		mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		_, err := service.GetWalletByID(999)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_GetWalletByUserID(t *testing.T) {
	t.Run("should get wallet by user id successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock expectations
		mockRepo.On("GetByUserID", uint(1)).Return(wallet, nil)

		// Execute
		result, err := service.GetWalletByUserID(1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, wallet.ID, result.ID)
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_UpdateWalletBalance(t *testing.T) {
	t.Run("should credit balance successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.UpdateWalletBalanceRequest{
			Amount:          50000,
			TransactionType: "credit",
			Description:     "Top up",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(wallet, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.Wallet")).Return(nil)

		// Execute
		result, err := service.UpdateWalletBalance(1, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(100000), result.PreviousBalance)
		assert.Equal(t, float64(150000), result.NewBalance)
		assert.Equal(t, req.Amount, result.Amount)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should debit balance successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.UpdateWalletBalanceRequest{
			Amount:          30000,
			TransactionType: "debit",
			Description:     "Purchase",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(wallet, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.Wallet")).Return(nil)

		// Execute
		result, err := service.UpdateWalletBalance(1, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, float64(100000), result.PreviousBalance)
		assert.Equal(t, float64(70000), result.NewBalance)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if insufficient balance", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   50000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.UpdateWalletBalanceRequest{
			Amount:          100000,
			TransactionType: "debit",
			Description:     "Purchase",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(wallet, nil)

		// Execute
		_, err := service.UpdateWalletBalance(1, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "insufficient balance", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if invalid transaction type", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.UpdateWalletBalanceRequest{
			Amount:          50000,
			TransactionType: "invalid",
			Description:     "Test",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(wallet, nil)

		// Execute
		_, err := service.UpdateWalletBalance(1, req)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, "invalid transaction type", err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_GetWallets(t *testing.T) {
	t.Run("should get wallets with pagination", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallets := []entity.Wallet{
			{
				ID:        1,
				UserID:    1,
				Balance:   100000,
				Currency:  "IDR",
				Status:    entity.WalletStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			{
				ID:        2,
				UserID:    2,
				Balance:   200000,
				Currency:  "IDR",
				Status:    entity.WalletStatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		filter := &dto.WalletFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", filter).Return(wallets, int64(2), nil)

		// Execute
		result, err := service.GetWallets(filter)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.TotalCount)
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_DeleteWallet(t *testing.T) {
	t.Run("should delete wallet successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		wallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(wallet, nil)
		mockRepo.On("Delete", uint(1)).Return(nil)

		// Execute
		err := service.DeleteWallet(1)

		// Assert
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if wallet not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		// Mock expectations
		mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		err := service.DeleteWallet(999)

		// Assert
		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_TransferFunds(t *testing.T) {
	t.Run("should transfer funds successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		fromWallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   1000000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		toWallet := &entity.Wallet{
			ID:        2,
			UserID:    2,
			Balance:   500000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   2,
			Amount:       250000,
			Description:  "Test transfer",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", uint(2)).Return(toWallet, nil)
		mockRepo.On("Update", mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 1 && w.Balance == 750000
		})).Return(nil)
		mockRepo.On("Update", mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 2 && w.Balance == 750000
		})).Return(nil)

		// Execute
		result, err := service.TransferFunds(req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), result.FromWalletID)
		assert.Equal(t, uint(2), result.ToWalletID)
		assert.Equal(t, float64(250000), result.Amount)
		assert.Equal(t, float64(1000000), result.FromPrevBalance)
		assert.Equal(t, float64(750000), result.FromNewBalance)
		assert.Equal(t, float64(500000), result.ToPrevBalance)
		assert.Equal(t, float64(750000), result.ToNewBalance)
		assert.Equal(t, "success", result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if transferring to same wallet", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   1, // Same wallet
			Amount:       250000,
			Description:  "Invalid transfer",
		}

		// Execute
		result, err := service.TransferFunds(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "cannot transfer to the same wallet", err.Error())
	})

	t.Run("should return error if source wallet not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		req := &dto.TransferRequest{
			FromWalletID: 999,
			ToWalletID:   2,
			Amount:       250000,
			Description:  "Test transfer",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		result, err := service.TransferFunds(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "source wallet not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if destination wallet not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		fromWallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   1000000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   999,
			Amount:       250000,
			Description:  "Test transfer",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		result, err := service.TransferFunds(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "destination wallet not found", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if insufficient balance", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		fromWallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		toWallet := &entity.Wallet{
			ID:        2,
			UserID:    2,
			Balance:   500000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   2,
			Amount:       250000, // More than balance
			Description:  "Test transfer",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", uint(2)).Return(toWallet, nil)

		// Execute
		result, err := service.TransferFunds(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "insufficient balance for transfer", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should rollback source wallet on destination update failure", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		fromWallet := &entity.Wallet{
			ID:        1,
			UserID:    1,
			Balance:   1000000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		toWallet := &entity.Wallet{
			ID:        2,
			UserID:    2,
			Balance:   500000,
			Currency:  "IDR",
			Status:    entity.WalletStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		req := &dto.TransferRequest{
			FromWalletID: 1,
			ToWalletID:   2,
			Amount:       250000,
			Description:  "Test transfer",
		}

		// Mock expectations
		mockRepo.On("GetByID", uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", uint(2)).Return(toWallet, nil)
		mockRepo.On("Update", mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 1
		})).Return(nil).Once()
		mockRepo.On("Update", mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 2
		})).Return(errors.New("database error")).Once()
		// Rollback update
		mockRepo.On("Update", mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 1 && w.Balance == 1000000
		})).Return(nil).Once()

		// Execute
		result, err := service.TransferFunds(req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}
