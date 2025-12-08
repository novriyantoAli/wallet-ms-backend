package entity

import (
	"time"

	"gorm.io/gorm"
)

// WalletStatus represents the status of a wallet
type WalletStatus string

const (
	WalletStatusActive    WalletStatus = "active"
	WalletStatusInactive  WalletStatus = "inactive"
	WalletStatusSuspended WalletStatus = "suspended"
	WalletStatusClosed    WalletStatus = "closed"
)

// Wallet represents a user wallet
type Wallet struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"uniqueIndex;not null"` // One wallet per user
	Balance   float64        `json:"balance" gorm:"default:0;not null"`   // Current balance
	Currency  string         `json:"currency" gorm:"size:3;default:IDR"`  // Currency code (IDR, USD, etc.)
	Status    WalletStatus   `json:"status" gorm:"default:active"`        // Wallet status
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (w Wallet) TableName() string {
	return "wallets"
}

// String implements the Stringer interface for WalletStatus
func (ws WalletStatus) String() string {
	return string(ws)
}

// IsValid checks if the wallet status is valid
func (ws WalletStatus) IsValid() bool {
	switch ws {
	case WalletStatusActive, WalletStatusInactive, WalletStatusSuspended, WalletStatusClosed:
		return true
	default:
		return false
	}
}
