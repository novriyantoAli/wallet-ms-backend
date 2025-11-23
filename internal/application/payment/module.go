package payment

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/service"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment/worker"
	"github.com/novriyantoAli/wallet-ms-backend/internal/pkg/queue"

	"go.uber.org/fx"
)

// Module provides all payment domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		service.NewPaymentService,
		handler.NewPaymentHandler,
		worker.NewPaymentWorker,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewPaymentRepository,
		service.NewPaymentService,
		// Provide the queue client as AsynqClient interface
		func(client *queue.Client) worker.AsynqClient {
			return client
		},
		worker.NewPaymentWorker,
	),
)
