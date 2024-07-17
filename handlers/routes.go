package handlers

import "github.com/labstack/echo/v4"

func SetupRoutes(e *echo.Echo, ph *PriceCalcHandler) {
	e.GET("/", ph.index)
	e.PUT("/ingredient", ph.putIngredient)
	e.PUT("/ingredient-price/:ingredient-id", ph.putIngredientPrice)
	e.DELETE("/ingredient/:ingredient-id", ph.deleteIngredient)
	e.GET("/categories", ph.categories)
	e.GET("/products", ph.products)
	e.PUT("/category", ph.putCategory)
	e.GET("/category/:category-id", ph.getCategory)
	e.GET("/category/:category-id/edit", ph.getCategoryEdit)
	e.PUT("/category/:category-id", ph.updateCategory)
}
