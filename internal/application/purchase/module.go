package purchase

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/handler"
	purchaseRepo "github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase/worker"
	walletService "github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/client"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func providePurchaseRepository(db *gorm.DB, logger *zap.Logger) purchaseRepo.PurchaseRepository {
	return purchaseRepo.NewPurchaseRepository(db, logger)
}

func provideAuthServiceClient(logger *zap.Logger) (*client.AuthServiceClient, error) {
	authTarget := "localhost:50051"
	return client.NewAuthServiceClient(authTarget, logger)
}

func providePurchaseService(
	db *gorm.DB,
	purchaseRepo purchaseRepo.PurchaseRepository,
	productRepo repository.ProductRepository,
	walletSvc walletService.WalletService,
	authClient *client.AuthServiceClient,
	logger *zap.Logger,
) service.PurchaseService {
	return service.NewPurchaseService(db, purchaseRepo, productRepo, walletSvc, authClient, logger)
}

func providePurchaseHandler(svc service.PurchaseService, logger *zap.Logger) *handler.PurchaseHandler {
	return handler.NewPurchaseHandler(svc, logger)
}

func providePurchaseWorker(logger *zap.Logger) *worker.PurchaseWorker {
	return worker.NewPurchaseWorker(logger)
}

// Module provides all purchase domain dependencies
var Module = fx.Options(
	fx.Provide(
		provideAuthServiceClient,
		providePurchaseRepository,
		providePurchaseService,
		providePurchaseHandler,
		providePurchaseWorker,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		provideAuthServiceClient,
		providePurchaseRepository,
		providePurchaseService,
		providePurchaseWorker,
	),
)
