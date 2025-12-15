package product

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product/service"
	"go.uber.org/fx"
)

// Module provides all product domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewProductRepository,
		repository.NewWiFiProductRepository,
		service.NewProductService,
		service.NewWiFiProductService,
		handler.NewProductHandler,
		handler.NewWiFiProductHandler,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewProductRepository,
		repository.NewWiFiProductRepository,
		service.NewProductService,
		service.NewWiFiProductService,
	),
)
