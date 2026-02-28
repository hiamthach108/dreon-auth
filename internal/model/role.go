package model

import "gorm.io/datatypes"

type Role struct {
	BaseModel
	Code        string         `gorm:"type:varchar(255);not null;unique"`
	Name        string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text"`
	IsActive    bool           `gorm:"type:boolean;default:true"`
	ProjectID   *string        `gorm:"type:varchar(36)"` // may be null for system roles
	Permissions datatypes.JSON `gorm:"type:jsonb"`
}

func (Role) TableName() string {
	return "roles"
}
