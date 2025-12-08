package handler

import (
	"net/http"
	"strconv"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service     service.ProductService
	wifiService service.WiFiProductService
	wifiHandler *WiFiProductHandler
	logger      *zap.Logger
}

func NewProductHandler(service service.ProductService, wifiService service.WiFiProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service:     service,
		wifiService: wifiService,
		wifiHandler: NewWiFiProductHandler(wifiService, logger),
		logger:      logger,
	}
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with the provided information
// @Tags products
// @Accept json
// @Produce json
// @Param request body dto.CreateProductRequest true "Create product request"
// @Success 201 {object} dto.ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid create product request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateProduct(&req)
	if err != nil {
		h.logger.Error("Failed to create product", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetProduct godoc
// @Summary Get product by ID
// @Description Retrieve product information by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} dto.ProductResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/products/{id} [get]
func (h *ProductHandler) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	resp, err := h.service.GetProductByID(uint(id))
	if err != nil {
		h.logger.Warn("Product not found", zap.Uint("id", uint(id)), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetProductBySKU godoc
// @Summary Get product by SKU
// @Description Retrieve product information by SKU
// @Tags products
// @Accept json
// @Produce json
// @Param sku query string true "Product SKU"
// @Success 200 {object} dto.ProductResponse
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/products/sku/{sku} [get]
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU is required"})
		return
	}

	resp, err := h.service.GetProductBySKU(sku)
	if err != nil {
		h.logger.Warn("Product not found by SKU", zap.String("sku", sku), zap.Error(err))
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListProducts godoc
// @Summary List products with pagination and filtering
// @Description Get a paginated list of products with optional filtering
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param search query string false "Search by name or description"
// @Param sku query string false "Filter by SKU"
// @Success 200 {object} dto.ProductListResponse
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/products [get]
func (h *ProductHandler) ListProducts(c *gin.Context) {
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	filter := &dto.ProductFilter{
		Search: c.Query("search"),
		SKU:    c.Query("sku"),
		Page:   page,
		Limit:  limit,
	}

	resp, err := h.service.GetProducts(filter)
	if err != nil {
		h.logger.Error("Failed to list products", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateProduct godoc
// @Summary Update product information
// @Description Update product details by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param request body dto.UpdateProductRequest true "Update product request"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/products/{id} [put]
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Invalid update product request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.UpdateProduct(uint(id), &req)
	if err != nil {
		if err.Error() == "product not found" {
			h.logger.Warn("Product not found", zap.Uint("id", uint(id)))
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("Failed to update product", zap.Error(err), zap.Uint("id", uint(id)))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteProduct godoc
// @Summary Delete a product
// @Description Remove a product by ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/products/{id} [delete]
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	err = h.service.DeleteProduct(uint(id))
	if err != nil {
		if err.Error() == "product not found" {
			h.logger.Warn("Product not found", zap.Uint("id", uint(id)))
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		h.logger.Error("Failed to delete product", zap.Error(err), zap.Uint("id", uint(id)))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

// RegisterRoutes registers all product routes
func (h *ProductHandler) RegisterRoutes(api *gin.RouterGroup) {
	products := api.Group("/products")
	{
		products.POST("", h.CreateProduct)
		products.GET("", h.ListProducts)
		products.GET("/:id", h.GetProduct)
		products.PUT("/:id", h.UpdateProduct)
		products.DELETE("/:id", h.DeleteProduct)
		products.GET("/sku/:sku", h.GetProductBySKU)
	}

	// Register WiFi product routes
	h.wifiHandler.RegisterRoutes(api)
}
