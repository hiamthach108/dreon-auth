package constant

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusPending  UserStatus = "pending"
	UserStatusBlocked  UserStatus = "blocked"
)

func (s UserStatus) String() string {
	return string(s)
}

type UserAuthType string

const (
	UserAuthTypeEmail    UserAuthType = "email"
	UserAuthTypeGoogle   UserAuthType = "google"
	UserAuthTypeFacebook UserAuthType = "facebook"
	UserAuthTypeApple    UserAuthType = "apple"
)

func (a UserAuthType) String() string {
	return string(a)
}
