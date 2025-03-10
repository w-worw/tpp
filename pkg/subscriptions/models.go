package subscriptions

import (
	"time"

	"gorm.io/gorm"
)


type User struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Email     string `gorm:"unique;not null"`
	Password  string `gorm:"not null"`
	AvatarURL string `gorm:"not null"`
	PackageID uint
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Subscription struct {
	SubscriptionID     string         `json:"subscription_id" gorm:"primaryKey"`
	UserID             uint           `json:"user_id"`
	StartDate          string         `json:"start_date"`
	EndDate            string         `json:"end_date"`
	PaymentStatus      string         `json:"payment_status"`
	PaymentUrl         string         `json:"payment_url"`
	TypeID             uint           `json:"type_id"`
	CustomerID         string         `json:"customer_id"`
	SubscriptionStatus string         `json:"subscription_status"`
	DeletedAt          gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

type SubscriptionType struct {
	TypeID    uint    `gorm:"primaryKey"`
	Name      string  `gorm:"not null"`
	PriceID   string  `gorm:"not null"`
	Price     float64 `gorm:"not null;default:0"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
