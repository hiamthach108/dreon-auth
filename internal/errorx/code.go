package errorx

type AppErrCode int

const (
	// General errors
	ErrInternal      AppErrCode = 500
	ErrBadRequest    AppErrCode = 400
	ErrNotFound      AppErrCode = 404
	ErrUnauthorized  AppErrCode = 401
	ErrForbidden     AppErrCode = 403
	ErrConflict      AppErrCode = 409
	ErrUnprocessable AppErrCode = 422
	ErrRateLimit     AppErrCode = 429

	// User errors
	ErrUserNotFound        AppErrCode = 1001
	ErrUserConflict        AppErrCode = 1002
	ErrCreateUser          AppErrCode = 1003
	ErrUpdateUser          AppErrCode = 1004
	ErrUserInactive        AppErrCode = 1005
	ErrInvalidCredentials  AppErrCode = 1006
	ErrInvalidAuthType     AppErrCode = 1007
	ErrInvalidPassword     AppErrCode = 1008
	ErrInvalidRefreshToken AppErrCode = 1009
	ErrRefreshTokenExpired AppErrCode = 1010
)

var errorMsgs = map[AppErrCode]string{
	ErrInternal:      "Internal server error",
	ErrBadRequest:    "Bad request",
	ErrNotFound:      "Resource not found",
	ErrUnauthorized:  "Unauthorized access",
	ErrForbidden:     "Forbidden access",
	ErrConflict:      "Resource conflict",
	ErrUnprocessable: "Unprocessable entity",
	ErrRateLimit:     "Too many requests",

	ErrUserNotFound:        "User not found",
	ErrUserConflict:        "User already exists",
	ErrCreateUser:          "Failed to create user",
	ErrUpdateUser:          "Failed to update user",
	ErrUserInactive:        "User is inactive",
	ErrInvalidCredentials:  "Invalid credentials",
	ErrInvalidAuthType:     "Invalid auth type",
	ErrInvalidPassword:     "Invalid password",
	ErrInvalidRefreshToken: "Invalid refresh token",
	ErrRefreshTokenExpired: "Refresh token expired",
}

// GetErrorMessage returns a user-friendly error message for a given error code.
func GetErrorMessage(code int) string {
	if msg, exists := errorMsgs[AppErrCode(code)]; exists {
		return msg
	}
	return "An unknown error occurred."
}
