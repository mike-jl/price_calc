package viewmodels

import "github.com/mike-jl/price_calc/db"

type IngredientWithPrice struct {
	ID    int64              `json:"id"`
	Name  string             `json:"name"`
	Price db.IngredientPrice `json:"price"`
}

type IngredientsViewModel struct {
	Ingredients  []IngredientWithPrice `json:"ingredients"`
	Units        []db.Unit             `json:"units"`
	ProductNames map[int64]string      `json:"product_names"`
}
