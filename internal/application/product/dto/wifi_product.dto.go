package dto

import "time"

type CreateWiFiProductRequest struct {
	ProductID  uint    `json:"product_id" binding:"required"`
	Quota      float64 `json:"quota" binding:"required,gt=0"`
	Duration   int     `json:"duration" binding:"required,gt=0"`
	SpeedLimit float64 `json:"speed_limit" binding:"required,gt=0"`
}

type UpdateWiFiProductRequest struct {
	Quota      float64 `json:"quota" binding:"gt=0"`
	Duration   int     `json:"duration" binding:"gt=0"`
	SpeedLimit float64 `json:"speed_limit" binding:"gt=0"`
}

type WiFiProductResponse struct {
	ID         uint      `json:"id"`
	ProductID  uint      `json:"product_id"`
	Quota      float64   `json:"quota"`
	Duration   int       `json:"duration"`
	SpeedLimit float64   `json:"speed_limit"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type WiFiProductListResponse struct {
	Data       []WiFiProductResponse `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int64                 `json:"total_pages"`
}

type WiFiProductFilter struct {
	ProductID uint
	Page      int
	Limit     int
}
