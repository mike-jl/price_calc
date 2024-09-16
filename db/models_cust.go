package db

type IngredientWithPrices struct {
	Ingredient Ingredient
	Prices     []IngredientPrice
}

type IngredientUsageWithPrice struct {
	Ingredient      Ingredient
	IngredientUsage IngredientUsage
	IngredientPrice IngredientPrice
}

type ProductWithIngredient struct {
	Product                  Product
	IngredientUsageWithPrice []IngredientUsageWithPrice
}

type ProductWithCost struct {
	Product Product
	Cost    float64
}
