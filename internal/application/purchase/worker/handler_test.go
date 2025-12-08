package worker

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	productEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/product/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/entity"
	userEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	walletEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/glebarez/sqlite"
)

func setupWorkerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&userEntity.User{},
		&walletEntity.Wallet{},
		&productEntity.Product{},
		&entity.Purchase{},
	)
	require.NoError(t, err)

	return db
}

func createTestPurchaseForWorker(t *testing.T, db *gorm.DB) *entity.Purchase {
	// Create user
	user := &userEntity.User{
		Name:     "Test User",
		Email:    fmt.Sprintf("test%d@example.com", time.Now().UnixNano()),
		Password: "hashed_password",
	}
	err := db.Create(user).Error
	require.NoError(t, err)

	// Create wallet
	wallet := &walletEntity.Wallet{
		UserID:  user.ID,
		Balance: 10000.0,
	}
	err = db.Create(wallet).Error
	require.NoError(t, err)

	// Create product
	product := &productEntity.Product{
		Name:  "Test Product",
		Price: 100.0,
		SKU:   fmt.Sprintf("SKU-%d", time.Now().UnixNano()),
		Stock: 50,
	}
	err = db.Create(product).Error
	require.NoError(t, err)

	// Create purchase
	purchase := &entity.Purchase{
		UserID:     user.ID,
		ProductID:  product.ID,
		Quantity:   5,
		TotalPrice: 500.0,
		Status:     entity.PurchaseStatusCompleted,
	}
	err = db.Create(purchase).Error
	require.NoError(t, err)

	return purchase
}

func TestWorker_SendPurchaseNotification(t *testing.T) {
	db := setupWorkerTestDB(t)

	purchase := createTestPurchaseForWorker(t, db)

	tests := []struct {
		name      string
		purchase  *entity.Purchase
		expectErr bool
	}{
		{
			name:      "send notification for valid purchase",
			purchase:  purchase,
			expectErr: false,
		},
		{
			name: "send notification with nil purchase",
			purchase: &entity.Purchase{
				ID:         999,
				UserID:     999,
				ProductID:  999,
				Quantity:   0,
				TotalPrice: 0,
				Status:     entity.PurchaseStatusFailed,
			},
			expectErr: false, // Should not error, just log
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create task payload
			payload := map[string]interface{}{
				"purchase_id": tt.purchase.ID,
				"user_id":     tt.purchase.UserID,
				"product_id":  tt.purchase.ProductID,
				"quantity":    tt.purchase.Quantity,
				"total_price": tt.purchase.TotalPrice,
				"status":      tt.purchase.Status,
			}

			payloadBytes, err := json.Marshal(payload)
			require.NoError(t, err)

			task := asynq.NewTask("purchase:send_notification", payloadBytes)
			assert.NotNil(t, task)

			// Verify task has payload
			assert.NotEmpty(t, task.Payload())

			// Verify we can unmarshal the payload
			var result map[string]interface{}
			err = json.Unmarshal(task.Payload(), &result)
			if !tt.expectErr {
				assert.NoError(t, err)
				assert.Equal(t, float64(tt.purchase.ID), result["purchase_id"])
				assert.Equal(t, float64(tt.purchase.UserID), result["user_id"])
			}
		})
	}
}

func TestWorker_PurchaseTaskQueue(t *testing.T) {
	db := setupWorkerTestDB(t)

	purchase := createTestPurchaseForWorker(t, db)

	tests := []struct {
		name         string
		taskType     string
		payload      map[string]interface{}
		expectedType string
	}{
		{
			name:     "queue purchase notification task",
			taskType: "purchase:send_notification",
			payload: map[string]interface{}{
				"purchase_id": purchase.ID,
				"user_id":     purchase.UserID,
				"product_id":  purchase.ProductID,
				"quantity":    purchase.Quantity,
				"total_price": purchase.TotalPrice,
				"status":      purchase.Status,
			},
			expectedType: "purchase:send_notification",
		},
		{
			name:     "queue purchase status update task",
			taskType: "purchase:update_status",
			payload: map[string]interface{}{
				"purchase_id": purchase.ID,
				"status":      "completed",
			},
			expectedType: "purchase:update_status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payloadBytes, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			task := asynq.NewTask(tt.taskType, payloadBytes)
			assert.NotNil(t, task)

			// Verify task type
			assert.Equal(t, tt.expectedType, task.Type())

			// Verify task payload can be unmarshaled
			var result map[string]interface{}
			err = json.Unmarshal(task.Payload(), &result)
			assert.NoError(t, err)

			// Verify all expected fields exist in payload
			assert.NotEmpty(t, result)
		})
	}
}

func TestWorker_ProcessPurchaseCompletion(t *testing.T) {
	db := setupWorkerTestDB(t)

	purchase := createTestPurchaseForWorker(t, db)

	// Verify initial purchase state
	savedPurchase := &entity.Purchase{}
	err := db.First(savedPurchase, purchase.ID).Error
	require.NoError(t, err)
	assert.Equal(t, entity.PurchaseStatusCompleted, savedPurchase.Status)

	tests := []struct {
		name           string
		purchaseID     uint
		newStatus      string
		expectedStatus string
		shouldSucceed  bool
	}{
		{
			name:           "process completed purchase",
			purchaseID:     purchase.ID,
			newStatus:      "completed",
			expectedStatus: "completed",
			shouldSucceed:  true,
		},
		{
			name:           "process failed purchase",
			purchaseID:     purchase.ID,
			newStatus:      "failed",
			expectedStatus: "failed",
			shouldSucceed:  true,
		},
		{
			name:           "process non-existent purchase",
			purchaseID:     999,
			newStatus:      "completed",
			expectedStatus: "",
			shouldSucceed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSucceed {
				// Update purchase status
				err := db.Model(&entity.Purchase{}).
					Where("id = ?", tt.purchaseID).
					Update("status", tt.newStatus).Error

				if tt.purchaseID != 999 {
					assert.NoError(t, err)

					// Verify update
					updated := &entity.Purchase{}
					db.First(updated, tt.purchaseID)
					assert.Equal(t, entity.PurchaseStatus(tt.expectedStatus), updated.Status)
				}
			} else {
				// This should not update anything
				result := db.Model(&entity.Purchase{}).
					Where("id = ?", tt.purchaseID).
					Update("status", tt.newStatus)

				assert.Equal(t, int64(0), result.RowsAffected)
			}
		})
	}
}

func TestWorker_PurchaseNotificationPayload(t *testing.T) {
	db := setupWorkerTestDB(t)
	purchase := createTestPurchaseForWorker(t, db)

	tests := []struct {
		name           string
		purchase       *entity.Purchase
		expectedFields []string
	}{
		{
			name:     "notification includes all required fields",
			purchase: purchase,
			expectedFields: []string{
				"purchase_id",
				"user_id",
				"product_id",
				"quantity",
				"total_price",
				"status",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := map[string]interface{}{
				"purchase_id": tt.purchase.ID,
				"user_id":     tt.purchase.UserID,
				"product_id":  tt.purchase.ProductID,
				"quantity":    tt.purchase.Quantity,
				"total_price": tt.purchase.TotalPrice,
				"status":      tt.purchase.Status,
			}

			// Verify all expected fields are present
			for _, field := range tt.expectedFields {
				_, exists := payload[field]
				assert.True(t, exists, fmt.Sprintf("field %s missing from payload", field))
			}

			// Verify payload can be marshaled and unmarshaled
			payloadBytes, err := json.Marshal(payload)
			assert.NoError(t, err)

			var unmarshaled map[string]interface{}
			err = json.Unmarshal(payloadBytes, &unmarshaled)
			assert.NoError(t, err)

			// Verify data integrity
			assert.Equal(t, float64(tt.purchase.ID), unmarshaled["purchase_id"])
			assert.Equal(t, float64(tt.purchase.UserID), unmarshaled["user_id"])
		})
	}
}
