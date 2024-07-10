package handlers

import "github.com/labstack/echo/v4"

func SetupRoutes(e *echo.Echo, ph *PriceCalcHandler) {
	e.GET("/", ph.index)
	e.PUT("/base-product", ph.putBaseProduct)
}
