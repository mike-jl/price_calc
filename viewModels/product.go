package viewmodels

import "github.com/mike-jl/price_calc/db"

type ProductEditViewModel struct {
	Product          db.ProductWithCost        `json:"product"`
	Categories       []db.Category             `json:"categories"`
	IngredientUsages []db.IngredientUsage      `json:"ingredient_usages"`
	Ingredients      []db.IngredientWithPrices `json:"ingredients"`
	Units            []db.Unit                 `json:"units"`
}
