package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
	"github.com/hiamthach108/dreon-auth/pkg/jwt"
	"github.com/labstack/echo/v4"
)

// VerifyJWTMiddleware is the Echo middleware that validates JWT. Use NewVerifyJWTMiddleware for fx injection.
type VerifyJWTMiddleware echo.MiddlewareFunc

// NewVerifyJWTMiddleware creates the JWT verification middleware with jwtManager injected by fx.
// Register in fx.Provide(middleware.NewVerifyJWTMiddleware) and inject VerifyJWTMiddleware where needed.
func NewVerifyJWTMiddleware(jwtManager jwt.IJwtTokenManager) VerifyJWTMiddleware {
	return VerifyJWTMiddleware(verifyJWT(jwtManager))
}

// verifyJWT returns an Echo middleware that validates the Bearer JWT and sets the payload on the context.
// Expects "Authorization: Bearer <token>". Returns 401 when the header is missing or the token is invalid.
func verifyJWT(jwtManager jwt.IJwtTokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get(echo.HeaderAuthorization)
			if auth == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
					"message": "missing authorization header",
					"code":    http.StatusUnauthorized,
				})
			}
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
					"message": "invalid authorization format",
					"code":    http.StatusUnauthorized,
				})
			}
			tokenString := strings.TrimSpace(auth[len(prefix):])
			if tokenString == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
					"message": "missing token",
					"code":    http.StatusUnauthorized,
				})
			}

			payload, err := jwtManager.Verify(c.Request().Context(), tokenString)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, echo.Map{
					"message": err.Error(),
					"code":    http.StatusUnauthorized,
				})
			}
			ctx := context.WithValue(c.Request().Context(), constant.JWT_PAYLOAD_CONTEXT_KEY, payload)
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	}
}

// GetJWTPayload returns the JWT payload set by VerifyJWT middleware. Returns nil if not set.
func GetJWTPayload(ctx context.Context) *jwt.Payload {
	v := ctx.Value(constant.JWT_PAYLOAD_CONTEXT_KEY)
	if v == nil {
		return nil
	}
	p, _ := v.(*jwt.Payload)
	return p
}
