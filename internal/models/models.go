package models

import (
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID           uint           `gorm:"primaryKey"`
	UserID       uint           `gorm:"not null"`
	RefreshToken string         `gorm:"not null"`
	ExpiredDate  string         `gorm:"type:date"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	AvatarURL string `gorm:"not null"`
	PackageID uint
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Package struct {
	ID                 uint   `gorm:"primaryKey"`
	Name               string `gorm:"not null"`
	MaximumProjectSlot int    `gorm:"not null"`
	Price              int
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

type ProjectStatus struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Project struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	PublishedAt string `gorm:"type:date"`
	StatusID    uint
	DeletedAt   gorm.DeletedAt `gorm:"index"`
}

type ProjectAttribute struct {
	ID        uint   `gorm:"primaryKey"`
	ProjectID uint   `gorm:"not null"`
	Key       string `gorm:"not null"`
	Value     string
	Type      string
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ProjectManagement struct {
	ID        uint           `gorm:"primaryKey"`
	UserID    uint           `gorm:"not null"`
	ProjectID uint           `gorm:"not null"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ProjectManagementAttribute struct {
	ID                  uint   `gorm:"primaryKey"`
	ProjectManagementID uint   `gorm:"not null"`
	Key                 string `gorm:"not null"`
	Value               string
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

type Activity struct {
	ID            uint   `gorm:"primaryKey"`
	ProjectID     uint   `gorm:"not null"`
	Name          string `gorm:"not null"`
	TypeID        uint
	StatusID      uint
	StartDate     string `gorm:"type:date"`
	EndDate       string `gorm:"type:date"`
	EstimatedTime int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

type ActivityType struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ActivityStatus struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ActivityCost struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"not null"`
	Cost       float64
	CostTypeID uint
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

type ActivityCostType struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type ActivityTracking struct {
	ID                  uint `gorm:"primaryKey"`
	ProjectManagementID uint `gorm:"not null"`
	ActivityID          uint `gorm:"not null"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	DeletedAt           gorm.DeletedAt `gorm:"index"`
}

type ActivityTrackingCost struct {
	ID                 uint `gorm:"primaryKey"`
	ActivityTrackingID uint `gorm:"not null"`
	Amount             int
	CostTypeID         uint
	CreatedAt          time.Time
	UpdatedAt          time.Time
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}
