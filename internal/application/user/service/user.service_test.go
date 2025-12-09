package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/entity"
	walletEntity "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/jwt"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func newTestJWTManager() *jwt.JWTManager {
	return jwt.NewJWTManager(jwt.JWTConfig{
		SecretKey: "test-secret-key",
		Expiry:    24 * time.Hour,
	})
}

func TestUserService_CreateUser(t *testing.T) {
	t.Run("should create user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(0).(*entity.User)
			user.ID = 1
		})
		mockWalletRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(nil)

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, req.Email, response.Email)
		assert.Equal(t, req.Level, response.Level)
		mockRepo.AssertExpectations(t)
		mockWalletRepo.AssertExpectations(t)
	})

	t.Run("should create user successfully even if wallet creation fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(0).(*entity.User)
			user.ID = 1
		})
		mockWalletRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(errors.New("wallet creation failed"))

		// When
		response, err := service.CreateUser(req)

		// Then - User creation should still succeed even if wallet fails
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		mockRepo.AssertExpectations(t)
		mockWalletRepo.AssertExpectations(t)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(true, nil)

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "email already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when email check fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, errors.New("database error"))

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user creation fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := testutil.CreateUserRequestFixture()

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(errors.New("create failed"))

		// When
		response, err := service.CreateUser(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "create failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	t.Run("should get user by ID successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(1)
		user := testutil.CreateUserFixture()
		user.ID = userID

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(user, nil)

		// When
		response, err := service.GetUserByID(userID)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, user.Name, response.Name)
		assert.Equal(t, user.Email, response.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(999)

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetUserByID(userID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(1)

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, errors.New("database error"))

		// When
		response, err := service.GetUserByID(userID)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByEmail(t *testing.T) {
	t.Run("should get user by email successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		email := "test@example.com"
		user := testutil.CreateUserFixture()
		user.Email = email

		// Mock expectations
		mockRepo.On("GetByEmail", email).Return(user, nil)

		// When
		response, err := service.GetUserByEmail(email)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, email, response.Email)
		assert.Equal(t, user.Name, response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		email := "nonexistent@example.com"

		// Mock expectations
		mockRepo.On("GetByEmail", email).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetUserByEmail(email)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUsers(t *testing.T) {
	t.Run("should get users with pagination successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		filter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		users := []entity.User{
			*testutil.CreateUserFixture(),
			*testutil.CreateUserFixture(),
		}
		users[0].ID = 1
		users[1].ID = 2
		users[1].Email = "user2@example.com"

		// Mock expectations
		mockRepo.On("GetAll", filter).Return(users, int64(2), nil)

		// When
		response, err := service.GetUsers(filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, int64(2), response.TotalCount)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should set default pagination values", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		filter := &dto.UserFilter{
			Page:     0,
			PageSize: 0,
		}

		expectedFilter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", expectedFilter).Return([]entity.User{}, int64(0), nil)

		// When
		response, err := service.GetUsers(filter)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 1, response.Page)
		assert.Equal(t, 10, response.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		filter := &dto.UserFilter{
			Page:     1,
			PageSize: 10,
		}

		// Mock expectations
		mockRepo.On("GetAll", filter).Return(nil, int64(0), errors.New("database error"))

		// When
		response, err := service.GetUsers(filter)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	t.Run("should update user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(1)
		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID

		req := testutil.CreateUpdateUserRequestFixture()

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)

		// When
		response, err := service.UpdateUser(userID, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, req.Name, response.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(999)
		req := testutil.CreateUpdateUserRequestFixture()

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.UpdateUser(userID, req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should update only name field", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(1)
		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Email = "original@example.com"
		originalEmail := existingUser.Email

		req := &dto.UpdateUserRequest{
			Name: "Updated Name",
		}

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)
		mockRepo.On("Update", mock.MatchedBy(func(u *entity.User) bool {
			return u.ID == userID && u.Name == req.Name && u.Email == originalEmail
		})).Return(nil)

		// When
		response, err := service.UpdateUser(userID, req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, originalEmail, response.Email)
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUserPassword(t *testing.T) {
	t.Run("should update user password successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(1)
		currentPassword := "currentpassword"
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)

		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Password = string(hashedPassword)

		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: currentPassword,
			NewPassword:     "newpassword123",
		}

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)
		mockRepo.On("Update", mock.AnythingOfType("*entity.User")).Return(nil)

		// When
		err := service.UpdateUserPassword(userID, req)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(999)
		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: "password",
			NewPassword:     "newpassword",
		}

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.UpdateUserPassword(userID, req)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when current password is incorrect", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		userID := uint(1)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)

		existingUser := testutil.CreateUserFixture()
		existingUser.ID = userID
		existingUser.Password = string(hashedPassword)

		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword123",
		}

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(existingUser, nil)

		// When
		err := service.UpdateUserPassword(userID, req)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "current password is incorrect")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	t.Run("should delete user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		userID := uint(1)
		user := testutil.CreateUserFixture()
		user.ID = userID

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(user, nil)
		mockWalletRepo.On("GetByUserID", userID).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Delete", userID).Return(nil)

		// When
		err := service.DeleteUser(userID)

		// Then
		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockWalletRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		userID := uint(999)

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(nil, gorm.ErrRecordNotFound)

		// When
		err := service.DeleteUser(userID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when delete fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		userID := uint(1)
		user := testutil.CreateUserFixture()
		user.ID = userID

		// Mock expectations
		mockRepo.On("GetByID", userID).Return(user, nil)
		mockWalletRepo.On("GetByUserID", userID).Return(nil, gorm.ErrRecordNotFound)
		mockRepo.On("Delete", userID).Return(errors.New("delete failed"))

		// When
		err := service.DeleteUser(userID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "delete failed")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_entityToResponse(t *testing.T) {
	t.Run("should convert entity to response correctly", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger).(*userService)

		user := testutil.CreateUserFixture()
		user.ID = 1
		user.Name = "Test User"
		user.Email = "test@example.com"
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()

		// When
		response := service.entityToResponse(user)

		// Then
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Name, response.Name)
		assert.Equal(t, user.Email, response.Email)
		assert.Equal(t, user.CreatedAt, response.CreatedAt)
		assert.Equal(t, user.UpdatedAt, response.UpdatedAt)
		// Password should not be included in response (UserResponse doesn't have Password field)
	})
}

func TestUserService_Register(t *testing.T) {
	t.Run("should register user successfully with wallet creation", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := &dto.RegisterRequest{
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Password: "password123",
		}

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(0).(*entity.User)
			user.ID = 1
		})
		mockWalletRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(nil)

		// When
		response, err := service.Register(req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(1), response.ID)
		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, req.Email, response.Email)
		mockRepo.AssertExpectations(t)
		mockWalletRepo.AssertExpectations(t)
	})

	t.Run("should register user successfully even if wallet creation fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := &dto.RegisterRequest{
			Name:     "Jane Doe",
			Email:    "jane.doe@example.com",
			Password: "password123",
		}

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*entity.User")).Return(nil).Run(func(args mock.Arguments) {
			user := args.Get(0).(*entity.User)
			user.ID = 2
		})
		mockWalletRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(errors.New("wallet creation failed"))

		// When
		response, err := service.Register(req)

		// Then - Registration should succeed even if wallet creation fails
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, uint(2), response.ID)
		mockRepo.AssertExpectations(t)
		mockWalletRepo.AssertExpectations(t)
	})

	t.Run("should return error when email already exists on register", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, mockWalletRepo, newTestJWTManager(), logger)

		req := &dto.RegisterRequest{
			Name:     "Jane Doe",
			Email:    "jane@example.com",
			Password: "password123",
		}

		// Mock expectations
		mockRepo.On("EmailExists", req.Email).Return(true, nil)

		// When
		response, err := service.Register(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "email already exists")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Login(t *testing.T) {
	t.Run("should login user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		req := &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		// Create a user with hashed password
		user := testutil.CreateUserFixture()
		user.ID = 1
		user.Email = req.Email
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		user.Password = string(hashedPassword)

		// Mock expectations
		mockRepo.On("GetByEmail", req.Email).Return(user, nil)

		// When
		response, err := service.Login(req)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Name, response.Name)
		assert.Equal(t, user.Email, response.Email)
		assert.NotEmpty(t, response.Token)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user email not found on login", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		req := &dto.LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		// Mock expectations
		mockRepo.On("GetByEmail", req.Email).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.Login(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid email or password")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when password is incorrect", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, newTestJWTManager(), logger)

		req := &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		// Create a user with a different hashed password
		user := testutil.CreateUserFixture()
		user.ID = 1
		user.Email = req.Email
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
		user.Password = string(hashedPassword)

		// Mock expectations
		mockRepo.On("GetByEmail", req.Email).Return(user, nil)

		// When
		response, err := service.Login(req)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid email or password")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetCurrentUser(t *testing.T) {
	t.Run("should get current user successfully with valid token", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		jwtManager := newTestJWTManager()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, jwtManager, logger)

		user := &entity.User{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			Password:  "hashed_password",
			Level:     "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Generate a valid token
		token, err := jwtManager.GenerateToken(user.ID, user.Email, string(user.Level))
		assert.NoError(t, err)

		// Mock expectations
		mockRepo.On("GetByID", user.ID).Return(user, nil)

		// When
		response, err := service.GetCurrentUser(token)

		// Then
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, user.ID, response.ID)
		assert.Equal(t, user.Email, response.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error with invalid token", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		jwtManager := newTestJWTManager()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, jwtManager, logger)

		// When
		response, err := service.GetCurrentUser("invalid-token")

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid or expired token")
		mockRepo.AssertNotCalled(t, "GetByID")
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		jwtManager := newTestJWTManager()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, jwtManager, logger)

		user := &entity.User{
			ID:    999,
			Email: "notfound@example.com",
			Level: "user",
		}

		// Generate a valid token with non-existent user ID
		token, err := jwtManager.GenerateToken(user.ID, user.Email, string(user.Level))
		assert.NoError(t, err)

		// Mock expectations
		mockRepo.On("GetByID", user.ID).Return(nil, gorm.ErrRecordNotFound)

		// When
		response, err := service.GetCurrentUser(token)

		// Then
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "user not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_Logout(t *testing.T) {
	t.Run("should logout user successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		jwtManager := newTestJWTManager()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, jwtManager, logger)

		// Generate a valid token
		token, err := jwtManager.GenerateToken(1, "test@example.com", "user")
		assert.NoError(t, err)

		// When
		err = service.Logout(context.Background(), token)

		// Then
		assert.NoError(t, err)
	})
	t.Run("should return error when token is invalid", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()
		jwtManager := newTestJWTManager()
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, jwtManager, logger)

		// When
		err := service.Logout(context.Background(), "invalid-token")

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("should return error when token is expired", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		logger := testutil.NewSilentLogger()

		// Create JWT manager with very short expiry
		config := jwt.JWTConfig{
			SecretKey: "test-secret-key",
			Expiry:    -time.Second, // Already expired
		}
		jwtManager := jwt.NewJWTManager(config)
		service := NewUserService(mockRepo, &testutil.MockWalletRepository{}, jwtManager, logger)

		// Generate a token that's already expired
		token, err := jwtManager.GenerateToken(1, "test@example.com", "user")
		assert.NoError(t, err)

		// When
		err = service.Logout(context.Background(), token)

		// Then
		assert.Error(t, err)
	})
}

func TestUserService_CreateWalletForUser(t *testing.T) {
	t.Run("should create wallet successfully", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := &userService{
			repo:       mockRepo,
			walletRepo: mockWalletRepo,
			jwtManager: newTestJWTManager(),
			logger:     logger,
		}

		userID := uint(1)

		// Mock expectations
		mockWalletRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(nil).Run(func(args mock.Arguments) {
			wallet := args.Get(0).(*walletEntity.Wallet)
			assert.Equal(t, userID, wallet.UserID)
			assert.Equal(t, 0.0, wallet.Balance)
			assert.Equal(t, "IDR", wallet.Currency)
			assert.Equal(t, walletEntity.WalletStatusActive, wallet.Status)
		})

		// When
		err := service.createWalletForUser(userID)

		// Then
		assert.NoError(t, err)
		mockWalletRepo.AssertExpectations(t)
	})

	t.Run("should return error when wallet creation fails", func(t *testing.T) {
		// Setup
		mockRepo := &testutil.MockUserRepository{}
		mockWalletRepo := &testutil.MockWalletRepository{}
		logger := testutil.NewSilentLogger()
		service := &userService{
			repo:       mockRepo,
			walletRepo: mockWalletRepo,
			jwtManager: newTestJWTManager(),
			logger:     logger,
		}

		userID := uint(1)

		// Mock expectations
		mockWalletRepo.On("Create", mock.AnythingOfType("*entity.Wallet")).Return(errors.New("database error"))

		// When
		err := service.createWalletForUser(userID)

		// Then
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockWalletRepo.AssertExpectations(t)
	})
}
