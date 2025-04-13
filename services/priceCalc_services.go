package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mike-jl/price_calc/db"
	"github.com/mike-jl/price_calc/internal/utils"
)

type UnitsMap map[int64]db.Unit

type PriceCalcService struct {
	queries *db.Queries
	db      *sql.DB
	logger  *slog.Logger

	baseProductPriceResolver baseProductPriceResolver
}

type baseProductPriceResolver interface {
	resolveBaseProductPrices([]db.IngredientWithPrices, context.Context) error
}

var ErrNoRowsAffected = errors.New("no rows affected")

func NewPriceCalcService(log *slog.Logger, dbName string) (*PriceCalcService, error) {
	ctx := context.Background()

	sql, err := sql.Open("sqlite3", "db.sqlite3?_foreign_keys=on")
	if err != nil {
		return nil, err
	}

	provider, err := goose.NewProvider(
		database.DialectSQLite3,
		sql,
		os.DirFS("data/sql/migrations"),
	)
	if err != nil {
		return nil, err
	}

	// List migration sources the provider is aware of.
	log.Info("\n=== migration list ===")
	sources := provider.ListSources()
	for _, s := range sources {
		log.Info(fmt.Sprintf("%-3s %-2v %v\n", s.Type, s.Version, filepath.Base(s.Path)))
	}

	// List status of migrations before applying them.
	stats, err := provider.Status(ctx)
	if err != nil {
		return nil, err
	}
	log.Info("\n=== migration status ===")
	for _, s := range stats {
		log.Info(fmt.Sprintf("%-3s %-2v %v\n", s.Source.Type, s.Source.Version, s.State))
	}

	log.Info("\n=== log migration output  ===")
	results, err := provider.Up(ctx)
	if err != nil {
		return nil, err
	}
	log.Info("\n=== migration results  ===")
	for _, r := range results {
		log.Info(fmt.Sprintf("%-3s %-2v done: %v\n", r.Source.Type, r.Source.Version, r.Duration))
	}

	queries := db.New(sql)

	service := PriceCalcService{
		queries: queries,
		db:      sql,
		logger:  log,
	}

	service.baseProductPriceResolver = &service

	return &service, nil
}

func (pc *PriceCalcService) CheckCircularDependency(
	productId, ingredientId int64,
	ctx context.Context,
) (bool, error) {
	return pc.checkCircular(productId, ingredientId, make(map[int64]bool), ctx)
}

func (pc *PriceCalcService) checkCircular(
	targetProductId, currentIngredientId int64,
	visited map[int64]bool,
	ctx context.Context,
) (bool, error) {
	if visited[currentIngredientId] {
		return false, nil // already checked
	}
	visited[currentIngredientId] = true

	ingredients, err := pc.queries.GetIngredientsWithPriceUnit(
		ctx,
		db.GetIngredientsWithPriceUnitParams{
			IngredientID: currentIngredientId,
			PriceLimit:   1,
		},
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	ingredient := ingredients[0]

	if err != nil {
		return false, err
	}

	if ingredient.BaseProductID == nil {
		return false, nil
	}
	if *ingredient.BaseProductID == targetProductId {
		return true, nil
	}

	return pc.checkCircular(targetProductId, *ingredient.BaseProductID, visited, ctx)
}

func (pc *PriceCalcService) parseIngredientsWithPriceUnitRow(
	ctx context.Context,
	ingredients []db.GetIngredientsWithPriceUnitRow,
) ([]db.IngredientWithPrices, error) {
	out := []db.IngredientWithPrices{}
	for _, ingredientRow := range ingredients {

		var target *db.IngredientWithPrices
		// check if the ingredient is already in the out struct
		ing, ok := utils.FirstPtr(out, func(ip db.IngredientWithPrices) bool {
			return ingredientRow.ID == ip.Ingredient.ID
		})
		if ok {
			target = ing
		} else {
			target = utils.AppendAndGetPtr(&out, db.IngredientWithPrices{
				Ingredient: db.Ingredient{ID: ingredientRow.ID, Name: ingredientRow.Name},
			})
		}

		// check if a price is in the row, if so, fill out price struct
		if ingredientRow.PriceID != nil &&
			ingredientRow.TimeStamp != nil &&
			ingredientRow.Quantity != nil &&
			ingredientRow.UnitID != nil {
			target.Prices = append(target.Prices, db.IngredientPrice{
				ID:            *ingredientRow.PriceID,
				TimeStamp:     *ingredientRow.TimeStamp,
				IngredientID:  ingredientRow.ID,
				Price:         ingredientRow.Price,
				Quantity:      *ingredientRow.Quantity,
				UnitID:        *ingredientRow.UnitID,
				BaseProductID: ingredientRow.BaseProductID,
			})
		} else if ingredientRow.PriceID != nil {
			return nil, fmt.Errorf("missing fields in ingredient price row: %d", ingredientRow.ID)
		}

	}

	err := pc.baseProductPriceResolver.resolveBaseProductPrices(out, ctx)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (pc *PriceCalcService) resolveBaseProductPrices(
	ingredients []db.IngredientWithPrices,
	ctx context.Context,
) error {
	// check if the ingredient has a base product
	for i, ingredient := range ingredients {
		if len(ingredient.Prices) <= 0 {
			continue
		}
		for j, price := range ingredient.Prices {
			if price.BaseProductID != nil {
				prodCost, err := pc.queries.GetProductCost(ctx, *price.BaseProductID)
				if err == sql.ErrNoRows {
					// this should not happen, but if it does, calculate the cost and create the row
					var cost float64
					cost, err = pc.UpdateProductCost(ctx, pc.queries, *price.BaseProductID)
					if err != nil {
						return err
					}
					ingredients[i].Prices[j].Price = &cost
				} else if err != nil {
					return err
				} else {
					ingredients[i].Prices[j].Price = &prodCost.Cost
				}
			}
		}
	}

	return nil
}

func (pc *PriceCalcService) UpdateProductCost(
	ctx context.Context,
	qtx *db.Queries,
	productID int64,
) (float64, error) {
	visited := make(map[int64]bool)
	cost, err := pc.calculateProductCost(productID, visited, ctx)
	if err != nil {
		return 0, err
	}

	_, err = qtx.InsertProductCost(ctx, db.InsertProductCostParams{
		ProductID: productID,
		Cost:      cost,
	})
	if err != nil {
		return 0, err
	}

	for id := range visited {
		if id == productID {
			continue
		}

		cost, err = pc.calculateProductCost(id, map[int64]bool{}, ctx)
		if err != nil {
			return 0, err
		}
		_, err = qtx.InsertProductCost(ctx, db.InsertProductCostParams{
			ProductID: id,
			Cost:      cost,
		})
		if err != nil {
			return 0, err
		}
	}

	return cost, nil
}

func (pc *PriceCalcService) calculateProductCost(
	productID int64,
	visited map[int64]bool,
	c context.Context,
) (float64, error) {
	if visited[productID] {
		return 0, fmt.Errorf("circular dependency detected on product %d", productID)
	}
	visited[productID] = true

	ingredientUsages, err := pc.queries.GetIngredientUsageForProductWithPrice(c, productID)
	if err != nil {
		return 0, err
	}
	totalCost := 0.0
	for _, ingredientUsage := range ingredientUsages {
		if ingredientUsage.BaseProductID != nil {
			subCost, err := pc.calculateProductCost(*ingredientUsage.BaseProductID, visited, c)
			if err != nil {
				return 0, err
			}
			totalCost += subCost * ingredientUsage.Quantity
		} else if ingredientUsage.Price != nil {
			totalCost += *ingredientUsage.Price * ingredientUsage.Quantity
		} else {
			return 0, fmt.Errorf("no price found for ingredient %d", ingredientUsage.IngredientID)
		}
	}

	return totalCost, nil
}

func (pc *PriceCalcService) GetIngredientsWithPrice(
	ctx context.Context,
) ([]db.IngredientWithPrices, error) {
	return pc.GetIngredientsWithPrices(ctx, 1)
}

func (pc *PriceCalcService) GetIngredientsWithPrices(
	ctx context.Context,
	priceLimit int64,
) ([]db.IngredientWithPrices, error) {
	ingredients, err := pc.queries.GetIngredientsWithPriceUnit(
		ctx,
		db.GetIngredientsWithPriceUnitParams{
			IngredientID: nil,
			PriceLimit:   priceLimit,
		},
	)
	if err != nil {
		return nil, err
	}
	return pc.parseIngredientsWithPriceUnitRow(ctx, ingredients)
}

func (pc *PriceCalcService) GetIngredientWithPrice(
	ctx context.Context,
	ingredientId int64,
) (*db.IngredientWithPrices, error) {
	return pc.GetIngredientWithPrices(ctx, ingredientId, 1)
}

func (pc *PriceCalcService) GetIngredientWithPrices(
	ctx context.Context,
	ingredientId,
	priceLimit int64,
) (*db.IngredientWithPrices, error) {
	ingredient, err := pc.queries.GetIngredientsWithPriceUnit(
		ctx,
		db.GetIngredientsWithPriceUnitParams{
			IngredientID: ingredientId,
			PriceLimit:   priceLimit,
		},
	)
	if err != nil {
		return nil, err
	}

	out, err := pc.parseIngredientsWithPriceUnitRow(
		ctx,
		ingredient,
	)
	if err != nil {
		return nil, err
	}

	return &out[0], nil
}

func (pc *PriceCalcService) NewIngredient(
	ctx context.Context,
	params UpdateIngredientParams,
) (*db.IngredientWithPrices, error) {
	tx, err := pc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := pc.queries.WithTx(tx)

	ingredient, err := qtx.InsertIngredient(ctx, params.Name)
	if err != nil {
		return nil, err
	}

	priceRow := db.GetIngredientsWithPriceUnitRow{
		ID:      ingredient.ID,
		Name:    ingredient.Name,
		PriceID: nil,
	}
	params.ID = ingredient.ID

	err = pc.insertIngredientPrice(ctx, qtx, &priceRow, params)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	ingredients, err := pc.parseIngredientsWithPriceUnitRow(
		ctx,
		[]db.GetIngredientsWithPriceUnitRow{priceRow},
	)

	return &ingredients[0], nil
}

func (pc *PriceCalcService) syncIngredientName(
	ctx context.Context,
	qtx *db.Queries,
	row *db.GetIngredientsWithPriceUnitRow,
	name string,
) error {
	if row.Name != name {
		var ingredient db.Ingredient
		ingredient, err := qtx.UpdateIngredient(ctx, db.UpdateIngredientParams{
			ID:   row.ID,
			Name: name,
		})
		if err != nil {
			return err
		}

		row.Name = ingredient.Name
	}

	return nil
}

type insertIngredientPriceQtx interface {
	PutIngredientPrice(
		ctx context.Context,
		arg db.PutIngredientPriceParams,
	) (db.IngredientPrice, error)
	GetUnit(ctx context.Context, unitID int64) (db.Unit, error)
}

func (pc *PriceCalcService) insertIngredientPrice(
	ctx context.Context,
	qtx insertIngredientPriceQtx,
	row *db.GetIngredientsWithPriceUnitRow,
	params UpdateIngredientParams,
) error {
	unit, err := qtx.GetUnit(ctx, params.UnitID)
	if err != nil {
		return err
	}

	if row.ID != params.ID {
		return fmt.Errorf("ingredient id %d does not match row id %d", params.ID, row.ID)
	}

	if params.Price == nil && params.BaseProductID == nil ||
		params.Price != nil && params.BaseProductID != nil {
		return errors.New("either price or baseProductId must be set but not both")
	}

	baseUnitQuantity := params.Quantity / unit.Factor
	var baseUnitPrice *float64 = nil
	if params.Price != nil {
		baseUnitPrice = utils.Ptr(*params.Price / baseUnitQuantity)
	}

	var ingredientPrice db.IngredientPrice
	if row.PriceID == nil ||
		!utils.PtrsEqual(row.Price, baseUnitPrice) ||
		!utils.PtrsEqual(row.BaseProductID, params.BaseProductID) ||
		*row.Quantity != params.Quantity ||
		*row.UnitID != params.UnitID {

		pc.logger.Debug(
			"update ingredient price",
			"baseUnitPrice",
			baseUnitPrice,
			"BaseProductID",
			params.BaseProductID,
			"old price",
			row.Price,
		)

		ingredientPrice, err = qtx.PutIngredientPrice(
			ctx,
			db.PutIngredientPriceParams{
				IngredientID:  params.ID,
				Price:         baseUnitPrice,
				BaseProductID: params.BaseProductID,
				Quantity:      params.Quantity,
				UnitID:        params.UnitID,
			},
		)
		if err != nil {
			return err
		}

		row.PriceID = &ingredientPrice.ID
		row.TimeStamp = &ingredientPrice.TimeStamp
		row.Price = ingredientPrice.Price
		row.Quantity = &ingredientPrice.Quantity
		row.UnitID = &ingredientPrice.UnitID
		row.BaseProductID = ingredientPrice.BaseProductID
	}

	return nil
}

type UpdateIngredientParams struct {
	ID            int64
	Name          string
	Price         *float64
	Quantity      float64
	UnitID        int64
	BaseProductID *int64
}

func (pc *PriceCalcService) UpdateIngredientWithPrice(
	ctx context.Context,
	params UpdateIngredientParams,
) (*db.IngredientWithPrices, error) {
	tx, err := pc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := pc.queries.WithTx(tx)

	ingredientWithPriceRows, err := qtx.GetIngredientsWithPriceUnit(
		ctx,
		db.GetIngredientsWithPriceUnitParams{
			IngredientID: params.ID,
			PriceLimit:   1,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(ingredientWithPriceRows) == 0 {
		return nil, fmt.Errorf("ingredient with id %d not found", params.ID)
	}

	ingredientWithPriceRow := ingredientWithPriceRows[0]

	err = pc.syncIngredientName(ctx, qtx, &ingredientWithPriceRow, params.Name)
	if err != nil {
		return nil, err
	}

	err = pc.insertIngredientPrice(ctx, qtx, &ingredientWithPriceRow, params)
	if err != nil {
		return nil, err
	}

	// find all products that use this ingredient
	products, err := qtx.GetProductsFromIngredient(ctx, params.ID)
	if err != nil {
		return nil, err
	}

	// update the cost of all products that use this ingredient
	for _, product := range products {
		_, err = pc.UpdateProductCost(ctx, qtx, product.ID)
		if err != nil {
			return nil, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	out, err := pc.parseIngredientsWithPriceUnitRow(
		ctx,
		[]db.GetIngredientsWithPriceUnitRow{
			db.GetIngredientsWithPriceUnitRow(ingredientWithPriceRow),
		},
	)
	if err != nil {
		return nil, err
	}

	return &out[0], nil
}

func (pc *PriceCalcService) GetUnits(ctx context.Context) ([]db.Unit, error) {
	units, err := pc.queries.GetUnits(ctx)
	if err != nil {
		return nil, err
	}
	return units, err
}

func (pc *PriceCalcService) GetUnitsMap(ctx context.Context) (UnitsMap, error) {
	units, err := pc.GetUnits(ctx)
	if err != nil {
		return nil, err
	}
	out := UnitsMap{}
	for _, unit := range units {
		out[unit.ID] = unit
	}
	return out, nil
}

func (pc *PriceCalcService) GetUnit(unitId int64) (*db.Unit, error) {
	ctx := context.Background()
	unit, err := pc.queries.GetUnit(ctx, unitId)
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (pc *PriceCalcService) DeleteIngredient(ingredientId int64) error {
	ctx := context.Background()
	num, err := pc.queries.DeleteIngredient(ctx, ingredientId)
	if err != nil {
		return err
	}
	if num < 1 {
		return ErrNoRowsAffected
	}
	return nil
}

func (pc *PriceCalcService) GetProductsWithCost() ([]db.ProductWithCost, error) {
	ctx := context.Background()
	products, err := pc.queries.GetProductsWithCost(ctx)
	if err != nil {
		return nil, err
	}

	out := []db.ProductWithCost{}
	for _, product := range products {
		if product.Cost == nil {
			// again, not supposed to happen, but if it does, calculate the cost and create the row
			newCost, err := pc.UpdateProductCost(ctx, pc.queries, product.ID)
			if err != nil {
				return nil, err
			}
			product.Cost = &newCost
		}
		// fmt.Printf("*baseProductID = %f\n", *product.Cost)
		out = append(out, db.ProductWithCost{
			Product: db.Product{
				ID:            product.ID,
				Name:          product.Name,
				CategoryID:    product.CategoryID,
				Price:         product.Price,
				Multiplicator: product.Multiplicator,
			},
			Cost: *product.Cost,
		})
	}
	return out, nil
}

func (pc *PriceCalcService) GetProductNames(ctx context.Context) (map[int64]string, error) {
	products, err := pc.queries.GetProductNames(ctx)
	if err != nil {
		return nil, err
	}
	out := map[int64]string{}
	for _, product := range products {
		out[product.ID] = product.Name
	}
	return out, nil
}

func (pc *PriceCalcService) GetProductWithCost(productId int64) (*db.ProductWithCost, error) {
	ctx := context.Background()
	product, err := pc.queries.GetProductWithCost(ctx, productId)
	if err != nil {
		return nil, err
	}
	productWithCost := db.ProductWithCost{
		Product: db.Product{
			ID:            product.ID,
			Name:          product.Name,
			CategoryID:    product.CategoryID,
			Price:         product.Price,
			Multiplicator: product.Multiplicator,
		},
		Cost: product.Cost,
	}
	return &productWithCost, nil
}

func (pc *PriceCalcService) UpdateProduct(
	productId, categoryId int64,
	price, multiplicator float64,
	name string,
) (*db.Product, error) {
	ctx := context.Background()
	product, err := pc.queries.UpdateProduct(ctx, db.UpdateProductParams{
		ID:            productId,
		CategoryID:    categoryId,
		Price:         price,
		Name:          name,
		Multiplicator: multiplicator,
	})
	if err != nil {
		return nil, err
	}

	return &product, nil
}

func (pc *PriceCalcService) DeleteProduct(productId int64) error {
	ctx := context.Background()
	num, err := pc.queries.DeleteProduct(ctx, productId)
	if err != nil {
		return nil
	}
	if num < 1 {
		return ErrNoRowsAffected
	}
	return nil
}

func (pc *PriceCalcService) GetIngredientUsageForProduct(
	productId int64,
) ([]db.IngredientUsage, error) {
	ctx := context.Background()
	ingredientUsage, err := pc.queries.GetIngredientUsageForProduct(ctx, productId)
	if err != nil {
		return nil, err
	}
	return ingredientUsage, nil
}

func (pc *PriceCalcService) GetIngredientUsage(
	ingredientUsageId int64,
) (*db.IngredientUsage, error) {
	ctx := context.Background()
	ingredientUsage, err := pc.queries.GetIngredientUsage(ctx, ingredientUsageId)
	if err != nil {
		return nil, err
	}
	return &ingredientUsage, nil
}

func (pc *PriceCalcService) GetCategories() ([]db.Category, error) {
	ctx := context.Background()
	categories, err := pc.queries.GetCategories(ctx)
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (pc *PriceCalcService) PutCategory(name string, vat int64) (*db.Category, error) {
	ctx := context.Background()
	category, err := pc.queries.PutCategory(ctx, db.PutCategoryParams{Name: name, Vat: vat})
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (pc *PriceCalcService) UpdateCategory(id int64, name string, vat int64) (*db.Category, error) {
	ctx := context.Background()
	category, err := pc.queries.UpdateCategory(
		ctx,
		db.UpdateCategoryParams{ID: id, Name: name, Vat: vat},
	)
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (pc *PriceCalcService) GetCategory(id int64) (*db.Category, error) {
	ctx := context.Background()
	category, err := pc.queries.GetCategory(ctx, id)
	if err != nil {
		return nil, err
	}
	return &category, err
}

func (pc *PriceCalcService) PutProduct(name string, categoryId int64) (*db.Product, error) {
	ctx := context.Background()
	product, err := pc.queries.PutProduct(
		ctx,
		db.PutProductParams{Name: name, CategoryID: categoryId},
	)
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (pc *PriceCalcService) PutIngredientUsage(
	ctx context.Context,
	ingredientId, productId, unitId int64,
	quantity float64,
) (*db.IngredientUsage, error) {
	units, err := pc.queries.GetUnits(ctx)
	if err != nil {
		return nil, err
	}
	unit, ok := utils.First(units, func(u db.Unit) bool {
		return u.ID == unitId
	})
	if !ok {
		return nil, fmt.Errorf("unit with id %d not found", unitId)
	}

	baseQuantity := quantity / unit.Factor
	ingredientUsage, err := pc.queries.PutIngredeintUsage(ctx, db.PutIngredeintUsageParams{
		IngredientID: ingredientId,
		ProductID:    productId,
		UnitID:       unitId,
		Quantity:     baseQuantity,
	})
	if err != nil {
		return nil, err
	}
	_, err = pc.UpdateProductCost(ctx, pc.queries, productId)
	if err != nil {
		return nil, err
	}
	return &ingredientUsage, nil
}

func (pc *PriceCalcService) UpdateIngredientUsage(
	ingredientUsageId, unitId int64,
	quantity float64,
	ctx context.Context,
) (*db.IngredientUsage, error) {
	units, err := pc.queries.GetUnits(ctx)
	if err != nil {
		return nil, err
	}
	unit, ok := utils.First(units, func(u db.Unit) bool {
		return u.ID == unitId
	})

	if !ok {
		return nil, fmt.Errorf("unit with id %d not found", unitId)
	}
	baseQuantity := quantity / unit.Factor
	ingredientUsage, err := pc.queries.UpdateIngredientUsage(ctx, db.UpdateIngredientUsageParams{
		ID:       ingredientUsageId,
		UnitID:   unitId,
		Quantity: baseQuantity,
	})
	if err != nil {
		return nil, err
	}

	_, err = pc.UpdateProductCost(ctx, pc.queries, ingredientUsage.ProductID)
	if err != nil {
		return nil, err
	}

	return &ingredientUsage, nil
}

func (pc *PriceCalcService) DeleteIngredientUsage(
	ctx context.Context,
	ingredientUsageId int64,
) error {
	productID, err := pc.queries.DeleteIngredientUsage(ctx, ingredientUsageId)
	if err != nil {
		return err
	}

	_, err = pc.UpdateProductCost(ctx, pc.queries, productID)
	if err != nil {
		return err
	}

	return nil
}

func (pc *PriceCalcService) GetProductsWithIngredient(
	ingredientId int64,
	ctx ...context.Context,
) ([]string, error) {
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}

	products, err := pc.queries.GetProductsFromIngredient(c, ingredientId)
	if err != nil {
		return nil, err
	}
	productNames := make([]string, len(products))
	for i, product := range products {
		productNames[i] = product.Name
	}
	return productNames, nil
}

func (pc *PriceCalcService) InsertUnit(
	name string,
	baseUnitId *int64,
	factor float64,
	ctx context.Context,
) (*db.Unit, error) {
	unit, err := pc.queries.InsertUnit(ctx, db.InsertUnitParams{
		Name:       name,
		BaseUnitID: baseUnitId,
		Factor:     factor,
	})
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (pc *PriceCalcService) UpdateUnit(
	id int64,
	name string,
	baseUnitId *int64,
	factor float64,
	ctx context.Context,
) (*db.Unit, error) {
	unit, err := pc.queries.UpdateUnit(ctx, db.UpdateUnitParams{
		ID:         id,
		Name:       name,
		BaseUnitID: baseUnitId,
		Factor:     factor,
	})
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (pc *PriceCalcService) DeleteUnit(id int64, ctx context.Context) error {
	num, err := pc.queries.DeleteUnit(ctx, id)
	if err != nil {
		return err
	}
	if num < 1 {
		return ErrNoRowsAffected
	}
	return nil
}

func (pc *PriceCalcService) GetIngredientsFromUnit(
	unitId int64,
	ctx context.Context,
) ([]string, error) {
	ingredients, err := pc.queries.GetIngredientsFromUnit(ctx, unitId)
	if err != nil {
		return nil, err
	}
	ingredientNames := make([]string, len(ingredients))
	for i, ingredient := range ingredients {
		ingredientNames[i] = ingredient.Name
	}
	return ingredientNames, nil
}

func (pc *PriceCalcService) GetProductsFromUnit(
	unitId int64,
	ctx context.Context,
) ([]string, error) {
	products, err := pc.queries.GetProductsFromUnit(ctx, unitId)
	if err != nil {
		return nil, err
	}
	productNames := make([]string, len(products))
	for i, product := range products {
		productNames[i] = product.Name
	}
	return productNames, nil
}
