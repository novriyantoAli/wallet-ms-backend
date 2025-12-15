package client

import (
	"context"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/api/proto/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthServiceClient provides gRPC client for auth service
type AuthServiceClient struct {
	client auth.AuthServiceClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

// NewAuthServiceClient creates a new auth service client
func NewAuthServiceClient(target string, logger *zap.Logger) (*AuthServiceClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10*1024*1024)),
	)
	if err != nil {
		logger.Error("Failed to connect to auth service", zap.String("target", target), zap.Error(err))
		return nil, err
	}

	logger.Info("Connected to auth service", zap.String("target", target))

	return &AuthServiceClient{
		client: auth.NewAuthServiceClient(conn),
		conn:   conn,
		logger: logger,
	}, nil
}

// CreateAuth creates authentication credentials with radcheck and radreply entries
func (a *AuthServiceClient) CreateAuth(ctx context.Context, req *auth.CreateAuthRequest) (*auth.CreateAuthResponse, error) {
	a.logger.Debug("Creating auth", zap.String("username", req.Username))

	resp, err := a.client.CreateAuth(ctx, req)
	if err != nil {
		a.logger.Error("Failed to create auth", zap.String("username", req.Username), zap.Error(err))
		return nil, err
	}

	a.logger.Debug("Auth created successfully", zap.String("username", resp.Username))
	return resp, nil
}

// Close closes the gRPC connection
func (a *AuthServiceClient) Close() error {
	a.logger.Info("Closing auth service connection")
	return a.conn.Close()
}
