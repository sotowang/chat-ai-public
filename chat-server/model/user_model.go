package model

import "time"

type VerificationCode struct {
	ID         uint `gorm:"primaryKey"`
	UserID     uint
	Code       string
	ExpireTime time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
