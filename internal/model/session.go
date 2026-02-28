package model

import "time"

type Session struct {
	BaseModel
	UserID       string    `gorm:"type:varchar(36);not null"`
	Email        string    `gorm:"type:varchar(255);default:null"`
	RefreshToken string    `gorm:"type:varchar(255);not null"`
	ExpiresAt    time.Time `gorm:"type:timestamp;not null"`
	IsActive     bool      `gorm:"type:boolean;default:true"`
	IsSuperAdmin bool      `gorm:"type:boolean;default:false"`
}

func (Session) TableName() string {
	return "sessions"
}
