package service

import (
	"context"
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
		mockRepo.On("GetByUserID", context.Background(), req.UserID).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Create", context.Background(), mock.AnythingOfType("*entity.Wallet")).Return(nil).Run(func(args mock.Arguments) {
			wallet := args.Get(1).(*entity.Wallet)
			wallet.ID = 1
		})
		mockRepo.On("GetByUserIDWithUser", context.Background(), req.UserID).Return(walletWithUser, nil)

		// Execute
		result, err := service.CreateWallet(context.Background(), req)

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
		_, err := service.CreateWallet(context.Background(), req)

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
		mockRepo.On("GetByUserID", context.Background(), req.UserID).Return(existingWallet, nil)

		// Execute
		_, err := service.CreateWallet(context.Background(), req)

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

		walletWithUser := &repository.WalletWithUserData{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserName:  "Test User",
			UserEmail: "test@example.com",
		}

		// Mock expectations - use GetByIDWithUser instead of GetByID
		mockRepo.On("GetByIDWithUser", context.Background(), uint(1)).Return(walletWithUser, nil)

		// Execute
		result, err := service.GetWalletByID(context.Background(), 1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, walletWithUser.ID, result.ID)
		assert.Equal(t, walletWithUser.UserID, result.UserID)
		assert.Equal(t, walletWithUser.Balance, result.Balance)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error if wallet not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

		// Mock expectations
		mockRepo.On("GetByIDWithUser", context.Background(), uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		_, err := service.GetWalletByID(context.Background(), 999)

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

		walletWithUser := &repository.WalletWithUserData{
			ID:        1,
			UserID:    1,
			Balance:   100000,
			Currency:  "IDR",
			Status:    "active",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserName:  "Test User",
			UserEmail: "test@example.com",
		}

		// Mock expectations
		mockRepo.On("GetByUserIDWithUser", context.Background(), uint(1)).Return(walletWithUser, nil)

		// Execute
		result, err := service.GetWalletByUserID(context.Background(), 1)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, walletWithUser.ID, result.ID)
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_UpdateWalletBalance(t *testing.T) {
	t.Run("should credit balance successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)
		mockRepo.On("Update", context.Background(), mock.AnythingOfType("*entity.Wallet")).Return(nil)

		// Execute
		result, err := service.UpdateWalletBalance(context.Background(), 1, req)

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
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)
		mockRepo.On("Update", context.Background(), mock.AnythingOfType("*entity.Wallet")).Return(nil)

		// Execute
		result, err := service.UpdateWalletBalance(context.Background(), 1, req)

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
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, &testutil.MockTransactionRepository{}, mockUserService, logger)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)

		// Execute
		_, err := service.UpdateWalletBalance(context.Background(), 1, req)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)

		// Execute
		_, err := service.UpdateWalletBalance(context.Background(), 1, req)

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
		mockRepo.On("GetAllWithUser", context.Background(), filter).Return(wallets, int64(2), nil)

		// Execute
		result, err := service.GetWallets(context.Background(), filter)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)
		mockRepo.On("Delete", context.Background(), uint(1)).Return(nil)

		// Execute
		err := service.DeleteWallet(context.Background(), 1)

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
		mockRepo.On("GetByID", context.Background(), uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		err := service.DeleteWallet(context.Background(), 999)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", context.Background(), uint(2)).Return(toWallet, nil)
		mockRepo.On("Update", context.Background(), mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 1 && w.Balance == 750000
		})).Return(nil)
		mockRepo.On("Update", context.Background(), mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 2 && w.Balance == 750000
		})).Return(nil)

		// Execute
		result, err := service.TransferFunds(context.Background(), req)

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
		result, err := service.TransferFunds(context.Background(), req)

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
		mockRepo.On("GetByID", context.Background(), uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		result, err := service.TransferFunds(context.Background(), req)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", context.Background(), uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// Execute
		result, err := service.TransferFunds(context.Background(), req)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", context.Background(), uint(2)).Return(toWallet, nil)

		// Execute
		result, err := service.TransferFunds(context.Background(), req)

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
		mockRepo.On("GetByID", context.Background(), uint(1)).Return(fromWallet, nil)
		mockRepo.On("GetByID", context.Background(), uint(2)).Return(toWallet, nil)
		mockRepo.On("Update", context.Background(), mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 1
		})).Return(nil).Once()
		mockRepo.On("Update", context.Background(), mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 2
		})).Return(errors.New("database error")).Once()
		// Rollback update
		mockRepo.On("Update", context.Background(), mock.MatchedBy(func(w *entity.Wallet) bool {
			return w.ID == 1 && w.Balance == 1000000
		})).Return(nil).Once()

		// Execute
		result, err := service.TransferFunds(context.Background(), req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestWalletService_Transfer(t *testing.T) {
	t.Run("should reject transfer when sender is regular user", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		mockTxRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, mockTxRepo, mockUserService, logger)

		senderID := uint(5)
		recipientID := uint(6)

		senderUserResp := &userDto.UserResponse{
			ID:    senderID,
			Name:  "Regular User",
			Email: "user@example.com",
			Level: "user",
		}

		mockUserService.On("GetUserByID", senderID).Return(senderUserResp, nil).Once()

		req := &dto.TransferWalletRequest{
			RecipientUserID: recipientID,
			Amount:          50000,
			Description:     "Should fail",
		}

		// Execute
		result, err := service.Transfer(context.Background(), senderID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "users cannot transfer funds", err.Error())
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return error when sender user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		mockTxRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, mockTxRepo, mockUserService, logger)

		senderID := uint(7)
		recipientID := uint(8)

		mockUserService.On("GetUserByID", senderID).Return(nil, errors.New("user not found")).Once()

		req := &dto.TransferWalletRequest{
			RecipientUserID: recipientID,
			Amount:          50000,
			Description:     "Should fail",
		}

		// Execute
		result, err := service.Transfer(context.Background(), senderID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "sender user not found", err.Error())
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return error when recipient user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		mockTxRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, mockTxRepo, mockUserService, logger)

		senderID := uint(9)
		recipientID := uint(10)

		senderUserResp := &userDto.UserResponse{
			ID:    senderID,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		mockUserService.On("GetUserByID", senderID).Return(senderUserResp, nil).Once()
		mockUserService.On("GetUserByID", recipientID).Return(nil, errors.New("user not found")).Once()

		req := &dto.TransferWalletRequest{
			RecipientUserID: recipientID,
			Amount:          50000,
			Description:     "Should fail",
		}

		// Execute
		result, err := service.Transfer(context.Background(), senderID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "recipient user not found", err.Error())
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return error when transferring to same user", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		mockTxRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, mockTxRepo, mockUserService, logger)

		senderID := uint(11)

		senderUserResp := &userDto.UserResponse{
			ID:    senderID,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		// GetUserByID is called twice: once for sender, once for recipient (even though same ID)
		mockUserService.On("GetUserByID", senderID).Return(senderUserResp, nil).Twice()

		req := &dto.TransferWalletRequest{
			RecipientUserID: senderID,
			Amount:          50000,
			Description:     "Self transfer",
		}

		// Execute
		result, err := service.Transfer(context.Background(), senderID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "cannot transfer to the same user", err.Error())
		mockUserService.AssertExpectations(t)
	})

	t.Run("should return error when sender wallet not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		mockTxRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, mockTxRepo, mockUserService, logger)

		senderID := uint(12)
		recipientID := uint(13)

		senderUserResp := &userDto.UserResponse{
			ID:    senderID,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		recipientUserResp := &userDto.UserResponse{
			ID:    recipientID,
			Name:  "Recipient",
			Email: "recipient@example.com",
			Level: "user",
		}

		mockUserService.On("GetUserByID", senderID).Return(senderUserResp, nil).Once()
		mockUserService.On("GetUserByID", recipientID).Return(recipientUserResp, nil).Once()
		mockRepo.On("GetByUserID", context.Background(), senderID).Return(nil, gorm.ErrRecordNotFound).Once()

		req := &dto.TransferWalletRequest{
			RecipientUserID: recipientID,
			Amount:          50000,
			Description:     "Should fail",
		}

		// Execute
		result, err := service.Transfer(context.Background(), senderID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "sender wallet not found", err.Error())
		mockUserService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when recipient wallet not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		mockTxRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, mockTxRepo, mockUserService, logger)

		senderID := uint(14)
		recipientID := uint(15)

		senderUserResp := &userDto.UserResponse{
			ID:    senderID,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		recipientUserResp := &userDto.UserResponse{
			ID:    recipientID,
			Name:  "Recipient",
			Email: "recipient@example.com",
			Level: "user",
		}

		senderWallet := &entity.Wallet{
			ID:       1,
			UserID:   senderID,
			Balance:  100000,
			Currency: "IDR",
			Status:   "active",
		}

		mockUserService.On("GetUserByID", senderID).Return(senderUserResp, nil).Once()
		mockUserService.On("GetUserByID", recipientID).Return(recipientUserResp, nil).Once()
		mockRepo.On("GetByUserID", context.Background(), senderID).Return(senderWallet, nil).Once()
		mockRepo.On("GetByUserID", context.Background(), recipientID).Return(nil, gorm.ErrRecordNotFound).Once()

		req := &dto.TransferWalletRequest{
			RecipientUserID: recipientID,
			Amount:          50000,
			Description:     "Should fail",
		}

		// Execute
		result, err := service.Transfer(context.Background(), senderID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "recipient wallet not found", err.Error())
		mockUserService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when insufficient balance", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockWalletRepository{}
		mockUserService := &testutil.MockUserService{}
		mockTxRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		db, _ := testutil.SetupTestDB()
		service := NewWalletService(db, mockRepo, mockTxRepo, mockUserService, logger)

		senderID := uint(16)
		recipientID := uint(17)

		senderUserResp := &userDto.UserResponse{
			ID:    senderID,
			Name:  "Reseller",
			Email: "reseller@example.com",
			Level: "reseller",
		}

		recipientUserResp := &userDto.UserResponse{
			ID:    recipientID,
			Name:  "Recipient",
			Email: "recipient@example.com",
			Level: "user",
		}

		senderWallet := &entity.Wallet{
			ID:       2,
			UserID:   senderID,
			Balance:  10000, // Only 10k, trying to transfer 50k
			Currency: "IDR",
			Status:   "active",
		}

		recipientWallet := &entity.Wallet{
			ID:       3,
			UserID:   recipientID,
			Balance:  50000,
			Currency: "IDR",
			Status:   "active",
		}

		mockUserService.On("GetUserByID", senderID).Return(senderUserResp, nil).Once()
		mockUserService.On("GetUserByID", recipientID).Return(recipientUserResp, nil).Once()
		mockRepo.On("GetByUserID", context.Background(), senderID).Return(senderWallet, nil).Once()
		mockRepo.On("GetByUserID", context.Background(), recipientID).Return(recipientWallet, nil).Once()

		req := &dto.TransferWalletRequest{
			RecipientUserID: recipientID,
			Amount:          50000,
			Description:     "Should fail",
		}

		// Execute
		result, err := service.Transfer(context.Background(), senderID, req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "insufficient balance for transfer", err.Error())
		mockUserService.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}
