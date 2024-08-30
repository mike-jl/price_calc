package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mike-jl/price_calc/components"
	"github.com/mike-jl/price_calc/db"
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
	ingredients, err := ph.service.GetIngredientsWithPrice()
	if err != nil {
		ph.log.Error(err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	// ph.log.Info(fmt.Sprintf("aaaaaaaaa: %+v\n", ingredients))
	return render(
		c,
		http.StatusOK,
		components.Index(components.IngredientsTable(ingredients, ph.service.Units)),
	)
}

func (ph *PriceCalcHandler) putIngredient(c echo.Context) error {
	name := c.FormValue("name")
	ingredient, err := ph.service.PutIngredient(name)
	if err != nil {
		return err
	}
	ingredientWithPrice := db.IngredientWithPrice{Ingredient: *ingredient}
	return render(
		c,
		http.StatusCreated,
		components.IngredientOob(ingredientWithPrice, ph.service.Units),
	)
}

func (ph *PriceCalcHandler) putIngredientPrice(c echo.Context) error {
	ingredientId, err := strconv.ParseInt(c.Param("ingredient-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient id "+err.Error())
	}
	price, err := strconv.ParseFloat(c.FormValue("price"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse price "+err.Error())
	}
	quantity, err := strconv.ParseFloat(c.FormValue("quantity"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse quantity "+err.Error())
	}
	unitId, err := strconv.ParseInt(c.FormValue("unit"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}

	ingredientPrice, err := ph.service.PutIngredientPrice(ingredientId, price, quantity, unitId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not insert price "+err.Error())
	}
	return render(
		c,
		http.StatusCreated,
		components.IngredientPriceOob(*ingredientPrice, ph.service.Units),
	)
}

func (ph *PriceCalcHandler) deleteIngredient(c echo.Context) error {
	ingredientId, err := strconv.ParseInt(c.Param("ingredient-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient id "+err.Error())
	}
	err = ph.service.DeleteIngredient(ingredientId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not delete ingredient "+err.Error())
	}
	return nil
}

func (ph *PriceCalcHandler) products(c echo.Context) error {
	products, err := ph.service.GetProductsWithPrice()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredients "+err.Error())
	}
	categories, err := ph.service.GetCategories()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get categories "+err.Error())
	}
	return render(
		c,
		http.StatusOK,
		components.Index(components.ProductsTable(products, categories)),
	)
}

func (ph *PriceCalcHandler) categories(c echo.Context) error {
	categories, err := ph.service.GetCategories()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get categories "+err.Error())
	}
	return render(c, http.StatusOK, components.Index(components.Categories(categories)))
}

func (ph *PriceCalcHandler) putCategory(c echo.Context) error {
	name := c.FormValue("name")
	vat, err := strconv.ParseInt(c.FormValue("vat"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse vat "+err.Error())
	}
	_, err = ph.service.PutCategory(name, vat)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not insert category "+err.Error())
	}

	return c.NoContent(http.StatusCreated)
}

func (ph *PriceCalcHandler) updateCategory(c echo.Context) error {
	name := c.FormValue("name")
	vat, err := strconv.ParseInt(c.FormValue("vat"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse vat "+err.Error())
	}
	categoryId, err := strconv.ParseInt(c.Param("category-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse category id "+err.Error())
	}
	category, err := ph.service.UpdateCategory(categoryId, name, vat)
	if err != nil {
		return c.String(http.StatusInternalServerError, "couold not update category "+err.Error())
	}
	return render(c, http.StatusCreated, components.CategoryRow(*category))
}

func (ph *PriceCalcHandler) getCategory(c echo.Context) error {
	categoryId, err := strconv.ParseInt(c.Param("category-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse category id "+err.Error())
	}
	category, err := ph.service.GetCategory(categoryId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get category "+err.Error())
	}
	return render(c, http.StatusOK, components.CategoryRow(*category))
}

func (ph *PriceCalcHandler) getCategoryEdit(c echo.Context) error {
	categoryId, err := strconv.ParseInt(c.Param("category-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse category id "+err.Error())
	}
	category, err := ph.service.GetCategory(categoryId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get category "+err.Error())
	}
	return render(c, http.StatusOK, components.CategoryRowEdit(*category))
}

func (ph *PriceCalcHandler) putProduct(c echo.Context) error {
	categoryId, err := strconv.ParseInt(c.FormValue("category-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse category id "+err.Error())
	}
	name := c.FormValue("name")
	product, err := ph.service.PutProduct(name, categoryId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not insert product "+err.Error())
	}
	productWithPrice := db.ProductWithPrice{Product: *product, Price: 0}
	categories, err := ph.service.GetCategories()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get categories "+err.Error())
	}
	return render(c, http.StatusOK, components.ProductRow(productWithPrice, categories))
}
