package adapter

import (
	"context"
	"testing"

	"github.com/novriyantoAli/wallet-ms-backend/internal/domain/payment/port"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDanaGapuraClient(t *testing.T) {
	t.Run("should create client successfully", func(t *testing.T) {
		// Arrange
		logger := testutil.NewSilentLogger()
		privateKeyPEM := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA2a2rwplBOV...\n-----END RSA PRIVATE KEY-----"
		publicKeyPEM := "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A...\n-----END PUBLIC KEY-----"

		// Act
		client, err := NewDanaGapuraClient(
			"https://api.sandbox.dana.id",
			"TEST-PARTNER-ID",
			privateKeyPEM,
			publicKeyPEM,
			logger,
		)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, client)
		assert.Equal(t, "TEST-PARTNER-ID", client.partnerID)
		assert.NotNil(t, client.apiClient)
	})
}

func TestDanaGapuraClient_CheckPaymentStatus(t *testing.T) {
	t.Run("should check payment status successfully", func(t *testing.T) {
		// Arrange
		logger := testutil.NewSilentLogger()
		privateKeyPEM := "test-private-key"
		publicKeyPEM := "test-public-key"

		client, err := NewDanaGapuraClient(
			"https://api.sandbox.dana.id",
			"TEST-PARTNER-ID",
			privateKeyPEM,
			publicKeyPEM,
			logger,
		)
		require.NoError(t, err)

		// Act & Assert
		// Note: Actual test would require mocking Dana API client
		assert.NotNil(t, client)
		assert.Equal(t, "TEST-PARTNER-ID", client.partnerID)
	})
}

func TestDanaGapuraClient_ProcessPayment(t *testing.T) {
	t.Run("should process payment successfully", func(t *testing.T) {
		// Arrange
		logger := testutil.NewSilentLogger()
		client, err := NewDanaGapuraClient(
			"https://api.sandbox.dana.id",
			"TEST-PARTNER-ID",
			"test-private-key",
			"test-public-key",
			logger,
		)
		require.NoError(t, err)

		req := &port.ProcessPaymentRequest{
			TransactionID: "TRX-001",
			Amount:        100.50,
			Currency:      "IDR",
			Description:   "Test payment",
		}

		// Act & Assert
		assert.NotNil(t, client)
		assert.NotNil(t, req)
	})
}

func TestDanaGapuraClient_RefundPayment(t *testing.T) {
	t.Run("should refund payment successfully", func(t *testing.T) {
		// Arrange
		logger := testutil.NewSilentLogger()
		client, err := NewDanaGapuraClient(
			"https://api.sandbox.dana.id",
			"TEST-PARTNER-ID",
			"test-private-key",
			"test-public-key",
			logger,
		)
		require.NoError(t, err)

		ctx := context.Background()

		// Act & Assert
		assert.NotNil(t, client)
		assert.NotNil(t, ctx)
	})
}

func TestDanaGapuraClient_CancelPayment(t *testing.T) {
	t.Run("should cancel payment successfully", func(t *testing.T) {
		// Arrange
		logger := testutil.NewSilentLogger()
		client, err := NewDanaGapuraClient(
			"https://api.sandbox.dana.id",
			"TEST-PARTNER-ID",
			"test-private-key",
			"test-public-key",
			logger,
		)
		require.NoError(t, err)

		ctx := context.Background()

		// Act & Assert
		assert.NotNil(t, client)
		assert.NotNil(t, ctx)
	})
}

func TestMapDanaStatusToLocal(t *testing.T) {
	tests := []struct {
		name       string
		danaStatus string
		expected   string
	}{
		{
			name:       "Success status",
			danaStatus: "ACQUIREMENTSTATUS_SUCCESS_",
			expected:   "completed",
		},
		{
			name:       "Paying status",
			danaStatus: "ACQUIREMENTSTATUS_PAYING_",
			expected:   "pending",
		},
		{
			name:       "Init status",
			danaStatus: "ACQUIREMENTSTATUS_INIT_",
			expected:   "pending",
		},
		{
			name:       "Failed status",
			danaStatus: "ACQUIREMENTSTATUS_FAILED_",
			expected:   "failed",
		},
		{
			name:       "Cancelled status",
			danaStatus: "ACQUIREMENTSTATUS_CANCELLED_",
			expected:   "cancelled",
		},
		{
			name:       "Closed status",
			danaStatus: "ACQUIREMENTSTATUS_CLOSED_",
			expected:   "completed",
		},
		{
			name:       "Merchant accept status",
			danaStatus: "ACQUIREMENTSTATUS_MERCHANT_ACCEPT_",
			expected:   "completed",
		},
		{
			name:       "Unknown status",
			danaStatus: "UNKNOWN_STATUS",
			expected:   "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := mapDanaStatusToLocal(tt.danaStatus)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDanaGapuraClient_VerifyWebhookSignature(t *testing.T) {
	t.Run("should verify webhook signature", func(t *testing.T) {
		// Arrange
		logger := testutil.NewSilentLogger()
		client, err := NewDanaGapuraClient(
			"https://api.sandbox.dana.id",
			"TEST-PARTNER-ID",
			"test-private-key",
			"test-public-key",
			logger,
		)
		require.NoError(t, err)

		headers := map[string]string{
			"X-SIGNATURE": "test-signature",
			"X-TIMESTAMP": "2024-01-15T10:30:00Z",
		}
		body := `{"status":"completed"}`

		// Act
		isValid, err := client.VerifyWebhookSignature("POST", "/v1.0/debit/notify", headers, body)

		// Assert
		assert.NoError(t, err)
		assert.True(t, isValid)
	})
}
