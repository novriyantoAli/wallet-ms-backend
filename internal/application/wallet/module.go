package wallet

import (
	userservice "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Module = fx.Options(
	fx.Provide(
		provideWalletRepository,
		provideTransactionRepository,
		provideWalletService,
		provideWalletHandler,
	),
)

func provideWalletRepository(db *gorm.DB, logger *zap.Logger) repository.WalletRepository {
	return repository.NewWalletRepository(db, logger)
}

func provideTransactionRepository(db *gorm.DB, logger *zap.Logger) repository.TransactionRepository {
	return repository.NewTransactionRepository(db, logger)
}

func provideWalletService(
	repo repository.WalletRepository,
	transactionRepo repository.TransactionRepository,
	userService userservice.UserService,
	logger *zap.Logger,
) service.WalletService {
	return service.NewWalletService(repo, transactionRepo, userService, logger)
}

func provideWalletHandler(service service.WalletService, logger *zap.Logger) *handler.WalletHandler {
	return handler.NewWalletHandler(service, logger)
}
