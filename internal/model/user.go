package model

import (
	"time"

	"github.com/hiamthach108/dreon-auth/internal/shared/constant"
)

type User struct {
	BaseModel
	Username    string                `gorm:"type:varchar(255);not null;unique"`
	Email       string                `gorm:"type:varchar(255);not null;unique"`
	Password    string                `gorm:"type:varchar(255);not null"`
	Status      constant.UserStatus   `gorm:"type:varchar(50);default:active"`
	AuthType    constant.UserAuthType `gorm:"type:varchar(50);default:email"`
	LastLoginAt time.Time             `gorm:"type:timestamp;default:null"`
}

func (User) TableName() string {
	return "users"
}
