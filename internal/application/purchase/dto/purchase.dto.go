package dto

type CreatePurchaseRequest struct {
	UserID    uint `json:"user_id" binding:"required,gt=0"`
	ProductID uint `json:"product_id" binding:"required,gt=0"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdatePurchaseStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending completed failed"`
}

type PurchaseResponse struct {
	ID         uint    `json:"id"`
	UserID     uint    `json:"user_id"`
	ProductID  uint    `json:"product_id"`
	Quantity   int     `json:"quantity"`
	TotalPrice float64 `json:"total_price"`
	Status     string  `json:"status"`
	Notes      string  `json:"notes"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

// PurchaseDetailResponse includes user and product information
type PurchaseDetailResponse struct {
	ID         uint                 `json:"id"`
	Quantity   int                  `json:"quantity"`
	TotalPrice float64              `json:"total_price"`
	Status     string               `json:"status"`
	Notes      string               `json:"notes"`
	CreatedAt  string               `json:"created_at"`
	UpdatedAt  string               `json:"updated_at"`
	User       *PurchaseUserInfo    `json:"user"`
	Product    *PurchaseProductInfo `json:"product"`
}

// PurchaseUserInfo contains user details for a purchase
type PurchaseUserInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Level string `json:"level"`
}

// PurchaseProductInfo contains product details for a purchase
type PurchaseProductInfo struct {
	ID    uint    `json:"id"`
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

type PurchaseListResponseWithDetails struct {
	Data       []PurchaseDetailResponse `json:"data"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	Limit      int                      `json:"limit"`
	TotalPages int64                    `json:"total_pages"`
}

type PurchaseListResponse struct {
	Data       []PurchaseResponse `json:"data"`
	Total      int64              `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int64              `json:"total_pages"`
}

type PurchaseFilter struct {
	UserID    uint
	ProductID uint
	Status    string
	Page      int
	Limit     int
}
