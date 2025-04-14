package services

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/mike-jl/price_calc/db"
	"github.com/mike-jl/price_calc/internal/utils"
	viewmodels "github.com/mike-jl/price_calc/viewModels"
	"github.com/stretchr/testify/assert"
)

type mockBaseProductPriceResolver struct{}

func (m *mockBaseProductPriceResolver) resolveBaseProductPrices(
	rows []viewmodels.IngredientWithPrices,
	ctx context.Context,
) error {
	return nil
}

func TestParseIngredientsWithPriceUnitRow(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                    string
		input                   []db.GetIngredientsWithPriceUnitRow
		expectedIngredientCount int
		expectedPriceCount      int
		expectedFirstName       string
		expectError             bool
	}{
		{
			name: "Single ingredient with single price",
			input: []db.GetIngredientsWithPriceUnitRow{
				{
					ID:        1,
					Name:      "Flour",
					PriceID:   utils.Ptr(int64(101)),
					TimeStamp: utils.Ptr(int64(1001)),
					Price:     utils.Ptr(float64(1.5)),
					Quantity:  utils.Ptr(float64(1)),
					UnitID:    utils.Ptr(int64(1)),
				},
			},
			expectedIngredientCount: 1,
			expectedPriceCount:      1,
			expectedFirstName:       "Flour",
			expectError:             false,
		},
		{
			name: "Multiple ingredients with no price",
			input: []db.GetIngredientsWithPriceUnitRow{
				{
					ID:        1,
					Name:      "Flour",
					PriceID:   nil,
					TimeStamp: utils.Ptr(int64(1001)),
					Price:     nil,
					Quantity:  nil,
					UnitID:    nil,
				},
				{
					ID:        2,
					Name:      "Sugar",
					PriceID:   nil,
					TimeStamp: utils.Ptr(int64(1001)),
					Price:     nil,
					Quantity:  nil,
					UnitID:    nil,
				},
			},
			expectedIngredientCount: 2,
			expectedPriceCount:      0,
			expectedFirstName:       "Flour",
			expectError:             false,
		},
		{
			name: "Single ingredient with multiple prices",
			input: []db.GetIngredientsWithPriceUnitRow{
				{
					ID:        1,
					Name:      "Flour",
					PriceID:   utils.Ptr(int64(101)),
					TimeStamp: utils.Ptr(int64(1001)),
					Price:     utils.Ptr(float64(1.5)),
					Quantity:  utils.Ptr(float64(1)),
					UnitID:    utils.Ptr(int64(1)),
				},
				{
					ID:        1,
					Name:      "Flour",
					PriceID:   utils.Ptr(int64(102)),
					TimeStamp: utils.Ptr(int64(1002)),
					Price:     utils.Ptr(float64(2.0)),
					Quantity:  utils.Ptr(float64(2)),
					UnitID:    utils.Ptr(int64(1)),
				},
			},
			expectedIngredientCount: 1,
			expectedPriceCount:      2,
			expectedFirstName:       "Flour",
			expectError:             false,
		},
		{
			name: "Multiple ingredients",
			input: []db.GetIngredientsWithPriceUnitRow{
				{
					ID:        1,
					Name:      "Flour",
					PriceID:   utils.Ptr(int64(101)),
					TimeStamp: utils.Ptr(int64(1001)),
					Price:     utils.Ptr(float64(1.5)),
					Quantity:  utils.Ptr(float64(1)),
					UnitID:    utils.Ptr(int64(1)),
				},
				{
					ID:        2,
					Name:      "Sugar",
					PriceID:   utils.Ptr(int64(201)),
					TimeStamp: utils.Ptr(int64(2001)),
					Price:     utils.Ptr(float64(3.0)),
					Quantity:  utils.Ptr(float64(1)),
					UnitID:    utils.Ptr(int64(2)),
				},
			},
			expectedIngredientCount: 2,
			expectedPriceCount:      1,
			expectedFirstName:       "Flour",
			expectError:             false,
		},
		{
			name: "Malformed row",
			input: []db.GetIngredientsWithPriceUnitRow{
				{
					ID:        1,
					Name:      "Flour",
					PriceID:   utils.Ptr(int64(101)),
					TimeStamp: nil,
					Price:     nil,
					Quantity:  nil,
					UnitID:    nil,
				},
			},
			expectError: true,
		},
	}

	pc := &PriceCalcService{
		baseProductPriceResolver: &mockBaseProductPriceResolver{},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := pc.parseIngredientsWithPriceUnitRow(ctx, tc.input)
			if tc.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, result, tc.expectedIngredientCount)
			assert.Len(t, result[0].Prices, tc.expectedPriceCount)
			assert.Equal(t, tc.expectedFirstName, result[0].Ingredient.Name)
		})
	}
}

type mockSyncIngredientPriceDb struct {
	putIngredientPriceCalled bool
	unit                     db.Unit
}

func (m *mockSyncIngredientPriceDb) PutIngredientPrice(
	ctx context.Context,
	arg db.PutIngredientPriceParams,
) (db.IngredientPrice, error) {
	m.putIngredientPriceCalled = true
	return db.IngredientPrice{
		ID:            1,
		IngredientID:  arg.IngredientID,
		TimeStamp:     1,
		Price:         arg.Price,
		Quantity:      arg.Quantity,
		UnitID:        arg.UnitID,
		BaseProductID: arg.BaseProductID,
	}, nil
}

func (m *mockSyncIngredientPriceDb) GetUnit(ctx context.Context, unitID int64) (db.Unit, error) {
	return m.unit, nil
}

func TestSyncIngredientPrice(t *testing.T) {
	tests := []struct {
		name                    string
		row                     db.GetIngredientsWithPriceUnitRow
		params                  UpdateIngredientParams
		unit                    db.Unit
		expectError             bool
		expectPriceInsertCalled bool
	}{
		{
			name:                    "No existing price, should insert new price",
			expectError:             false,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       nil,
				TimeStamp:     nil,
				Price:         nil,
				Quantity:      nil,
				UnitID:        nil,
				BaseProductID: nil,
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         utils.Ptr(1.5),
				Quantity:      1,
				UnitID:        1,
				BaseProductID: nil,
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "Same price, no base product, same quantity, factor 1, should not insert",
			expectError:             false,
			expectPriceInsertCalled: false,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         utils.Ptr(1.5),
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: nil,
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         utils.Ptr(1.5),
				Quantity:      1,
				UnitID:        1,
				BaseProductID: nil,
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "Same price, no base product, same quantity, factor 10, should insert",
			expectError:             false,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         utils.Ptr(1.5),
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: nil,
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         utils.Ptr(1.5),
				Quantity:      1,
				UnitID:        1,
				BaseProductID: nil,
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 10,
			},
		},
		{
			name:                    "Same price, no base product, same quantity, factor 0.1, should insert",
			expectError:             false,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         utils.Ptr(1.5),
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: nil,
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         utils.Ptr(1.5),
				Quantity:      1,
				UnitID:        1,
				BaseProductID: nil,
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 0.1,
			},
		},
		{
			name:                    "Different price, no base product, same quantity, factor 1, should insert",
			expectError:             false,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         utils.Ptr(1.5),
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: nil,
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         utils.Ptr(1.51),
				Quantity:      1,
				UnitID:        1,
				BaseProductID: nil,
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "both price and baseproduct nil, should error",
			expectError:             true,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         utils.Ptr(1.5),
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: nil,
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         nil,
				Quantity:      1,
				UnitID:        1,
				BaseProductID: nil,
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "both price and baseproduct set, should error",
			expectError:             true,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         utils.Ptr(1.5),
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: nil,
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         utils.Ptr(1.5),
				Quantity:      1,
				UnitID:        1,
				BaseProductID: utils.Ptr(int64(1)),
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "price is nil, baseproduct and quantity is unchanged, should not insert",
			expectError:             false,
			expectPriceInsertCalled: false,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         nil,
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: utils.Ptr(int64(16)),
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         nil,
				Quantity:      1,
				UnitID:        1,
				BaseProductID: utils.Ptr(int64(16)),
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "price is nil, baseproduct changed, quantity is unchanged, should insert",
			expectError:             false,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         nil,
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: utils.Ptr(int64(16)),
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         nil,
				Quantity:      1.0,
				UnitID:        1,
				BaseProductID: utils.Ptr(int64(17)),
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "row id and params id are different, should error",
			expectError:             true,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         nil,
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: utils.Ptr(int64(16)),
			},
			params: UpdateIngredientParams{
				ID:            10,
				Name:          "Flour",
				Price:         nil,
				Quantity:      1.1,
				UnitID:        1,
				BaseProductID: utils.Ptr(int64(16)),
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
		{
			name:                    "price is nil, baseproduct unchanged, quantity changed, should insert",
			expectError:             false,
			expectPriceInsertCalled: true,
			row: db.GetIngredientsWithPriceUnitRow{
				ID:            1,
				Name:          "Flour",
				PriceID:       utils.Ptr(int64(101)),
				TimeStamp:     nil,
				Price:         nil,
				Quantity:      utils.Ptr(1.0),
				UnitID:        utils.Ptr(int64(1)),
				BaseProductID: utils.Ptr(int64(16)),
			},
			params: UpdateIngredientParams{
				ID:            1,
				Name:          "Flour",
				Price:         nil,
				Quantity:      1.1,
				UnitID:        1,
				BaseProductID: utils.Ptr(int64(16)),
			},
			unit: db.Unit{
				ID:     1,
				Name:   "unit",
				Factor: 1,
			},
		},
	}

	ctx := context.Background()
	pc := &PriceCalcService{
		baseProductPriceResolver: &mockBaseProductPriceResolver{},
		logger: slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			qtx := &mockSyncIngredientPriceDb{
				unit: tc.unit,
			}
			err := pc.insertIngredientPrice(ctx, qtx, &tc.row, tc.params)
			if tc.expectError {
				assert.Error(t, err, "expected error")
				return
			}
			assert.NoError(t, err, "unexpected error")
			assert.Equal(
				t,
				tc.expectPriceInsertCalled,
				qtx.putIngredientPriceCalled,
				"expected price insert to be called",
			)
			if assert.NotNil(t, tc.row.Quantity, "tc.row.Quantity should not be nil") {
				assert.InDelta(
					t,
					tc.params.Quantity,
					*tc.row.Quantity,
					0.0001,
					"expected quantity to be equal",
				)
			}
			if tc.params.Price != nil {
				if assert.NotNil(
					t,
					tc.row.Price,
					"if tc.params.Price is not nil, tc.row.Price should not be nil",
				) {
					expectedPrice := (*tc.params.Price * tc.unit.Factor) / tc.params.Quantity
					assert.InDelta(
						t,
						expectedPrice,
						*tc.row.Price,
						0.0001,
						"expected price to be equal",
					)
				}
			} else {
				assert.Nil(t, tc.row.Price, "tc.input.Price should nil", "if tc.params.Price is nil, tc.row.Price should be nil")
			}
			if tc.params.BaseProductID == nil {
				assert.Nil(t, tc.row.BaseProductID, "tc.row.BaseProductID should be nil")
			} else {
				if assert.NotNil(t, tc.row.BaseProductID, "tc.row.BaseProductID should not be nil") {
					assert.Equal(t, *tc.params.BaseProductID, *tc.row.BaseProductID, "expected BaseProductID to be equal")
				}
			}
		})
	}
}
