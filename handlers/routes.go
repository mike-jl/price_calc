package handlers

import "github.com/labstack/echo/v4"

func SetupRoutes(e *echo.Echo, ph *PriceCalcHandler) {
	e.GET("/", ph.index)
	e.PUT("/ingredient", ph.putIngredient)
	e.PUT("/ingredient-price/:ingredient-id", ph.putIngredientPrice)
	e.GET("/ingredient/:ingredient-id/edit", ph.getIngredientEdit)
	e.GET("/ingredient/:ingredient-id", ph.getIngredient)
	e.DELETE("/ingredient/:ingredient-id", ph.deleteIngredient)
	e.GET("/categories", ph.categories)
	e.GET("/products", ph.products)
	e.PUT("/category", ph.putCategory)
	e.GET("/category/:category-id", ph.getCategory)
	e.GET("/category/:category-id/edit", ph.getCategoryEdit)
	e.PUT("/category/:category-id", ph.updateCategory)
	e.PUT("/product", ph.putProduct)
	e.GET("/product/:product-id/edit", ph.getProductEditPage)
	e.POST("/product/:product-id", ph.postProduct)
	e.DELETE("/product/:product-id", ph.deleteProduct)
	e.PUT("/ingredient-usage/:product-id", ph.putIngredientUsage)
	e.GET("/ingredient-usage-edit/:ingredient-usage-id", ph.getIngredientUsageEdit)
	e.POST("/ingredient-usage/:ingredient-usage-id", ph.postIngredientUsage)
	e.DELETE("/ingredient-usage/:ingredient-usage-id", ph.deleteIngredientUsage)
	e.GET("/units", ph.getUnits)
	e.PUT("/unit", ph.putUnit)
	e.GET("/unit/:unit-id/edit", ph.getUnitEdit)
	e.POST("/unit/:unit-id", ph.postUnit)
	e.DELETE("/unit/:unit-id", ph.deleteUnit)
}
