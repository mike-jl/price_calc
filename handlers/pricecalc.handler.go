package handlers

import (
	"log/slog"
	"net/http"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mike-jl/price_calc/components"
	"github.com/mike-jl/price_calc/services"
)

type PriceCalcHandler struct {
	log     *slog.Logger
	service *services.PriceCalcService
}

func NewPriceCalcHandler(log *slog.Logger, service *services.PriceCalcService) *PriceCalcHandler {
	return &PriceCalcHandler{log, service}
}

// This custom render replaces Echo's echo.Context.render() with templ's templ.Component.render().
func render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}

func (ph *PriceCalcHandler) index(c echo.Context) error {
	return render(c, http.StatusOK, components.Index())
}

func (ph *PriceCalcHandler) putBaseProduct(c echo.Context) error {
	name := c.FormValue("name")
	_, err := ph.service.AddBaseProduct(name)
	if err != nil {
		return err
	}
	return nil
}
