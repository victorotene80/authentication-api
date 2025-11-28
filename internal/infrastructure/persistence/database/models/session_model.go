package models

import (
	"time"

	"gorm.io/gorm"
)

type SessionModel struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)"`
	UserID       string         `gorm:"not null;type:varchar(36);index"`
	RefreshToken string         `gorm:"uniqueIndex;not null;type:text"`
	AccessToken  string         `gorm:"not null;type:text"`
	IPAddress    string         `gorm:"type:varchar(45)"`
	UserAgent    string         `gorm:"type:text"`
	ExpiresAt    time.Time      `gorm:"not null;index"`
	IsRevoked    bool           `gorm:"not null;default:false;index"`
	RevokedAt    *time.Time     `gorm:"type:timestamp"`
	CreatedAt    time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`

	User UserModel `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (SessionModel) TableName() string {
	return "sessions"
}
