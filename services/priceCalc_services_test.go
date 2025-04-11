package services

import (
	"context"
	"testing"

	"github.com/mike-jl/price_calc/db"
	"github.com/mike-jl/price_calc/internal/utils"
	"github.com/stretchr/testify/assert"
)

type mockBaseProductPriceResolver struct{}

func (m *mockBaseProductPriceResolver) resolveBaseProductPrices(
	rows []db.IngredientWithPrices,
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
		expectedError           bool
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
			expectedError:           false,
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
			expectedError:           false,
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
			expectedError:           false,
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
			expectedError:           false,
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
			expectedError: true,
		},
	}

	pc := &PriceCalcService{
		baseProductPriceResolver: &mockBaseProductPriceResolver{},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := pc.parseIngredientsWithPriceUnitRow(tc.input, ctx)
			if tc.expectedError {
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
