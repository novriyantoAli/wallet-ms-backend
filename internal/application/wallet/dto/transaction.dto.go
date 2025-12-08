package dto

import "time"

// CreateTransactionRequest contains the request payload for creating a transaction
type CreateTransactionRequest struct {
	WalletID    uint    `json:"wallet_id" binding:"required"`
	Type        string  `json:"type" binding:"required,oneof=deposit withdrawal transfer topup payment refund adjustment"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

// TransactionResponse contains the response for transaction operations
type TransactionResponse struct {
	ID              uint      `json:"id"`
	WalletID        uint      `json:"wallet_id"`
	Type            string    `json:"type"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	Description     string    `json:"description"`
	BalanceAfter    float64   `json:"balance_after"`
	ReferenceID     string    `json:"reference_id"`
	RelatedWalletID *uint     `json:"related_wallet_id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// GetTransactionRequest contains filter parameters for transaction queries
type GetTransactionRequest struct {
	WalletID  uint   `form:"wallet_id" binding:"required"`
	Type      string `form:"type"`
	Status    string `form:"status"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Page      int    `form:"page"`
	PageSize  int    `form:"page_size"`
}

// TransactionListResponse contains the response for listing transactions
type TransactionListResponse struct {
	Data       []TransactionResponse `json:"data"`
	TotalCount int64                 `json:"total_count"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
}

// TransactionFilter contains filter parameters for transaction queries
type TransactionFilter struct {
	WalletID  uint
	Type      string
	Status    string
	StartDate *time.Time
	EndDate   *time.Time
	Page      int
	PageSize  int
}
