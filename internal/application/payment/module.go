package payment

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/worker"
	userService "github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/config"
	"github.com/novriyantoAli/wallet-ms-backend/internal/domain/payment/port"
	"github.com/novriyantoAli/wallet-ms-backend/internal/infrastructure/payment/adapter"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/queue"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Module provides all payment domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		// Provide PaymentGateway (Dana Gapura) menggunakan Dana Go SDK
		func(cfg *config.Config, logger *zap.Logger) (port.PaymentGateway, error) {
			return adapter.NewDanaGapuraClient(
				cfg.DanaGapura.BaseURL,
				cfg.DanaGapura.PartnerID,
				cfg.DanaGapura.PrivateKey,
				"", // publicKeyPEM tidak digunakan lagi, Dana SDK menangani key parsing
				logger,
			)
		},
		// Provide a nil JobScheduler for API (PaymentWorker is only used in worker service)
		func() service.JobScheduler {
			return nil
		},
		service.NewPaymentService,
		handler.NewPaymentHandler,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		// Provide PaymentGateway (Dana Gapura) menggunakan Dana Go SDK
		func(cfg *config.Config, logger *zap.Logger) (port.PaymentGateway, error) {
			return adapter.NewDanaGapuraClient(
				cfg.DanaGapura.BaseURL,
				cfg.DanaGapura.PartnerID,
				cfg.DanaGapura.PrivateKey,
				"",
				logger,
			)
		},
		// Provide the queue client as AsynqClient interface
		func(client *queue.Client) worker.AsynqClient {
			return client
		},
		// Provide PaymentService dengan nil JobScheduler untuk worker context
		func(
			repo repository.PaymentRepository,
			userSvc userService.UserService,
			logger *zap.Logger,
		) service.PaymentService {
			return service.NewPaymentService(repo, userSvc, nil, logger)
		},
		worker.NewPaymentWorker,
	),
	// Decorate JobScheduler untuk return dari PaymentWorker
	fx.Decorate(func(w *worker.PaymentWorker) service.JobScheduler {
		return w
	}),
)
