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

	// Business errors
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
	ErrProjectNotFound     AppErrCode = 1011
	ErrProjectConflict     AppErrCode = 1012
	ErrCreateProject       AppErrCode = 1013
	ErrUpdateProject       AppErrCode = 1014
	ErrPermissionDenied    AppErrCode = 1015
	ErrPermissionNotFound  AppErrCode = 1016
	ErrPermissionConflict  AppErrCode = 1017
	ErrInvalidPermission   AppErrCode = 1018
	ErrPermissionExpired   AppErrCode = 1019
	ErrGrantPermission     AppErrCode = 1020
	ErrRevokePermission    AppErrCode = 1021
	ErrInvalidTupleFormat  AppErrCode = 1022
	ErrRoleNotFound        AppErrCode = 1023
	ErrRoleConflict        AppErrCode = 1024
	ErrCreateRole          AppErrCode = 1025
	ErrUpdateRole          AppErrCode = 1026
	ErrDeleteRole          AppErrCode = 1027
	ErrSystemRoleProtected AppErrCode = 1028
	ErrInvalidRole         AppErrCode = 1029
	ErrRoleAssignment      AppErrCode = 1030
	ErrInvalidRefreshState AppErrCode = 1031
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
	ErrInvalidRefreshState: "Invalid or expired refresh state",

	ErrProjectNotFound: "Project not found",
	ErrProjectConflict: "Project with this code already exists",
	ErrCreateProject:   "Failed to create project",
	ErrUpdateProject:   "Failed to update project",

	ErrPermissionDenied:   "Permission denied",
	ErrPermissionNotFound: "Permission not found",
	ErrPermissionConflict: "Permission already exists",
	ErrInvalidPermission:  "Invalid permission",
	ErrPermissionExpired:  "Permission has expired",
	ErrGrantPermission:    "Failed to grant permission",
	ErrRevokePermission:   "Failed to revoke permission",
	ErrInvalidTupleFormat: "Invalid relation tuple format",

	ErrRoleNotFound:        "Role not found",
	ErrRoleConflict:        "Role with this code already exists",
	ErrCreateRole:          "Failed to create role",
	ErrUpdateRole:          "Failed to update role",
	ErrDeleteRole:          "Failed to delete role",
	ErrSystemRoleProtected: "System roles can only be modified by super admins",
	ErrInvalidRole:         "Invalid role data",
	ErrRoleAssignment:      "Failed to assign/remove role",
}

// GetErrorMessage returns a user-friendly error message for a given error code.
func GetErrorMessage(code int) string {
	if msg, exists := errorMsgs[AppErrCode(code)]; exists {
		return msg
	}
	return "An unknown error occurred."
}
