package repository

import (
	"fmt"
	"testing"
	"time"

	productEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	purchaseDto "github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/dto"
	purchaseEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/entity"
	userEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	walletEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupPurchaseTestDB(t *testing.T) (*gorm.DB, *zap.Logger) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Migrate all required entities
	err = db.AutoMigrate(
		&userEntity.User{},
		&walletEntity.Wallet{},
		&productEntity.Product{},
		&purchaseEntity.Purchase{},
	)
	require.NoError(t, err)

	zapLogger := zap.NewNop()

	return db, zapLogger
}

func createTestUser(t *testing.T, db *gorm.DB) *userEntity.User {
	user := &userEntity.User{
		Name:     "Test User",
		Email:    fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		Password: "hashed_password",
	}
	err := db.Create(user).Error
	require.NoError(t, err)
	return user
}

func createTestWallet(t *testing.T, db *gorm.DB, userID uint) *walletEntity.Wallet {
	wallet := &walletEntity.Wallet{
		UserID:  userID,
		Balance: 10000.0,
	}
	err := db.Create(wallet).Error
	require.NoError(t, err)
	return wallet
}

func createTestProduct(t *testing.T, db *gorm.DB) *productEntity.Product {
	product := &productEntity.Product{
		Name:  "Test Product",
		Price: 100.0,
		SKU:   fmt.Sprintf("SKU-%d", time.Now().UnixNano()),
		Stock: 50,
	}
	err := db.Create(product).Error
	require.NoError(t, err)
	return product
}

func TestPurchaseRepository_Create(t *testing.T) {
	db, zapLogger := setupPurchaseTestDB(t)
	repo := NewPurchaseRepository(db, zapLogger)

	user := createTestUser(t, db)
	createTestWallet(t, db, user.ID)
	product := createTestProduct(t, db)

	tests := []struct {
		name      string
		purchase  *purchaseEntity.Purchase
		expectErr bool
	}{
		{
			name: "create valid purchase",
			purchase: &purchaseEntity.Purchase{
				UserID:     user.ID,
				ProductID:  product.ID,
				Quantity:   5,
				TotalPrice: 500.0,
				Status:     purchaseEntity.PurchaseStatusCompleted,
			},
			expectErr: false,
		},
		{
			name: "create purchase with pending status",
			purchase: &purchaseEntity.Purchase{
				UserID:     user.ID,
				ProductID:  product.ID,
				Quantity:   3,
				TotalPrice: 300.0,
				Status:     purchaseEntity.PurchaseStatusPending,
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.purchase)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.purchase.ID)

				// Verify it was saved
				saved, err := repo.GetByID(tt.purchase.ID)
				assert.NoError(t, err)
				assert.Equal(t, tt.purchase.UserID, saved.UserID)
				assert.Equal(t, tt.purchase.ProductID, saved.ProductID)
				assert.Equal(t, tt.purchase.Quantity, saved.Quantity)
			}
		})
	}
}

func TestPurchaseRepository_GetByID(t *testing.T) {
	db, zapLogger := setupPurchaseTestDB(t)
	repo := NewPurchaseRepository(db, zapLogger)

	user := createTestUser(t, db)
	createTestWallet(t, db, user.ID)
	product := createTestProduct(t, db)

	purchase := &purchaseEntity.Purchase{
		UserID:     user.ID,
		ProductID:  product.ID,
		Quantity:   2,
		TotalPrice: 200.0,
		Status:     purchaseEntity.PurchaseStatusCompleted,
	}
	err := repo.Create(purchase)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		expectErr bool
	}{
		{
			name:      "get existing purchase",
			id:        purchase.ID,
			expectErr: false,
		},
		{
			name:      "get non-existent purchase",
			id:        999,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.id)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, purchase.UserID, result.UserID)
				assert.Equal(t, purchase.ProductID, result.ProductID)
				assert.Equal(t, 2, result.Quantity)
			}
		})
	}
}

func TestPurchaseRepository_GetByUserID(t *testing.T) {
	db, zapLogger := setupPurchaseTestDB(t)
	repo := NewPurchaseRepository(db, zapLogger)

	user := createTestUser(t, db)
	createTestWallet(t, db, user.ID)
	product := createTestProduct(t, db)

	// Create multiple purchases for the same user
	purchases := []purchaseEntity.Purchase{
		{UserID: user.ID, ProductID: product.ID, Quantity: 1, TotalPrice: 100.0, Status: purchaseEntity.PurchaseStatusCompleted},
		{UserID: user.ID, ProductID: product.ID, Quantity: 2, TotalPrice: 200.0, Status: purchaseEntity.PurchaseStatusCompleted},
	}

	for i := range purchases {
		err := repo.Create(&purchases[i])
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		userID        uint
		expectedCount int
		expectErr     bool
	}{
		{
			name:          "get purchases by existing user",
			userID:        user.ID,
			expectedCount: 2,
			expectErr:     false,
		},
		{
			name:          "get purchases by non-existent user",
			userID:        999,
			expectedCount: 0,
			expectErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _, err := repo.GetByUserID(tt.userID, &purchaseDto.PurchaseFilter{
				Page:  1,
				Limit: 10,
			})

			if !tt.expectErr {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(results))
			}
		})
	}
}

func TestPurchaseRepository_Update(t *testing.T) {
	db, zapLogger := setupPurchaseTestDB(t)
	repo := NewPurchaseRepository(db, zapLogger)

	user := createTestUser(t, db)
	createTestWallet(t, db, user.ID)
	product := createTestProduct(t, db)

	purchase := &purchaseEntity.Purchase{
		UserID:     user.ID,
		ProductID:  product.ID,
		Quantity:   5,
		TotalPrice: 500.0,
		Status:     purchaseEntity.PurchaseStatusPending,
	}
	err := repo.Create(purchase)
	require.NoError(t, err)

	tests := []struct {
		name      string
		update    func(*purchaseEntity.Purchase)
		expectErr bool
	}{
		{
			name: "update purchase status",
			update: func(p *purchaseEntity.Purchase) {
				p.Status = purchaseEntity.PurchaseStatusCompleted
			},
			expectErr: false,
		},
		{
			name: "update purchase notes",
			update: func(p *purchaseEntity.Purchase) {
				p.Notes = "Updated notes"
			},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.update(purchase)
			err := repo.Update(purchase)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify update
				updated, err := repo.GetByID(purchase.ID)
				assert.NoError(t, err)
				assert.Equal(t, purchase.Status, updated.Status)
			}
		})
	}
}

func TestPurchaseRepository_Delete(t *testing.T) {
	db, zapLogger := setupPurchaseTestDB(t)
	repo := NewPurchaseRepository(db, zapLogger)

	user := createTestUser(t, db)
	createTestWallet(t, db, user.ID)
	product := createTestProduct(t, db)

	purchase := &purchaseEntity.Purchase{
		UserID:     user.ID,
		ProductID:  product.ID,
		Quantity:   2,
		TotalPrice: 200.0,
		Status:     purchaseEntity.PurchaseStatusCompleted,
	}
	err := repo.Create(purchase)
	require.NoError(t, err)

	tests := []struct {
		name      string
		id        uint
		expectErr bool
	}{
		{
			name:      "delete existing purchase",
			id:        purchase.ID,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(tt.id)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify deletion
				_, err := repo.GetByID(tt.id)
				assert.Error(t, err)
			}
		})
	}
}
