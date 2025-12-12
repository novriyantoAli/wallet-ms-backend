package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service service.UserService
	logger  *zap.Logger
}

func NewUserHandler(service service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service: service,
		logger:  logger,
	}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User creation request"
// @Success 201 {object} map[string]interface{} "Created user"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Email already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [post]
func (h *UserHandler) CreateUser(ctx *gin.Context) {
	var req dto.CreateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.CreateUser(&req)
	if err != nil {
		h.logger.Error("Failed to create user", zap.Error(err))
		if err.Error() == "email already exists" {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": user})
}

// GetUser godoc
// @Summary Get a user by ID
// @Description Get a single user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User details"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.service.GetUserByID(uint(id))
	if err != nil {
		h.logger.Error("Failed to get user", zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

// GetUsers godoc
// @Summary Get all users
// @Description Get a list of users with optional filtering and pagination
// @Tags users
// @Accept json
// @Produce json
// @Param name query string false "Filter by name"
// @Param email query string false "Filter by email"
// @Param level query string false "Filter by level (user, reseller, admin)"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Number of items per page" default(10)
// @Success 200 {object} dto.UserListResponse "List of users"
// @Failure 400 {object} map[string]interface{} "Invalid query parameters"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users [get]
func (h *UserHandler) GetUsers(ctx *gin.Context) {
	var filter dto.UserFilter
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		h.logger.Error("Invalid query parameters", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	users, err := h.service.GetUsers(&filter)
	if err != nil {
		h.logger.Error("Failed to get users", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get users"})
		return
	}

	ctx.JSON(http.StatusOK, users)
}

// UpdateUser godoc
// @Summary Update a user
// @Description Update a user's name by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User update request"
// @Success 200 {object} map[string]interface{} "Updated user"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update user", zap.Error(err))
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

// UpdateUserPassword godoc
// @Summary Update user password
// @Description Update a user's password by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param password body dto.UpdateUserPasswordRequest true "Password update request"
// @Success 200 {object} map[string]interface{} "Password updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 401 {object} map[string]interface{} "Current password is incorrect"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{id}/password [put]
func (h *UserHandler) UpdateUserPassword(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req dto.UpdateUserPasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.service.UpdateUserPassword(uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update user password", zap.Error(err))
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "current password is incorrect" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// DeleteUser godoc
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]interface{} "Invalid user ID"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.service.DeleteUser(uint(id))
	if err != nil {
		h.logger.Error("Failed to delete user", zap.Error(err))
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// GetCurrentUser godoc
// @Summary Get current user details
// @Description Get the current authenticated user's details using JWT token
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Current user details"
// @Failure 400 {object} map[string]interface{} "Missing or invalid token"
// @Failure 401 {object} map[string]interface{} "Unauthorized - Invalid token"
// @Failure 404 {object} map[string]interface{} "User not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/me [get]
func (h *UserHandler) GetCurrentUser(ctx *gin.Context) {
	// Extract token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn("Missing authorization header")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		h.logger.Warn("Invalid authorization header format")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		h.logger.Warn("Empty token in authorization header")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	user, err := h.service.GetCurrentUser(token)
	if err != nil {
		h.logger.Error("Failed to get current user", zap.Error(err))
		if err.Error() == "invalid or expired token" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if err.Error() == "user not found" {
			ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": user})
}

// Logout godoc
// @Summary Logout user
// @Description Revoke the user's JWT token using Redis
// @Tags auth
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} map[string]interface{} "Logout successful"
// @Failure 400 {object} map[string]interface{} "Missing or invalid authorization header"
// @Failure 401 {object} map[string]interface{} "Invalid token"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/logout [post]
func (h *UserHandler) Logout(ctx *gin.Context) {
	// Extract token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn("Missing authorization header for logout")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		h.logger.Warn("Invalid authorization header format for logout")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		h.logger.Warn("Empty token in authorization header for logout")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	err := h.service.Logout(ctx, token)
	if err != nil {
		h.logger.Error("Failed to logout user", zap.Error(err))
		if err.Error() == "invalid token" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

func (h *UserHandler) RegisterRoutes(api *gin.RouterGroup) {
	users := api.Group("/users")
	{
		users.POST("", h.CreateUser)
		users.GET("", h.GetUsers)
		users.GET("/:id", h.GetUser)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
		users.PUT("/:id/password", h.UpdateUserPassword)
	}

	auth := api.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.GET("/me", h.GetCurrentUser)
		auth.DELETE("/logout", h.Logout)
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with name, email, password and optional level (user/reseller/admin, defaults to user)
// @Tags auth
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequest true "Registration request"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 409 {object} map[string]interface{} "Email already exists"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/register [post]
func (h *UserHandler) Register(ctx *gin.Context) {
	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid register request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.service.Register(&req)
	if err != nil {
		h.logger.Error("Failed to register user", zap.Error(err))
		if err.Error() == "email already exists" {
			ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": user})
}

// Login godoc
// @Summary Login user
// @Description Authenticate user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login request"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]interface{} "Invalid request body"
// @Failure 401 {object} map[string]interface{} "Invalid email or password"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /auth/login [post]
func (h *UserHandler) Login(ctx *gin.Context) {
	var req dto.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Invalid login request body", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loginResp, err := h.service.Login(&req)
	if err != nil {
		h.logger.Error("Failed to login user", zap.Error(err))
		if err.Error() == "invalid email or password" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to login"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": loginResp})
}
