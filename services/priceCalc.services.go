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

var NoRowsAffectedError = errors.New("No Rows Affected")

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
	log.Info(fmt.Sprintf("\n=== migration list ==="))
	sources := provider.ListSources()
	for _, s := range sources {
		log.Info(fmt.Sprintf("%-3s %-2v %v\n", s.Type, s.Version, filepath.Base(s.Path)))
	}

	// List status of migrations before applying them.
	stats, err := provider.Status(ctx)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("\n=== migration status ==="))
	for _, s := range stats {
		log.Info(fmt.Sprintf("%-3s %-2v %v\n", s.Source.Type, s.Source.Version, s.State))
	}

	log.Info(fmt.Sprintf("\n=== log migration output  ==="))
	results, err := provider.Up(ctx)
	if err != nil {
		return nil, err
	}
	log.Info(fmt.Sprintf("\n=== migration results  ==="))
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

func (pc *PriceCalcService) parseIngredientsWithPriceUnitRow(
	ingredients []db.GetIngredientsWithPriceUnitRow,
) []db.IngredientWithPrices {
	out := []db.IngredientWithPrices{}
	for _, ingredientRow := range ingredients {
		var price *db.IngredientPrice = nil
		// check if a price is in the row, if so, fill out price struct
		if ingredientRow.PriceID != nil {
			displayPrice := *ingredientRow.Price * *ingredientRow.Quantity
			displayQuantity := *ingredientRow.Quantity * pc.Units[*ingredientRow.UnitID].Factor

			price = &db.IngredientPrice{
				ID:           *ingredientRow.PriceID,
				TimeStamp:    *ingredientRow.TimeStamp,
				IngredientID: ingredientRow.ID,
				Price:        &displayPrice,
				Quantity:     displayQuantity,
				UnitID:       *ingredientRow.UnitID,
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
	return out
}

func (pc *PriceCalcService) GetIngredientsWithPrice() ([]db.IngredientWithPrices, error) {
	return pc.GetIngredientsWithPrices(1)
}

func (pc *PriceCalcService) GetIngredientsWithPrices(
	priceLimit int64,
) ([]db.IngredientWithPrices, error) {
	ctx := context.Background()
	ingredients, err := pc.queries.GetIngredientsWithPriceUnit(ctx, priceLimit)
	if err != nil {
		return nil, err
	}
	return pc.parseIngredientsWithPriceUnitRow(ingredients), nil
}

func (pc *PriceCalcService) GetIngredientWithPrice(
	ingredientId int64,
) (*db.IngredientWithPrices, error) {
	return pc.GetIngredientWithPrices(ingredientId, 1)
}

func (pc *PriceCalcService) GetIngredientWithPrices(
	ingredientId,
	priceLimit int64,
) (*db.IngredientWithPrices, error) {
	ctx := context.Background()
	ingredient, err := pc.queries.GetIngredientWithPriceUnit(
		ctx,
		db.GetIngredientWithPriceUnitParams{
			ID:    ingredientId,
			Limit: priceLimit,
		},
	)
	if err != nil {
		return nil, err
	}

	out := pc.parseIngredientsWithPriceUnitRow(
		[]db.GetIngredientsWithPriceUnitRow{db.GetIngredientsWithPriceUnitRow(ingredient)},
	)[0]

	return &out, nil
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
	price float64,
	quantity float64,
	unitId int64,
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

	ingredientWithPrice, err := qtx.GetIngredientWithPriceUnit(
		ctx,
		db.GetIngredientWithPriceUnitParams{
			ID:    ingredientId,
			Limit: 1,
		},
	)
	if err != nil {
		return nil, err
	}

	if ingredientWithPrice.Name != ingredientName {
		_, err = qtx.UpdateIngredient(ctx, db.UpdateIngredientParams{
			ID:   ingredientId,
			Name: ingredientName,
		})
		if err != nil {
			return nil, err
		}
	}

	ingredientPrice := db.IngredientPrice{
		ID:        *ingredientWithPrice.PriceID,
		TimeStamp: *ingredientWithPrice.TimeStamp,
		Price:     ingredientWithPrice.Price,
		Quantity:  *ingredientWithPrice.Quantity,
	}

	if ingredientWithPrice.PriceID == nil ||
		*ingredientWithPrice.Price != price ||
		*ingredientWithPrice.Quantity != quantity ||
		*ingredientWithPrice.UnitID != unitId {
		ingredientPrice, err = qtx.PutIngredientPrice(
			ctx,
			db.PutIngredientPriceParams{
				IngredientID: ingredientId,
				Price:        &price,
				Quantity:     quantity,
				UnitID:       unitId,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &db.IngredientWithPrices{
		Ingredient: db.Ingredient{
			ID:   ingredientWithPrice.ID,
			Name: ingredientWithPrice.Name,
		},
		Prices: []db.IngredientPrice{
			ingredientPrice,
		},
	}, nil
}

func (pc *PriceCalcService) PutIngredientPrice(
	ingredientId int64,
	price float64,
	quantity float64,
	unitId int64,
) (*db.IngredientPrice, error) {
	ctx := context.Background()

	units, err := pc.GetUnits()
	if err != nil {
		return nil, err
	}

	i := slices.IndexFunc(units, func(unit db.Unit) bool {
		return unit.ID == unitId
	})

	if i == -1 {
		return nil, errors.New("unit not found")
	}

	unit := units[i]
	baseUnitQuantity := quantity / unit.Factor
	baseUnitPrice := price / baseUnitQuantity

	ingredientPrice, err := pc.queries.PutIngredientPrice(
		ctx,
		db.PutIngredientPriceParams{
			IngredientID: ingredientId,
			Price:        &baseUnitPrice,
			Quantity:     baseUnitQuantity,
			UnitID:       unitId,
		},
	)
	if err != nil {
		return nil, err
	}
	// calculate price and quantity in display units
	// ingredientPrice.Price *= ingredientPrice.Quantity
	// ingredientPrice.Quantity *= unit.Factor
	return &ingredientPrice, nil
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
		return NoRowsAffectedError
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
		out = append(out, db.ProductWithCost{
			Product: db.Product{
				ID:            product.ID,
				Name:          product.Name,
				CategoryID:    product.CategoryID,
				Price:         product.Price,
				Multiplicator: product.Multiplicator,
			},
			Cost: product.Cost,
		})
	}
	return out, nil
}

func (pc *PriceCalcService) GetProductWithCost(productId int64) (*db.ProductWithCost, error) {
	ctx := context.Background()
	product, err := pc.queries.GetProductWithCost(ctx, productId)
	if err != nil {
		return nil, err
	}
	pc.logger.Info("bbbbb", product)
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
	pc.logger.Info("cccc", productWithCost)
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
		return NoRowsAffectedError
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
) (*db.IngredientUsage, error) {
	ctx := context.Background()
	ingredientUsage, err := pc.queries.PutIngredeintUsage(ctx, db.PutIngredeintUsageParams{
		IngredientID: ingredientId,
		ProductID:    productId,
		UnitID:       unitId,
		Quantity:     quantity,
	})
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
	// calculate quantity for display unit
	// ingredientUsage.Quantity *= pc.Units[unitId].Factor
	return &ingredientUsage, nil
}

func (pc *PriceCalcService) DeleteIngredientUsage(ingredientUsageId int64) error {
	ctx := context.Background()
	_, err := pc.queries.DeleteIngredientUsage(ctx, ingredientUsageId)
	if err != nil {
		return err
	}
	return nil
}
