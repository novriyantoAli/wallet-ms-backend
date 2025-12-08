package entity

import (
	"time"

	"gorm.io/gorm"
)

// WiFiProduct represents a WiFi service product with quota, duration, and speed
type WiFiProduct struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	ProductID  uint           `json:"product_id" gorm:"not null;uniqueIndex"` // Link to Product table
	Quota      float64        `json:"quota" gorm:"not null"`                  // Data quota in GB
	Duration   int            `json:"duration" gorm:"not null"`               // Duration in days
	SpeedLimit float64        `json:"speed_limit" gorm:"not null"`            // Speed in Mbps
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (w WiFiProduct) TableName() string {
	return "wifi_products"
}

func (w *WiFiProduct) IsValid() bool {
	return w.ProductID > 0 && w.Quota > 0 && w.Duration > 0 && w.SpeedLimit > 0
}
