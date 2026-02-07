package model

type Project struct {
	BaseModel
	Code        string `gorm:"type:varchar(255);not null;unique"`
	Name        string `gorm:"type:varchar(255);not null"`
	Description string `gorm:"type:text"`
}

func (Project) TableName() string {
	return "clients"
}
