package model

type UserRole struct {
	BaseModel
	UserID    string  `gorm:"type:varchar(36);not null"`
	RoleID    string  `gorm:"type:varchar(36);not null"`
	ProjectID *string `gorm:"type:varchar(36)"` // may be null for system user roles

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Role Role `gorm:"foreignKey:RoleID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

func (UserRole) TableName() string {
	return "user_roles"
}
