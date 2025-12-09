package service

import (
	"context"
	"errors"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/repository"
	walletEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	walletrepo "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/jwt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService interface {
	Register(req *dto.RegisterRequest) (*dto.UserResponse, error)
	Login(req *dto.LoginRequest) (*dto.LoginResponse, error)
	GetCurrentUser(token string) (*dto.UserResponse, error)
	CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error)
	GetUserByID(id uint) (*dto.UserResponse, error)
	GetUserByEmail(email string) (*dto.UserResponse, error)
	GetUsers(filter *dto.UserFilter) (*dto.UserListResponse, error)
	UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error)
	UpdateUserPassword(id uint, req *dto.UpdateUserPasswordRequest) error
	DeleteUser(id uint) error
	Logout(ctx context.Context, token string) error
}

type userService struct {
	repo       repository.UserRepository
	walletRepo walletrepo.WalletRepository
	jwtManager *jwt.JWTManager
	logger     *zap.Logger
}

func NewUserService(repo repository.UserRepository, walletRepo walletrepo.WalletRepository, jwtManager *jwt.JWTManager, logger *zap.Logger) UserService {
	return &userService{
		repo:       repo,
		walletRepo: walletRepo,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

func (s *userService) Register(req *dto.RegisterRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.EmailExists(req.Email)
	if err != nil {
		s.logger.Error("Failed to check email existence", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	user := &entity.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.Create(user)
	if err != nil {
		s.logger.Error("Failed to register user", zap.Error(err))
		return nil, err
	}

	// Create wallet for user automatically
	if err := s.createWalletForUser(user.ID); err != nil {
		s.logger.Error("Failed to create wallet for user", zap.Error(err), zap.Uint("user_id", user.ID))
		// Don't fail registration if wallet creation fails, but log the error
	}

	s.logger.Info("User registered successfully", zap.String("email", user.Email))

	return s.entityToResponse(user), nil
}

func (s *userService) Login(req *dto.LoginRequest) (*dto.LoginResponse, error) {
	user, err := s.repo.GetByEmail(req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Login attempt with non-existent email", zap.String("email", req.Email))
			return nil, errors.New("invalid email or password")
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		s.logger.Warn("Login attempt with incorrect password", zap.String("email", req.Email))
		return nil, errors.New("invalid email or password")
	}

	s.logger.Info("User logged in successfully", zap.String("email", user.Email), zap.Uint("user_id", user.ID))

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Email, string(user.Level))
	if err != nil {
		s.logger.Error("Failed to generate JWT token", zap.Error(err))
		return nil, err
	}

	return &dto.LoginResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Level: string(user.Level),
		Token: token,
	}, nil
}

func (s *userService) GetCurrentUser(token string) (*dto.UserResponse, error) {
	// Verify and extract claims from JWT token
	claims, err := s.jwtManager.VerifyToken(token)
	if err != nil {
		s.logger.Warn("Invalid token provided", zap.Error(err))
		return nil, errors.New("invalid or expired token")
	}

	// Retrieve user by ID from claims
	user, err := s.repo.GetByID(claims.UserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found for token", zap.Uint("user_id", claims.UserID))
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to retrieve user", zap.Error(err))
		return nil, err
	}

	s.logger.Info("Current user retrieved", zap.String("email", user.Email), zap.Uint("user_id", user.ID))

	return s.entityToResponse(user), nil
}

func (s *userService) CreateUser(req *dto.CreateUserRequest) (*dto.UserResponse, error) {
	exists, err := s.repo.EmailExists(req.Email)
	if err != nil {
		s.logger.Error("Failed to check email existence", zap.Error(err))
		return nil, err
	}
	if exists {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, err
	}

	user := &entity.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  string(hashedPassword),
		Level:     entity.UserLevel(req.Level),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.Create(user)
	if err != nil {
		s.logger.Error("Failed to create user", zap.Error(err))
		return nil, err
	}

	// Create wallet for user automatically
	if err := s.createWalletForUser(user.ID); err != nil {
		s.logger.Error("Failed to create wallet for user", zap.Error(err), zap.Uint("user_id", user.ID))
		// Don't fail user creation if wallet creation fails, but log the error
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByID(id uint) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUserByEmail(email string) (*dto.UserResponse, error) {
	user, err := s.repo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) GetUsers(filter *dto.UserFilter) (*dto.UserListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	users, totalCount, err := s.repo.GetAll(filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, *s.entityToResponse(&user))
	}

	return &dto.UserListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *userService) UpdateUser(id uint, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	user.Name = req.Name
	user.UpdatedAt = time.Now()

	err = s.repo.Update(user)
	if err != nil {
		s.logger.Error("Failed to update user", zap.Error(err))
		return nil, err
	}

	return s.entityToResponse(user), nil
}

func (s *userService) UpdateUserPassword(id uint, req *dto.UpdateUserPasswordRequest) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash new password", zap.Error(err))
		return err
	}

	user.Password = string(hashedPassword)
	user.UpdatedAt = time.Now()

	return s.repo.Update(user)
}

func (s *userService) DeleteUser(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Delete user's wallet if exists
	wallet, err := s.walletRepo.GetByUserID(id)
	if err == nil && wallet != nil {
		if err := s.walletRepo.Delete(wallet.ID); err != nil {
			s.logger.Error("Failed to delete user's wallet", zap.Uint("user_id", id), zap.Uint("wallet_id", wallet.ID), zap.Error(err))
			// Continue with user deletion even if wallet deletion fails
		}
	}

	return s.repo.Delete(id)
}

// createWalletForUser creates a wallet for a newly registered user
func (s *userService) createWalletForUser(userID uint) error {
	wallet := &walletEntity.Wallet{
		UserID:    userID,
		Balance:   0,
		Currency:  "IDR",
		Status:    walletEntity.WalletStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.walletRepo.Create(wallet); err != nil {
		return err
	}

	s.logger.Info("Wallet created for user", zap.Uint("user_id", userID))
	return nil
}

func (s *userService) Logout(ctx context.Context, token string) error {
	// Verify token to get expiration time
	claims, err := s.jwtManager.VerifyToken(token)
	if err != nil {
		s.logger.Warn("Failed to verify token for logout", zap.Error(err))
		return errors.New("invalid token")
	}

	// Revoke the token using Redis
	expirationTime := claims.ExpiresAt.Time
	if err := s.jwtManager.RevokeToken(ctx, token, expirationTime); err != nil {
		s.logger.Error("Failed to revoke token", zap.Error(err))
		return errors.New("failed to revoke token")
	}

	s.logger.Info("Token revoked successfully", zap.Uint("user_id", claims.UserID))
	return nil
}

func (s *userService) entityToResponse(user *entity.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Level:     string(user.Level),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
