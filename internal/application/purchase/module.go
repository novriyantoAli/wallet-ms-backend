package purchase

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/handler"
	purchaseRepo "github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/worker"
	walletService "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/service"
)

func providePurchaseRepository(db *gorm.DB, logger *zap.Logger) purchaseRepo.PurchaseRepository {
	return purchaseRepo.NewPurchaseRepository(db, logger)
}

func providePurchaseService(
	db *gorm.DB,
	purchaseRepo purchaseRepo.PurchaseRepository,
	productRepo repository.ProductRepository,
	walletSvc walletService.WalletService,
	logger *zap.Logger,
) service.PurchaseService {
	return service.NewPurchaseService(db, purchaseRepo, productRepo, walletSvc, logger)
}

func providePurchaseHandler(svc service.PurchaseService, logger *zap.Logger) *handler.PurchaseHandler {
	return handler.NewPurchaseHandler(svc, logger)
}

func providePurchaseWorker(logger *zap.Logger) *worker.PurchaseWorker {
	return worker.NewPurchaseWorker(logger)
}

var Module = fx.Options(
	fx.Provide(providePurchaseRepository),
	fx.Provide(providePurchaseService),
	fx.Provide(providePurchaseHandler),
	fx.Provide(providePurchaseWorker),
)
