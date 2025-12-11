package dto

import "time"

// CreateWalletRequest contains the request payload for creating a wallet
type CreateWalletRequest struct {
	UserID         uint    `json:"user_id" binding:"required"`
	Currency       string  `json:"currency" binding:"required,len=3"`
	InitialBalance float64 `json:"initial_balance" binding:"gte=0"`
}

// UpdateWalletBalanceRequest contains the request payload for updating wallet balance
type UpdateWalletBalanceRequest struct {
	Amount          float64 `json:"amount" binding:"required,ne=0"`
	Description     string  `json:"description"`
	TransactionType string  `json:"transaction_type" binding:"required,oneof=credit debit"`
}

// UserInfo contains basic user information for wallet response
type UserInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Level string `json:"level"`
}

// WalletResponse contains the response for wallet operations
type WalletResponse struct {
	ID        uint      `json:"id"`
	UserID    uint      `json:"user_id"`
	User      *UserInfo `json:"user"`
	Balance   float64   `json:"balance"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WalletListResponse contains the response for listing wallets
type WalletListResponse struct {
	Data       []WalletResponse `json:"data"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
}

// WalletFilter contains filter parameters for wallet queries
type WalletFilter struct {
	UserID   uint   `form:"user_id"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

// WalletBalanceResponse contains response for balance operations
type WalletBalanceResponse struct {
	WalletID        uint      `json:"wallet_id"`
	PreviousBalance float64   `json:"previous_balance"`
	NewBalance      float64   `json:"new_balance"`
	Amount          float64   `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TransferRequest contains the request payload for wallet transfer
type TransferRequest struct {
	FromWalletID uint    `json:"from_wallet_id" binding:"required"`
	ToWalletID   uint    `json:"to_wallet_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	Description  string  `json:"description"`
}

// TransferResponse contains the response for transfer operations
type TransferResponse struct {
	TransferID      uint      `json:"transfer_id"`
	FromWalletID    uint      `json:"from_wallet_id"`
	ToWalletID      uint      `json:"to_wallet_id"`
	Amount          float64   `json:"amount"`
	Description     string    `json:"description"`
	FromPrevBalance float64   `json:"from_prev_balance"`
	FromNewBalance  float64   `json:"from_new_balance"`
	ToPrevBalance   float64   `json:"to_prev_balance"`
	ToNewBalance    float64   `json:"to_new_balance"`
	Status          string    `json:"status"`
	TransferredAt   time.Time `json:"transferred_at"`
}
