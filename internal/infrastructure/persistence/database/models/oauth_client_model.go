package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type OAuthClientModel struct {
	ID           string         `gorm:"primaryKey;type:varchar(36)"`
	ClientID     string         `gorm:"uniqueIndex;not null;type:varchar(255)"`
	ClientSecret string         `gorm:"not null;type:text"`
	Provider     string         `gorm:"not null;type:varchar(50);index"`
	Name         string         `gorm:"not null;type:varchar(255)"`
	RedirectURIs datatypes.JSON `gorm:"type:json"`
	Scopes       datatypes.JSON `gorm:"type:json"`
	IsActive     bool           `gorm:"not null;default:true;index"`
	Version      int            `gorm:"not null;default:1"`
	CreatedAt    time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (OAuthClientModel) TableName() string {
	return "oauth_clients"
}
