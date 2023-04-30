package model

import (
	"gorm.io/gorm"
	"time"
)

// 定义产品结构体
type Product struct {
	ID           uint    `gorm:"primaryKey"`
	ProductID    int     `gorm:"not null"`
	ProductName  string  `gorm:"not null"`
	Description  string  `gorm:"default:''"`
	Price        float64 `gorm:"type:decimal(10,2);default:0.00;not null"`
	Stock        int     `gorm:"default:0;not null"`
	Image        string  `gorm:"default:''"`
	ProductValue int     `gorm:"default:0;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Status       int `gorm:"default:0" gorm:"comment:'产品状态：0-可用，1-不可用'"`
}

type ProductVO struct {
	Products []struct {
		ID          uint
		Price       float64
		ProductName string
	}
}

type Order struct {
	gorm.Model
	UserID     uint64  `gorm:"not null" json:"user_id"`
	OrderNo    string  `gorm:"unique;not null" json:"order_no"`
	ProductID  uint64  `gorm:"default:0;not null" json:"product_id"`
	Status     int     `gorm:"default:0;not null" json:"status"`
	TotalPrice float64 `gorm:"not null" json:"total_price"`
	Source     string  `json:"source"`
	PayType    int     `json:"pay_type"`
	QrUrl      string  ` json:"qr_url"`
}

type OrderVO struct {
	Status     int
	QR         string
	QrURL      string
	TotalPrice float64
	OrderNo    string
	CreatedAt  string
}
