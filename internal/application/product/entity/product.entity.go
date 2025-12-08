package entity

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID          uint    `gorm:"primaryKey"`
	Name        string  `gorm:"not null"`
	Description string  `gorm:"size:1000"`
	Price       float64 `gorm:"not null"`
	SKU         string  `gorm:"size:100;uniqueIndex;not null"`
	Stock       int     `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

func (p Product) TableName() string {
	return "products"
}

func (p *Product) IsValid() bool {
	return p.Name != "" && p.Price > 0 && p.SKU != ""
}
