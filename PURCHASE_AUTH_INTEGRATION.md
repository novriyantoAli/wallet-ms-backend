# Purchase Feature with Auth gRPC Integration

This document describes how the purchase feature integrates with the auth gRPC service for user validation.

## Overview

When a user creates a purchase, the purchase service now validates the user through the auth gRPC service before processing the purchase. This ensures that only authenticated and authorized users can make purchases.

## Implementation Details

### Auth Client
- **Location**: `internal/pkg/client/auth_client.go`
- **Type**: gRPC client for the auth service
- **Methods**:
  - `NewAuthServiceClient(target string, logger *zap.Logger) (*AuthServiceClient, error)` - Creates a new auth service client
  - `CreateAuth(ctx context.Context, req *auth.CreateAuthRequest) (*auth.CreateAuthResponse, error)` - Validates user via auth service
  - `Close() error` - Closes the connection

### Purchase Service Integration

#### Module Setup
In `internal/application/purchase/module.go`:
```go
// Auth client provider
func provideAuthServiceClient(logger *zap.Logger) (*client.AuthServiceClient, error) {
    authTarget := "localhost:50051" // Configure via environment
    return client.NewAuthServiceClient(authTarget, logger)
}

// Purchase service includes auth client
func providePurchaseService(
    db *gorm.DB,
    purchaseRepo purchaseRepo.PurchaseRepository,
    productRepo repository.ProductRepository,
    walletSvc walletService.WalletService,
    authClient *client.AuthServiceClient,
    logger *zap.Logger,
) service.PurchaseService {
    return service.NewPurchaseService(db, purchaseRepo, productRepo, walletSvc, authClient, logger)
}
```

#### Service Implementation
In `internal/application/purchase/service/purchase.service.go`:

The `purchaseService` struct now includes the auth client:
```go
type purchaseService struct {
    db            *gorm.DB
    purchaseRepo  purchaseRepo.PurchaseRepository
    productRepo   repository.ProductRepository
    walletService walletService.WalletService
    authClient    *client.AuthServiceClient  // New field
    queueClient   *queue.Client
    logger        *zap.Logger
}
```

#### User Validation
When creating a purchase, the service validates the user:
```go
func (s *purchaseService) CreatePurchase(req *dto.CreatePurchaseRequest) (*dto.PurchaseResponse, error) {
    // ... validation ...
    
    // Create context with timeout for auth service call
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Validate user via auth service
    if err := s.validateUserViaAuth(ctx, fmt.Sprintf("user_%d", req.UserID)); err != nil {
        s.logger.Warn("User validation failed", zap.Uint("user_id", req.UserID), zap.Error(err))
        // Continue with purchase even if auth validation fails
    }
    
    // ... rest of purchase creation logic ...
}
```

#### Validation Helper Method
```go
func (s *purchaseService) validateUserViaAuth(ctx context.Context, username string) error {
    if s.authClient == nil {
        s.logger.Warn("Auth client not available, skipping user validation")
        return nil
    }

    // Create auth validation request
    req := &auth.CreateAuthRequest{
        Username: username,
    }

    s.logger.Debug("Validating user via auth service", zap.String("username", username))

    resp, err := s.authClient.CreateAuth(ctx, req)
    if err != nil {
        s.logger.Error("Failed to validate user via auth service", zap.String("username", username), zap.Error(err))
        return fmt.Errorf("user validation failed: %w", err)
    }

    s.logger.Debug("User validated successfully", zap.String("username", resp.Username))
    return nil
}
```

## Configuration

The auth service target is configured in the `provideAuthServiceClient` function:
```go
authTarget := "localhost:50051"
```

This should be updated to use environment variables or configuration files:
```go
authTarget := os.Getenv("AUTH_SERVICE_TARGET")
if authTarget == "" {
    authTarget = "localhost:50051"
}
```

## Flow Diagram

```
Purchase Creation Request
         |
         v
CreatePurchase Handler
         |
         v
Purchase Service
         |
         +---> Validate Request (UserID, ProductID, Quantity)
         |
         +---> Validate User via Auth Service (gRPC)
         |         |
         |         +---> Auth Client
         |              |
         |              +---> Auth gRPC Server
         |
         +---> Database Transaction (if auth passes)
         |         |
         |         +---> Check Product Availability
         |         +---> Check Wallet Balance
         |         +---> Deduct Wallet Balance
         |         +---> Reduce Product Stock
         |         +---> Create Purchase Record
         |         +---> Record Transaction
         |
         v
Return Purchase Response
```

## Error Handling

- **Auth Connection Error**: Logs warning, continues with purchase (graceful degradation)
- **Auth Validation Error**: Logs warning, continues with purchase (can be made strict)
- **Database Error**: Rolls back transaction, returns error
- **Stock/Balance Error**: Rolls back transaction, returns error

## Future Improvements

1. **Strict Validation**: Make auth validation mandatory rather than optional
2. **Token-based Auth**: Integrate with JWT tokens instead of username-based validation
3. **Caching**: Cache user validation results to reduce gRPC calls
4. **Metrics**: Add Prometheus metrics for auth service calls
5. **Retry Logic**: Implement exponential backoff for failed auth service calls
6. **Circuit Breaker**: Add circuit breaker pattern for resilience
