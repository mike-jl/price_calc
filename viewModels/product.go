package viewmodels

import "github.com/mike-jl/price_calc/db"

type ProductWithCost struct {
	Product db.Product `json:"product"`
	Cost    float64    `json:"cost"`
}

type ProductEditViewModel struct {
	Product          ProductWithCost                `json:"product"`
	Categories       []db.Category                  `json:"categories"`
	IngredientUsages []db.IngredientUsage           `json:"ingredient_usages"`
	Ingredients      map[int64]IngredientWithPrices `json:"ingredients"`
	Units            map[int64]db.Unit              `json:"units"`
}
