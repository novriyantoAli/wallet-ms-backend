package wallet

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/service"

	"go.uber.org/fx"
)

// Module provides all wallet domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewWalletRepository,
		repository.NewTransactionRepository,
		service.NewWalletService,
		handler.NewWalletHandler,
	),
)
