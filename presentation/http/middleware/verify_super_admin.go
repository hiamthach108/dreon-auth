package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// VerifySuperAdminMiddleware is the Echo middleware that ensures the request context has a JWT payload with IsSuperAdmin true.
// Must be used after VerifyJWTMiddleware so the payload is set on the context.
type VerifySuperAdminMiddleware echo.MiddlewareFunc

// NewVerifySuperAdminMiddleware creates the super-admin verification middleware.
// It reads the JWT payload from context (set by VerifyJWTMiddleware) and returns 403 if the user is not a super admin.
func NewVerifySuperAdminMiddleware() VerifySuperAdminMiddleware {
	return VerifySuperAdminMiddleware(verifySuperAdmin)
}

// verifySuperAdmin returns an Echo middleware that requires payload.IsSuperAdmin.
// Returns 403 Forbidden when payload is missing or IsSuperAdmin is false.
func verifySuperAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		payload := GetJWTPayload(c.Request().Context())
		if payload == nil || !payload.IsSuperAdmin {
			return echo.NewHTTPError(http.StatusForbidden, echo.Map{
				"message": "super admin access required",
				"code":    http.StatusForbidden,
			})
		}
		return next(c)
	}
}
