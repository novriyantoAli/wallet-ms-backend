package handler

import (
	"net/http"
	"strconv"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type WiFiProductHandler struct {
	service service.WiFiProductService
	logger  *zap.Logger
}

func NewWiFiProductHandler(svc service.WiFiProductService, logger *zap.Logger) *WiFiProductHandler {
	return &WiFiProductHandler{
		service: svc,
		logger:  logger,
	}
}

// CreateWiFiProduct godoc
// @Summary Create WiFi product
// @Description Create a new WiFi product with quota, duration, and speed limit
// @Tags WiFi Product
// @Accept json
// @Produce json
// @Param request body dto.CreateWiFiProductRequest true "WiFi product request"
// @Success 201 {object} dto.WiFiProductResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/products/wifi [post]
func (h *WiFiProductHandler) CreateWiFiProduct(c *gin.Context) {
	var req dto.CreateWiFiProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create wifi product request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.service.CreateWiFiProduct(&req)
	if err != nil {
		h.logger.Warn("Failed to create wifi product", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("WiFi product created", zap.Uint("id", resp.ID))
	c.JSON(http.StatusCreated, resp)
}

// GetWiFiProduct godoc
// @Summary Get WiFi product by ID
// @Description Get WiFi product details by ID
// @Tags WiFi Product
// @Produce json
// @Param id path int true "WiFi Product ID"
// @Success 200 {object} dto.WiFiProductResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/products/wifi/{id} [get]
func (h *WiFiProductHandler) GetWiFiProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid wifi product ID", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wifi product id"})
		return
	}

	resp, err := h.service.GetWiFiProductByID(uint(id))
	if err != nil {
		h.logger.Debug("WiFi product not found", zap.Uint("id", uint(id)))
		c.JSON(http.StatusNotFound, gin.H{"error": "wifi product not found"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetWiFiProductByProductID godoc
// @Summary Get WiFi product by Product ID
// @Description Get WiFi product by associated product ID
// @Tags WiFi Product
// @Produce json
// @Param product_id query int true "Product ID"
// @Success 200 {object} dto.WiFiProductResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/products/wifi/product/{product_id} [get]
func (h *WiFiProductHandler) GetWiFiProductByProductID(c *gin.Context) {
	productIDStr := c.Param("product_id")
	productID, err := strconv.ParseUint(productIDStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID", zap.String("product_id", productIDStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product_id"})
		return
	}

	resp, err := h.service.GetWiFiProductByProductID(uint(productID))
	if err != nil {
		h.logger.Debug("WiFi product not found", zap.Uint("product_id", uint(productID)))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListWiFiProducts godoc
// @Summary List WiFi products
// @Description Get list of WiFi products with pagination
// @Tags WiFi Product
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Success 200 {object} dto.WiFiProductListResponse
// @Router /api/v1/products/wifi [get]
func (h *WiFiProductHandler) ListWiFiProducts(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	filter := &dto.WiFiProductFilter{
		Page:  page,
		Limit: limit,
	}

	resp, err := h.service.GetWiFiProducts(filter)
	if err != nil {
		h.logger.Error("Failed to list wifi products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list wifi products"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateWiFiProduct godoc
// @Summary Update WiFi product
// @Description Update WiFi product details
// @Tags WiFi Product
// @Accept json
// @Produce json
// @Param id path int true "WiFi Product ID"
// @Param request body dto.UpdateWiFiProductRequest true "Update request"
// @Success 200 {object} dto.WiFiProductResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/products/wifi/{id} [put]
func (h *WiFiProductHandler) UpdateWiFiProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid wifi product ID", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wifi product id"})
		return
	}

	var req dto.UpdateWiFiProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update wifi product request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.service.UpdateWiFiProduct(uint(id), &req)
	if err != nil {
		if err.Error() == "wifi product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wifi product not found"})
			return
		}
		h.logger.Warn("Failed to update wifi product", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("WiFi product updated", zap.Uint("id", uint(id)))
	c.JSON(http.StatusOK, resp)
}

// DeleteWiFiProduct godoc
// @Summary Delete WiFi product
// @Description Delete WiFi product
// @Tags WiFi Product
// @Param id path int true "WiFi Product ID"
// @Success 204
// @Failure 404 {object} map[string]string
// @Router /api/v1/products/wifi/{id} [delete]
func (h *WiFiProductHandler) DeleteWiFiProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid wifi product ID", zap.String("id", idStr))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wifi product id"})
		return
	}

	err = h.service.DeleteWiFiProduct(uint(id))
	if err != nil {
		if err.Error() == "wifi product not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "wifi product not found"})
			return
		}
		h.logger.Error("Failed to delete wifi product", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete wifi product"})
		return
	}

	h.logger.Info("WiFi product deleted", zap.Uint("id", uint(id)))
	c.JSON(http.StatusNoContent, nil)
}

// RegisterRoutes registers WiFi product routes
func (h *WiFiProductHandler) RegisterRoutes(api *gin.RouterGroup) {
	wifi := api.Group("/products/wifi")
	{
		wifi.POST("", h.CreateWiFiProduct)
		wifi.GET("", h.ListWiFiProducts)
		// Static route must come before parameterized routes
		// Use GET with explicit "product" path component
		wifi.GET("/product/:product_id", h.GetWiFiProductByProductID)
		// Generic parameterized routes after specific ones
		wifi.GET("/:id", h.GetWiFiProduct)
		wifi.PUT("/:id", h.UpdateWiFiProduct)
		wifi.DELETE("/:id", h.DeleteWiFiProduct)
	}
}
