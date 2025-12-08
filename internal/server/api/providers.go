package api

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/payment"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/product"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/purchase"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/wallet"

	"go.uber.org/fx"
)

var Module = fx.Options(
	// Include all domain modules
	user.Module,
	payment.Module,
	product.Module,
	wallet.Module,
	purchase.Module,

	// API api
	fx.Provide(NewServer),
)
