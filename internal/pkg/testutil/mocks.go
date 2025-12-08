package testutil

import (
	"context"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/entity"
	userDto "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	userEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	walletDto "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	walletEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *userEntity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (*userEntity.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userEntity.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (*userEntity.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userEntity.User), args.Error(1)
}

func (m *MockUserRepository) GetAll(filter *userDto.UserFilter) ([]userEntity.User, int64, error) {
	args := m.Called(filter)
	var users []userEntity.User
	if args.Get(0) != nil {
		users = args.Get(0).([]userEntity.User)
	}

	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return users, count, args.Error(2)
}

func (m *MockUserRepository) Update(user *userEntity.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) EmailExists(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

// MockPaymentRepository is a mock implementation of PaymentRepository
type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(payment *entity.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(id uint) (*entity.Payment, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetAll(filter *dto.PaymentFilter) ([]entity.Payment, int64, error) {
	args := m.Called(filter)
	var payments []entity.Payment
	if args.Get(0) != nil {
		payments = args.Get(0).([]entity.Payment)
	}

	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return payments, count, args.Error(2)
}

func (m *MockPaymentRepository) Update(payment *entity.Payment) error {
	args := m.Called(payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByUserID(userID uint) ([]entity.Payment, error) {
	args := m.Called(userID)
	var payments []entity.Payment
	if args.Get(0) != nil {
		payments = args.Get(0).([]entity.Payment)
	}
	return payments, args.Error(1)
}

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(req *userDto.CreateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uint) (*userDto.UserResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUserByEmail(email string) (*userDto.UserResponse, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUsers(filter *userDto.UserFilter) (*userDto.UserListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserListResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(id uint, req *userDto.UpdateUserRequest) (*userDto.UserResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUserPassword(id uint, req *userDto.UpdateUserPasswordRequest) error {
	args := m.Called(id, req)
	return args.Error(0)
}

func (m *MockUserService) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) Register(req *userDto.RegisterRequest) (*userDto.UserResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) Login(req *userDto.LoginRequest) (*userDto.LoginResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetCurrentUser(token string) (*userDto.UserResponse, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*userDto.UserResponse), args.Error(1)
}

func (m *MockUserService) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

// MockWalletRepository is a mock implementation of WalletRepository
type MockWalletRepository struct {
	mock.Mock
}

func (m *MockWalletRepository) Create(wallet *walletEntity.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) GetByID(id uint) (*walletEntity.Wallet, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletEntity.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetByUserID(userID uint) (*walletEntity.Wallet, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletEntity.Wallet), args.Error(1)
}

func (m *MockWalletRepository) GetAll(filter *walletDto.WalletFilter) ([]walletEntity.Wallet, int64, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]walletEntity.Wallet), args.Get(1).(int64), args.Error(2)
}

func (m *MockWalletRepository) Update(wallet *walletEntity.Wallet) error {
	args := m.Called(wallet)
	return args.Error(0)
}

func (m *MockWalletRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockWalletRepository) UpdateBalance(id uint, amount float64) error {
	args := m.Called(id, amount)
	return args.Error(0)
}

// MockRedisClient is a mock implementation of redis.Client
type MockRedisClient struct {
	mock.Mock
	data map[string]string
}

func (m *MockRedisClient) Get(ctx context.Context, key string) *redis.StringCmd {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*redis.StringCmd)
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	args := m.Called(ctx, key, value, expiration)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*redis.StatusCmd)
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*redis.IntCmd)
}

func (m *MockRedisClient) Ping(ctx context.Context) *redis.StatusCmd {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*redis.StatusCmd)
}

// NewMockRedisClient returns a mock Redis client for testing
// It accepts token revocation operations but doesn't enforce them
func NewMockRedisClient() *redis.Client {
	// For testing, we return nil which will be handled gracefully by the JWT manager
	// The JWT manager now returns nil when Redis client is not configured
	return nil
}

// MockTransactionRepository is a mock implementation of TransactionRepository
type MockTransactionRepository struct {
	mock.Mock
}

func (m *MockTransactionRepository) Create(transaction *walletEntity.WalletTransaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByID(id uint) (*walletEntity.WalletTransaction, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletEntity.WalletTransaction), args.Error(1)
}

func (m *MockTransactionRepository) GetByWalletID(filter *walletDto.TransactionFilter) ([]walletEntity.WalletTransaction, int64, error) {
	args := m.Called(filter)
	var transactions []walletEntity.WalletTransaction
	if args.Get(0) != nil {
		transactions = args.Get(0).([]walletEntity.WalletTransaction)
	}
	var count int64
	if args.Get(1) != nil {
		count = args.Get(1).(int64)
	}
	return transactions, count, args.Error(2)
}

func (m *MockTransactionRepository) Update(transaction *walletEntity.WalletTransaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *MockTransactionRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTransactionRepository) GetByReferenceID(referenceID string) (*walletEntity.WalletTransaction, error) {
	args := m.Called(referenceID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletEntity.WalletTransaction), args.Error(1)
}

// MockWalletService is a mock implementation of WalletService
type MockWalletService struct {
	mock.Mock
}

func (m *MockWalletService) CreateWallet(req *walletDto.CreateWalletRequest) (*walletDto.WalletResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWalletByID(id uint) (*walletDto.WalletResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWalletByUserID(userID uint) (*walletDto.WalletResponse, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.WalletResponse), args.Error(1)
}

func (m *MockWalletService) GetWallets(filter *walletDto.WalletFilter) (*walletDto.WalletListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.WalletListResponse), args.Error(1)
}

func (m *MockWalletService) UpdateWalletBalance(walletID uint, req *walletDto.UpdateWalletBalanceRequest) (*walletDto.WalletBalanceResponse, error) {
	args := m.Called(walletID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.WalletBalanceResponse), args.Error(1)
}

func (m *MockWalletService) DeleteWallet(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockWalletService) TransferFunds(req *walletDto.TransferRequest) (*walletDto.TransferResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.TransferResponse), args.Error(1)
}

func (m *MockWalletService) CreateTransaction(req *walletDto.CreateTransactionRequest) (*walletDto.TransactionResponse, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.TransactionResponse), args.Error(1)
}

func (m *MockWalletService) GetTransactions(filter *walletDto.TransactionFilter) (*walletDto.TransactionListResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.TransactionListResponse), args.Error(1)
}

func (m *MockWalletService) GetTransactionByID(id uint) (*walletDto.TransactionResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*walletDto.TransactionResponse), args.Error(1)
}

// NewMockUserService returns a new mock user service
func NewMockUserService() *MockUserService {
	return &MockUserService{}
}
