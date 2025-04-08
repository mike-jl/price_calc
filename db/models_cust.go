package db

type IngredientWithPrices struct {
	Ingredient Ingredient        `json:"ingredient"`
	Prices     []IngredientPrice `json:"prices"`
}

type ProductWithCost struct {
	Product Product `json:"product"`
	Cost    float64 `json:"cost"`
}
