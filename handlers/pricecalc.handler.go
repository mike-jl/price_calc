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
	"github.com/mike-jl/price_calc/internal/utils"
	"github.com/mike-jl/price_calc/services"
	viewmodels "github.com/mike-jl/price_calc/viewModels"
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

func (ph *PriceCalcHandler) getIngredients(c echo.Context) error {
	ingredients, err := ph.service.GetIngredientsWithPrice(c.Request().Context())
	if err != nil {
		ph.log.Error(err.Error())
		return c.NoContent(http.StatusInternalServerError)
	}
	products, err := ph.service.GetProductNames(c.Request().Context())
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get products "+err.Error())
	}

	units, err := ph.service.GetUnitsMap(c.Request().Context())
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	ph.log.Info("get ingredients", "ingredients", ingredients, "products", products, "units", units)

	// Convert the slice of db.IngredientWithPrices to a slice of viewmodels.IngredientWithPrice
	ingredientsWithPrice := make([]viewmodels.IngredientWithPrice, len(ingredients))
	for i, ingredient := range ingredients {
		if len(ingredient.Prices) > 0 {
			ingredientsWithPrice[i] = viewmodels.IngredientWithPrice{
				ID:    ingredient.Ingredient.ID,
				Name:  ingredient.Ingredient.Name,
				Price: ingredient.Prices[0],
			}
		}
	}

	viewModel := viewmodels.IngredientsViewModel{
		Ingredients:  ingredientsWithPrice,
		Units:        units,
		ProductNames: products,
	}

	return render(
		c,
		http.StatusOK,
		components.Index(
			components.Ingredients(viewModel),
		),
	)
}

func (ph *PriceCalcHandler) postIngredient(c echo.Context) error {
	name := strings.TrimSpace(c.FormValue("name"))
	if name == "" {
		return c.String(http.StatusBadRequest, "ingredient name is empty")
	}

	var price *float64 = nil
	var baseProductId *int64 = nil

	ingType := strings.TrimSpace(c.FormValue("type"))
	switch ingType {
	case "price":
		priceValue, err := strconv.ParseFloat(c.FormValue("price"), 64)
		if err != nil {
			return c.String(http.StatusBadRequest, "could not parse price "+err.Error())
		}
		price = &priceValue
	case "product":
		baseProductIdValue, err := strconv.ParseInt(c.FormValue("base-product-id"), 10, 64)
		if err != nil {
			c.String(http.StatusBadRequest, "could not parse base product id "+err.Error())
		}
		baseProductId = &baseProductIdValue
	default:
		return c.String(http.StatusBadRequest, "invalid ingredient type "+ingType)
	}

	quantity, err := strconv.ParseFloat(c.FormValue("quantity"), 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse quantity "+err.Error())
	}

	unitId, err := strconv.ParseInt(c.FormValue("unit"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse unit id "+err.Error())
	}

	ingredient, err := ph.service.NewIngredient(
		c.Request().Context(),
		services.UpdateIngredientParams{
			ID:            0,
			Name:          name,
			Price:         price,
			Quantity:      quantity,
			UnitID:        unitId,
			BaseProductID: baseProductId,
		},
	)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not insert ingredient "+err.Error())
	}

	ingredientWithPrice := viewmodels.IngredientWithPrice{
		ID:    ingredient.Ingredient.ID,
		Name:  ingredient.Ingredient.Name,
		Price: ingredient.Prices[0],
	}

	return render(c, http.StatusCreated, components.NewIngredient(ingredientWithPrice))
}

func (ph *PriceCalcHandler) postIngredientPrice(c echo.Context) error {
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

	_, err = ph.service.UpdateIngredientWithPrice(
		c.Request().Context(),
		services.UpdateIngredientParams{
			ID:            ingredientId,
			Name:          name,
			Price:         pricePtr,
			Quantity:      quantity,
			UnitID:        unitId,
			BaseProductID: baseProductIdPtr,
		},
	)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not update ingredient "+err.Error())
	}

	return c.NoContent(http.StatusOK)
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
	productWithCost := viewmodels.ProductWithCost{Product: *product, Cost: 0}
	categories, err := ph.service.GetCategories()
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get categories "+err.Error())
	}
	return render(c, http.StatusOK, components.ProductRow(productWithCost, categories))
}

func (ph *PriceCalcHandler) getProductEditPage(c echo.Context) error {
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
	ingredients, err := ph.service.GetIngredientsWithPrice(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredients "+err.Error())
	}

	ingredientsMap := make(map[int64]viewmodels.IngredientWithPrices, len(ingredients))
	for _, ingredient := range ingredients {
		ingredientsMap[ingredient.Ingredient.ID] = ingredient
	}

	units, err := ph.service.GetUnitsMap(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	viewModel := viewmodels.ProductEditViewModel{
		Product:          *productWithCost,
		Categories:       categories,
		IngredientUsages: ingredientUsage,
		Ingredients:      ingredientsMap,
		Units:            units,
	}

	return render(
		c,
		http.StatusOK,
		components.Index(
			components.ProductEdit(
				viewModel,
			),
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
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not update product "+err.Error())
	}

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
	circ, err := ph.service.CheckCircularDependency(productId, ingredientId, c.Request().Context())
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

	units, err := ph.service.GetUnitsMap(c.Request().Context())
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	baseQuantity := quantity / units[unitId].Factor
	ingredientUsage, err := ph.service.PutIngredientUsage(
		c.Request().Context(),
		ingredientId,
		productId,
		unitId,
		baseQuantity,
	)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not insert ingredient usage "+err.Error()+" values: "+
				"ingredient id: "+strconv.FormatInt(ingredientId, 10)+
				"product id: "+strconv.FormatInt(productId, 10)+
				"unit id: "+strconv.FormatInt(unitId, 10)+
				"base quantity: "+strconv.FormatFloat(baseQuantity, 'f', -1, 64),
		)
	}
	return render(
		c,
		http.StatusOK,
		components.NewIngredientUsage(*ingredientUsage),
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

	ingredient, err := ph.service.GetIngredientWithPrice(c.Request().Context(), ingredientId)
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get ingredient "+err.Error())
	}

	baseUnitID := ingredient.Prices[0].UnitID

	units, err := ph.service.GetUnitsMap(c.Request().Context())
	if err != nil {
		ph.log.Error(err.Error())
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	if units[baseUnitID].BaseUnitID != nil {
		baseUnitID = *units[baseUnitID].BaseUnitID
	}
	filteredUnits := services.UnitsMap{}

	// filteredUnits = append(filteredUnits, ph.service.Units[baseUnitID])
	filteredUnits[baseUnitID] = units[baseUnitID]
	for id, unit := range units {
		if unit.BaseUnitID != nil && *unit.BaseUnitID == baseUnitID {
			// filteredUnits = append(filteredUnits, unit)
			filteredUnits[id] = unit
		}
	}
	return render(
		c,
		http.StatusOK,
		components.IngredientUsageRowEdit(),
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
	// ph.log.Info("post ingredient usage", "unitId", unitId, "quantity", quantity)
	// return c.String(http.StatusOK, "could not parse quantity ")
	_, err = ph.service.UpdateIngredientUsage(
		ingredientUsageId,
		unitId,
		quantity,
		c.Request().Context(),
	)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not update ingredient usage "+err.Error(),
		)
	}

	return c.NoContent(http.StatusOK)
}

func (ph *PriceCalcHandler) deleteIngredientUsage(c echo.Context) error {
	ingredientUsageId, err := strconv.ParseInt(c.Param("ingredient-usage-id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "could not parse ingredient usage id "+err.Error())
	}
	err = ph.service.DeleteIngredientUsage(c.Request().Context(), ingredientUsageId)
	if err != nil {
		return c.String(
			http.StatusInternalServerError,
			"could not delete ingredient usage "+err.Error(),
		)
	}
	return c.NoContent(http.StatusOK)
}

func (ph *PriceCalcHandler) getUnits(c echo.Context) error {
	units, err := ph.service.GetUnits(c.Request().Context())
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
		units, err := ph.service.GetUnits(c.Request().Context())
		if err != nil {
			return c.String(http.StatusInternalServerError, "could not get base unit "+err.Error())
		}
		if bunit, ok := utils.First(units, func(u db.Unit) bool {
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

	units, err := ph.service.GetUnits(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	unit, ok := utils.First(units, func(u db.Unit) bool {
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

	units, err := ph.service.GetUnits(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "could not get units "+err.Error())
	}

	unit, ok := utils.First(units, func(u db.Unit) bool {
		return u.ID == unitId
	})
	if !ok {
		return c.String(
			http.StatusNotFound,
			"could not find unit with id "+strconv.FormatInt(unitId, 10),
		)
	}

	if baseUnitIdPtr != nil {
		if _, ok := utils.First(units, func(u db.Unit) bool {
			return u.ID == *baseUnitIdPtr
		}); !ok {
			return c.String(
				http.StatusBadRequest,
				"could not find base unit with id "+strconv.FormatInt(*baseUnitIdPtr, 10),
			)
		}
	}

	if baseUnitIdPtr != nil && unit.BaseUnitID == nil {
		dependentUnits := utils.Where(units, func(u db.Unit) bool {
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
		if bunit, ok := utils.First(units, func(u db.Unit) bool {
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
