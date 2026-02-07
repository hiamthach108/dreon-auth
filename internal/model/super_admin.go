package model

type SuperAdmin struct {
	BaseModel
	Name     string `gorm:"type:varchar(255);not null"`
	Email    string `gorm:"type:varchar(255);not null;unique"`
	Password string `gorm:"type:varchar(255);not null"`
	IsActive bool   `gorm:"type:boolean;default:false"`
}

func (SuperAdmin) TableName() string {
	return "super_admins"
}
