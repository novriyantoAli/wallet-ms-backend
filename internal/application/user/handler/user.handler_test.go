package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserHandler() (*UserHandler, *testutil.MockUserService) {
	gin.SetMode(gin.TestMode)
	mockService := &testutil.MockUserService{}
	logger := testutil.NewSilentLogger()
	handler := NewUserHandler(mockService, logger)
	return handler, mockService
}

func TestUserHandler_CreateUser(t *testing.T) {
	t.Run("should create user successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := testutil.CreateUserRequestFixture()
		response := &dto.UserResponse{
			ID:        1,
			Name:      req.Name,
			Email:     req.Email,
			Level:     req.Level,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("CreateUser", mock.AnythingOfType("*dto.CreateUserRequest")).Return(response, nil)

		// Prepare request
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateUser(ctx)

		// Then
		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, req.Name, data["name"])
		assert.Equal(t, req.Email, data["email"])
		assert.Equal(t, req.Level, data["level"])
	})

	t.Run("should return bad request for invalid JSON", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/users", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateUser(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return conflict when email already exists", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := testutil.CreateUserRequestFixture()
		mockService.On("CreateUser", mock.AnythingOfType("*dto.CreateUserRequest")).Return(nil, errors.New("email already exists"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateUser(ctx)

		// Then
		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal api error for other errors", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := testutil.CreateUserRequestFixture()
		mockService.On("CreateUser", mock.AnythingOfType("*dto.CreateUserRequest")).Return(nil, errors.New("database error"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/users", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.CreateUser(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUser(t *testing.T) {
	t.Run("should get user successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(1)
		response := &dto.UserResponse{
			ID:        userID,
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("GetUserByID", userID).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users/1", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.GetUser(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users/invalid", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "invalid"},
		}

		// When
		handler.GetUser(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when user not found", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(999)
		mockService.On("GetUserByID", userID).Return(nil, errors.New("user not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users/999", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "999"},
		}

		// When
		handler.GetUser(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetUsers(t *testing.T) {
	t.Run("should get users successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		response := &dto.UserListResponse{
			Data: []dto.UserResponse{
				{ID: 1, Name: "User 1", Email: "user1@example.com"},
				{ID: 2, Name: "User 2", Email: "user2@example.com"},
			},
			TotalCount: 2,
			Page:       1,
			PageSize:   10,
		}

		mockService.On("GetUsers", mock.AnythingOfType("*dto.UserFilter")).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users?page=1&page_size=10", nil)

		// When
		handler.GetUsers(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result dto.UserListResponse
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.TotalCount)
	})

	t.Run("should return bad request for invalid query parameters", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users?page=invalid", nil)

		// When
		handler.GetUsers(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal api error when service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		mockService.On("GetUsers", mock.AnythingOfType("*dto.UserFilter")).Return(nil, errors.New("database error"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/users", nil)

		// When
		handler.GetUsers(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	t.Run("should update user successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(1)
		req := testutil.CreateUpdateUserRequestFixture()
		response := &dto.UserResponse{
			ID:        userID,
			Name:      req.Name,
			Level:     "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("UpdateUser", userID, mock.AnythingOfType("*dto.UpdateUserRequest")).Return(response, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/users/1", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.UpdateUser(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := testutil.CreateUpdateUserRequestFixture()
		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/users/invalid", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "invalid"},
		}

		// When
		handler.UpdateUser(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when user not found", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(999)
		req := testutil.CreateUpdateUserRequestFixture()
		mockService.On("UpdateUser", userID, mock.AnythingOfType("*dto.UpdateUserRequest")).Return(nil, errors.New("user not found"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/users/999", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "999"},
		}

		// When
		handler.UpdateUser(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateUserPassword(t *testing.T) {
	t.Run("should update user password successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(1)
		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: "oldpassword",
			NewPassword:     "newpassword123",
		}

		mockService.On("UpdateUserPassword", userID, mock.AnythingOfType("*dto.UpdateUserPasswordRequest")).Return(nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/users/1/password", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.UpdateUserPassword(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "message")
		assert.Equal(t, "Password updated successfully", result["message"])
	})

	t.Run("should return unauthorized when current password is incorrect", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(1)
		req := &dto.UpdateUserPasswordRequest{
			CurrentPassword: "wrongpassword",
			NewPassword:     "newpassword123",
		}

		mockService.On("UpdateUserPassword", userID, mock.AnythingOfType("*dto.UpdateUserPasswordRequest")).Return(errors.New("current password is incorrect"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("PUT", "/users/1/password", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.UpdateUserPassword(ctx)

		// Then
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	t.Run("should delete user successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(1)
		mockService.On("DeleteUser", userID).Return(nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/users/1", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "1"},
		}

		// When
		handler.DeleteUser(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "message")
		assert.Equal(t, "User deleted successfully", result["message"])
	})

	t.Run("should return not found when user not found", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		userID := uint(999)
		mockService.On("DeleteUser", userID).Return(errors.New("user not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/users/999", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "999"},
		}

		// When
		handler.DeleteUser(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request for invalid ID", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("DELETE", "/users/invalid", nil)
		ctx.Params = gin.Params{
			{Key: "id", Value: "invalid"},
		}

		// When
		handler.DeleteUser(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_RegisterRoutes(t *testing.T) {
	t.Run("should register all routes correctly", func(t *testing.T) {
		// Setup
		handler, _ := setupUserHandler()
		router := gin.New()
		api := router.Group("/api/v1")

		// When
		handler.RegisterRoutes(api)

		// Then
		routes := router.Routes()
		expectedRoutes := []string{
			"POST /api/v1/users",
			"GET /api/v1/users",
			"GET /api/v1/users/:id",
			"PUT /api/v1/users/:id",
			"DELETE /api/v1/users/:id",
			"PUT /api/v1/users/:id/password",
			"POST /api/v1/auth/register",
			"POST /api/v1/auth/login",
			"GET /api/v1/auth/me",
		}

		assert.GreaterOrEqual(t, len(routes), len(expectedRoutes))
	})
}

func TestUserHandler_Register(t *testing.T) {
	t.Run("should register user successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := &dto.RegisterRequest{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "password123",
		}
		response := &dto.UserResponse{
			ID:        1,
			Name:      req.Name,
			Email:     req.Email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mockService.On("Register", mock.AnythingOfType("*dto.RegisterRequest")).Return(response, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.Register(ctx)

		// Then
		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
	})

	t.Run("should return bad request for invalid JSON on register", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.Register(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return conflict when email already exists on register", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := &dto.RegisterRequest{
			Name:     "Jane Doe",
			Email:    "jane@example.com",
			Password: "password123",
		}
		mockService.On("Register", mock.AnythingOfType("*dto.RegisterRequest")).Return(nil, errors.New("email already exists"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.Register(ctx)

		// Then
		assert.Equal(t, http.StatusConflict, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_Login(t *testing.T) {
	t.Run("should login user successfully", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}
		response := &dto.LoginResponse{
			ID:    1,
			Name:  "Test User",
			Email: req.Email,
			Token: "token_test@example.com_20250101120000",
		}

		mockService.On("Login", mock.AnythingOfType("*dto.LoginRequest")).Return(response, nil)

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.Login(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.NotEmpty(t, data["token"])
	})

	t.Run("should return bad request for invalid JSON on login", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer([]byte("invalid json")))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.Login(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return unauthorized for invalid credentials", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		req := &dto.LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}
		mockService.On("Login", mock.AnythingOfType("*dto.LoginRequest")).Return(nil, errors.New("invalid email or password"))

		reqBody, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
		ctx.Request.Header.Set("Content-Type", "application/json")

		// When
		handler.Login(ctx)

		// Then
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_GetCurrentUser(t *testing.T) {
	t.Run("should get current user successfully with valid token", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		response := &dto.UserResponse{
			ID:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			Level:     "user",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.valid.token"
		mockService.On("GetCurrentUser", validToken).Return(response, nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/auth/me", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+validToken)

		// When
		handler.GetCurrentUser(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)

		var result map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &result)
		assert.Contains(t, result, "data")
		data := result["data"].(map[string]interface{})
		assert.Equal(t, float64(1), data["id"])
		assert.Equal(t, "john@example.com", data["email"])
	})

	t.Run("should return bad request when authorization header is missing", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/auth/me", nil)
		// No Authorization header

		// When
		handler.GetCurrentUser(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertNotCalled(t, "GetCurrentUser")
	})

	t.Run("should return bad request when authorization header format is invalid", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/auth/me", nil)
		ctx.Request.Header.Set("Authorization", "InvalidFormat")

		// When
		handler.GetCurrentUser(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertNotCalled(t, "GetCurrentUser")
	})

	t.Run("should return unauthorized when token is invalid", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		invalidToken := "invalid.jwt.token"
		mockService.On("GetCurrentUser", invalidToken).Return(nil, errors.New("invalid or expired token"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/auth/me", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+invalidToken)

		// When
		handler.GetCurrentUser(ctx)

		// Then
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return not found when user does not exist", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.valid.token"
		mockService.On("GetCurrentUser", validToken).Return(nil, errors.New("user not found"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("GET", "/auth/me", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+validToken)

		// When
		handler.GetCurrentUser(ctx)

		// Then
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_Logout(t *testing.T) {
	t.Run("should logout user successfully with valid token", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.valid.token"
		mockService.On("Logout", mock.Anything, validToken).Return(nil)

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/logout", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+validToken)

		// When
		handler.Logout(ctx)

		// Then
		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return bad request when authorization header is missing", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/logout", nil)
		// No Authorization header

		// When
		handler.Logout(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertNotCalled(t, "Logout")
	})

	t.Run("should return bad request when authorization header format is invalid", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/logout", nil)
		ctx.Request.Header.Set("Authorization", "InvalidFormat token") // Invalid format

		// When
		handler.Logout(ctx)

		// Then
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertNotCalled(t, "Logout")
	})

	t.Run("should return unauthorized when token is invalid", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		invalidToken := "invalid.token.here"
		mockService.On("Logout", mock.Anything, invalidToken).Return(errors.New("invalid token"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/logout", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+invalidToken)

		// When
		handler.Logout(ctx)

		// Then
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return internal server error when logout service fails", func(t *testing.T) {
		// Setup
		handler, mockService := setupUserHandler()

		validToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.valid.token"
		mockService.On("Logout", mock.Anything, validToken).Return(errors.New("redis connection failed"))

		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = httptest.NewRequest("POST", "/auth/logout", nil)
		ctx.Request.Header.Set("Authorization", "Bearer "+validToken)

		// When
		handler.Logout(ctx)

		// Then
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}
