package constant

type UserStatus string

const (
	UserStatusActive   UserStatus = "ACTIVE"
	UserStatusInactive UserStatus = "INACTIVE"
	UserStatusPending  UserStatus = "PENDING"
	UserStatusBlocked  UserStatus = "BLOCKED"
)

func (s UserStatus) String() string {
	return string(s)
}

type UserAuthType string

const (
	UserAuthTypeEmail      UserAuthType = "EMAIL"
	UserAuthTypeSuperAdmin UserAuthType = "SUPER_ADMIN"
	UserAuthTypeGoogle     UserAuthType = "GOOGLE"
	UserAuthTypeFacebook   UserAuthType = "FACEBOOK"
	UserAuthTypeApple      UserAuthType = "APPLE"
)

func (a UserAuthType) String() string {
	return string(a)
}

// Context keys for request-scoped values (use with context.WithValue / context.Value).
// Typed keys avoid collisions with other packages.
type ContextKey string

const (
	JWT_PAYLOAD_CONTEXT_KEY ContextKey = "user_payload"

	// Request metadata for session (ip, user agent, referer)
	ContextKeyClientIP  ContextKey = "ip"
	ContextKeyUserAgent ContextKey = "user_agent"
	ContextKeyReferer   ContextKey = "referer"
)

// Role codes for system roles
const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleUser   = "user"
)

const SystemProjectID = "system"
