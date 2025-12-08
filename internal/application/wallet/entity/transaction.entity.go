package entity

import (
	"time"

	"gorm.io/gorm"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeDeposit    TransactionType = "deposit"
	TransactionTypeWithdrawal TransactionType = "withdrawal"
	TransactionTypeTransfer   TransactionType = "transfer"
	TransactionTypeTopUp      TransactionType = "topup"
	TransactionTypePayment    TransactionType = "payment"
	TransactionTypeRefund     TransactionType = "refund"
	TransactionTypeAdjustment TransactionType = "adjustment"
)

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// WalletTransaction represents a wallet transaction
type WalletTransaction struct {
	ID              uint              `json:"id" gorm:"primaryKey"`
	WalletID        uint              `json:"wallet_id" gorm:"not null;index"`
	Type            TransactionType   `json:"type" gorm:"not null"`
	Amount          float64           `json:"amount" gorm:"not null"`
	Status          TransactionStatus `json:"status" gorm:"default:pending"`
	Description     string            `json:"description" gorm:"size:255"`
	BalanceAfter    float64           `json:"balance_after" gorm:"not null"`
	ReferenceID     string            `json:"reference_id" gorm:"size:100;index"`
	RelatedWalletID *uint             `json:"related_wallet_id"`
	Metadata        string            `json:"metadata" gorm:"type:json"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	DeletedAt       gorm.DeletedAt    `json:"deleted_at,omitempty" gorm:"index"`
}

func (wt WalletTransaction) TableName() string {
	return "wallet_transactions"
}

// IsValid checks if the transaction type is valid
func (tt TransactionType) IsValid() bool {
	switch tt {
	case TransactionTypeDeposit, TransactionTypeWithdrawal, TransactionTypeTransfer,
		TransactionTypeTopUp, TransactionTypePayment, TransactionTypeRefund,
		TransactionTypeAdjustment:
		return true
	default:
		return false
	}
}

// IsValid checks if the transaction status is valid
func (ts TransactionStatus) IsValid() bool {
	switch ts {
	case TransactionStatusPending, TransactionStatusCompleted, TransactionStatusFailed,
		TransactionStatusCancelled:
		return true
	default:
		return false
	}
}

// String implements the Stringer interface
func (tt TransactionType) String() string {
	return string(tt)
}

func (ts TransactionStatus) String() string {
	return string(ts)
}
