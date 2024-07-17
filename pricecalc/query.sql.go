// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query.sql

package pricecalc

import (
	"context"
)

const getIngredients = `-- name: GetIngredients :many
select id, name
from ingredients
`

func (q *Queries) GetIngredients(ctx context.Context) ([]Ingredient, error) {
	rows, err := q.db.QueryContext(ctx, getIngredients)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Ingredient
	for rows.Next() {
		var i Ingredient
		if err := rows.Scan(&i.ID, &i.Name); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
