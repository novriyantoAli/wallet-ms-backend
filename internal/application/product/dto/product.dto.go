package dto

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=255"`
	Description string  `json:"description" binding:"max=1000"`
	Price       float64 `json:"price" binding:"required,gt=0"`
	SKU         string  `json:"sku" binding:"required,min=1,max=100"`
	Stock       int     `json:"stock" binding:"min=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name" binding:"omitempty,min=1,max=255"`
	Description string  `json:"description" binding:"omitempty,max=1000"`
	Price       float64 `json:"price" binding:"omitempty,gt=0"`
	SKU         string  `json:"sku" binding:"omitempty,min=1,max=100"`
	Stock       int     `json:"stock" binding:"omitempty,min=0"`
}

type ProductResponse struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	SKU         string  `json:"sku"`
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
