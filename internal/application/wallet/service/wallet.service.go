package service

import (
	"context"
	"errors"
	"time"

	userservice "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/dto"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/entity"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/database"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type WalletService interface {
	CreateWallet(ctx context.Context, req *dto.CreateWalletRequest) (*dto.WalletResponse, error)
	GetWalletByID(ctx context.Context, id uint) (*dto.WalletResponse, error)
	GetWalletByUserID(ctx context.Context, userID uint) (*dto.WalletResponse, error)
	GetWallets(ctx context.Context, filter *dto.WalletFilter) (*dto.WalletListResponse, error)
	UpdateWalletBalance(ctx context.Context, walletID uint, req *dto.UpdateWalletBalanceRequest) (*dto.WalletBalanceResponse, error)
	DeleteWallet(ctx context.Context, id uint) error
	Transfer(ctx context.Context, senderID uint, req *dto.TransferWalletRequest) (*dto.TransferResponse, error)
	CreateTransaction(ctx context.Context, req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error)
	GetTransactions(ctx context.Context, filter *dto.TransactionFilter) (*dto.TransactionListResponse, error)
	GetTransactionByID(ctx context.Context, id uint) (*dto.TransactionResponse, error)
}

type walletService struct {
	txManager       database.TransactionManagerI
	repo            repository.WalletRepository
	transactionRepo repository.TransactionRepository
	userService     userservice.UserService
	logger          *zap.Logger
}

func NewWalletService(
	txManager database.TransactionManagerI,
	repo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	userService userservice.UserService,
	logger *zap.Logger,
) WalletService {
	return &walletService{
		txManager:       txManager,
		repo:            repo,
		transactionRepo: transactionRepo,
		userService:     userService,
		logger:          logger,
	}
}

func (s *walletService) CreateWallet(ctx context.Context, req *dto.CreateWalletRequest) (*dto.WalletResponse, error) {
	// Verify that user exists
	_, err := s.userService.GetUserByID(req.UserID)
	if err != nil {
		s.logger.Error("User not found for wallet creation", zap.Uint("user_id", req.UserID), zap.Error(err))
		return nil, errors.New("user not found")
	}

	// Check if user already has a wallet
	existing, err := s.repo.GetByUserID(ctx, req.UserID)
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

	err = s.repo.Create(ctx, wallet)
	if err != nil {
		s.logger.Error("Failed to create wallet", zap.Error(err))
		return nil, err
	}

	// Fetch wallet with user data
	walletWithUser, err := s.repo.GetByUserIDWithUser(ctx, req.UserID)
	if err != nil {
		s.logger.Error("Failed to get wallet with user data after creation", zap.Uint("user_id", req.UserID), zap.Error(err))
		// Return basic response without user data if fetch fails
		return s.entityToResponse(wallet), nil
	}

	return s.walletWithUserToResponse(walletWithUser), nil
}

func (s *walletService) GetWalletByID(ctx context.Context, id uint) (*dto.WalletResponse, error) {
	walletWithUser, err := s.repo.GetByIDWithUser(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}

	return s.walletWithUserToResponse(walletWithUser), nil
}

func (s *walletService) GetWalletByUserID(ctx context.Context, userID uint) (*dto.WalletResponse, error) {
	walletWithUser, err := s.repo.GetByUserIDWithUser(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found for user")
		}
		return nil, err
	}

	return s.walletWithUserToResponse(walletWithUser), nil
}

func (s *walletService) GetWallets(ctx context.Context, filter *dto.WalletFilter) (*dto.WalletListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	walletsWithUser, totalCount, err := s.repo.GetAllWithUser(ctx, filter)
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

func (s *walletService) UpdateWalletBalance(ctx context.Context, walletID uint, req *dto.UpdateWalletBalanceRequest) (*dto.WalletBalanceResponse, error) {
	var previousBalance float64
	var wallet *entity.Wallet

	// Execute within transaction using tx_manager
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		// Get wallet repository

		tx := database.GetDB(txCtx, nil)

		// Get wallet with row lock to prevent concurrent updates
		var w entity.Wallet
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&w, walletID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("wallet not found")
			}
			return err
		}

		previousBalance = w.Balance
		wallet = &w

		// Determine transaction type for transaction record
		var transactionType entity.TransactionType
		if req.TransactionType == "credit" {
			w.Balance += req.Amount
			transactionType = entity.TransactionTypeDeposit
		} else if req.TransactionType == "debit" {
			// Check if wallet has sufficient balance
			if w.Balance < req.Amount {
				s.logger.Warn("Insufficient balance", zap.Uint("wallet_id", walletID), zap.Float64("balance", w.Balance), zap.Float64("amount", req.Amount))
				return errors.New("insufficient balance")
			}
			w.Balance -= req.Amount
			transactionType = entity.TransactionTypeWithdrawal
		} else {
			return errors.New("invalid transaction type")
		}

		w.UpdatedAt = time.Now()

		// Update wallet within transaction
		if err := tx.Save(&w).Error; err != nil {
			s.logger.Error("Failed to update wallet balance", zap.Uint("wallet_id", walletID), zap.Error(err))
			return err
		}

		// Create transaction record within transaction
		transaction := &entity.WalletTransaction{
			WalletID:     walletID,
			Type:         transactionType,
			Amount:       req.Amount,
			Status:       entity.TransactionStatusCompleted,
			Description:  req.Description,
			BalanceAfter: w.Balance,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := tx.Create(transaction).Error; err != nil {
			s.logger.Error("Failed to create transaction record", zap.Uint("wallet_id", walletID), zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
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

func (s *walletService) DeleteWallet(ctx context.Context, id uint) error {
	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("wallet not found")
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

// Transfer performs a wallet transfer from sender to recipient with level-based access control
func (s *walletService) Transfer(ctx context.Context, senderID uint, req *dto.TransferWalletRequest) (*dto.TransferResponse, error) {
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
	senderWallet, err := s.repo.GetByUserID(ctx, senderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Sender wallet not found", zap.Uint("user_id", senderID))
			return nil, errors.New("sender wallet not found")
		}
		return nil, err
	}

	// Get recipient wallet
	recipientWallet, err := s.repo.GetByUserID(ctx, req.RecipientUserID)
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

	var lockedSenderWallet *entity.Wallet
	var lockedRecipientWallet *entity.Wallet
	var senderTransaction *entity.WalletTransaction
	var recipientTransaction *entity.WalletTransaction

	// Execute within transaction using tx_manager
	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {

		// Lock sender wallet for update
		lsw, err := s.repo.GetByIDForUpdate(txCtx, senderWallet.ID)
		if err != nil {
			s.logger.Error("Failed to lock sender wallet", zap.Error(err))
			return errors.New("failed to acquire lock on sender wallet")
		}
		lockedSenderWallet = lsw

		// Lock recipient wallet for update
		lrw, err := s.repo.GetByIDForUpdate(txCtx, recipientWallet.ID)
		if err != nil {
			s.logger.Error("Failed to lock recipient wallet", zap.Error(err))
			return errors.New("failed to acquire lock on recipient wallet")
		}
		lockedRecipientWallet = lrw

		// Perform transfer
		lockedSenderWallet.Balance -= req.Amount
		lockedRecipientWallet.Balance += req.Amount
		lockedSenderWallet.UpdatedAt = time.Now()
		lockedRecipientWallet.UpdatedAt = time.Now()

		// Update sender wallet
		if err := s.repo.Update(txCtx, lockedSenderWallet); err != nil {
			s.logger.Error("Failed to debit sender wallet", zap.Uint("wallet_id", senderWallet.ID), zap.Error(err))
			return err
		}

		// Update recipient wallet
		if err := s.repo.Update(txCtx, lockedRecipientWallet); err != nil {
			s.logger.Error("Failed to credit recipient wallet", zap.Uint("wallet_id", recipientWallet.ID), zap.Error(err))
			return err
		}

		// Create transaction record for sender (transfer_out)
		senderTransaction = &entity.WalletTransaction{
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

		// Create transaction record
		if err := s.transactionRepo.Create(txCtx, senderTransaction); err != nil {
			s.logger.Error("Failed to create sender transaction record", zap.Error(err))
			return err
		}

		// Create transaction record for recipient (transfer_in)
		recipientTransaction = &entity.WalletTransaction{
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

		// Create transaction record
		if err := s.transactionRepo.Create(txCtx, recipientTransaction); err != nil {
			s.logger.Error("Failed to create recipient transaction record", zap.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
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

func (s *walletService) CreateTransaction(ctx context.Context, req *dto.CreateTransactionRequest) (*dto.TransactionResponse, error) {
	// Validate transaction type
	transactionType := entity.TransactionType(req.Type)
	if !transactionType.IsValid() {
		s.logger.Warn("Invalid transaction type", zap.String("type", req.Type))
		return nil, errors.New("invalid transaction type")
	}

	// Get wallet
	wallet, err := s.repo.GetByID(ctx, req.WalletID)
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
	err = s.repo.Update(ctx, wallet)
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

	err = s.transactionRepo.Create(ctx, transaction)
	if err != nil {
		s.logger.Error("Failed to create transaction record", zap.Error(err))
		// Rollback wallet balance change
		wallet.Balance = previousBalance
		wallet.UpdatedAt = time.Now()
		s.repo.Update(ctx, wallet)
		return nil, err
	}

	s.logger.Info("Transaction created successfully",
		zap.Uint("wallet_id", req.WalletID),
		zap.String("type", req.Type),
		zap.Float64("amount", req.Amount),
		zap.Float64("balance_after", wallet.Balance))

	return s.transactionEntityToResponse(transaction), nil
}

func (s *walletService) GetTransactions(ctx context.Context, filter *dto.TransactionFilter) (*dto.TransactionListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 10
	}

	transactions, totalCount, err := s.transactionRepo.GetByWalletID(ctx, filter)
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

func (s *walletService) GetTransactionByID(ctx context.Context, id uint) (*dto.TransactionResponse, error) {
	transaction, err := s.transactionRepo.GetByID(ctx, id)
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
