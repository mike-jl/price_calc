package models

import (
	"time"
)

type Category struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Vat       uint
}

type BaseProduct struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Price     []BaseProductPrice
}

type BaseProductPrice struct {
	ID            uint `gorm:"primaryKey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	BaseProductID uint
	Price         float64
	Quantity      float64
}

type Product struct {
	ID         uint `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Name       string
	CategoryID uint
	Category   Category
	Ingedients []Ingedient
}

type Ingedient struct {
	ID            uint `gorm:"primaryKey"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	BaseProductID uint
	BaseProduct   BaseProduct
	ProductID     uint
	Quantity      float64
}
