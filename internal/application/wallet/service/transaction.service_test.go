package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestWalletService_CreateTransaction(t *testing.T) {
	t.Run("should create deposit transaction successfully", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		wallet := &entity.Wallet{
			ID:      1,
			UserID:  1,
			Balance: 100.0,
		}

		req := &dto.CreateTransactionRequest{
			WalletID:    1,
			Type:        "deposit",
			Amount:      50.0,
			Description: "Deposit",
		}

		mockWalletRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)
		mockWalletRepo.On("Update", context.Background(), mock.AnythingOfType("*entity.Wallet")).Return(nil).Run(func(args mock.Arguments) {
			updatedWallet := args.Get(1).(*entity.Wallet)
			assert.Equal(t, 150.0, updatedWallet.Balance)
		})
		mockTransactionRepo.On("Create", context.Background(), mock.AnythingOfType("*entity.WalletTransaction")).Return(nil)

		// When
		response, err := service.CreateTransaction(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.WalletID)
		assert.Equal(t, "deposit", response.Type)
		assert.Equal(t, 50.0, response.Amount)
		assert.Equal(t, 150.0, response.BalanceAfter)
		mockWalletRepo.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("should create withdrawal transaction successfully", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		wallet := &entity.Wallet{
			ID:      1,
			UserID:  1,
			Balance: 200.0,
		}

		req := &dto.CreateTransactionRequest{
			WalletID:    1,
			Type:        "withdrawal",
			Amount:      50.0,
			Description: "Withdrawal",
		}

		mockWalletRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)
		mockWalletRepo.On("Update", context.Background(), mock.AnythingOfType("*entity.Wallet")).Return(nil).Run(func(args mock.Arguments) {
			updatedWallet := args.Get(1).(*entity.Wallet)
			assert.Equal(t, 150.0, updatedWallet.Balance)
		})
		mockTransactionRepo.On("Create", context.Background(), mock.AnythingOfType("*entity.WalletTransaction")).Return(nil)

		// When
		response, err := service.CreateTransaction(context.Background(), req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 150.0, response.BalanceAfter)
		mockWalletRepo.AssertExpectations(t)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("should return error for insufficient balance on withdrawal", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		wallet := &entity.Wallet{
			ID:      1,
			UserID:  1,
			Balance: 30.0,
		}

		req := &dto.CreateTransactionRequest{
			WalletID:    1,
			Type:        "withdrawal",
			Amount:      50.0,
			Description: "Withdrawal",
		}

		mockWalletRepo.On("GetByID", context.Background(), uint(1)).Return(wallet, nil)

		// When
		response, err := service.CreateTransaction(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "insufficient balance")
		mockWalletRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid transaction type", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		req := &dto.CreateTransactionRequest{
			WalletID:    1,
			Type:        "invalid_type",
			Amount:      50.0,
			Description: "Invalid",
		}

		// When
		response, err := service.CreateTransaction(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid transaction type")
	})

	t.Run("should return error when wallet not found", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		req := &dto.CreateTransactionRequest{
			WalletID:    999,
			Type:        "deposit",
			Amount:      50.0,
			Description: "Deposit",
		}

		mockWalletRepo.On("GetByID", context.Background(), uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.CreateTransaction(context.Background(), req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "wallet not found")
	})
}

func TestWalletService_GetTransactions(t *testing.T) {
	t.Run("should get wallet transactions successfully", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		transactions := []entity.WalletTransaction{
			{
				ID:       1,
				WalletID: 1,
				Type:     entity.TransactionTypeDeposit,
				Amount:   50.0,
				Status:   entity.TransactionStatusCompleted,
			},
			{
				ID:       2,
				WalletID: 1,
				Type:     entity.TransactionTypeWithdrawal,
				Amount:   20.0,
				Status:   entity.TransactionStatusCompleted,
			},
		}

		filter := &dto.TransactionFilter{
			WalletID: 1,
			Page:     1,
			PageSize: 10,
		}

		mockTransactionRepo.On("GetByWalletID", context.Background(), filter).Return(transactions, int64(2), nil)

		// When
		response, err := service.GetTransactions(context.Background(), filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 2, len(response.Data))
		assert.Equal(t, int64(2), response.TotalCount)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("should return empty list when no transactions found", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		filter := &dto.TransactionFilter{
			WalletID: 1,
			Page:     1,
			PageSize: 10,
		}

		mockTransactionRepo.On("GetByWalletID", context.Background(), filter).Return([]entity.WalletTransaction{}, int64(0), nil)

		// When
		response, err := service.GetTransactions(context.Background(), filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 0, len(response.Data))
		assert.Equal(t, int64(0), response.TotalCount)
	})

	t.Run("should return error on repository failure", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		filter := &dto.TransactionFilter{
			WalletID: 1,
			Page:     1,
			PageSize: 10,
		}

		mockTransactionRepo.On("GetByWalletID", context.Background(), filter).Return(nil, int64(0), errors.New("database error"))

		// When
		response, err := service.GetTransactions(context.Background(), filter)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

func TestWalletService_GetTransactionByID(t *testing.T) {
	t.Run("should get transaction by ID successfully", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		transaction := &entity.WalletTransaction{
			ID:           1,
			WalletID:     1,
			Type:         entity.TransactionTypeDeposit,
			Amount:       50.0,
			Status:       entity.TransactionStatusCompleted,
			BalanceAfter: 150.0,
			CreatedAt:    time.Now(),
		}

		mockTransactionRepo.On("GetByID", context.Background(), uint(1)).Return(transaction, nil)

		// When
		response, err := service.GetTransactionByID(context.Background(), 1)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, uint(1), response.WalletID)
		assert.Equal(t, "deposit", response.Type)
		mockTransactionRepo.AssertExpectations(t)
	})

	t.Run("should return error when transaction not found", func(t *testing.T) {
		// Setup
		mockWalletRepo := &testutil.MockWalletRepository{}
		mockTransactionRepo := &testutil.MockTransactionRepository{}
		logger := testutil.NewSilentLogger()
		service := NewWalletService(nil, mockWalletRepo, mockTransactionRepo, testutil.NewMockUserService(), logger)

		mockTransactionRepo.On("GetByID", context.Background(), uint(999)).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetTransactionByID(context.Background(), 999)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "transaction not found")
	})
}
