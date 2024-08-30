package db

type IngredientWithPrice struct {
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

type ProductWithPrice struct {
	Product Product
	Price   float64
}
