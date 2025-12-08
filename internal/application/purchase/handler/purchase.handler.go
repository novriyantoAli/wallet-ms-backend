package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/service"
	"go.uber.org/zap"
)

type PurchaseHandler struct {
	service service.PurchaseService
	logger  *zap.Logger
}

func NewPurchaseHandler(svc service.PurchaseService, logger *zap.Logger) *PurchaseHandler {
	return &PurchaseHandler{
		service: svc,
		logger:  logger,
	}
}

// CreatePurchase godoc
// @Summary Create a new purchase
// @Description Create a new purchase with wallet deduction and stock reduction
// @Tags purchases
// @Accept json
// @Produce json
// @Param request body dto.CreatePurchaseRequest true "Purchase request"
// @Success 201 {object} dto.PurchaseResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/purchases [post]
func (h *PurchaseHandler) CreatePurchase(c *gin.Context) {
	var req dto.CreatePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purchase, err := h.service.CreatePurchase(&req)
	if err != nil {
		h.logger.Warn("Failed to create purchase", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("Purchase created", zap.Uint("id", purchase.ID))
	c.JSON(http.StatusCreated, purchase)
}

// GetPurchase godoc
// @Summary Get purchase by ID
// @Description Get a specific purchase by ID
// @Tags purchases
// @Produce json
// @Param id path int true "Purchase ID"
// @Success 200 {object} dto.PurchaseResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/purchases/{id} [get]
func (h *PurchaseHandler) GetPurchase(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase id"})
		return
	}

	purchase, err := h.service.GetPurchaseByID(uint(id))
	if err != nil {
		h.logger.Debug("Purchase not found", zap.Uint("id", uint(id)))
		c.JSON(http.StatusNotFound, gin.H{"error": "purchase not found"})
		return
	}

	c.JSON(http.StatusOK, purchase)
}

// GetUserPurchases godoc
// @Summary Get user purchases
// @Description Get all purchases for a specific user with pagination
// @Tags purchases
// @Produce json
// @Param user_id path int true "User ID"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Param status query string false "Filter by status (pending/completed/failed)"
// @Success 200 {object} dto.PurchaseListResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/purchases/user/{user_id} [get]
func (h *PurchaseHandler) GetUserPurchases(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	page := 1
	limit := 10
	if p := c.Query("page"); p != "" {
		if pageNum, err := strconv.Atoi(p); err == nil && pageNum > 0 {
			page = pageNum
		}
	}
	if l := c.Query("limit"); l != "" {
		if limitNum, err := strconv.Atoi(l); err == nil && limitNum > 0 {
			limit = limitNum
		}
	}

	filter := &dto.PurchaseFilter{
		Status: c.Query("status"),
		Page:   page,
		Limit:  limit,
	}

	purchases, err := h.service.GetUserPurchases(uint(userID), filter)
	if err != nil {
		h.logger.Error("Failed to get purchases", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get purchases"})
		return
	}

	c.JSON(http.StatusOK, purchases)
}

// UpdatePurchaseStatus godoc
// @Summary Update purchase status
// @Description Update the status of a purchase
// @Tags purchases
// @Accept json
// @Produce json
// @Param id path int true "Purchase ID"
// @Param request body dto.UpdatePurchaseStatusRequest true "Status update"
// @Success 200 {object} dto.PurchaseResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/purchases/{id}/status [put]
func (h *PurchaseHandler) UpdatePurchaseStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase id"})
		return
	}

	var req dto.UpdatePurchaseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purchase, err := h.service.UpdatePurchaseStatus(uint(id), req.Status)
	if err != nil {
		h.logger.Warn("Failed to update purchase status", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, purchase)
}

// RegisterRoutes registers purchase routes
func (h *PurchaseHandler) RegisterRoutes(api *gin.RouterGroup) {
	purchases := api.Group("/purchases")
	{
		purchases.POST("", h.CreatePurchase)
		purchases.GET("/:id", h.GetPurchase)
		purchases.PUT("/:id/status", h.UpdatePurchaseStatus)
		purchases.GET("/user/:user_id", h.GetUserPurchases)
	}
}
