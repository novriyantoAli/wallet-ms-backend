package api

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include all domain modules
	user.Module,
	payment.Module,

	// API api
	fx.Provide(NewServer),
)
