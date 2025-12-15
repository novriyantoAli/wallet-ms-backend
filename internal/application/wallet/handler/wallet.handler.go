package handler

import (
	"net/http"
	"strconv"
	"strings"

	userservice "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WalletHandler struct {
	service     service.WalletService
	userService userservice.UserService
	logger      *zap.Logger
}

func NewWalletHandler(
	service service.WalletService,
	userService userservice.UserService,
	logger *zap.Logger,
) *WalletHandler {
	return &WalletHandler{
		service:     service,
		userService: userService,
		logger:      logger,
	}
}

// CreateWallet godoc
// @Summary Create a new wallet for a user
// @Description Create a new wallet with initial balance for a user
// @Tags wallet
// @Accept json
// @Produce json
// @Param request body dto.CreateWalletRequest true "Create wallet request"
// @Success 201 {object} dto.WalletResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/wallets [post]
func (h *WalletHandler) CreateWallet(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.service.CreateWallet(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create wallet", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetWallet godoc
// @Summary Get wallet by ID
// @Description Retrieve wallet information by wallet ID
// @Tags wallet
// @Accept json
// @Produce json
// @Param id path int true "Wallet ID"
// @Success 200 {object} dto.WalletResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/wallets/{id} [get]
func (h *WalletHandler) GetWallet(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	resp, err := h.service.GetWalletByID(ctx, uint(id))
	if err != nil {
		h.logger.Error("Failed to get wallet", zap.Uint("id", uint(id)), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetUserWallet godoc
// @Summary Get wallet by user ID
// @Description Retrieve wallet information for a specific user
// @Tags wallet
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} dto.WalletResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/users/{id}/wallet [get]
func (h *WalletHandler) GetUserWallet(c *gin.Context) {
	ctx := c.Request.Context()
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	resp, err := h.service.GetWalletByUserID(ctx, uint(userID))
	if err != nil {
		h.logger.Error("Failed to get user wallet", zap.Uint("user_id", uint(userID)), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListWallets godoc
// @Summary List wallets with pagination
// @Description Get a paginated list of wallets with optional filtering
// @Tags wallet
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param user_id query int false "Filter by user ID"
// @Param status query string false "Filter by status"
// @Success 200 {object} dto.WalletListResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/wallets [get]
func (h *WalletHandler) ListWallets(c *gin.Context) {
	ctx := c.Request.Context()
	filter := &dto.WalletFilter{}

	if page := c.Query("page"); page != "" {
		p, err := strconv.Atoi(page)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page parameter"})
			return
		}
		filter.Page = p
	}

	if pageSize := c.Query("page_size"); pageSize != "" {
		ps, err := strconv.Atoi(pageSize)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid page_size parameter"})
			return
		}
		filter.PageSize = ps
	}

	if userID := c.Query("user_id"); userID != "" {
		uid, err := strconv.ParseUint(userID, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id parameter"})
			return
		}
		filter.UserID = uint(uid)
	}

	if status := c.Query("status"); status != "" {
		filter.Status = status
	}

	resp, err := h.service.GetWallets(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to list wallets", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateWalletBalance godoc
// @Summary Update wallet balance
// @Description Add or deduct amount from wallet balance
// @Tags wallet
// @Accept json
// @Produce json
// @Param id path int true "Wallet ID"
// @Param request body dto.UpdateWalletBalanceRequest true "Update balance request"
// @Success 200 {object} dto.WalletBalanceResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/wallets/{id}/balance [post]
func (h *WalletHandler) UpdateWalletBalance(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	var req dto.UpdateWalletBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.service.UpdateWalletBalance(ctx, uint(id), &req)
	if err != nil {
		h.logger.Error("Failed to update wallet balance", zap.Uint("id", uint(id)), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteWallet godoc
// @Summary Delete wallet
// @Description Delete a wallet (soft delete)
// @Tags wallet
// @Accept json
// @Produce json
// @Param id path int true "Wallet ID"
// @Success 204 {object} nil
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/wallets/{id} [delete]
func (h *WalletHandler) DeleteWallet(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet ID"})
		return
	}

	err = h.service.DeleteWallet(ctx, uint(id))
	if err != nil {
		h.logger.Error("Failed to delete wallet", zap.Uint("id", uint(id)), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet deleted successfully"})
}

// Transfer godoc
// @Summary Transfer funds between users with level validation
// @Description Transfer funds from sender to recipient user with level-based access control (only reseller and admin can transfer)
// @Tags wallet
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body dto.TransferWalletRequest true "Transfer wallet request with recipient user ID and amount"
// @Success 200 {object} dto.TransferResponse "Transfer successful"
// @Failure 400 {object} map[string]interface{} "Invalid request body or insufficient balance"
// @Failure 401 {object} map[string]interface{} "Unauthorized or insufficient access level"
// @Failure 404 {object} map[string]interface{} "User or wallet not found"
// @Router /api/v1/wallets/transfer-to-user [post]
func (h *WalletHandler) Transfer(c *gin.Context) {
	ctx := c.Request.Context()
	// Extract token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		h.logger.Warn("Missing authorization header for transfer")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract token from "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		h.logger.Warn("Invalid authorization header format for transfer")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	token := authHeader[len(bearerPrefix):]
	if token == "" {
		h.logger.Warn("Empty token in authorization header for transfer")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization header format"})
		return
	}

	// Get current user from token to extract sender ID
	currentUser, err := h.userService.GetCurrentUser(token)
	if err != nil {
		h.logger.Error("Failed to get current user from token", zap.Error(err))
		if err.Error() == "invalid or expired token" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to authenticate user"})
		return
	}

	senderID := currentUser.ID

	// Bind request body
	var req dto.TransferWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind transfer request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Call the transfer service
	resp, err := h.service.Transfer(ctx, senderID, &req)
	if err != nil {
		h.logger.Error("Failed to transfer funds", zap.String("error", err.Error()), zap.Uint("sender_id", senderID), zap.Uint("recipient_id", req.RecipientUserID))

		// Return appropriate error status codes
		switch err.Error() {
		case "users cannot transfer funds":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "sender user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "recipient user not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "sender wallet not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "recipient wallet not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "insufficient balance for transfer":
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	h.logger.Info("Transfer completed successfully", zap.Uint("sender_id", senderID), zap.Uint("recipient_id", req.RecipientUserID), zap.Float64("amount", req.Amount))
	c.JSON(http.StatusOK, resp)
}

// CreateTransaction godoc
// @Summary Create a wallet transaction
// @Description Create a transaction (deposit, withdrawal, payment, etc.) on a wallet
// @Tags wallet-transactions
// @Accept json
// @Produce json
// @Param request body dto.CreateTransactionRequest true "Create transaction request"
// @Success 201 {object} dto.TransactionResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/wallets/transactions [post]
func (h *WalletHandler) CreateTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("Failed to bind request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	resp, err := h.service.CreateTransaction(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create transaction", zap.Error(err))
		if err.Error() == "wallet not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetTransactions godoc
// @Summary Get wallet transactions
// @Description Get paginated list of transactions for a wallet with optional filters
// @Tags wallet-transactions
// @Accept json
// @Produce json
// @Param wallet_id query uint true "Wallet ID"
// @Param type query string false "Transaction type filter"
// @Param status query string false "Transaction status filter"
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} dto.TransactionListResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/wallets/transactions [get]
func (h *WalletHandler) GetTransactions(c *gin.Context) {
	ctx := c.Request.Context()
	var req dto.GetTransactionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Error("Failed to bind query", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	if req.WalletID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet_id is required"})
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	filter := &dto.TransactionFilter{
		WalletID: req.WalletID,
		Type:     req.Type,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	resp, err := h.service.GetTransactions(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to get transactions", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetTransaction godoc
// @Summary Get transaction by ID
// @Description Get details of a specific transaction
// @Tags wallet-transactions
// @Accept json
// @Produce json
// @Param id path uint true "Transaction ID"
// @Success 200 {object} dto.TransactionResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/wallets/transactions/{id} [get]
func (h *WalletHandler) GetTransaction(c *gin.Context) {
	ctx := c.Request.Context()
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Error("Invalid transaction ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction ID"})
		return
	}

	resp, err := h.service.GetTransactionByID(ctx, uint(id))
	if err != nil {
		h.logger.Error("Failed to get transaction", zap.Error(err))
		if err.Error() == "transaction not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes registers all wallet routes
func (h *WalletHandler) RegisterRoutes(api *gin.RouterGroup) {
	wallets := api.Group("/wallets")
	{
		wallets.POST("", h.CreateWallet)
		wallets.GET("", h.ListWallets)
		wallets.GET("/:id", h.GetWallet)
		wallets.POST("/:id/balance", h.UpdateWalletBalance)
		wallets.DELETE("/:id", h.DeleteWallet)
		wallets.POST("/transfer-to-user", h.Transfer)

		// Transaction routes
		wallets.POST("/transactions", h.CreateTransaction)
		wallets.GET("/transactions", h.GetTransactions)
		wallets.GET("/transactions/:id", h.GetTransaction)
	}

	users := api.Group("/users")
	{
		users.GET("/:id/wallet", h.GetUserWallet)
	}
}
