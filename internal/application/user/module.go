package user

import (
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/handler"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/repository"
	"github.com/novriyantoAli/wallet-ms-backend/internal/application/user/service"

	"go.uber.org/fx"
)

// Module provides all user domain dependencies
var Module = fx.Options(
	fx.Provide(
		repository.NewUserRepository,
		service.NewUserService,
		handler.NewUserHandler,
	),
)

// WorkerModule provides only worker dependencies for worker api
var WorkerModule = fx.Options(
	fx.Provide(
		repository.NewUserRepository,
		service.NewUserService,
	),
)
