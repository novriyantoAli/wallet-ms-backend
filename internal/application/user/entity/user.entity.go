package entity

import (
	"time"

	"gorm.io/gorm"
)

type UserLevel string

const (
	UserLevelUser     UserLevel = "user"
	UserLevelReseller UserLevel = "reseller"
	UserLevelAdmin    UserLevel = "admin"
)

type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Name      string         `json:"name" gorm:"not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"not null"`
	Level     UserLevel      `json:"level" gorm:"type:VARCHAR(20);default:'user';not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

func (u User) TableName() string {
	return "users"
}
