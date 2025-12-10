package dto

// CreateProductRequest contains the information needed to create a new product.
// SKU is auto-generated on the server using category, name, and price.
// The JSON tags are used to map the fields to the corresponding JSON fields.
// The binding tags are used to validate the input data.
//
// @example
//
//	{
//	 "name": "WiFi 10GB",
//	 "description": "WiFi package 10GB",
//	 "price": 50000,
//	 "category": "wifi",
//	 "stock": 100
//	}
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description string  `json:"description" binding:"max=1000"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required,oneof=wifi pulsa"` // wifi or pulsa
	Stock       int     `json:"stock" binding:"min=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=1,max=255"`
	Description string  `json:"description" binding:"omitempty,max=1000"`
	Price       float64 `json:"price" binding:"omitempty,gt=0"`
	Status      string  `json:"status" binding:"omitempty,oneof=active inactive"`
	Stock       int     `json:"stock" binding:"omitempty,min=0"`
}

type UpdateProductStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=active inactive"`
}

type ProductResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
	Category    string  `json:"category"`
	Status      string  `json:"status"`
	Stock       int     `json:"stock"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ProductFilter struct {
	Search string
	SKU    string
	Page   int
	Limit  int
}

type ProductListResponse struct {
	Data       []ProductResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}
