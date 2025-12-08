package port

import "context"

// PaymentGateway defines the interface for payment gateway operations
// This follows the Dependency Inversion Principle - the high-level module (worker)
// depends on abstractions (this interface) rather than concrete implementations
type PaymentGateway interface {
	// CheckPaymentStatus checks the current status of a payment with the gateway
	CheckPaymentStatus(ctx context.Context, transactionID string) (*PaymentStatusResponse, error)

	// ProcessPayment processes a payment with the gateway
	ProcessPayment(ctx context.Context, req *ProcessPaymentRequest) (*ProcessPaymentResponse, error)
}

// ProcessPaymentRequest contains data needed to process a payment
type ProcessPaymentRequest struct {
	TransactionID string
	Amount        float64
	Currency      string
	Description   string
}

// PaymentStatusResponse contains the response from checking payment status
type PaymentStatusResponse struct {
	Status   string // "success", "pending", "failed"
	Amount   float64
	Currency string
	Message  string
}

// ProcessPaymentResponse contains the response from processing a payment
type ProcessPaymentResponse struct {
	Status        string // "success", "pending", "failed"
	Message       string
	TransactionID string
}
