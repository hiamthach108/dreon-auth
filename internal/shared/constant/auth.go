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

const (
	JWT_PAYLOAD_CONTEXT_KEY = "user_payload"
)
