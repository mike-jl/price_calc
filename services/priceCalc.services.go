package services

import (
	"errors"
	"log/slog"

	"github.com/mike-jl/price_calc/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type PriceCalcService struct {
	db     *gorm.DB
	logger *slog.Logger
}

func NewPriceCalcService(log *slog.Logger, dbName string) (*PriceCalcService, error) {
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{TranslateError: true})
	if err != nil {
		return &PriceCalcService{}, err
	}
	return &PriceCalcService{db, log}, nil
}

func (pc *PriceCalcService) GetBaseProducts() ([]models.BaseProduct, error) {
	var baseProducts []models.BaseProduct
	err := pc.db.Preload("BaseProductPrices").Find(&baseProducts).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return baseProducts, nil
}

func (pc *PriceCalcService) AddBaseProduct(name string) (uint, error) {
	baseProduct := models.BaseProduct{Name: name}
	err := pc.db.Create(&baseProduct).Error
	if err != nil {
		return 0, err
	}
	return baseProduct.ID, nil
}

func (pc *PriceCalcService) AddBaseProductPrice(baseProductId uint, price float64) (uint, error) {
	baseProductPrice := models.BaseProductPrice{BaseProductID: baseProductId, Price: price}
	err := pc.db.Create(&baseProductPrice).Error
	if err != nil {
		return 0, err
	}
	return baseProductPrice.ID, nil
}
