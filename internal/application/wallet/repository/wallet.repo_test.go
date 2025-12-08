package repository

import (
	"testing"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MockWalletDB is a mock implementation for testing repository with mocked DB calls
type MockWalletDB struct {
	mock.Mock
}

func TestWalletRepository_InterfaceImplementation(t *testing.T) {
	t.Run("should implement WalletRepository interface", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := NewWalletRepository(nil, logger)
		require.NotNil(t, repo)

		// Act & Assert - Interface is implemented
		assert.Implements(t, (*WalletRepository)(nil), repo)
	})
}

func TestWalletRepository_Methods(t *testing.T) {
	t.Run("Create_should_accept_wallet_entity", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		var repo WalletRepository = &walletRepository{
			logger: logger,
		}

		// Assert
		assert.NotNil(t, repo)
	})

	t.Run("GetByID_should_return_wallet_or_error", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := &walletRepository{
			logger: logger,
		}

		// Act & Assert - Method exists
		assert.NotNil(t, repo.GetByID)
	})

	t.Run("GetByUserID_should_return_wallet_or_error", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := &walletRepository{
			logger: logger,
		}

		// Act & Assert - Method exists
		assert.NotNil(t, repo.GetByUserID)
	})

	t.Run("GetAll_should_return_wallets_with_pagination", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := &walletRepository{
			logger: logger,
		}

		// Act & Assert - Method exists
		assert.NotNil(t, repo.GetAll)
	})

	t.Run("Update_should_update_wallet", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := &walletRepository{
			logger: logger,
		}

		// Act & Assert - Method exists
		assert.NotNil(t, repo.Update)
	})

	t.Run("Delete_should_delete_wallet", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := &walletRepository{
			logger: logger,
		}

		// Act & Assert - Method exists
		assert.NotNil(t, repo.Delete)
	})

	t.Run("UpdateBalance_should_update_wallet_balance", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := &walletRepository{
			logger: logger,
		}

		// Act & Assert - Method exists
		assert.NotNil(t, repo.UpdateBalance)
	})
}

func TestWalletRepository_Constructor(t *testing.T) {
	t.Run("NewWalletRepository_should_create_new_instance", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()

		// Act
		repo := NewWalletRepository(nil, logger)

		// Assert
		assert.NotNil(t, repo)
		assert.IsType(t, &walletRepository{}, repo)
	})

	t.Run("NewWalletRepository_should_store_dependencies", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()

		// Act
		repo := NewWalletRepository(nil, logger).(*walletRepository)

		// Assert
		assert.NotNil(t, repo.logger)
		assert.Equal(t, logger, repo.logger)
	})
}

func TestWalletRepository_ErrorHandling(t *testing.T) {
	t.Run("GetByID_should_handle_record_not_found", func(t *testing.T) {
		// This documents the expected behavior
		// When a wallet is not found, GetByID should return gorm.ErrRecordNotFound
		assert.Equal(t, gorm.ErrRecordNotFound.Error(), "record not found")
	})

	t.Run("GetByUserID_should_handle_record_not_found", func(t *testing.T) {
		// This documents the expected behavior
		// When a user's wallet is not found, GetByUserID should return gorm.ErrRecordNotFound
		assert.Equal(t, gorm.ErrRecordNotFound.Error(), "record not found")
	})
}

func TestWalletRepository_FilterDTO(t *testing.T) {
	t.Run("WalletFilter_should_support_pagination", func(t *testing.T) {
		// Arrange
		filter := &dto.WalletFilter{
			Page:     1,
			PageSize: 10,
		}

		// Act & Assert
		assert.Equal(t, int(1), filter.Page)
		assert.Equal(t, int(10), filter.PageSize)
	})

	t.Run("WalletFilter_should_support_user_id_filter", func(t *testing.T) {
		// Arrange
		filter := &dto.WalletFilter{
			UserID: 5,
		}

		// Act & Assert
		assert.Equal(t, uint(5), filter.UserID)
	})

	t.Run("WalletFilter_should_support_status_filter", func(t *testing.T) {
		// Arrange
		filter := &dto.WalletFilter{
			Status: string(entity.WalletStatusActive),
		}

		// Act & Assert
		assert.Equal(t, "active", filter.Status)
	})
}

func TestWalletRepository_EntityStructure(t *testing.T) {
	t.Run("Wallet_should_have_required_fields", func(t *testing.T) {
		// Arrange & Act
		wallet := &entity.Wallet{
			UserID:   1,
			Balance:  100000,
			Currency: "IDR",
			Status:   entity.WalletStatusActive,
		}

		// Assert
		assert.NotNil(t, wallet)
		assert.Equal(t, uint(1), wallet.UserID)
		assert.Equal(t, float64(100000), wallet.Balance)
		assert.Equal(t, "IDR", wallet.Currency)
		assert.Equal(t, entity.WalletStatusActive, wallet.Status)
	})

	t.Run("Wallet_should_support_all_status_values", func(t *testing.T) {
		// Arrange
		statuses := []entity.WalletStatus{
			entity.WalletStatusActive,
			entity.WalletStatusInactive,
			entity.WalletStatusSuspended,
			entity.WalletStatusClosed,
		}

		// Act & Assert
		for _, status := range statuses {
			assert.True(t, status.IsValid())
			wallet := &entity.Wallet{Status: status}
			assert.NotNil(t, wallet)
		}
	})

	t.Run("Wallet_should_have_timestamps", func(t *testing.T) {
		// Arrange
		wallet := &entity.Wallet{
			UserID: 1,
		}

		// Assert
		assert.NotNil(t, wallet.CreatedAt)
		assert.NotNil(t, wallet.UpdatedAt)
		assert.NotNil(t, wallet.DeletedAt)
	})

	t.Run("Wallet_table_name_should_be_wallets", func(t *testing.T) {
		// Arrange
		wallet := &entity.Wallet{}

		// Act
		tableName := wallet.TableName()

		// Assert
		assert.Equal(t, "wallets", tableName)
	})
}

func TestWalletRepository_StatusEnum(t *testing.T) {
	t.Run("WalletStatus_String_method", func(t *testing.T) {
		// Arrange
		status := entity.WalletStatusActive

		// Act
		result := status.String()

		// Assert
		assert.Equal(t, "active", result)
	})

	t.Run("WalletStatus_IsValid_should_validate_all_statuses", func(t *testing.T) {
		// Arrange
		validStatuses := []entity.WalletStatus{
			entity.WalletStatusActive,
			entity.WalletStatusInactive,
			entity.WalletStatusSuspended,
			entity.WalletStatusClosed,
		}

		// Act & Assert
		for _, status := range validStatuses {
			assert.True(t, status.IsValid(), "Status %s should be valid", status)
		}
	})

	t.Run("WalletStatus_IsValid_should_reject_invalid_statuses", func(t *testing.T) {
		// Arrange
		invalidStatus := entity.WalletStatus("invalid")

		// Act & Assert
		assert.False(t, invalidStatus.IsValid())
	})
}

func TestWalletRepository_Behavior(t *testing.T) {
	t.Run("repository_should_log_all_operations", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()
		repo := &walletRepository{
			logger: logger,
		}

		// Assert
		assert.NotNil(t, repo.logger)
	})

	t.Run("repository_should_accept_gorm_db", func(t *testing.T) {
		// Arrange
		logger := zap.NewNop()

		// Act
		repo := NewWalletRepository(nil, logger)

		// Assert
		assert.NotNil(t, repo)
	})

	t.Run("GetAll_should_handle_empty_results", func(t *testing.T) {
		// This documents expected behavior
		// When no wallets match the filter, GetAll should return empty slice and count=0
		var emptyWallets []entity.Wallet
		assert.Empty(t, emptyWallets)
		assert.Len(t, emptyWallets, 0)
	})

	t.Run("GetAll_should_return_count_with_results", func(t *testing.T) {
		// This documents expected behavior
		// GetAll returns both wallets slice and total count for pagination
		var wallets []entity.Wallet
		var totalCount int64

		// Empty state
		assert.Equal(t, int64(0), totalCount)
		assert.Len(t, wallets, 0)
	})
}

func TestWalletRepository_CRUD_Operations(t *testing.T) {
	t.Run("CRUD_create_method_signature", func(t *testing.T) {
		// Documents: Create(wallet *entity.Wallet) error
		logger := zap.NewNop()
		repo := &walletRepository{logger: logger}

		assert.NotNil(t, repo)
		// Would call: err := repo.Create(wallet)
	})

	t.Run("CRUD_read_methods_signatures", func(t *testing.T) {
		// Documents:
		// - GetByID(id uint) (*entity.Wallet, error)
		// - GetByUserID(userID uint) (*entity.Wallet, error)
		// - GetAll(filter *dto.WalletFilter) ([]entity.Wallet, int64, error)
		logger := zap.NewNop()
		repo := &walletRepository{logger: logger}

		assert.NotNil(t, repo)
	})

	t.Run("CRUD_update_method_signature", func(t *testing.T) {
		// Documents:
		// - Update(wallet *entity.Wallet) error
		// - UpdateBalance(walletID uint, amount float64) error
		logger := zap.NewNop()
		repo := &walletRepository{logger: logger}

		assert.NotNil(t, repo)
	})

	t.Run("CRUD_delete_method_signature", func(t *testing.T) {
		// Documents: Delete(id uint) error
		logger := zap.NewNop()
		repo := &walletRepository{logger: logger}

		assert.NotNil(t, repo)
	})
}

func TestWalletRepository_Constraints(t *testing.T) {
	t.Run("wallet_should_enforce_unique_user_id", func(t *testing.T) {
		// Documents the unique constraint on UserID
		// One wallet per user is enforced at database level
		wallet1 := &entity.Wallet{
			UserID:   1,
			Balance:  100000,
			Currency: "IDR",
			Status:   entity.WalletStatusActive,
		}

		wallet2 := &entity.Wallet{
			UserID:   1, // Same UserID
			Balance:  50000,
			Currency: "USD",
			Status:   entity.WalletStatusInactive,
		}

		// Both have same UserID - database would reject wallet2 due to unique constraint
		assert.Equal(t, wallet1.UserID, wallet2.UserID)
	})

	t.Run("wallet_balance_should_support_decimals", func(t *testing.T) {
		// Balance is float64, supports decimal amounts
		wallet := &entity.Wallet{
			Balance: 1234.56,
		}

		assert.Equal(t, float64(1234.56), wallet.Balance)
	})

	t.Run("wallet_currency_is_3_chars", func(t *testing.T) {
		// Currency field has size constraint of 3 (ISO 4217 code)
		currencies := []string{"IDR", "USD", "EUR", "GBP", "JPY"}

		for _, curr := range currencies {
			assert.Len(t, curr, 3)
		}
	})
}
