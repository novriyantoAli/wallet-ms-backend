package repository

import (
	"context"
	"testing"

	userEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "failed to create test database")

	err = db.AutoMigrate(&userEntity.User{}, &entity.Wallet{})
	require.NoError(t, err, "failed to migrate test database")

	return db
}

func TestWalletRepository_InterfaceImplementation(t *testing.T) {
	t.Run("should implement WalletRepository interface", func(t *testing.T) {
		logger := zap.NewNop()
		repo := NewWalletRepository(nil, logger)
		require.NotNil(t, repo)
		assert.Implements(t, (*WalletRepository)(nil), repo)
	})
}

func TestWalletRepository_Create(t *testing.T) {
	tests := []struct {
		name      string
		wallet    *entity.Wallet
		assertion func(t *testing.T, err error, wallet *entity.Wallet)
	}{
		{
			name: "should_create_wallet_successfully",
			wallet: &entity.Wallet{
				UserID:   1,
				Balance:  1000.0,
				Currency: "IDR",
				Status:   entity.WalletStatusActive,
			},
			assertion: func(t *testing.T, err error, wallet *entity.Wallet) {
				require.NoError(t, err)
				assert.NotZero(t, wallet.ID)
				assert.Equal(t, uint(1), wallet.UserID)
				assert.Equal(t, 1000.0, wallet.Balance)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			err := repo.Create(context.Background(), tt.wallet)
			tt.assertion(t, err, tt.wallet)
		})
	}
}

func TestWalletRepository_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB) uint
		assertion func(t *testing.T, err error, wallet *entity.Wallet)
	}{
		{
			name: "should_get_wallet_by_id_successfully",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   1,
					Balance:  1000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.ID
			},
			assertion: func(t *testing.T, err error, wallet *entity.Wallet) {
				require.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, uint(1), wallet.UserID)
				assert.Equal(t, 1000.0, wallet.Balance)
			},
		},
		{
			name: "should_return_error_for_nonexistent_wallet",
			setup: func(db *gorm.DB) uint {
				return 9999
			},
			assertion: func(t *testing.T, err error, wallet *entity.Wallet) {
				require.Error(t, err)
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			id := tt.setup(db)
			wallet, err := repo.GetByID(context.Background(), id)
			tt.assertion(t, err, wallet)
		})
	}
}

func TestWalletRepository_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB) uint
		assertion func(t *testing.T, err error, wallet *entity.Wallet)
	}{
		{
			name: "should_get_wallet_by_user_id_successfully",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   100,
					Balance:  5000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.UserID
			},
			assertion: func(t *testing.T, err error, wallet *entity.Wallet) {
				require.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, uint(100), wallet.UserID)
				assert.Equal(t, 5000.0, wallet.Balance)
			},
		},
		{
			name: "should_return_error_for_user_without_wallet",
			setup: func(db *gorm.DB) uint {
				return 9999
			},
			assertion: func(t *testing.T, err error, wallet *entity.Wallet) {
				require.Error(t, err)
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			userID := tt.setup(db)
			wallet, err := repo.GetByUserID(context.Background(), userID)
			tt.assertion(t, err, wallet)
		})
	}
}

func TestWalletRepository_GetByIDForUpdate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB) uint
		assertion func(t *testing.T, err error, wallet *entity.Wallet)
	}{
		{
			name: "should_get_wallet_by_id_with_update_lock",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   1,
					Balance:  1000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.ID
			},
			assertion: func(t *testing.T, err error, wallet *entity.Wallet) {
				require.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, uint(1), wallet.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			id := tt.setup(db)
			wallet, err := repo.GetByIDForUpdate(context.Background(), id)
			tt.assertion(t, err, wallet)
		})
	}
}

func TestWalletRepository_GetByUserIDForUpdate(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB) uint
		assertion func(t *testing.T, err error, wallet *entity.Wallet)
	}{
		{
			name: "should_get_wallet_by_user_id_with_update_lock",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   200,
					Balance:  3000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.UserID
			},
			assertion: func(t *testing.T, err error, wallet *entity.Wallet) {
				require.NoError(t, err)
				assert.NotNil(t, wallet)
				assert.Equal(t, uint(200), wallet.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			userID := tt.setup(db)
			wallet, err := repo.GetByUserIDForUpdate(context.Background(), userID)
			tt.assertion(t, err, wallet)
		})
	}
}

func TestWalletRepository_GetAll(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB)
		filter    *dto.WalletFilter
		assertion func(t *testing.T, err error, wallets []entity.Wallet, count int64)
	}{
		{
			name: "should_get_all_wallets_without_filter",
			setup: func(db *gorm.DB) {
				wallets := []entity.Wallet{
					{UserID: 1, Balance: 1000.0, Currency: "IDR", Status: entity.WalletStatusActive},
					{UserID: 2, Balance: 2000.0, Currency: "IDR", Status: entity.WalletStatusActive},
					{UserID: 3, Balance: 3000.0, Currency: "IDR", Status: entity.WalletStatusInactive},
				}
				for _, w := range wallets {
					require.NoError(t, db.Create(&w).Error)
				}
			},
			filter: &dto.WalletFilter{Page: 1, PageSize: 10},
			assertion: func(t *testing.T, err error, wallets []entity.Wallet, count int64) {
				require.NoError(t, err)
				assert.Equal(t, int64(3), count)
				assert.Len(t, wallets, 3)
			},
		},
		{
			name: "should_get_wallets_with_user_id_filter",
			setup: func(db *gorm.DB) {
				wallets := []entity.Wallet{
					{UserID: 1, Balance: 1000.0, Currency: "IDR", Status: entity.WalletStatusActive},
					{UserID: 2, Balance: 2000.0, Currency: "IDR", Status: entity.WalletStatusActive},
				}
				for _, w := range wallets {
					require.NoError(t, db.Create(&w).Error)
				}
			},
			filter: &dto.WalletFilter{UserID: 1, Page: 1, PageSize: 10},
			assertion: func(t *testing.T, err error, wallets []entity.Wallet, count int64) {
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)
				assert.Len(t, wallets, 1)
				assert.Equal(t, uint(1), wallets[0].UserID)
			},
		},
		{
			name: "should_handle_pagination_correctly",
			setup: func(db *gorm.DB) {
				for i := 1; i <= 25; i++ {
					wallet := entity.Wallet{
						UserID:   uint(i),
						Balance:  float64(i * 1000),
						Currency: "IDR",
						Status:   entity.WalletStatusActive,
					}
					require.NoError(t, db.Create(&wallet).Error)
				}
			},
			filter: &dto.WalletFilter{Page: 2, PageSize: 10},
			assertion: func(t *testing.T, err error, wallets []entity.Wallet, count int64) {
				require.NoError(t, err)
				assert.Equal(t, int64(25), count)
				assert.Len(t, wallets, 10)
				assert.Equal(t, uint(11), wallets[0].UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			tt.setup(db)
			wallets, count, err := repo.GetAll(context.Background(), tt.filter)
			tt.assertion(t, err, wallets, count)
		})
	}
}

func TestWalletRepository_Update(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB) *entity.Wallet
		update    func(w *entity.Wallet)
		assertion func(t *testing.T, err error, db *gorm.DB, wallet *entity.Wallet)
	}{
		{
			name: "should_update_wallet_successfully",
			setup: func(db *gorm.DB) *entity.Wallet {
				wallet := &entity.Wallet{
					UserID:   1,
					Balance:  1000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet
			},
			update: func(w *entity.Wallet) {
				w.Balance = 2000.0
				w.Status = entity.WalletStatusInactive
			},
			assertion: func(t *testing.T, err error, db *gorm.DB, wallet *entity.Wallet) {
				require.NoError(t, err)
				var updated entity.Wallet
				require.NoError(t, db.First(&updated, wallet.ID).Error)
				assert.Equal(t, 2000.0, updated.Balance)
				assert.Equal(t, entity.WalletStatusInactive, updated.Status)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			wallet := tt.setup(db)
			tt.update(wallet)
			err := repo.Update(context.Background(), wallet)
			tt.assertion(t, err, db, wallet)
		})
	}
}

func TestWalletRepository_Delete(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB) uint
		assertion func(t *testing.T, err error, db *gorm.DB, id uint)
	}{
		{
			name: "should_delete_wallet_successfully",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   1,
					Balance:  1000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.ID
			},
			assertion: func(t *testing.T, err error, db *gorm.DB, id uint) {
				require.NoError(t, err)
				var wallet entity.Wallet
				err = db.First(&wallet, id).Error
				assert.Equal(t, gorm.ErrRecordNotFound, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			id := tt.setup(db)
			err := repo.Delete(context.Background(), id)
			tt.assertion(t, err, db, id)
		})
	}
}

func TestWalletRepository_UpdateBalance(t *testing.T) {
	tests := []struct {
		name            string
		setup           func(db *gorm.DB) uint
		amount          float64
		expectedBalance float64
		assertion       func(t *testing.T, err error, db *gorm.DB, walletID uint, expectedBalance float64)
	}{
		{
			name: "should_increase_wallet_balance",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   1,
					Balance:  1000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.ID
			},
			amount:          500.0,
			expectedBalance: 1500.0,
			assertion: func(t *testing.T, err error, db *gorm.DB, walletID uint, expectedBalance float64) {
				require.NoError(t, err)
				var wallet entity.Wallet
				require.NoError(t, db.First(&wallet, walletID).Error)
				assert.Equal(t, expectedBalance, wallet.Balance)
			},
		},
		{
			name: "should_decrease_wallet_balance",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   2,
					Balance:  1000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.ID
			},
			amount:          -300.0,
			expectedBalance: 700.0,
			assertion: func(t *testing.T, err error, db *gorm.DB, walletID uint, expectedBalance float64) {
				require.NoError(t, err)
				var wallet entity.Wallet
				require.NoError(t, db.First(&wallet, walletID).Error)
				assert.Equal(t, expectedBalance, wallet.Balance)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			walletID := tt.setup(db)
			err := repo.UpdateBalance(context.Background(), walletID, tt.amount)
			tt.assertion(t, err, db, walletID, tt.expectedBalance)
		})
	}
}

func TestWalletRepository_GetByIDWithUser(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB) uint
		assertion func(t *testing.T, err error, walletWithUser *WalletWithUserData)
	}{
		{
			name: "should_get_wallet_with_user_data",
			setup: func(db *gorm.DB) uint {
				wallet := &entity.Wallet{
					UserID:   1,
					Balance:  1000.0,
					Currency: "IDR",
					Status:   entity.WalletStatusActive,
				}
				require.NoError(t, db.Create(wallet).Error)
				return wallet.ID
			},
			assertion: func(t *testing.T, err error, walletWithUser *WalletWithUserData) {
				require.NoError(t, err)
				assert.NotNil(t, walletWithUser)
				assert.Equal(t, uint(1), walletWithUser.UserID)
				assert.Equal(t, 1000.0, walletWithUser.Balance)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			id := tt.setup(db)
			walletWithUser, err := repo.GetByIDWithUser(context.Background(), id)
			tt.assertion(t, err, walletWithUser)
		})
	}
}

func TestWalletRepository_GetAllWithUser(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(db *gorm.DB)
		filter    *dto.WalletFilter
		assertion func(t *testing.T, err error, wallets []WalletWithUserData, count int64)
	}{
		{
			name: "should_get_all_wallets_with_user_without_filter",
			setup: func(db *gorm.DB) {
				wallets := []entity.Wallet{
					{UserID: 1, Balance: 1000.0, Currency: "IDR", Status: entity.WalletStatusActive},
					{UserID: 2, Balance: 2000.0, Currency: "IDR", Status: entity.WalletStatusActive},
				}
				for _, w := range wallets {
					require.NoError(t, db.Create(&w).Error)
				}
			},
			filter: &dto.WalletFilter{Page: 1, PageSize: 10},
			assertion: func(t *testing.T, err error, wallets []WalletWithUserData, count int64) {
				require.NoError(t, err)
				assert.Equal(t, int64(2), count)
				assert.Len(t, wallets, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			repo := NewWalletRepository(db, zap.NewNop())
			tt.setup(db)
			wallets, count, err := repo.GetAllWithUser(context.Background(), tt.filter)
			tt.assertion(t, err, wallets, count)
		})
	}
}

func TestWalletRepository_WithTransactionContext(t *testing.T) {
	t.Run("should_use_transaction_db_when_provided_in_context", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewWalletRepository(db, zap.NewNop())

		wallet := &entity.Wallet{
			UserID:   1,
			Balance:  1000.0,
			Currency: "IDR",
			Status:   entity.WalletStatusActive,
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			ctx := database.WithTx(context.Background(), tx)
			return repo.Create(ctx, wallet)
		})

		require.NoError(t, err)

		var retrieved entity.Wallet
		require.NoError(t, db.First(&retrieved, wallet.ID).Error)
		assert.Equal(t, uint(1), retrieved.UserID)
	})

	t.Run("should_rollback_on_transaction_error", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewWalletRepository(db, zap.NewNop())

		wallet := &entity.Wallet{
			UserID:   2,
			Balance:  1000.0,
			Currency: "IDR",
			Status:   entity.WalletStatusActive,
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			ctx := database.WithTx(context.Background(), tx)
			err := repo.Create(ctx, wallet)
			if err != nil {
				return err
			}
			return gorm.ErrInvalidData
		})

		require.Error(t, err)

		var retrieved entity.Wallet
		err = db.First(&retrieved, wallet.ID).Error
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})

	t.Run("should_perform_balance_update_in_transaction", func(t *testing.T) {
		db := setupTestDB(t)
		repo := NewWalletRepository(db, zap.NewNop())

		wallet := &entity.Wallet{
			UserID:   3,
			Balance:  1000.0,
			Currency: "IDR",
			Status:   entity.WalletStatusActive,
		}
		require.NoError(t, db.Create(wallet).Error)

		err := db.Transaction(func(tx *gorm.DB) error {
			ctx := database.WithTx(context.Background(), tx)
			return repo.UpdateBalance(ctx, wallet.ID, 500.0)
		})

		require.NoError(t, err)

		var updated entity.Wallet
		require.NoError(t, db.First(&updated, wallet.ID).Error)
		assert.Equal(t, 1500.0, updated.Balance)
	})
}
