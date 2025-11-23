package worker

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include domain worker modules
	payment.WorkerModule,
	user.WorkerModule,

	// Worker api
	fx.Provide(NewServer),
)
