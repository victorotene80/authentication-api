package models

import (
    "time"
    "gorm.io/gorm"
    "gorm.io/datatypes"
)

type AuditLogModel struct {
    ID           string         `gorm:"primaryKey;type:varchar(36)"`
    UserID       string         `gorm:"not null;type:varchar(36);index"`
    Action       string         `gorm:"not null;type:varchar(100);index"`
    ResourceType string         `gorm:"not null;type:varchar(100);index"`
    ResourceID   string         `gorm:"type:varchar(36);index"`
    IPAddress    string         `gorm:"type:varchar(45)"`
    UserAgent    string         `gorm:"type:text"`
    Status       string         `gorm:"not null;type:varchar(20);index"`
    ErrorMessage string         `gorm:"type:text"`
    Metadata     datatypes.JSON `gorm:"type:json"`
    Timestamp    time.Time      `gorm:"not null;index"`
    CreatedAt    time.Time      `gorm:"not null;autoCreateTime"`
    DeletedAt    gorm.DeletedAt `gorm:"index"`
    
    User         UserModel      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (AuditLogModel) TableName() string {
    return "audit_logs"
}