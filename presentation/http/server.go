package http

import (
	"context"
	"net/http"
	"time"

	"github.com/hiamthach108/dreon-auth/config"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	"github.com/hiamthach108/dreon-auth/presentation/http/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/fx"
)

type HttpServer struct {
	config config.AppConfig
	logger logger.ILogger
	echo   *echo.Echo
}

func NewHttpServer(
	config *config.AppConfig,
	logger logger.ILogger,
	userHandler *handler.UserHandler,
	authHandler *handler.AuthHandler,
) *HttpServer {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	// Use middleware with your logger
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logger.Info("Request",
				"ip", c.RealIP(),
				"method", c.Request().Method,
				"path", c.Request().URL.Path,
				"user-agent", c.Request().UserAgent(),
				"referer", c.Request().Referer(),
			)
			return next(c)
		}
	})
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderAccessControlMaxAge,
			echo.HeaderAcceptEncoding,
			echo.HeaderAccessControlAllowCredentials,
			echo.HeaderAccessControlAllowHeaders,
			echo.HeaderCacheControl,
			echo.HeaderContentLength,
			echo.HeaderUpgrade,
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Content-Type", "application/json;charset=UTF-8")
			return next(c)
		}
	})

	// Healthcheck route
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	v1 := e.Group("/api/v1")

	// Register user routes (middleware applied inside RegisterRoutes)
	userHandler.RegisterRoutes(v1.Group("/users"))
	authHandler.RegisterRoutes(v1.Group("/auth"))

	return &HttpServer{
		config: *config,
		logger: logger,
		echo:   e,
	}
}

func RegisterHooks(lc fx.Lifecycle, server *HttpServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				addr := server.config.Server.Host + ":" + server.config.Server.Port
				server.logger.Info("Starting HTTP server", "addr", addr)
				if err := server.echo.Start(addr); err != nil && err != http.ErrServerClosed {
					server.logger.Fatal("Failed to start server", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.logger.Info("Shutting down HTTP server...")
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			return server.echo.Shutdown(ctx)
		},
	})
}
