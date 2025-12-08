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
