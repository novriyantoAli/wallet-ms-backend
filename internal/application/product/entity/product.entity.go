package entity

import (
	"time"

	"gorm.io/gorm"
)

// ProductCategory represents the category of a product
type ProductCategory string

const (
	ProductCategoryWiFi  ProductCategory = "wifi"
	ProductCategoryPulsa ProductCategory = "pulsa"
)

// ProductStatus represents the status of a product (active or inactive)
type ProductStatus string

const (
	ProductStatusActive   ProductStatus = "active"
	ProductStatusInactive ProductStatus = "inactive"
)

// Product represents a product in the system
type Product struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Name        string          `json:"name" gorm:"not null"`
	Description string          `json:"description" gorm:"size:1000"`
	Price       float64         `json:"price" gorm:"not null"`
	SKU         string          `json:"sku" gorm:"size:100;uniqueIndex;not null"`
	Category    ProductCategory `json:"category" gorm:"type:VARCHAR(20);default:'wifi';not null"` // wifi or pulsa
	Status      ProductStatus   `json:"status" gorm:"type:VARCHAR(20);default:'active';not null"` // active or inactive
	Stock       int             `json:"stock" gorm:"default:0"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"deleted_at,omitempty" gorm:"index"`
}

func (p Product) TableName() string {
	return "products"
}

// IsValid checks if the product has valid required fields
func (p *Product) IsValid() bool {
	return p.Name != "" && p.Price > 0 && p.SKU != "" && p.IsValidCategory()
}

// IsValidCategory checks if the category is valid
func (p *Product) IsValidCategory() bool {
	switch p.Category {
	case ProductCategoryWiFi, ProductCategoryPulsa:
		return true
	default:
		return false
	}
}

// String implements the Stringer interface
func (pc ProductCategory) String() string {
	return string(pc)
}
