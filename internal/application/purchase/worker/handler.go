package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type PurchaseWorker struct {
	logger *zap.Logger
}

func NewPurchaseWorker(logger *zap.Logger) *PurchaseWorker {
	return &PurchaseWorker{
		logger: logger,
	}
}

// HandlePurchaseNotification handles purchase completion notification
func (w *PurchaseWorker) HandlePurchaseNotification(ctx context.Context, t *asynq.Task) error {
	var payload map[string]interface{}

	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		w.logger.Error("Failed to unmarshal payload", zap.Error(err))
		return fmt.Errorf("unmarshal failed: %w", err)
	}

	purchaseID := uint(payload["purchase_id"].(float64))
	userID := uint(payload["user_id"].(float64))
	productID := uint(payload["product_id"].(float64))

	w.logger.Info("Processing purchase notification",
		zap.Uint("purchase_id", purchaseID),
		zap.Uint("user_id", userID),
		zap.Uint("product_id", productID),
	)

	// TODO: Call gRPC payment service to notify purchase completion
	// This would be: client.NotifyPurchase(ctx, &payment.NotifyPurchaseRequest{...})

	w.logger.Info("Purchase notification processed successfully",
		zap.Uint("purchase_id", purchaseID),
	)

	return nil
}
