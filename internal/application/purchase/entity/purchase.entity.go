package entity

import (
	"time"

	"gorm.io/gorm"
)

type PurchaseStatus string

const (
	PurchaseStatusPending   PurchaseStatus = "pending"
	PurchaseStatusCompleted PurchaseStatus = "completed"
	PurchaseStatusFailed    PurchaseStatus = "failed"
)

// Purchase represents a product purchase transaction
type Purchase struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"user_id" gorm:"not null;index"`
	ProductID  uint           `json:"product_id" gorm:"not null;index"`
	Quantity   int            `json:"quantity" gorm:"not null"`
	TotalPrice float64        `json:"total_price" gorm:"not null"`
	Status     PurchaseStatus `json:"status" gorm:"default:'pending'"`
	Notes      string         `json:"notes"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (p Purchase) TableName() string {
	return "purchases"
}

func (p *Purchase) IsValid() bool {
	return p.UserID > 0 && p.ProductID > 0 && p.Quantity > 0 && p.TotalPrice > 0
}
