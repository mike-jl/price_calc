package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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
	products, err := ph.service.GetProductNames()
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get products "+err.Error())
	}

	return render(
		c,
		http.StatusOK,
		components.Index(components.IngredientsTable(ingredients, ph.service.Units, products)),
	)
}

func (ph *PriceCalcHandler) putIngredient(c echo.Context) error {
	ph.log.Info("put ingredient")
	name := c.FormValue("name")
	ingredient, err := ph.service.PutIngredient(name)
	if err != nil {
		return err
	}
	ingredientWithPrice := db.IngredientWithPrices{
		Ingredient: db.Ingredient{
			ID:   ingredient.ID,
			Name: ingredient.Name,
		},
	}
	products, err := ph.service.GetProductNames()
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get products "+err.Error())
	}
	return render(
		c,
		http.StatusCreated,
		components.IngredientRow(ingredientWithPrice, ph.service.Units, products),
	)
}

func (ph *PriceCalcHandler) putIngredientPrice(c echo.Context) error {
	ingredientId, err := strconv.ParseInt(c.Param("ingredient-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient id "+err.Error())
	}
	quantity, err := strconv.ParseFloat(c.FormValue("quantity"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse quantity "+err.Error())
	}
	unitId, err := strconv.ParseInt(c.FormValue("unit"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}

	ingType := c.FormValue("type")
	if ingType != "price" && ingType != "product" {
		return c.String(http.StatusBadRequest, "could not parse type "+ingType)
	}
	price := float64(0)
	pricePtr := &price
	baseProductId := int64(0)
	baseProductIdPtr := &baseProductId

	switch ingType {
	case "price":
		price, err = strconv.ParseFloat(c.FormValue("price"), 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "could not parse price "+err.Error())
		}
		baseProductIdPtr = nil
	case "product":
		baseProductId, err = strconv.ParseInt(c.FormValue("base-product-id"), 10, 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "could not parse base product id "+err.Error())
		}
		pricePtr = nil
	}

	name := c.FormValue("name")

	ingredientWithPrice, err := ph.service.UpdateIngredientWithPrice(
		ingredientId, name, pricePtr, quantity, unitId, baseProductIdPtr, c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not update ingredient "+err.Error())
	}

	products, err := ph.service.GetProductNames()
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get products "+err.Error())
	}

	return render(
		c,
		http.StatusCreated,
		components.IngredientRow(*ingredientWithPrice, ph.service.Units, products),
	)
}

func (ph *PriceCalcHandler) getIngredientEdit(c echo.Context) error {
	ph.log.Info("get ingredient edit")
	ingredientId, err := strconv.ParseInt(c.Param("ingredient-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient id "+err.Error())
	}

	ingredient, err := ph.service.GetIngredientWithPrices(ingredientId, 10)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredient "+err.Error())
	}

	products, err := ph.service.GetProductNames()
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get products "+err.Error())
	}

	return render(
		c,
		http.StatusOK,
		components.IngredientRowEdit(*ingredient, ph.service.Units, products),
	)
}

func (ph *PriceCalcHandler) deleteIngredient(c echo.Context) error {
	ingredientId, err := strconv.ParseInt(c.Param("ingredient-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient id "+err.Error())
	}

	// Check if the ingredient is used in any product
	products, err := ph.service.GetProductsWithIngredient(ingredientId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get products "+err.Error())
	}

	if len(products) > 0 {
		return c.String(
			http.StatusConflict,
			"Cannot delete ingredient because its still used in the following products:\n"+strings.Join(
				products,
				", ",
			),
		)
	}

	err = ph.service.DeleteIngredient(ingredientId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not delete ingredient "+err.Error())
	}

	return c.String(http.StatusOK, "")
}

func (ph *PriceCalcHandler) products(c echo.Context) error {
	products, err := ph.service.GetProductsWithCost()
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
	category, err := ph.service.PutCategory(name, vat)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not insert category "+err.Error())
	}

	return render(c, http.StatusOK, components.CategoryRow(*category))
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
	productWithCost := db.ProductWithCost{Product: *product, Cost: 0}
	categories, err := ph.service.GetCategories()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get categories "+err.Error())
	}
	return render(c, http.StatusOK, components.ProductRow(productWithCost, categories))
}

func (ph *PriceCalcHandler) getProductEdit(c echo.Context) error {
	productId, err := strconv.ParseInt(c.Param("product-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse product id "+err.Error())
	}
	productWithCost, err := ph.service.GetProductWithCost(productId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get product "+err.Error())
	}
	ingredientUsage, err := ph.service.GetIngredientUsageForProduct(productId)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not get ingredient usage "+err.Error(),
		)
	}

	categories, err := ph.service.GetCategories()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get categories "+err.Error())
	}
	ingredients, err := ph.service.GetIngredientsWithPrice()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredients "+err.Error())
	}
	return render(
		c,
		http.StatusOK,
		components.ProductRowEdit(
			*productWithCost,
			categories,
			ingredientUsage,
			ingredients,
			ph.service.Units,
		),
	)
}

func (ph *PriceCalcHandler) postProduct(c echo.Context) error {
	productId, err := strconv.ParseInt(c.Param("product-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse product id "+err.Error())
	}

	name := c.FormValue("name")
	price, err := strconv.ParseFloat(c.FormValue("price"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse price "+err.Error())
	}
	multiplicator, err := strconv.ParseFloat(c.FormValue("multiplicator"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse multiplicator "+err.Error())
	}
	categoryId, err := strconv.ParseInt(c.FormValue("category"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse category id "+err.Error())
	}
	_, err = ph.service.UpdateProduct(productId, categoryId, price, multiplicator, name)

	product, err := ph.service.GetProductWithCost(productId)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not get updated product "+err.Error(),
		)
	}
	categories, err := ph.service.GetCategories()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get categories "+err.Error())
	}

	return render(c, http.StatusOK, components.ProductRow(*product, categories))
}

func (ph *PriceCalcHandler) deleteProduct(c echo.Context) error {
	productId, err := strconv.ParseInt(c.Param("product-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse product id "+err.Error())
	}

	err = ph.service.DeleteProduct(productId)
	if err != nil {
		return c.String(http.StatusBadRequest, "error when deleting product "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (ph *PriceCalcHandler) getUnitListFiltered(c echo.Context) error {
	unitId, err := strconv.ParseInt(c.FormValue("unitId"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}
	baseUnitID := unitId
	if ph.service.Units[unitId].BaseUnitID != nil {
		baseUnitID = *ph.service.Units[unitId].BaseUnitID
	}
	filteredUnits := []db.Unit{}
	filteredUnits = append(filteredUnits, ph.service.Units[baseUnitID])
	for _, unit := range ph.service.Units {
		if unit.BaseUnitID != nil && *unit.BaseUnitID == baseUnitID {
			filteredUnits = append(filteredUnits, unit)
		}
	}
	return render(c, http.StatusOK, components.UnitSelect(filteredUnits, unitId))
}

func (ph *PriceCalcHandler) putIngredientUsage(c echo.Context) error {
	productId, err := strconv.ParseInt(c.Param("product-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse product id "+err.Error())
	}
	ingredientId, err := strconv.ParseInt(c.FormValue("ingredient"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient id "+err.Error())
	}
	unitId, err := strconv.ParseInt(c.FormValue("unit"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}
	quantity, err := strconv.ParseFloat(c.FormValue("amount"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse quantity "+err.Error())
	}

	// check for circular dependencies
	circ, err := ph.service.CheckCircularDependency(productId, ingredientId)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not check circular dependency "+err.Error(),
		)
	}
	if circ {
		return c.String(
			http.StatusConflict,
			"Can't add ingredient usage because it would create a circular dependency!",
		)
	}

	baseQuantity := quantity / ph.service.Units[unitId].Factor
	ingredientUsage, err := ph.service.PutIngredientUsage(
		ingredientId,
		productId,
		unitId,
		baseQuantity,
	)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not insert ingredient usage "+err.Error(),
		)
	}
	ingredient, err := ph.service.GetIngredientWithPrice(ingredientId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredient "+err.Error())
	}
	return render(
		c,
		http.StatusOK,
		components.IngredientUsageRow(*ingredientUsage, *ingredient, ph.service.Units[unitId]),
	)
}

func (ph *PriceCalcHandler) getIngredientUsageEdit(c echo.Context) error {
	ingredientUsageId, err := strconv.ParseInt(c.Param("ingredient-usage-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient usage id "+err.Error())
	}

	ingredientUsage, err := ph.service.GetIngredientUsage(ingredientUsageId)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not get ingredient usage "+err.Error(),
		)
	}

	ingredientId := ingredientUsage.IngredientID

	ingredient, err := ph.service.GetIngredientWithPrice(ingredientId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredient "+err.Error())
	}

	baseUnitID := ingredient.Prices[0].UnitID
	if ph.service.Units[baseUnitID].BaseUnitID != nil {
		baseUnitID = *ph.service.Units[baseUnitID].BaseUnitID
	}
	filteredUnits := services.UnitsMap{}

	// filteredUnits = append(filteredUnits, ph.service.Units[baseUnitID])
	filteredUnits[baseUnitID] = ph.service.Units[baseUnitID]
	for id, unit := range ph.service.Units {
		if unit.BaseUnitID != nil && *unit.BaseUnitID == baseUnitID {
			// filteredUnits = append(filteredUnits, unit)
			filteredUnits[id] = unit
		}
	}
	return render(
		c,
		http.StatusOK,
		components.IngredientUsageRowEdit(*ingredientUsage, *ingredient, filteredUnits),
	)
}

func (ph *PriceCalcHandler) postIngredientUsage(c echo.Context) error {
	ingredientUsageId, err := strconv.ParseInt(c.Param("ingredient-usage-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient usage id "+err.Error())
	}
	unitId, err := strconv.ParseInt(c.FormValue("unit"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}
	quantity, err := strconv.ParseFloat(c.FormValue("amount"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse quantity "+err.Error())
	}
	ingredientUsage, err := ph.service.UpdateIngredientUsage(ingredientUsageId, unitId, quantity)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not update ingredient usage "+err.Error(),
		)
	}
	ingredientId := ingredientUsage.IngredientID

	ingredient, err := ph.service.GetIngredientWithPrice(ingredientId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredient "+err.Error())
	}
	return render(
		c,
		http.StatusOK,
		components.IngredientUsageRow(*ingredientUsage, *ingredient, ph.service.Units[unitId]),
	)
}

func (ph *PriceCalcHandler) deleteIngredientUsage(c echo.Context) error {
	ingredientUsageId, err := strconv.ParseInt(c.Param("ingredient-usage-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient usage id "+err.Error())
	}
	err = ph.service.DeleteIngredientUsage(ingredientUsageId)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not delete ingredient usage "+err.Error(),
		)
	}
	return c.NoContent(http.StatusOK)
}

func (ph *PriceCalcHandler) getUnits(c echo.Context) error {
	units, err := ph.service.GetUnits()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}
	return render(c, http.StatusOK, components.Index(components.UnitsTable(units)))
}

func (ph *PriceCalcHandler) putUnit(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.String(http.StatusBadRequest, "unit name is empty")
	}
	baseUnitId, err := strconv.ParseInt(c.FormValue("base-unit-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse base unit id "+err.Error())
	}

	factor := float64(1)
	if baseUnitId != 0 {
		factor, err = strconv.ParseFloat(c.FormValue("factor"), 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "could not parse factor "+err.Error())
		}
		if factor <= 0 {
			return c.String(http.StatusBadRequest, "factor must be greater than 0")
		}
	}

	var baseUnitIdPtr *int64 = nil
	if baseUnitId != 0 {
		baseUnitIdPtr = &baseUnitId
	}

	newUnit, err := ph.service.InsertUnit(name, baseUnitIdPtr, factor, c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not insert unit "+err.Error())
	}

	var baseUnit *db.Unit = nil
	if baseUnitId != 0 {
		units, err := ph.service.GetUnits()
		if err != nil {
			return c.String(http.StatusInternalServerError, "could not get base unit "+err.Error())
		}
		if bunit, ok := services.First(units, func(u db.Unit) bool {
			return u.ID == baseUnitId
		}); ok {
			baseUnit = &bunit
		}
	}

	return render(c, http.StatusOK, components.UnitRow(*newUnit, baseUnit))
}

func (ph *PriceCalcHandler) getUnitEdit(c echo.Context) error {
	unitId, err := strconv.ParseInt(c.Param("unit-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}

	units, err := ph.service.GetUnits()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	unit, ok := services.First(units, func(u db.Unit) bool {
		return u.ID == unitId
	})
	if !ok {
		return c.String(
			http.StatusNotFound,
			"could not find unit with id "+strconv.FormatInt(unitId, 10),
		)
	}

	return render(c, http.StatusOK, components.UnitRowEdit(unit, units))
}

func (ph *PriceCalcHandler) postUnit(c echo.Context) error {
	unitId, err := strconv.ParseInt(c.Param("unit-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}

	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.String(http.StatusBadRequest, "unit name is empty")
	}
	baseUnitId, err := strconv.ParseInt(c.FormValue("base-unit-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse base unit id "+err.Error())
	}

	factor := float64(1)
	if baseUnitId != 0 {
		factor, err = strconv.ParseFloat(c.FormValue("factor"), 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "could not parse factor "+err.Error())
		}
		if factor <= 0 {
			return c.String(http.StatusBadRequest, "factor must be greater than 0")
		}
	}

	var baseUnitIdPtr *int64 = nil
	if baseUnitId != 0 {
		baseUnitIdPtr = &baseUnitId
	}

	units, err := ph.service.GetUnits()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	unit, ok := services.First(units, func(u db.Unit) bool {
		return u.ID == unitId
	})
	if !ok {
		return c.String(
			http.StatusNotFound,
			"could not find unit with id "+strconv.FormatInt(unitId, 10),
		)
	}

	if baseUnitIdPtr != nil {
		if _, ok := services.First(units, func(u db.Unit) bool {
			return u.ID == *baseUnitIdPtr
		}); !ok {
			return c.String(
				http.StatusBadRequest,
				"could not find base unit with id "+strconv.FormatInt(*baseUnitIdPtr, 10),
			)
		}
	}

	if baseUnitIdPtr != nil && unit.BaseUnitID == nil {
		dependentUnits := services.Where(units, func(u db.Unit) bool {
			return u.BaseUnitID != nil && *u.BaseUnitID == unitId
		})
		if len(dependentUnits) > 0 {
			return c.String(
				http.StatusBadRequest,
				"could not remove base unit because there are dependent units",
			)
		}
	}

	newUnit, err := ph.service.UpdateUnit(
		unitId,
		name,
		baseUnitIdPtr,
		factor,
		c.Request().Context(),
	)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not update unit "+err.Error())
	}

	baseUnit := &db.Unit{}
	if baseUnitId != 0 {
		if bunit, ok := services.First(units, func(u db.Unit) bool {
			return u.ID == baseUnitId
		}); ok {
			baseUnit = &bunit
		} else {
			return c.String(
				http.StatusInternalServerError,
				"could not find base unit with id "+strconv.FormatInt(baseUnitId, 10),
			)
		}
	}

	return render(c, http.StatusOK, components.UnitRow(*newUnit, baseUnit))
}

func (ph *PriceCalcHandler) deleteUnit(c echo.Context) error {
	unitId, err := strconv.ParseInt(c.Param("unit-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}

	// Check if the unit is used in any ingredient
	ingredients, err := ph.service.GetIngredientsFromUnit(unitId, c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredients "+err.Error())
	}
	if len(ingredients) > 0 {
		return c.String(
			http.StatusConflict,
			"Cannot delete unit because its still used in the following ingredients:\n"+strings.Join(
				ingredients,
				", ",
			),
		)
	}

	// check if unit is used in any product
	products, err := ph.service.GetProductsFromUnit(unitId, c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get products "+err.Error())
	}
	if len(products) > 0 {
		return c.String(http.StatusConflict,
			"Cannot delete unit because its still used in the following products:\n"+strings.Join(
				products,
				", ",
			),
		)
	}

	err = ph.service.DeleteUnit(unitId, c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not delete unit "+err.Error())
	}
	return c.NoContent(http.StatusOK)
}
