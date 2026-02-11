package main

import (
	"github.com/hiamthach108/dreon-auth/config"
	"github.com/hiamthach108/dreon-auth/internal/repository"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/cache"
	"github.com/hiamthach108/dreon-auth/pkg/database"
	"github.com/hiamthach108/dreon-auth/pkg/jwt"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"github.com/hiamthach108/dreon-auth/presentation/http"
	"github.com/hiamthach108/dreon-auth/presentation/http/handler"
	echomw "github.com/hiamthach108/dreon-auth/presentation/http/middleware"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
)

func main() {
	app := fx.New(
		fx.WithLogger(func(appLogger logger.ILogger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: appLogger.GetZapLogger()}
		}),
		fx.Provide(
			// Core
			config.NewAppConfig,
			logger.NewLogger,
			cache.NewAppCache,
			database.NewDbClient,
			jwt.NewJwtTokenManagerFromConfig,
			echomw.NewVerifyJWTMiddleware,
			http.NewHttpServer,

			// Handlers
			handler.NewUserHandler,
			handler.NewAuthHandler,

			// Services
			service.NewUserSvc,
			service.NewAuthSvc,

			// Repositories
			repository.NewUserRepository,
			repository.NewSuperAdminRepository,
			repository.NewProjectRepository,
			repository.NewSessionRepository,
		),
		fx.Invoke(http.RegisterHooks),
	)

	app.Run()
}
