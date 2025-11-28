package models

import (
    "time"
    "gorm.io/gorm"
)

type OAuthTokenModel struct {
    ID           string         `gorm:"primaryKey;type:varchar(36)"`
    UserID       string         `gorm:"not null;type:varchar(36);index"`
    Provider     string         `gorm:"not null;type:varchar(50);index"`
    AccessToken  string         `gorm:"not null;type:text"`
    RefreshToken string         `gorm:"type:text"`
    TokenType    string         `gorm:"type:varchar(50)"`
    ExpiresAt    time.Time      `gorm:"not null;index"`
    Scope        string         `gorm:"type:text"`
    IsRevoked    bool           `gorm:"not null;default:false;index"`
    RevokedAt    *time.Time     `gorm:"type:timestamp"`
    CreatedAt    time.Time      `gorm:"not null;autoCreateTime"`
    UpdatedAt    time.Time      `gorm:"not null;autoUpdateTime"`
    DeletedAt    gorm.DeletedAt `gorm:"index"`
    
    User         UserModel      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (OAuthTokenModel) TableName() string {
    return "oauth_tokens"
}

/*func (OAuthTokenModel) BeforeSave(tx *gorm.DB) error {
    // Add unique constraint on (user_id, provider) combination
    return nil
}*/