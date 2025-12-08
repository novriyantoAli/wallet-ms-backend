package product

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/service"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var Module = fx.Options(
	fx.Provide(
		provideProductRepository,
		provideProductService,
		provideProductHandler,
		provideWiFiProductRepository,
		provideWiFiProductService,
		provideWiFiProductHandler,
	),
)

func provideProductRepository(db *gorm.DB, logger *zap.Logger) repository.ProductRepository {
	return repository.NewProductRepository(db, logger)
}

func provideProductService(repo repository.ProductRepository, logger *zap.Logger) service.ProductService {
	return service.NewProductService(repo, logger)
}

func provideProductHandler(service service.ProductService, wifiService service.WiFiProductService, logger *zap.Logger) *handler.ProductHandler {
	return handler.NewProductHandler(service, wifiService, logger)
}

func provideWiFiProductRepository(db *gorm.DB) repository.WiFiProductRepository {
	return repository.NewWiFiProductRepository(db)
}

func provideWiFiProductService(repo repository.WiFiProductRepository, logger *zap.Logger) service.WiFiProductService {
	return service.NewWiFiProductService(repo, logger)
}

func provideWiFiProductHandler(service service.WiFiProductService, logger *zap.Logger) *handler.WiFiProductHandler {
	return handler.NewWiFiProductHandler(service, logger)
}
