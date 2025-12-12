package service

import (
	"errors"
	"time"

	userservice "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/repository"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletService interface {
	CreateWallet(req *dto.CreateWalletRequest) (*dto.WalletResponse, error)
	GetWalletByID(id uint) (*dto.WalletResponse, error)
	GetWalletByUserID(userID uint) (*dto.WalletResponse, error)
	GetWallets(filter *dto.WalletFilter) (*dto.WalletListResponse, error)
	UpdateWalletBalance(walletID uint, req *dto.UpdateWalletBalanceRequest) (*dto.WalletBalanceResponse, error)
	DeleteWallet(id uint) error
	Transfer(senderID uint, req *dto.TransferWalletRequest) (*dto.TransferResponse, error)
	TransferFunds(req *dto.TransferRequest) (*dto.TransferResponse, error)
	CreateTransaction(req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error)
	GetTransactions(filter *dto.TransactionFilter) (*dto.TransactionListResponse, error)
	GetTransactionByID(id uint) (*dto.TransactionResponse, error)
}

type walletService struct {
	db              *gorm.DB
	repo            repository.WalletRepository
	transactionRepo repository.TransactionRepository
	userService     userservice.UserService
	logger          *zap.Logger
}

func NewWalletService(
	db *gorm.DB,
	repo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	userService userservice.UserService,
	logger *zap.Logger,
) WalletService {
	return &walletService{
		db:              db,
		repo:            repo,
		transactionRepo: transactionRepo,
		userService:     userService,
		logger:          logger,
	}
}

func (s *walletService) CreateWallet(req *dto.CreateWalletRequest) (*dto.WalletResponse, error) {
	// Verify that user exists
	_, err := s.userService.GetUserByID(req.UserID)
	if err != nil {
		s.logger.Error("User not found for wallet creation", zap.Uint("user_id", req.UserID), zap.Error(err))
		return nil, errors.New("user not found")
	}

	// Check if user already has a wallet
	existing, err := s.repo.GetByUserID(req.UserID)
	if err == nil && existing != nil {
		s.logger.Warn("User already has a wallet", zap.Uint("user_id", req.UserID))
		return nil, errors.New("user already has a wallet")
	}

	wallet := &entity.Wallet{
		UserID:    req.UserID,
		Balance:   req.InitialBalance,
		Currency:  req.Currency,
		Status:    entity.WalletStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.repo.Create(wallet)
	if err != nil {
		s.logger.Error("Failed to create wallet", zap.Error(err))
		return nil, err
	}

	// Fetch wallet with user data
	walletWithUser, err := s.repo.GetByUserIDWithUser(req.UserID)
	if err != nil {
		s.logger.Error("Failed to get wallet with user data after creation", zap.Uint("user_id", req.UserID), zap.Error(err))
		// Return basic response without user data if fetch fails
		return s.entityToResponse(wallet), nil
	}

	return s.walletWithUserToResponse(walletWithUser), nil
}

func (s *walletService) GetWalletByID(id uint) (*dto.WalletResponse, error) {
	walletWithUser, err := s.repo.GetByIDWithUser(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}

	return s.walletWithUserToResponse(walletWithUser), nil
}

func (s *walletService) GetWalletByUserID(userID uint) (*dto.WalletResponse, error) {
	walletWithUser, err := s.repo.GetByUserIDWithUser(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found for user")
		}
		return nil, err
	}

	return s.walletWithUserToResponse(walletWithUser), nil
}

func (s *walletService) GetWallets(filter *dto.WalletFilter) (*dto.WalletListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	walletsWithUser, totalCount, err := s.repo.GetAllWithUser(filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.WalletResponse, 0, len(walletsWithUser))
	for _, wallet := range walletsWithUser {
		responses = append(responses, *s.walletWithUserToResponse(&wallet))
	}

	return &dto.WalletListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *walletService) UpdateWalletBalance(walletID uint, req *dto.UpdateWalletBalanceRequest) (*dto.WalletBalanceResponse, error) {
	// Start database transaction for atomicity
	tx := s.db.Begin()
	if tx.Error != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return nil, errors.New("failed to begin transaction")
	}

	// Get wallet with row lock to prevent concurrent updates
	var wallet entity.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, walletID).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}

	previousBalance := wallet.Balance

	// Determine transaction type for transaction record
	var transactionType entity.TransactionType
	if req.TransactionType == "credit" {
		wallet.Balance += req.Amount
		transactionType = entity.TransactionTypeDeposit
	} else if req.TransactionType == "debit" {
		// Check if wallet has sufficient balance
		if wallet.Balance < req.Amount {
			tx.Rollback()
			s.logger.Warn("Insufficient balance", zap.Uint("wallet_id", walletID), zap.Float64("balance", wallet.Balance), zap.Float64("amount", req.Amount))
			return nil, errors.New("insufficient balance")
		}
		wallet.Balance -= req.Amount
		transactionType = entity.TransactionTypeWithdrawal
	} else {
		tx.Rollback()
		return nil, errors.New("invalid transaction type")
	}

	wallet.UpdatedAt = time.Now()

	// Update wallet within transaction
	if err := tx.Save(&wallet).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to update wallet balance", zap.Uint("wallet_id", walletID), zap.Error(err))
		return nil, err
	}

	// Create transaction record within transaction
	transaction := &entity.WalletTransaction{
		WalletID:     walletID,
		Type:         transactionType,
		Amount:       req.Amount,
		Status:       entity.TransactionStatusCompleted,
		Description:  req.Description,
		BalanceAfter: wallet.Balance,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := tx.Create(transaction).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create transaction record", zap.Uint("wallet_id", walletID), zap.Error(err))
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transaction", zap.Uint("wallet_id", walletID), zap.Error(err))
		return nil, errors.New("failed to commit transaction")
	}

	s.logger.Info("Wallet balance updated atomically",
		zap.Uint("wallet_id", walletID),
		zap.Float64("previous_balance", previousBalance),
		zap.Float64("new_balance", wallet.Balance),
		zap.String("type", req.TransactionType))

	return &dto.WalletBalanceResponse{
		WalletID:        wallet.ID,
		PreviousBalance: previousBalance,
		NewBalance:      wallet.Balance,
		Amount:          req.Amount,
		TransactionType: req.TransactionType,
		UpdatedAt:       wallet.UpdatedAt,
	}, nil
}

func (s *walletService) DeleteWallet(id uint) error {
	_, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("wallet not found")
		}
		return err
	}

	return s.repo.Delete(id)
}

func (s *walletService) TransferFunds(req *dto.TransferRequest) (*dto.TransferResponse, error) {
	// Validate that source and destination wallets are different
	if req.FromWalletID == req.ToWalletID {
		s.logger.Warn("Transfer to same wallet attempted", zap.Uint("wallet_id", req.FromWalletID))
		return nil, errors.New("cannot transfer to the same wallet")
	}

	// Get source wallet
	fromWallet, err := s.repo.GetByID(req.FromWalletID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Source wallet not found", zap.Uint("wallet_id", req.FromWalletID))
			return nil, errors.New("source wallet not found")
		}
		return nil, err
	}

	// Get destination wallet
	toWallet, err := s.repo.GetByID(req.ToWalletID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Destination wallet not found", zap.Uint("wallet_id", req.ToWalletID))
			return nil, errors.New("destination wallet not found")
		}
		return nil, err
	}

	// Check if source wallet has sufficient balance
	if fromWallet.Balance < req.Amount {
		s.logger.Warn("Insufficient balance for transfer",
			zap.Uint("from_wallet_id", req.FromWalletID),
			zap.Float64("balance", fromWallet.Balance),
			zap.Float64("amount", req.Amount))
		return nil, errors.New("insufficient balance for transfer")
	}

	// Store previous balances
	fromPrevBalance := fromWallet.Balance
	toPrevBalance := toWallet.Balance

	// Perform transfer
	fromWallet.Balance -= req.Amount
	toWallet.Balance += req.Amount
	fromWallet.UpdatedAt = time.Now()
	toWallet.UpdatedAt = time.Now()

	// Update source wallet
	err = s.repo.Update(fromWallet)
	if err != nil {
		s.logger.Error("Failed to debit source wallet",
			zap.Uint("from_wallet_id", req.FromWalletID),
			zap.Error(err))
		return nil, err
	}

	// Update destination wallet
	err = s.repo.Update(toWallet)
	if err != nil {
		// Rollback source wallet debit by crediting it back
		fromWallet.Balance += req.Amount
		fromWallet.UpdatedAt = time.Now()
		s.repo.Update(fromWallet)
		s.logger.Error("Failed to credit destination wallet, rolled back source wallet",
			zap.Uint("to_wallet_id", req.ToWalletID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("Funds transferred successfully",
		zap.Uint("from_wallet_id", req.FromWalletID),
		zap.Uint("to_wallet_id", req.ToWalletID),
		zap.Float64("amount", req.Amount),
		zap.Float64("from_prev_balance", fromPrevBalance),
		zap.Float64("from_new_balance", fromWallet.Balance),
		zap.Float64("to_prev_balance", toPrevBalance),
		zap.Float64("to_new_balance", toWallet.Balance))

	return &dto.TransferResponse{
		TransferID:      req.FromWalletID, // In production, you'd want to track transfers separately
		FromWalletID:    req.FromWalletID,
		ToWalletID:      req.ToWalletID,
		Amount:          req.Amount,
		Description:     req.Description,
		FromPrevBalance: fromPrevBalance,
		FromNewBalance:  fromWallet.Balance,
		ToPrevBalance:   toPrevBalance,
		ToNewBalance:    toWallet.Balance,
		Status:          "success",
		TransferredAt:   time.Now(),
	}, nil
}

// Transfer performs a wallet transfer from sender to recipient with level-based access control
func (s *walletService) Transfer(senderID uint, req *dto.TransferWalletRequest) (*dto.TransferResponse, error) {
	// Get sender user to check level
	senderUser, err := s.userService.GetUserByID(senderID)
	if err != nil {
		s.logger.Error("Sender user not found", zap.Uint("user_id", senderID), zap.Error(err))
		return nil, errors.New("sender user not found")
	}

	// Validate user level - only reseller and admin can transfer
	if senderUser.Level == "user" {
		s.logger.Warn("User level cannot perform transfer",
			zap.Uint("user_id", senderID),
			zap.String("level", senderUser.Level))
		return nil, errors.New("users cannot transfer funds")
	}

	// Get recipient user to verify existence
	_, err = s.userService.GetUserByID(req.RecipientUserID)
	if err != nil {
		s.logger.Error("Recipient user not found", zap.Uint("user_id", req.RecipientUserID), zap.Error(err))
		return nil, errors.New("recipient user not found")
	}

	// Validate that sender and recipient are different
	if senderID == req.RecipientUserID {
		s.logger.Warn("Transfer to same user attempted", zap.Uint("user_id", senderID))
		return nil, errors.New("cannot transfer to the same user")
	}

	// Get sender wallet
	senderWallet, err := s.repo.GetByUserID(senderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Sender wallet not found", zap.Uint("user_id", senderID))
			return nil, errors.New("sender wallet not found")
		}
		return nil, err
	}

	// Get recipient wallet
	recipientWallet, err := s.repo.GetByUserID(req.RecipientUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Recipient wallet not found", zap.Uint("user_id", req.RecipientUserID))
			return nil, errors.New("recipient wallet not found")
		}
		return nil, err
	}

	// Check if sender wallet has sufficient balance
	if senderWallet.Balance < req.Amount {
		s.logger.Warn("Insufficient balance for transfer",
			zap.Uint("user_id", senderID),
			zap.Float64("balance", senderWallet.Balance),
			zap.Float64("amount", req.Amount))
		return nil, errors.New("insufficient balance for transfer")
	}

	// Store previous balances
	senderPrevBalance := senderWallet.Balance
	recipientPrevBalance := recipientWallet.Balance

	// Begin transaction
	tx := s.db.Begin()
	if tx.Error != nil {
		s.logger.Error("Failed to begin transaction", zap.Error(tx.Error))
		return nil, errors.New("failed to begin transaction")
	}

	// Lock sender wallet for update
	var lockedSenderWallet entity.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&lockedSenderWallet, senderWallet.ID).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to lock sender wallet", zap.Error(err))
		return nil, errors.New("failed to acquire lock on sender wallet")
	}

	// Lock recipient wallet for update
	var lockedRecipientWallet entity.Wallet
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&lockedRecipientWallet, recipientWallet.ID).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to lock recipient wallet", zap.Error(err))
		return nil, errors.New("failed to acquire lock on recipient wallet")
	}

	// Perform transfer
	lockedSenderWallet.Balance -= req.Amount
	lockedRecipientWallet.Balance += req.Amount
	lockedSenderWallet.UpdatedAt = time.Now()
	lockedRecipientWallet.UpdatedAt = time.Now()

	// Update sender wallet
	if err := tx.Save(&lockedSenderWallet).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to debit sender wallet", zap.Uint("wallet_id", senderWallet.ID), zap.Error(err))
		return nil, err
	}

	// Update recipient wallet
	if err := tx.Save(&lockedRecipientWallet).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to credit recipient wallet", zap.Uint("wallet_id", recipientWallet.ID), zap.Error(err))
		return nil, err
	}

	// Create transaction record for sender (transfer_out)
	senderTransaction := &entity.WalletTransaction{
		WalletID:        senderWallet.ID,
		Type:            entity.TransactionTypeTransfer,
		Amount:          req.Amount,
		Status:          entity.TransactionStatusCompleted,
		Description:     "transfer_out: " + req.Description,
		BalanceAfter:    lockedSenderWallet.Balance,
		RelatedWalletID: &recipientWallet.ID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := tx.Create(senderTransaction).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create sender transaction record", zap.Error(err))
		return nil, err
	}

	// Create transaction record for recipient (transfer_in)
	recipientTransaction := &entity.WalletTransaction{
		WalletID:        recipientWallet.ID,
		Type:            entity.TransactionTypeTransfer,
		Amount:          req.Amount,
		Status:          entity.TransactionStatusCompleted,
		Description:     "transfer_in: " + req.Description,
		BalanceAfter:    lockedRecipientWallet.Balance,
		RelatedWalletID: &senderWallet.ID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := tx.Create(recipientTransaction).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to create recipient transaction record", zap.Error(err))
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		s.logger.Error("Failed to commit transfer transaction", zap.Error(err))
		return nil, errors.New("failed to commit transfer transaction")
	}

	s.logger.Info("Wallet transfer completed successfully with level validation",
		zap.Uint("sender_user_id", senderID),
		zap.String("sender_level", senderUser.Level),
		zap.Uint("recipient_user_id", req.RecipientUserID),
		zap.Float64("amount", req.Amount),
		zap.Float64("sender_prev_balance", senderPrevBalance),
		zap.Float64("sender_new_balance", lockedSenderWallet.Balance),
		zap.Float64("recipient_prev_balance", recipientPrevBalance),
		zap.Float64("recipient_new_balance", lockedRecipientWallet.Balance))

	return &dto.TransferResponse{
		TransferID:      senderTransaction.ID,
		FromWalletID:    senderWallet.ID,
		ToWalletID:      recipientWallet.ID,
		Amount:          req.Amount,
		Description:     req.Description,
		FromPrevBalance: senderPrevBalance,
		FromNewBalance:  lockedSenderWallet.Balance,
		ToPrevBalance:   recipientPrevBalance,
		ToNewBalance:    lockedRecipientWallet.Balance,
		Status:          "success",
		TransferredAt:   time.Now(),
	}, nil
}

func (s *walletService) entityToResponse(wallet *entity.Wallet) *dto.WalletResponse {
	return &dto.WalletResponse{
		ID:        wallet.ID,
		UserID:    wallet.UserID,
		User:      nil, // No user data when using this method
		Balance:   wallet.Balance,
		Currency:  wallet.Currency,
		Status:    wallet.Status.String(),
		CreatedAt: wallet.CreatedAt,
		UpdatedAt: wallet.UpdatedAt,
	}
}

func (s *walletService) walletWithUserToResponse(walletWithUser *repository.WalletWithUserData) *dto.WalletResponse {
	var userInfo *dto.UserInfo
	if walletWithUser.UserName != "" || walletWithUser.UserEmail != "" {
		userInfo = &dto.UserInfo{
			ID:    walletWithUser.UserID,
			Name:  walletWithUser.UserName,
			Email: walletWithUser.UserEmail,
			Level: walletWithUser.UserLevel,
		}
	}

	return &dto.WalletResponse{
		ID:        walletWithUser.ID,
		UserID:    walletWithUser.UserID,
		User:      userInfo,
		Balance:   walletWithUser.Balance,
		Currency:  walletWithUser.Currency,
		Status:    walletWithUser.Status,
		CreatedAt: walletWithUser.CreatedAt,
		UpdatedAt: walletWithUser.UpdatedAt,
	}
}

func (s *walletService) CreateTransaction(req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	// Validate transaction type
	transactionType := entity.TransactionType(req.Type)
	if !transactionType.IsValid() {
		s.logger.Warn("Invalid transaction type", zap.String("type", req.Type))
		return nil, errors.New("invalid transaction type")
	}

	// Get wallet
	wallet, err := s.repo.GetByID(req.WalletID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Wallet not found for transaction", zap.Uint("wallet_id", req.WalletID))
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}

	// Check balance for debit transactions
	if transactionType == entity.TransactionTypeWithdrawal ||
		transactionType == entity.TransactionTypePayment {
		if wallet.Balance < req.Amount {
			s.logger.Warn("Insufficient balance",
				zap.Uint("wallet_id", req.WalletID),
				zap.Float64("balance", wallet.Balance),
				zap.Float64("amount", req.Amount))
			return nil, errors.New("insufficient balance")
		}
	}

	// Update wallet balance
	previousBalance := wallet.Balance
	if transactionType == entity.TransactionTypeWithdrawal ||
		transactionType == entity.TransactionTypePayment ||
		transactionType == entity.TransactionTypeTransfer {
		wallet.Balance -= req.Amount
	} else {
		wallet.Balance += req.Amount
	}

	wallet.UpdatedAt = time.Now()

	// Update wallet
	err = s.repo.Update(wallet)
	if err != nil {
		s.logger.Error("Failed to update wallet", zap.Uint("wallet_id", req.WalletID), zap.Error(err))
		return nil, err
	}

	// Create transaction record
	transaction := &entity.WalletTransaction{
		WalletID:     req.WalletID,
		Type:         transactionType,
		Amount:       req.Amount,
		Status:       entity.TransactionStatusCompleted,
		Description:  req.Description,
		BalanceAfter: wallet.Balance,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = s.transactionRepo.Create(transaction)
	if err != nil {
		s.logger.Error("Failed to create transaction record", zap.Error(err))
		// Rollback wallet balance change
		wallet.Balance = previousBalance
		wallet.UpdatedAt = time.Now()
		s.repo.Update(wallet)
		return nil, err
	}

	s.logger.Info("Transaction created successfully",
		zap.Uint("wallet_id", req.WalletID),
		zap.String("type", req.Type),
		zap.Float64("amount", req.Amount),
		zap.Float64("balance_after", wallet.Balance))

	return s.transactionEntityToResponse(transaction), nil
}

func (s *walletService) GetTransactions(filter *dto.TransactionFilter) (*dto.TransactionListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	transactions, totalCount, err := s.transactionRepo.GetByWalletID(filter)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.TransactionResponse, 0, len(transactions))
	for _, transaction := range transactions {
		responses = append(responses, *s.transactionEntityToResponse(&transaction))
	}

	return &dto.TransactionListResponse{
		Data:       responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}, nil
}

func (s *walletService) GetTransactionByID(id uint) (*dto.TransactionResponse, error) {
	transaction, err := s.transactionRepo.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	return s.transactionEntityToResponse(transaction), nil
}

func (s *walletService) transactionEntityToResponse(transaction *entity.WalletTransaction) *dto.TransactionResponse {
	return &dto.TransactionResponse{
		ID:              transaction.ID,
		WalletID:        transaction.WalletID,
		Type:            transaction.Type.String(),
		Amount:          transaction.Amount,
		Status:          transaction.Status.String(),
		Description:     transaction.Description,
		BalanceAfter:    transaction.BalanceAfter,
		ReferenceID:     transaction.ReferenceID,
		RelatedWalletID: transaction.RelatedWalletID,
		CreatedAt:       transaction.CreatedAt,
		UpdatedAt:       transaction.UpdatedAt,
	}
}
