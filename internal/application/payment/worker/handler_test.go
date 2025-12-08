package worker

import (
"context"
"encoding/json"
"errors"
"testing"
"time"

"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/dto"
"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/entity"
"github.com/novriyantoAli/wallet-ms-backend/internal/config"
"github.com/novriyantoAli/wallet-ms-backend/internal/domain/payment/port"
"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

"github.com/hibiken/asynq"
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"
)

type MockPaymentService struct {
mock.Mock
}

func (m *MockPaymentService) CreatePayment(req *dto.CreatePaymentRequest) (*dto.PaymentResponse, error) {
args := m.Called(req)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPaymentByID(id uint) (*dto.PaymentResponse, error) {
args := m.Called(id)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) GetPayments(filter *dto.PaymentFilter) (*dto.PaymentListResponse, error) {
args := m.Called(filter)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*dto.PaymentListResponse), args.Error(1)
}

func (m *MockPaymentService) UpdatePayment(id uint, req *dto.UpdatePaymentRequest) (*dto.PaymentResponse, error) {
args := m.Called(id, req)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*dto.PaymentResponse), args.Error(1)
}

func (m *MockPaymentService) DeletePayment(id uint) error {
args := m.Called(id)
return args.Error(0)
}

func (m *MockPaymentService) GetPaymentsByUser(userID uint) ([]dto.PaymentResponse, error) {
args := m.Called(userID)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).([]dto.PaymentResponse), args.Error(1)
}

type MockAsynqClient struct {
mock.Mock
}

func (m *MockAsynqClient) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
args := m.Called(task, opts)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*asynq.TaskInfo), args.Error(1)
}

// MockPaymentGateway implements port.PaymentGateway interface
type MockPaymentGateway struct {
mock.Mock
}

func (m *MockPaymentGateway) CheckPaymentStatus(ctx context.Context, transactionID string) (*port.PaymentStatusResponse, error) {
args := m.Called(ctx, transactionID)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*port.PaymentStatusResponse), args.Error(1)
}

func (m *MockPaymentGateway) ProcessPayment(ctx context.Context, req *port.ProcessPaymentRequest) (*port.ProcessPaymentResponse, error) {
args := m.Called(ctx, req)
if args.Get(0) == nil {
return nil, args.Error(1)
}
return args.Get(0).(*port.ProcessPaymentResponse), args.Error(1)
}

func setupPaymentWorker() (*PaymentWorker, *MockPaymentService, *MockAsynqClient, *MockPaymentGateway) {
mockService := &MockPaymentService{}
mockClient := &MockAsynqClient{}
mockGateway := &MockPaymentGateway{}
logger := testutil.NewSilentLogger()
cfg := &config.Config{
Worker: config.WorkerConfig{
PaymentCheckInterval: 5 * time.Minute,
RetryMaxAttempts:     3,
},
}

worker := NewPaymentWorker(mockService, mockGateway, mockClient, logger, cfg)

return worker, mockService, mockClient, mockGateway
}

func TestPaymentWorker_HandleCheckPaymentStatus(t *testing.T) {
t.Run("should handle check payment status successfully when status needs update", func(t *testing.T) {
// Setup
worker, mockService, _, mockGateway := setupPaymentWorker()
paymentID := uint(1)
payment := &dto.PaymentResponse{
ID:          paymentID,
Amount:      100000,
Currency:    "IDR",
Status:      entity.PaymentStatusPending.String(),
Description: "Test payment",
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}

payload := CheckPaymentStatusPayload{PaymentID: paymentID}
payloadBytes, _ := json.Marshal(payload)
task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

ctx := context.Background()

// Expectations
mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
mockGateway.On("CheckPaymentStatus", mock.Anything, "1").Return(&port.PaymentStatusResponse{
Status:   "success",
Amount:   100000,
Currency: "IDR",
}, nil)
mockService.On("UpdatePayment", paymentID, mock.Anything).Return(payment, nil)

// Execute
err := worker.HandleCheckPaymentStatus(ctx, task)

// Assert
assert.NoError(t, err)
mockService.AssertExpectations(t)
mockGateway.AssertExpectations(t)
})

t.Run("should return error when payment service fails to get payment", func(t *testing.T) {
// Setup
worker, mockService, _, _ := setupPaymentWorker()
paymentID := uint(1)

payload := CheckPaymentStatusPayload{PaymentID: paymentID}
payloadBytes, _ := json.Marshal(payload)
task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

ctx := context.Background()

// Expectations
mockService.On("GetPaymentByID", paymentID).Return(nil, errors.New("database error"))

// Execute
err := worker.HandleCheckPaymentStatus(ctx, task)

// Assert
assert.Error(t, err)
mockService.AssertExpectations(t)
})

t.Run("should skip if payment already in final state", func(t *testing.T) {
// Setup
worker, mockService, _, _ := setupPaymentWorker()
paymentID := uint(1)
payment := &dto.PaymentResponse{
ID:          paymentID,
Amount:      100000,
Currency:    "IDR",
Status:      entity.PaymentStatusCompleted.String(),
Description: "Test payment",
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}

payload := CheckPaymentStatusPayload{PaymentID: paymentID}
payloadBytes, _ := json.Marshal(payload)
task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

ctx := context.Background()

// Expectations
mockService.On("GetPaymentByID", paymentID).Return(payment, nil)

// Execute
err := worker.HandleCheckPaymentStatus(ctx, task)

// Assert
assert.NoError(t, err)
mockService.AssertExpectations(t)
// Verify UpdatePayment was not called
mockService.AssertNotCalled(t, "UpdatePayment")
})

t.Run("should return error when gateway check fails", func(t *testing.T) {
// Setup
worker, mockService, _, mockGateway := setupPaymentWorker()
paymentID := uint(1)
payment := &dto.PaymentResponse{
ID:          paymentID,
Amount:      100000,
Currency:    "IDR",
Status:      entity.PaymentStatusPending.String(),
Description: "Test payment",
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}

payload := CheckPaymentStatusPayload{PaymentID: paymentID}
payloadBytes, _ := json.Marshal(payload)
task := asynq.NewTask(TypeCheckPaymentStatus, payloadBytes)

ctx := context.Background()

// Expectations
mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
mockGateway.On("CheckPaymentStatus", mock.Anything, "1").Return(nil, errors.New("gateway error"))

// Execute
err := worker.HandleCheckPaymentStatus(ctx, task)

// Assert
assert.Error(t, err)
mockService.AssertExpectations(t)
mockGateway.AssertExpectations(t)
})
}

func TestPaymentWorker_HandleProcessPayment(t *testing.T) {
t.Run("should process payment successfully with success response", func(t *testing.T) {
// Setup
worker, mockService, _, mockGateway := setupPaymentWorker()
paymentID := uint(1)
payment := &dto.PaymentResponse{
ID:          paymentID,
Amount:      100000,
Currency:    "IDR",
Status:      entity.PaymentStatusPending.String(),
Description: "Test payment",
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}

payload := ProcessPaymentPayload{PaymentID: paymentID}
payloadBytes, _ := json.Marshal(payload)
task := asynq.NewTask(TypeProcessPayment, payloadBytes)

ctx := context.Background()

// Expectations
mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
mockGateway.On("ProcessPayment", mock.Anything, mock.Anything).Return(&port.ProcessPaymentResponse{
Status:        "success",
Message:       "Payment processed",
TransactionID: "1",
}, nil)
mockService.On("UpdatePayment", paymentID, mock.Anything).Return(payment, nil)

// Execute
err := worker.HandleProcessPayment(ctx, task)

// Assert
assert.NoError(t, err)
mockService.AssertExpectations(t)
mockGateway.AssertExpectations(t)
})

t.Run("should process payment with failed response", func(t *testing.T) {
// Setup
worker, mockService, _, mockGateway := setupPaymentWorker()
paymentID := uint(1)
payment := &dto.PaymentResponse{
ID:          paymentID,
Amount:      100000,
Currency:    "IDR",
Status:      entity.PaymentStatusPending.String(),
Description: "Test payment",
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}

payload := ProcessPaymentPayload{PaymentID: paymentID}
payloadBytes, _ := json.Marshal(payload)
task := asynq.NewTask(TypeProcessPayment, payloadBytes)

ctx := context.Background()

// Expectations
mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
mockGateway.On("ProcessPayment", mock.Anything, mock.Anything).Return(&port.ProcessPaymentResponse{
Status:        "failed",
Message:       "Payment failed",
TransactionID: "1",
}, nil)
mockService.On("UpdatePayment", paymentID, mock.Anything).Return(payment, nil)

// Execute
err := worker.HandleProcessPayment(ctx, task)

// Assert
assert.NoError(t, err)
mockService.AssertExpectations(t)
mockGateway.AssertExpectations(t)
})

t.Run("should return error when gateway process fails", func(t *testing.T) {
// Setup
worker, mockService, _, mockGateway := setupPaymentWorker()
paymentID := uint(1)
payment := &dto.PaymentResponse{
ID:          paymentID,
Amount:      100000,
Currency:    "IDR",
Status:      entity.PaymentStatusPending.String(),
Description: "Test payment",
CreatedAt:   time.Now(),
UpdatedAt:   time.Now(),
}

payload := ProcessPaymentPayload{PaymentID: paymentID}
payloadBytes, _ := json.Marshal(payload)
task := asynq.NewTask(TypeProcessPayment, payloadBytes)

ctx := context.Background()

// Expectations
mockService.On("GetPaymentByID", paymentID).Return(payment, nil)
mockGateway.On("ProcessPayment", mock.Anything, mock.Anything).Return(nil, errors.New("gateway error"))

// Execute
err := worker.HandleProcessPayment(ctx, task)

// Assert
assert.Error(t, err)
mockService.AssertExpectations(t)
mockGateway.AssertExpectations(t)
})
}

func TestPaymentWorker_MapGatewayStatusToLocal(t *testing.T) {
t.Run("should map success to completed", func(t *testing.T) {
worker, _, _, _ := setupPaymentWorker()
status := worker.mapGatewayStatusToLocal("success")
assert.Equal(t, entity.PaymentStatusCompleted.String(), status)
})

t.Run("should map failed to failed", func(t *testing.T) {
worker, _, _, _ := setupPaymentWorker()
status := worker.mapGatewayStatusToLocal("failed")
assert.Equal(t, entity.PaymentStatusFailed.String(), status)
})

t.Run("should map pending to pending", func(t *testing.T) {
worker, _, _, _ := setupPaymentWorker()
status := worker.mapGatewayStatusToLocal("pending")
assert.Equal(t, entity.PaymentStatusPending.String(), status)
})

t.Run("should map unknown status to pending", func(t *testing.T) {
worker, _, _, _ := setupPaymentWorker()
status := worker.mapGatewayStatusToLocal("unknown")
assert.Equal(t, entity.PaymentStatusPending.String(), status)
})
}

func TestPaymentWorker_SchedulePaymentStatusCheck(t *testing.T) {
t.Run("should schedule payment status check successfully", func(t *testing.T) {
worker, _, mockClient, _ := setupPaymentWorker()
paymentID := uint(1)
delay := 5 * time.Minute

mockClient.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{
ID:    "task-id",
Type:  TypeCheckPaymentStatus,
State: asynq.TaskStatePending,
}, nil)

err := worker.SchedulePaymentStatusCheck(paymentID, delay)

assert.NoError(t, err)
mockClient.AssertExpectations(t)
})

t.Run("should return error when enqueue fails", func(t *testing.T) {
worker, _, mockClient, _ := setupPaymentWorker()
paymentID := uint(1)
delay := 5 * time.Minute

mockClient.On("Enqueue", mock.Anything, mock.Anything).Return(nil, errors.New("enqueue error"))

err := worker.SchedulePaymentStatusCheck(paymentID, delay)

assert.Error(t, err)
mockClient.AssertExpectations(t)
})
}

func TestPaymentWorker_SchedulePaymentProcessing(t *testing.T) {
t.Run("should schedule payment processing successfully", func(t *testing.T) {
worker, _, mockClient, _ := setupPaymentWorker()
paymentID := uint(1)

mockClient.On("Enqueue", mock.Anything, mock.Anything).Return(&asynq.TaskInfo{
ID:    "task-id",
Type:  TypeProcessPayment,
State: asynq.TaskStatePending,
}, nil)

err := worker.SchedulePaymentProcessing(paymentID)

assert.NoError(t, err)
mockClient.AssertExpectations(t)
})

t.Run("should return error when enqueue fails", func(t *testing.T) {
worker, _, mockClient, _ := setupPaymentWorker()
paymentID := uint(1)

mockClient.On("Enqueue", mock.Anything, mock.Anything).Return(nil, errors.New("enqueue error"))

err := worker.SchedulePaymentProcessing(paymentID)

assert.Error(t, err)
mockClient.AssertExpectations(t)
})
}
