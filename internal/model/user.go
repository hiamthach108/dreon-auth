package model

type User struct {
	BaseModel
	Username string `gorm:"type:varchar(255);not null"`
	Email    string `gorm:"type:varchar(255);not null"`
	Password string `gorm:"type:varchar(255);not null"`
}

func (User) TableName() string {
	return "users"
}
