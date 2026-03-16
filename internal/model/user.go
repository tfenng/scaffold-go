package model

import "time"

type User struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	UID       string     `gorm:"column:uid;type:varchar(64);not null;uniqueIndex"`
	Email     *string    `gorm:"column:email;type:varchar(255);uniqueIndex"`
	Name      string     `gorm:"column:name;type:varchar(255);not null"`
	UsedName  string     `gorm:"column:used_name;type:varchar(255);not null;default:''"`
	Company   string     `gorm:"column:company;type:varchar(255);not null;default:''"`
	Birth     *time.Time `gorm:"column:birth;type:date"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (User) TableName() string {
	return "users"
}
