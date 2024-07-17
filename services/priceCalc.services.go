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
	logger  *slog.Logger
	Units   UnitsMap
}

var NoRowsAffectedError = errors.New("No Rows Affected")

func NewPriceCalcService(log *slog.Logger, dbName string) (*PriceCalcService, error) {
	ctx := context.Background()

	sql, err := sql.Open("sqlite3", "test123.db?_foreign_keys=on")
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

	return &PriceCalcService{queries, log, unitsMap}, nil
}

func (pc *PriceCalcService) GetIngredientsWithPrice() ([]db.IngredientWithPrice, error) {
	ctx := context.Background()
	ingredientsRows, err := pc.queries.GetIngredients(ctx)
	if err != nil {
		return nil, err
	}

	var out []db.IngredientWithPrice
	for _, ingredientRow := range ingredientsRows {
		var price *db.IngredientPrice = nil
		if ingredientRow.ID_2 != nil {
			price = &db.IngredientPrice{
				ID:           *ingredientRow.ID_2,
				Price:        *ingredientRow.Price,
				TimeStamp:    *ingredientRow.TimeStamp,
				IngredientID: *ingredientRow.IngredientID,
				Quantity:     *ingredientRow.Quantity,
				UnitID:       *ingredientRow.UnitID,
			}
		}
		i := slices.IndexFunc(out, func(ip db.IngredientWithPrice) bool {
			return ingredientRow.ID == ip.Ingredient.ID
		})
		if i == -1 {
			out = append(
				out,
				db.IngredientWithPrice{
					Ingredient: db.Ingredient{ID: ingredientRow.ID, Name: ingredientRow.Name},
				},
			)
			i = len(out) - 1
		}
		if price != nil {
			out[i].Prices = append(out[i].Prices, *price)
		}
	}

	return out, nil
}

func (pc *PriceCalcService) PutIngredient(name string) (*db.Ingredient, error) {
	ctx := context.Background()
	ingredient, err := pc.queries.PutIngredient(ctx, name)
	if err != nil {
		return nil, err
	}
	return &ingredient, nil
}

func (pc *PriceCalcService) PutIngredientPrice(
	ingredientId int64,
	price float64,
	quantity float64,
	unitId int64,
) (*db.IngredientPrice, error) {
	ctx := context.Background()
	ingredientPrice, err := pc.queries.PutIngredientPrice(
		ctx,
		db.PutIngredientPriceParams{
			IngredientID: ingredientId,
			Price:        price,
			Quantity:     quantity,
			UnitID:       unitId,
		},
	)
	if err != nil {
		return nil, err
	}
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
					Price:        *product.Price,
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
