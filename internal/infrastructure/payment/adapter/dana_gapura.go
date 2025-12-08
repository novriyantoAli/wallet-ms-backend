package adapter

import (
	"context"
	"fmt"
	"os"

	dana "github.com/dana-id/dana-go"
	"github.com/dana-id/dana-go/config"
	payment_gateway "github.com/dana-id/dana-go/payment_gateway/v1"
	"github.com/novriyantoAli/wallet-ms-backend/internal/domain/payment/port"
	"go.uber.org/zap"
)

// DanaGapuraClient implements the PaymentGateway interface
// menggunakan official Dana Go SDK
type DanaGapuraClient struct {
	apiClient  *dana.APIClient
	partnerID  string
	merchantID string
	logger     *zap.Logger
}

// NewDanaGapuraClient creates new Dana payment gateway client
// Menggunakan official Dana Go SDK
func NewDanaGapuraClient(
	baseURL string,
	partnerID string,
	privateKeyPEM string,
	publicKeyPEM string,
	logger *zap.Logger,
) (*DanaGapuraClient, error) {
	// Configure Dana API client sesuai dokumentasi
	configuration := config.NewConfiguration()
	configuration.APIKey = &config.APIKey{
		DANA_ENV:     config.ENV_SANDBOX, // Use ENV_PRODUCTION for production
		X_PARTNER_ID: partnerID,
		PRIVATE_KEY:  privateKeyPEM,
		// PRIVATE_KEY_PATH: privateKeyPath, // Alternative: use file path
		ORIGIN: os.Getenv("ORIGIN"),
	}

	// Create API client
	apiClient := dana.NewAPIClient(configuration)

	client := &DanaGapuraClient{
		apiClient:  apiClient,
		partnerID:  partnerID,
		merchantID: partnerID, // Menggunakan partnerID sebagai merchantID untuk Dana API
		logger:     logger,
	}

	return client, nil
}

// CheckPaymentStatus implements PaymentGateway.CheckPaymentStatus
// Mengecek status pembayaran menggunakan QueryPayment API dari Dana
func (c *DanaGapuraClient) CheckPaymentStatus(
	ctx context.Context,
	transactionID string,
) (*port.PaymentStatusResponse, error) {
	c.logger.Debug("Checking payment status",
		zap.String("transaction_id", transactionID))

	// Build QueryPaymentRequest sesuai dokumentasi Dana
	request := payment_gateway.NewQueryPaymentRequest(
		transactionID, // originalPartnerReferenceNo
		c.merchantID,  // merchantId
	)

	// Call Dana QueryPayment API
	resp, _, err := c.apiClient.PaymentGatewayAPI.
		QueryPayment(ctx).
		QueryPaymentRequest(*request).
		Execute()

	if err != nil {
		c.logger.Error("Failed to query payment status from Dana",
			zap.String("transaction_id", transactionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to query payment status: %w", err)
	}

	// Parse response
	if resp == nil {
		return nil, fmt.Errorf("empty response from Dana API")
	}

	// Get LatestTransactionStatus dari response
	paymentStatus := mapDanaStatusToLocal(resp.LatestTransactionStatus)

	return &port.PaymentStatusResponse{
		Status:  paymentStatus,
		Message: "Payment status retrieved successfully",
	}, nil
}

// ProcessPayment implements PaymentGateway.ProcessPayment
// Memproses pembayaran menggunakan CreateOrder API dari Dana
func (c *DanaGapuraClient) ProcessPayment(
	ctx context.Context,
	req *port.ProcessPaymentRequest,
) (*port.ProcessPaymentResponse, error) {
	c.logger.Debug("Processing payment",
		zap.String("transaction_id", req.TransactionID),
		zap.Float64("amount", req.Amount))

	// Build Money object
	money := payment_gateway.NewMoney(
		fmt.Sprintf("%.2f", req.Amount), // value
		req.Currency,                    // currency
	)

	// Build CreateOrderByApiRequest dengan field yang benar
	urlParams := []payment_gateway.UrlParam{}

	danaReq := payment_gateway.NewCreateOrderByApiRequest(
		[]payment_gateway.PayOptionDetail{}, // payOptionDetails (empty list)
		req.TransactionID,                   // partnerReferenceNo
		c.merchantID,                        // merchantId
		*money,                              // amount
		urlParams,                           // urlParams
	)

	// Call Dana CreateOrder API
	resp, _, err := c.apiClient.PaymentGatewayAPI.
		CreateOrder(ctx).
		CreateOrderRequest(payment_gateway.CreateOrderByApiRequestAsCreateOrderRequest(danaReq)).
		Execute()

	if err != nil {
		c.logger.Error("Failed to process payment with Dana",
			zap.String("transaction_id", req.TransactionID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

	// Parse response
	if resp == nil {
		return nil, fmt.Errorf("empty response from Dana API")
	}

	// Get status dari ResponseCode (00 = success, others = pending/failed)
	paymentStatus := mapDanaResponseCodeToStatus(resp.ResponseCode)

	// Get transaction ID dari response
	transactionID := req.TransactionID
	if resp.ReferenceNo != nil {
		transactionID = *resp.ReferenceNo
	}

	return &port.ProcessPaymentResponse{
		Status:        paymentStatus,
		Message:       "Payment processed successfully",
		TransactionID: transactionID,
	}, nil
}

// RefundPayment implements payment refund using Dana RefundOrder API
func (c *DanaGapuraClient) RefundPayment(
	ctx context.Context,
	transactionID string,
	refundAmount float64,
	reason string,
) error {
	c.logger.Debug("Refunding payment",
		zap.String("transaction_id", transactionID),
		zap.Float64("amount", refundAmount))

	// Build Money object for refund amount
	refundMoney := payment_gateway.NewMoney(
		fmt.Sprintf("%.2f", refundAmount), // value
		"IDR",                             // currency
	)

	// Build RefundOrderRequest with all required parameters
	refundReq := payment_gateway.NewRefundOrderRequest(
		c.merchantID,                            // merchantId
		transactionID,                           // originalPartnerReferenceNo
		fmt.Sprintf("refund_%s", transactionID), // partnerRefundNo
		*refundMoney,                            // refundAmount
	)

	// Call Dana RefundOrder API
	_, _, err := c.apiClient.PaymentGatewayAPI.
		RefundOrder(ctx).
		RefundOrderRequest(*refundReq).
		Execute()

	if err != nil {
		c.logger.Error("Failed to refund payment with Dana",
			zap.String("transaction_id", transactionID),
			zap.Error(err))
		return fmt.Errorf("failed to refund payment: %w", err)
	}

	c.logger.Debug("Payment refunded successfully",
		zap.String("transaction_id", transactionID))

	return nil
}

// CancelPayment implements payment cancellation using Dana CancelOrder API
func (c *DanaGapuraClient) CancelPayment(
	ctx context.Context,
	transactionID string,
) error {
	c.logger.Debug("Canceling payment",
		zap.String("transaction_id", transactionID))

	// Build CancelOrderRequest
	cancelReq := payment_gateway.NewCancelOrderRequest(
		transactionID, // originalPartnerReferenceNo
		c.merchantID,  // merchantId
	)

	// Call Dana CancelOrder API
	_, _, err := c.apiClient.PaymentGatewayAPI.
		CancelOrder(ctx).
		CancelOrderRequest(*cancelReq).
		Execute()

	if err != nil {
		c.logger.Error("Failed to cancel payment with Dana",
			zap.String("transaction_id", transactionID),
			zap.Error(err))
		return fmt.Errorf("failed to cancel payment: %w", err)
	}

	c.logger.Debug("Payment canceled successfully",
		zap.String("transaction_id", transactionID))

	return nil
}

// mapDanaStatusToLocal maps Dana API status ke local payment status
// Dana SDK returns status values like ACQUIREMENTSTATUS_SUCCESS_, ACQUIREMENTSTATUS_PAYING_, etc.
func mapDanaStatusToLocal(danaStatus string) string {
	switch danaStatus {
	case "00", "ACQUIREMENTSTATUS_SUCCESS_": // Success
		return "completed"
	case "01", "ACQUIREMENTSTATUS_INIT_": // Initiated
		return "pending"
	case "02", "ACQUIREMENTSTATUS_PAYING_": // Paying
		return "pending"
	case "05", "ACQUIREMENTSTATUS_CANCELLED_": // Cancelled
		return "cancelled"
	case "07", "ACQUIREMENTSTATUS_FAILED_": // Not found / Failed
		return "failed"
	case "ACQUIREMENTSTATUS_CLOSED_": // Closed = completed
		return "completed"
	case "ACQUIREMENTSTATUS_MERCHANT_ACCEPT_": // Merchant accepted = completed
		return "completed"
	default:
		return "pending"
	}
}

// mapDanaResponseCodeToStatus maps Dana response code to payment status
func mapDanaResponseCodeToStatus(responseCode string) string {
	// Dana response codes:
	// "00" = Success
	// Other codes = Error/Pending
	if responseCode == "00" {
		return "completed"
	}
	return "pending"
}

// VerifyWebhookSignature verifies webhook signature dari Dana
// Menggunakan webhook.WebhookParser dari Dana SDK
func (c *DanaGapuraClient) VerifyWebhookSignature(
	httpMethod string,
	path string,
	headers map[string]string,
	body string,
) (bool, error) {
	// Note: Implementasi webhook verification menggunakan Dana SDK
	// dapat dilakukan dengan webhook.NewWebhookParser()
	// Ini adalah contoh placeholder untuk integrasi webhook

	c.logger.Debug("Verifying webhook signature",
		zap.String("path", path))

	// TODO: Implement webhook verification menggunakan Dana SDK
	// Contoh:
	// danaPublicKey := os.Getenv("DANA_PUBLIC_KEY")
	// parser, err := webhook.NewWebhookParser(&danaPublicKey, nil)
	// parsedData, err := parser.ParseWebhook(httpMethod, path, headers, body)

	return true, nil
}
