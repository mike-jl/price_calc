// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: query.sql

package db

import (
	"context"
)

const deleteIngredient = `-- name: DeleteIngredient :execrows
;

delete from ingredients
where (id = ?)
`

func (q *Queries) DeleteIngredient(ctx context.Context, id int64) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteIngredient, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteIngredientUsage = `-- name: DeleteIngredientUsage :execrows
;

delete from ingredient_usage
where (id = ?)
`

func (q *Queries) DeleteIngredientUsage(ctx context.Context, id int64) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteIngredientUsage, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteProduct = `-- name: DeleteProduct :execrows
;

delete from products
where id = ?
`

func (q *Queries) DeleteProduct(ctx context.Context, id int64) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteProduct, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getCategories = `-- name: GetCategories :many
;

select id, name, vat
from categories
`

func (q *Queries) GetCategories(ctx context.Context) ([]Category, error) {
	rows, err := q.db.QueryContext(ctx, getCategories)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Category
	for rows.Next() {
		var i Category
		if err := rows.Scan(&i.ID, &i.Name, &i.Vat); err != nil {
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

const getCategory = `-- name: GetCategory :one
;

select id, name, vat
from categories
where id = ?
`

func (q *Queries) GetCategory(ctx context.Context, id int64) (Category, error) {
	row := q.db.QueryRowContext(ctx, getCategory, id)
	var i Category
	err := row.Scan(&i.ID, &i.Name, &i.Vat)
	return i, err
}

const getIngredientUsage = `-- name: GetIngredientUsage :one
;

select id, quantity, unit_id, ingredient_id, product_id
from ingredient_usage iu
where iu.id = ?
`

func (q *Queries) GetIngredientUsage(ctx context.Context, id int64) (IngredientUsage, error) {
	row := q.db.QueryRowContext(ctx, getIngredientUsage, id)
	var i IngredientUsage
	err := row.Scan(
		&i.ID,
		&i.Quantity,
		&i.UnitID,
		&i.IngredientID,
		&i.ProductID,
	)
	return i, err
}

const getIngredientUsageForProduct = `-- name: GetIngredientUsageForProduct :many
;

select id, quantity, unit_id, ingredient_id, product_id
from ingredient_usage iu
where iu.product_id = ?
`

func (q *Queries) GetIngredientUsageForProduct(ctx context.Context, productID int64) ([]IngredientUsage, error) {
	rows, err := q.db.QueryContext(ctx, getIngredientUsageForProduct, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []IngredientUsage
	for rows.Next() {
		var i IngredientUsage
		if err := rows.Scan(
			&i.ID,
			&i.Quantity,
			&i.UnitID,
			&i.IngredientID,
			&i.ProductID,
		); err != nil {
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

const getIngredientWithPriceUnit = `-- name: GetIngredientWithPriceUnit :one
;

select i.id, i.name, ip.id as price_id, ip.price, ip.unit_id, ip.quantity, ip.time_stamp
from ingredients i
left join
    ingredient_prices ip
    on ip.id = (
        select id
        from ingredient_prices as ip2
        where ip2.ingredient_id = i.id
        order by time_stamp desc
        limit ?
    )
where i.id = ?
`

type GetIngredientWithPriceUnitParams struct {
	Limit int64
	ID    int64
}

type GetIngredientWithPriceUnitRow struct {
	ID        int64
	Name      string
	PriceID   *int64
	Price     *float64
	UnitID    *int64
	Quantity  *float64
	TimeStamp *int64
}

func (q *Queries) GetIngredientWithPriceUnit(ctx context.Context, arg GetIngredientWithPriceUnitParams) (GetIngredientWithPriceUnitRow, error) {
	row := q.db.QueryRowContext(ctx, getIngredientWithPriceUnit, arg.Limit, arg.ID)
	var i GetIngredientWithPriceUnitRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.PriceID,
		&i.Price,
		&i.UnitID,
		&i.Quantity,
		&i.TimeStamp,
	)
	return i, err
}

const getIngredientsWithPriceUnit = `-- name: GetIngredientsWithPriceUnit :many
select i.id, i.name, ip.id as price_id, ip.price, ip.unit_id, ip.quantity, ip.time_stamp
from ingredients i
left join
    ingredient_prices ip
    on ip.id = (
        select id
        from ingredient_prices as ip2
        where ip2.ingredient_id = i.id
        order by time_stamp desc
        limit ?
    )
`

type GetIngredientsWithPriceUnitRow struct {
	ID        int64
	Name      string
	PriceID   *int64
	Price     *float64
	UnitID    *int64
	Quantity  *float64
	TimeStamp *int64
}

func (q *Queries) GetIngredientsWithPriceUnit(ctx context.Context, limit int64) ([]GetIngredientsWithPriceUnitRow, error) {
	rows, err := q.db.QueryContext(ctx, getIngredientsWithPriceUnit, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetIngredientsWithPriceUnitRow
	for rows.Next() {
		var i GetIngredientsWithPriceUnitRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.PriceID,
			&i.Price,
			&i.UnitID,
			&i.Quantity,
			&i.TimeStamp,
		); err != nil {
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

const getProductWithCost = `-- name: GetProductWithCost :one
;

select
    p.id,
    p.name,
    p.price,
    p.multiplicator,
    p.category_id,
    cast(ifnull(sum(ip.price * iu.quantity), 0) as real) as cost
from products p
left join ingredient_usage iu on iu.product_id = p.id
left join
    (
        select price, ingredient_id, max(time_stamp)
        from ingredient_prices
        group by ingredient_id
    ) ip
    on iu.ingredient_id = ip.ingredient_id
where p.id = ?
`

type GetProductWithCostRow struct {
	ID            int64
	Name          string
	Price         float64
	Multiplicator float64
	CategoryID    int64
	Cost          float64
}

func (q *Queries) GetProductWithCost(ctx context.Context, id int64) (GetProductWithCostRow, error) {
	row := q.db.QueryRowContext(ctx, getProductWithCost, id)
	var i GetProductWithCostRow
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.Multiplicator,
		&i.CategoryID,
		&i.Cost,
	)
	return i, err
}

const getProductsWithCost = `-- name: GetProductsWithCost :many
;

select
    p.id,
    p.name,
    p.price,
    p.multiplicator,
    p.category_id,
    cast(ifnull(sum(ip.price * ip.quantity * iu.quantity), 0) as real) as cost
from products p
left join ingredient_usage iu on iu.product_id = p.id
left join
    (
        select price, ingredient_id, quantity, max(time_stamp)
        from ingredient_prices
        group by ingredient_id
    ) ip
    on iu.ingredient_id = ip.ingredient_id
group by p.id
`

type GetProductsWithCostRow struct {
	ID            int64
	Name          string
	Price         float64
	Multiplicator float64
	CategoryID    int64
	Cost          float64
}

func (q *Queries) GetProductsWithCost(ctx context.Context) ([]GetProductsWithCostRow, error) {
	rows, err := q.db.QueryContext(ctx, getProductsWithCost)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetProductsWithCostRow
	for rows.Next() {
		var i GetProductsWithCostRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Price,
			&i.Multiplicator,
			&i.CategoryID,
			&i.Cost,
		); err != nil {
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

const getProductsWithIngredients = `-- name: GetProductsWithIngredients :many
;

select p.id, p.name, p.price, p.multiplicator, p.category_id, iu.id, iu.quantity, iu.unit_id, iu.ingredient_id, iu.product_id, i.id, i.name, ip.id, ip.time_stamp, ip.price, ip.quantity, ip.unit_id, ip.ingredient_id, ip.base_product_id
from products p
left join ingredient_usage iu on iu.product_id = p.id
left join ingredients i on i.id = iu.ingredient_id
left join
    ingredient_prices ip
    on ip.id = (
        select id
        from ingredient_prices as ip2
        where ip2.ingredient_id = i.id
        order by time_stamp desc
        limit 1
    )
`

type GetProductsWithIngredientsRow struct {
	ID             int64
	Name           string
	Price          float64
	Multiplicator  float64
	CategoryID     int64
	ID_2           *int64
	Quantity       *float64
	UnitID         *int64
	IngredientID   *int64
	ProductID      *int64
	ID_3           *int64
	Name_2         *string
	ID_4           *int64
	TimeStamp      *int64
	Price_2        *float64
	Quantity_2     *float64
	UnitID_2       *int64
	IngredientID_2 *int64
	BaseProductID  *int64
}

func (q *Queries) GetProductsWithIngredients(ctx context.Context) ([]GetProductsWithIngredientsRow, error) {
	rows, err := q.db.QueryContext(ctx, getProductsWithIngredients)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetProductsWithIngredientsRow
	for rows.Next() {
		var i GetProductsWithIngredientsRow
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Price,
			&i.Multiplicator,
			&i.CategoryID,
			&i.ID_2,
			&i.Quantity,
			&i.UnitID,
			&i.IngredientID,
			&i.ProductID,
			&i.ID_3,
			&i.Name_2,
			&i.ID_4,
			&i.TimeStamp,
			&i.Price_2,
			&i.Quantity_2,
			&i.UnitID_2,
			&i.IngredientID_2,
			&i.BaseProductID,
		); err != nil {
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

const getUnits = `-- name: GetUnits :many
;

select id, name, base_unit_id, factor
from units
`

func (q *Queries) GetUnits(ctx context.Context) ([]Unit, error) {
	rows, err := q.db.QueryContext(ctx, getUnits)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Unit
	for rows.Next() {
		var i Unit
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.BaseUnitID,
			&i.Factor,
		); err != nil {
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

const putCategory = `-- name: PutCategory :one
;

insert into categories (name, vat)
values (?,?)
returning id, name, vat
`

type PutCategoryParams struct {
	Name string
	Vat  int64
}

func (q *Queries) PutCategory(ctx context.Context, arg PutCategoryParams) (Category, error) {
	row := q.db.QueryRowContext(ctx, putCategory, arg.Name, arg.Vat)
	var i Category
	err := row.Scan(&i.ID, &i.Name, &i.Vat)
	return i, err
}

const putIngredeintUsage = `-- name: PutIngredeintUsage :one
;

insert into ingredient_usage (quantity, unit_id, ingredient_id, product_id)
values (?, ?, ?, ?)
returning id, quantity, unit_id, ingredient_id, product_id
`

type PutIngredeintUsageParams struct {
	Quantity     float64
	UnitID       int64
	IngredientID int64
	ProductID    int64
}

func (q *Queries) PutIngredeintUsage(ctx context.Context, arg PutIngredeintUsageParams) (IngredientUsage, error) {
	row := q.db.QueryRowContext(ctx, putIngredeintUsage,
		arg.Quantity,
		arg.UnitID,
		arg.IngredientID,
		arg.ProductID,
	)
	var i IngredientUsage
	err := row.Scan(
		&i.ID,
		&i.Quantity,
		&i.UnitID,
		&i.IngredientID,
		&i.ProductID,
	)
	return i, err
}

const putIngredient = `-- name: PutIngredient :one
;

insert into ingredients(name)
values (?)
returning id, name
`

func (q *Queries) PutIngredient(ctx context.Context, name string) (Ingredient, error) {
	row := q.db.QueryRowContext(ctx, putIngredient, name)
	var i Ingredient
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const putIngredientPrice = `-- name: PutIngredientPrice :one
;

insert into ingredient_prices (ingredient_id, price, quantity, unit_id)
values (?, ?, ?, ?)
returning id, time_stamp, price, quantity, unit_id, ingredient_id, base_product_id
`

type PutIngredientPriceParams struct {
	IngredientID int64
	Price        *float64
	Quantity     float64
	UnitID       int64
}

func (q *Queries) PutIngredientPrice(ctx context.Context, arg PutIngredientPriceParams) (IngredientPrice, error) {
	row := q.db.QueryRowContext(ctx, putIngredientPrice,
		arg.IngredientID,
		arg.Price,
		arg.Quantity,
		arg.UnitID,
	)
	var i IngredientPrice
	err := row.Scan(
		&i.ID,
		&i.TimeStamp,
		&i.Price,
		&i.Quantity,
		&i.UnitID,
		&i.IngredientID,
		&i.BaseProductID,
	)
	return i, err
}

const putProduct = `-- name: PutProduct :one
;

insert into products(name, category_id)
values (?, ?)
returning id, name, price, multiplicator, category_id
`

type PutProductParams struct {
	Name       string
	CategoryID int64
}

func (q *Queries) PutProduct(ctx context.Context, arg PutProductParams) (Product, error) {
	row := q.db.QueryRowContext(ctx, putProduct, arg.Name, arg.CategoryID)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.Multiplicator,
		&i.CategoryID,
	)
	return i, err
}

const updateCategory = `-- name: UpdateCategory :one
;


update categories
set name=?, vat=?
where id=?
returning id, name, vat
`

type UpdateCategoryParams struct {
	Name string
	Vat  int64
	ID   int64
}

func (q *Queries) UpdateCategory(ctx context.Context, arg UpdateCategoryParams) (Category, error) {
	row := q.db.QueryRowContext(ctx, updateCategory, arg.Name, arg.Vat, arg.ID)
	var i Category
	err := row.Scan(&i.ID, &i.Name, &i.Vat)
	return i, err
}

const updateIngredient = `-- name: UpdateIngredient :one
;

update ingredients
set name=?
where ( id = ? )
returning id, name
`

type UpdateIngredientParams struct {
	Name string
	ID   int64
}

func (q *Queries) UpdateIngredient(ctx context.Context, arg UpdateIngredientParams) (Ingredient, error) {
	row := q.db.QueryRowContext(ctx, updateIngredient, arg.Name, arg.ID)
	var i Ingredient
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const updateIngredientUsage = `-- name: UpdateIngredientUsage :one
;

update ingredient_usage
set quantity=?, unit_id=?
where ( id = ? )
returning id, quantity, unit_id, ingredient_id, product_id
`

type UpdateIngredientUsageParams struct {
	Quantity float64
	UnitID   int64
	ID       int64
}

func (q *Queries) UpdateIngredientUsage(ctx context.Context, arg UpdateIngredientUsageParams) (IngredientUsage, error) {
	row := q.db.QueryRowContext(ctx, updateIngredientUsage, arg.Quantity, arg.UnitID, arg.ID)
	var i IngredientUsage
	err := row.Scan(
		&i.ID,
		&i.Quantity,
		&i.UnitID,
		&i.IngredientID,
		&i.ProductID,
	)
	return i, err
}

const updateProduct = `-- name: UpdateProduct :one
;

update products
set name=?, category_id=?, price=?, multiplicator=?
where id=?
returning id, name, price, multiplicator, category_id
`

type UpdateProductParams struct {
	Name          string
	CategoryID    int64
	Price         float64
	Multiplicator float64
	ID            int64
}

func (q *Queries) UpdateProduct(ctx context.Context, arg UpdateProductParams) (Product, error) {
	row := q.db.QueryRowContext(ctx, updateProduct,
		arg.Name,
		arg.CategoryID,
		arg.Price,
		arg.Multiplicator,
		arg.ID,
	)
	var i Product
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Price,
		&i.Multiplicator,
		&i.CategoryID,
	)
	return i, err
}
