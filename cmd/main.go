package main

import (
	"github.com/hiamthach108/dreon-auth/config"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/cache"
	"github.com/hiamthach108/dreon-auth/pkg/database"

	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"github.com/hiamthach108/dreon-auth/presentation/http"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			// Core
			config.NewAppConfig,
			logger.NewLogger,
			cache.NewAppCache,
			database.NewDbClient,
			http.NewHttpServer,

			// Services
			service.NewUserSvc,

			// Repositories
			repository.NewUserRepository,
		),
		fx.Invoke(http.RegisterHooks),
	)

	app.Run()
}
