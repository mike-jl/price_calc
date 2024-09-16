package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mike-jl/price_calc/handlers"
	"github.com/mike-jl/price_calc/services"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	app := echo.New()
	app.Static("/", "assets")
	app.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	service, err := services.NewPriceCalcService(logger, "test123.db")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(-1)
	}

	handler := handlers.NewPriceCalcHandler(logger, service)
	handlers.SetupRoutes(app, handler)

	app.Logger.Fatal(app.Start("localhost:42069"))
}
