package worker

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include domain modules
	wallet.Module,
	user.WorkerModule,
	payment.WorkerModule,

	// Worker api
	fx.Provide(NewServer),
)
