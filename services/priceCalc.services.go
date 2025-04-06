package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mike-jl/price_calc/db"
)

type UnitsMap map[int64]db.Unit

type PriceCalcService struct {
	queries *db.Queries
	db      *sql.DB
	logger  *slog.Logger
	Units   UnitsMap
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

	// get units, needs to be done only once
	units, err := queries.GetUnits(ctx)
	if err != nil {
		log.Error("error getting units " + err.Error())
	}

	unitsMap := UnitsMap{}
	for _, unit := range units {
		unitsMap[unit.ID] = unit
	}

	return &PriceCalcService{queries, sql, log, unitsMap}, nil
}

func (pc *PriceCalcService) CheckCircularDependency(
	productId, ingredientId int64,
	c ...context.Context,
) (bool, error) {
	ctx := context.Background()
	if len(c) > 0 {
		ctx = c[0]
	}
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

	ingredient, err := pc.queries.GetIngredientWithPriceUnit(
		ctx,
		db.GetIngredientWithPriceUnitParams{
			ID:    currentIngredientId,
			Limit: 1,
		},
	)
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
	ingredients []db.GetIngredientsWithPriceUnitRow,
	c context.Context,
) ([]db.IngredientWithPrices, error) {
	out := []db.IngredientWithPrices{}
	for _, ingredientRow := range ingredients {
		var price *db.IngredientPrice = nil
		// check if a price is in the row, if so, fill out price struct
		if ingredientRow.PriceID != nil {
			price = &db.IngredientPrice{
				ID:            *ingredientRow.PriceID,
				TimeStamp:     *ingredientRow.TimeStamp,
				IngredientID:  ingredientRow.ID,
				Price:         ingredientRow.Price,
				Quantity:      *ingredientRow.Quantity,
				UnitID:        *ingredientRow.UnitID,
				BaseProductID: ingredientRow.BaseProductID,
			}
		}

		// check if the ingredeient already exists in the out struct
		i := slices.IndexFunc(out, func(ip db.IngredientWithPrices) bool {
			return ingredientRow.ID == ip.Ingredient.ID
		})
		// append the ingredient
		if i == -1 {
			out = append(
				out,
				db.IngredientWithPrices{
					Ingredient: db.Ingredient{ID: ingredientRow.ID, Name: ingredientRow.Name},
				},
			)
			i = len(out) - 1
		}
		// append the price to the ingredient
		if price != nil {
			out[i].Prices = append(out[i].Prices, *price)
		}
	}

	// check if the ingredient has a base product
	for i, ingredient := range out {
		if len(ingredient.Prices) <= 0 {
			continue
		}
		for j, price := range ingredient.Prices {
			if price.BaseProductID != nil {
				prodCost, err := pc.queries.GetProductCost(c, *price.BaseProductID)
				if err == sql.ErrNoRows {
					// this should not happen, but if it does, calculate the cost and create the row
					cost, err := pc.UpdateProductCost(*price.BaseProductID, c)
					if err != nil {
						return nil, err
					}
					out[i].Prices[j].Price = &cost
				} else if err != nil {
					return nil, err
				} else {
					out[i].Prices[j].Price = &prodCost.Cost
				}
			}
		}
	}

	return out, nil
}

func (pc *PriceCalcService) UpdateProductCost(
	productID int64,
	ctx ...context.Context,
) (float64, error) {
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}

	visited := make(map[int64]bool)
	cost, err := pc.calculateProductCost(productID, visited, c)
	if err != nil {
		return 0, err
	}

	_, err = pc.queries.InsertProductCost(c, db.InsertProductCostParams{
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

		cost, err = pc.calculateProductCost(id, map[int64]bool{}, c)
		if err != nil {
			return 0, err
		}
		_, err = pc.queries.InsertProductCost(c, db.InsertProductCostParams{
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

func (pc *PriceCalcService) GetIngredientsWithPrice() ([]db.IngredientWithPrices, error) {
	return pc.GetIngredientsWithPrices(1)
}

func (pc *PriceCalcService) GetIngredientsWithPrices(
	priceLimit int64,
	ctx ...context.Context,
) ([]db.IngredientWithPrices, error) {
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}
	ingredients, err := pc.queries.GetIngredientsWithPriceUnit(c, priceLimit)
	if err != nil {
		return nil, err
	}
	return pc.parseIngredientsWithPriceUnitRow(ingredients, c)
}

func (pc *PriceCalcService) GetIngredientWithPrice(
	ingredientId int64,
) (*db.IngredientWithPrices, error) {
	return pc.GetIngredientWithPrices(ingredientId, 1)
}

func (pc *PriceCalcService) GetIngredientWithPrices(
	ingredientId,
	priceLimit int64,
	ctx ...context.Context,
) (*db.IngredientWithPrices, error) {
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}

	ingredient, err := pc.queries.GetIngredientWithPriceUnit(
		c,
		db.GetIngredientWithPriceUnitParams{
			ID:    ingredientId,
			Limit: priceLimit,
		},
	)
	if err != nil {
		return nil, err
	}

	out, err := pc.parseIngredientsWithPriceUnitRow(
		[]db.GetIngredientsWithPriceUnitRow{db.GetIngredientsWithPriceUnitRow(ingredient)},
		c,
	)
	if err != nil {
		return nil, err
	}

	return &out[0], nil
}

func (pc *PriceCalcService) PutIngredient(name string) (*db.Ingredient, error) {
	ctx := context.Background()
	ingredient, err := pc.queries.PutIngredient(ctx, name)
	if err != nil {
		return nil, err
	}
	return &ingredient, nil
}

func (pc *PriceCalcService) UpdateIngredientWithPrice(
	ingredientId int64,
	ingredientName string,
	price *float64,
	quantity float64,
	unitId int64,
	baseProductId *int64,
	c ...context.Context,
) (*db.IngredientWithPrices, error) {
	ctx := context.Background()
	if len(c) > 0 {
		ctx = c[0]
	}

	tx, err := pc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	qtx := pc.queries.WithTx(tx)

	ingredientWithPriceRow, err := qtx.GetIngredientWithPriceUnit(
		ctx,
		db.GetIngredientWithPriceUnitParams{
			ID:    ingredientId,
			Limit: 1,
		},
	)
	if err != nil {
		return nil, err
	}

	if ingredientWithPriceRow.Name != ingredientName {
		ingredient, err := qtx.UpdateIngredient(ctx, db.UpdateIngredientParams{
			ID:   ingredientId,
			Name: ingredientName,
		})
		if err != nil {
			return nil, err
		}

		ingredientWithPriceRow.Name = ingredient.Name
	}

	unit, ok := pc.Units[unitId]
	if !ok {
		return nil, errors.New("unit not found")
	}

	if price == nil && baseProductId == nil || price != nil && baseProductId != nil {
		return nil, errors.New("either price or baseProductId must be set but not both")
	}

	baseUnitQuantity := quantity / unit.Factor
	var baseUnitPrice *float64 = nil
	if price != nil {
		baseUnitPrice = new(float64)
		*baseUnitPrice = *price / baseUnitQuantity
	}

	var ingredientPrice db.IngredientPrice
	if ingredientWithPriceRow.PriceID == nil ||
		ingredientWithPriceRow.Price != baseUnitPrice ||
		(ingredientWithPriceRow.Price != nil && baseUnitPrice != nil && *ingredientWithPriceRow.Price != *baseUnitPrice) ||
		ingredientWithPriceRow.BaseProductID != baseProductId ||
		(ingredientWithPriceRow.BaseProductID != nil && baseProductId != nil && *ingredientWithPriceRow.BaseProductID != *baseProductId) ||
		*ingredientWithPriceRow.Quantity != quantity ||
		*ingredientWithPriceRow.UnitID != unitId {

		fmt.Printf(
			"Insert Price: %v, BaseProductID: %v (price==nil: %v, baseProductID==nil: %v)\n",
			baseUnitPrice,
			baseProductId,
			baseUnitPrice == nil,
			baseProductId == nil,
		)
		if baseProductId != nil {
			fmt.Printf("*baseProductID = %d\n", *baseProductId)
		}

		ingredientPrice, err = qtx.PutIngredientPrice(
			ctx,
			db.PutIngredientPriceParams{
				IngredientID:  ingredientId,
				Price:         baseUnitPrice,
				BaseProductID: baseProductId,
				Quantity:      quantity,
				UnitID:        unitId,
			},
		)
		if err != nil {
			return nil, err
		}

		ingredientWithPriceRow.PriceID = &ingredientPrice.ID
		ingredientWithPriceRow.TimeStamp = &ingredientPrice.TimeStamp
		ingredientWithPriceRow.Price = ingredientPrice.Price
		ingredientWithPriceRow.Quantity = &ingredientPrice.Quantity
		ingredientWithPriceRow.UnitID = &ingredientPrice.UnitID
	}

	// find all products that use this ingredient
	products, err := qtx.GetProductsFromIngredient(ctx, ingredientId)
	if err != nil {
		return nil, err
	}

	// update the cost of all products that use this ingredient
	for _, product := range products {
		_, err = pc.UpdateProductCost(product.ID, ctx)
		if err != nil {
			return nil, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	// out := pc.parseIngredientWithPriceUnitRow(ingredientWithPriceRow)
	out, err := pc.parseIngredientsWithPriceUnitRow(
		[]db.GetIngredientsWithPriceUnitRow{
			db.GetIngredientsWithPriceUnitRow(ingredientWithPriceRow),
		},
		ctx,
	)
	if err != nil {
		return nil, err
	}

	return &out[0], nil
}

func (pc *PriceCalcService) GetUnits() ([]db.Unit, error) {
	ctx := context.Background()
	units, err := pc.queries.GetUnits(ctx)
	if err != nil {
		return nil, err
	}
	return units, err
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

func (pc *PriceCalcService) GetProductsWithIngredients() ([]db.ProductWithIngredient, error) {
	ctx := context.Background()
	products, err := pc.queries.GetProductsWithIngredients(ctx)
	if err != nil {
		return nil, err
	}
	out := []db.ProductWithIngredient{}
	for _, product := range products {
		var ingredient *db.IngredientUsageWithPrice = nil
		if product.ID_2 != nil {
			ingredient = &db.IngredientUsageWithPrice{
				IngredientUsage: db.IngredientUsage{
					ID:           *product.ID_2,
					Quantity:     *product.Quantity,
					UnitID:       *product.UnitID,
					IngredientID: *product.IngredientID,
					ProductID:    *product.ProductID,
				},
				Ingredient: db.Ingredient{
					ID:   *product.ID_3,
					Name: *product.Name_2,
				},
				IngredientPrice: db.IngredientPrice{
					ID:           *product.ID_4,
					TimeStamp:    *product.TimeStamp,
					Price:        &product.Price,
					Quantity:     *product.Quantity_2,
					UnitID:       *product.UnitID_2,
					IngredientID: *product.IngredientID_2,
				},
			}
		}
		i := slices.IndexFunc(out, func(ip db.ProductWithIngredient) bool {
			return product.ID == ip.Product.ID
		})
		if i == -1 {
			out = append(out, db.ProductWithIngredient{
				Product: db.Product{
					ID:         product.ID,
					Name:       product.Name,
					CategoryID: product.CategoryID,
				},
			})
			i = len(out) - 1
		}
		if ingredient != nil {
			out[i].IngredientUsageWithPrice = append(out[i].IngredientUsageWithPrice, *ingredient)
		}
	}
	return out, nil
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
			newCost, err := pc.UpdateProductCost(product.ID, ctx)
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

func (pc *PriceCalcService) GetProductNames(ctx ...context.Context) (map[int64]string, error) {
	c := context.Background()
	if len(ctx) > 0 {
		c = ctx[0]
	}
	products, err := pc.queries.GetProductNames(c)
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
	ingredientId, productId, unitId int64,
	quantity float64,
	c ...context.Context,
) (*db.IngredientUsage, error) {
	ctx := context.Background()
	if len(c) > 0 {
		ctx = c[0]
	}
	baseQuantity := quantity / pc.Units[unitId].Factor
	ingredientUsage, err := pc.queries.PutIngredeintUsage(ctx, db.PutIngredeintUsageParams{
		IngredientID: ingredientId,
		ProductID:    productId,
		UnitID:       unitId,
		Quantity:     baseQuantity,
	})
	if err != nil {
		return nil, err
	}
	_, err = pc.UpdateProductCost(productId, ctx)
	if err != nil {
		return nil, err
	}
	return &ingredientUsage, nil
}

func (pc *PriceCalcService) UpdateIngredientUsage(
	ingredientUsageId, unitId int64,
	quantity float64,
) (*db.IngredientUsage, error) {
	ctx := context.Background()
	baseQuantity := quantity / pc.Units[unitId].Factor
	ingredientUsage, err := pc.queries.UpdateIngredientUsage(ctx, db.UpdateIngredientUsageParams{
		ID:       ingredientUsageId,
		UnitID:   unitId,
		Quantity: baseQuantity,
	})
	if err != nil {
		return nil, err
	}

	_, err = pc.UpdateProductCost(ingredientUsage.ProductID, ctx)
	if err != nil {
		return nil, err
	}

	return &ingredientUsage, nil
}

func (pc *PriceCalcService) DeleteIngredientUsage(ingredientUsageId int64) error {
	ctx := context.Background()
	productID, err := pc.queries.DeleteIngredientUsage(ctx, ingredientUsageId)
	if err != nil {
		return err
	}

	_, err = pc.UpdateProductCost(productID, ctx)
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

func (pc *PriceCalcService) GetUnit(id int64, ctx context.Context) (*db.Unit, error) {
	unit, err := pc.queries.GetUnit(ctx, id)
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

func Where[C ~[]T, T any](collection C, predicate func(T) bool) (out C) {
	for _, v := range collection {
		if predicate(v) {
			out = append(out, v)
		}
	}
	return
}

func First[C ~[]T, T any](collection C, predicate func(T) bool) (out T, ok bool) {
	for _, v := range collection {
		if predicate(v) {
			return v, true
		}
	}
	return out, false
}
