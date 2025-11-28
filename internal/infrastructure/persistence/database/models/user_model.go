package models

import (
    "time"
    "gorm.io/gorm"
)

type UserModel struct {
    ID           string         `gorm:"primaryKey;type:varchar(36)"`
    Username     string         `gorm:"uniqueIndex;not null;type:varchar(50)"`
    Email        string         `gorm:"uniqueIndex;not null;type:varchar(255)"`
    PasswordHash string         `gorm:"not null;type:text"`
    Phone        string         `gorm:"type:varchar(20)"`
    FirstName    string         `gorm:"not null;type:varchar(100)"`
    LastName     string         `gorm:"not null;type:varchar(100)"`
    Role         string         `gorm:"not null;type:varchar(20);index"`
    IsActive     bool           `gorm:"not null;default:true;index"`
    IsVerified   bool           `gorm:"not null;default:false"`
    LastLoginAt  *time.Time     `gorm:"type:timestamp"`
    Version      int            `gorm:"not null;default:1"`
    CreatedAt    time.Time      `gorm:"not null;autoCreateTime"`
    UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime"`
    DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
    return "users"
}