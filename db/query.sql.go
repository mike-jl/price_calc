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

const deleteIngredientUsage = `-- name: DeleteIngredientUsage :one
;

delete from ingredient_usage
where (id = ?)
returning product_id
`

func (q *Queries) DeleteIngredientUsage(ctx context.Context, id int64) (int64, error) {
	row := q.db.QueryRowContext(ctx, deleteIngredientUsage, id)
	var product_id int64
	err := row.Scan(&product_id)
	return product_id, err
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

const deleteUnit = `-- name: DeleteUnit :execrows
;

delete from units
where id = ?
`

func (q *Queries) DeleteUnit(ctx context.Context, id int64) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteUnit, id)
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

const getIngredient = `-- name: GetIngredient :one
;

select id, name
from ingredients
where id = ?
`

func (q *Queries) GetIngredient(ctx context.Context, id int64) (Ingredient, error) {
	row := q.db.QueryRowContext(ctx, getIngredient, id)
	var i Ingredient
	err := row.Scan(&i.ID, &i.Name)
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

const getIngredientUsageForProductWithPrice = `-- name: GetIngredientUsageForProductWithPrice :many
;

select iu.id, iu.quantity, iu.unit_id, iu.ingredient_id, iu.product_id, i.id, i.name, ip.id, ip.time_stamp, ip.price, ip.quantity, ip.unit_id, ip.ingredient_id, ip.base_product_id
from ingredient_usage iu
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
where iu.product_id = ?
`

type GetIngredientUsageForProductWithPriceRow struct {
	ID             int64    `json:"id"`
	Quantity       float64  `json:"quantity"`
	UnitID         int64    `json:"unit_id"`
	IngredientID   int64    `json:"ingredient_id"`
	ProductID      int64    `json:"product_id"`
	ID_2           *int64   `json:"id_2"`
	Name           *string  `json:"name"`
	ID_3           *int64   `json:"id_3"`
	TimeStamp      *int64   `json:"time_stamp"`
	Price          *float64 `json:"price"`
	Quantity_2     *float64 `json:"quantity_2"`
	UnitID_2       *int64   `json:"unit_id_2"`
	IngredientID_2 *int64   `json:"ingredient_id_2"`
	BaseProductID  *int64   `json:"base_product_id"`
}

func (q *Queries) GetIngredientUsageForProductWithPrice(ctx context.Context, productID int64) ([]GetIngredientUsageForProductWithPriceRow, error) {
	rows, err := q.db.QueryContext(ctx, getIngredientUsageForProductWithPrice, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetIngredientUsageForProductWithPriceRow
	for rows.Next() {
		var i GetIngredientUsageForProductWithPriceRow
		if err := rows.Scan(
			&i.ID,
			&i.Quantity,
			&i.UnitID,
			&i.IngredientID,
			&i.ProductID,
			&i.ID_2,
			&i.Name,
			&i.ID_3,
			&i.TimeStamp,
			&i.Price,
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

const getIngredientsFromUnit = `-- name: GetIngredientsFromUnit :many
;

select distinct i.id, i.name
from ingredient_prices ip
join ingredients i on i.id = ip.ingredient_id
where ip.unit_id = ?
`

func (q *Queries) GetIngredientsFromUnit(ctx context.Context, unitID int64) ([]Ingredient, error) {
	rows, err := q.db.QueryContext(ctx, getIngredientsFromUnit, unitID)
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

const getIngredientsWithPriceUnit = `-- name: GetIngredientsWithPriceUnit :many
select
    i.id, i.name,
    ip.id as price_id,
    ip.price,
    ip.unit_id,
    ip.quantity,
    ip.time_stamp,
    ip.base_product_id
from ingredients i
left join
    ingredient_prices ip
    on ip.id = (
        select id
        from ingredient_prices as ip2
        where ip2.ingredient_id = i.id
        order by time_stamp desc
        limit?1
    )
where (?2 is null or i.id =?2)
`

type GetIngredientsWithPriceUnitParams struct {
	PriceLimit   int64       `json:"price_limit"`
	IngredientID interface{} `json:"ingredient_id"`
}

type GetIngredientsWithPriceUnitRow struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	PriceID       *int64   `json:"price_id"`
	Price         *float64 `json:"price"`
	UnitID        *int64   `json:"unit_id"`
	Quantity      *float64 `json:"quantity"`
	TimeStamp     *int64   `json:"time_stamp"`
	BaseProductID *int64   `json:"base_product_id"`
}

func (q *Queries) GetIngredientsWithPriceUnit(ctx context.Context, arg GetIngredientsWithPriceUnitParams) ([]GetIngredientsWithPriceUnitRow, error) {
	rows, err := q.db.QueryContext(ctx, getIngredientsWithPriceUnit, arg.PriceLimit, arg.IngredientID)
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

const getProductCost = `-- name: GetProductCost :one
;

select product_id, cost
from product_cost_cache
where product_id = ?
`

func (q *Queries) GetProductCost(ctx context.Context, productID int64) (ProductCostCache, error) {
	row := q.db.QueryRowContext(ctx, getProductCost, productID)
	var i ProductCostCache
	err := row.Scan(&i.ProductID, &i.Cost)
	return i, err
}

const getProductNames = `-- name: GetProductNames :many
;

select p.id, p.name
from products p
`

type GetProductNamesRow struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (q *Queries) GetProductNames(ctx context.Context) ([]GetProductNamesRow, error) {
	rows, err := q.db.QueryContext(ctx, getProductNames)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetProductNamesRow
	for rows.Next() {
		var i GetProductNamesRow
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
	ID            int64   `json:"id"`
	Name          string  `json:"name"`
	Price         float64 `json:"price"`
	Multiplicator float64 `json:"multiplicator"`
	CategoryID    int64   `json:"category_id"`
	Cost          float64 `json:"cost"`
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

const getProductsFromIngredient = `-- name: GetProductsFromIngredient :many
;

select distinct p.id, p.name
from ingredient_usage iu
join products p on p.id = iu.product_id
where iu.ingredient_id = ?
`

type GetProductsFromIngredientRow struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (q *Queries) GetProductsFromIngredient(ctx context.Context, ingredientID int64) ([]GetProductsFromIngredientRow, error) {
	rows, err := q.db.QueryContext(ctx, getProductsFromIngredient, ingredientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetProductsFromIngredientRow
	for rows.Next() {
		var i GetProductsFromIngredientRow
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

const getProductsFromUnit = `-- name: GetProductsFromUnit :many
;

select distinct p.id, p.name
from ingredient_usage iu
join products p on p.id = iu.product_id
where iu.unit_id = ?
`

type GetProductsFromUnitRow struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (q *Queries) GetProductsFromUnit(ctx context.Context, unitID int64) ([]GetProductsFromUnitRow, error) {
	rows, err := q.db.QueryContext(ctx, getProductsFromUnit, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetProductsFromUnitRow
	for rows.Next() {
		var i GetProductsFromUnitRow
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

const getProductsWithCost = `-- name: GetProductsWithCost :many
;

select p.id, p.name, p.price, p.multiplicator, p.category_id, pc.cost
from products p
left join product_cost_cache pc on pc.product_id = p.id
`

type GetProductsWithCostRow struct {
	ID            int64    `json:"id"`
	Name          string   `json:"name"`
	Price         float64  `json:"price"`
	Multiplicator float64  `json:"multiplicator"`
	CategoryID    int64    `json:"category_id"`
	Cost          *float64 `json:"cost"`
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
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Price          float64  `json:"price"`
	Multiplicator  float64  `json:"multiplicator"`
	CategoryID     int64    `json:"category_id"`
	ID_2           *int64   `json:"id_2"`
	Quantity       *float64 `json:"quantity"`
	UnitID         *int64   `json:"unit_id"`
	IngredientID   *int64   `json:"ingredient_id"`
	ProductID      *int64   `json:"product_id"`
	ID_3           *int64   `json:"id_3"`
	Name_2         *string  `json:"name_2"`
	ID_4           *int64   `json:"id_4"`
	TimeStamp      *int64   `json:"time_stamp"`
	Price_2        *float64 `json:"price_2"`
	Quantity_2     *float64 `json:"quantity_2"`
	UnitID_2       *int64   `json:"unit_id_2"`
	IngredientID_2 *int64   `json:"ingredient_id_2"`
	BaseProductID  *int64   `json:"base_product_id"`
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

const getUnit = `-- name: GetUnit :one
;

select id, name, base_unit_id, factor
from units
where id = ?
`

func (q *Queries) GetUnit(ctx context.Context, id int64) (Unit, error) {
	row := q.db.QueryRowContext(ctx, getUnit, id)
	var i Unit
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.BaseUnitID,
		&i.Factor,
	)
	return i, err
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

const insertIngredient = `-- name: InsertIngredient :one
;

insert into ingredients(name)
values (?)
returning id, name
`

func (q *Queries) InsertIngredient(ctx context.Context, name string) (Ingredient, error) {
	row := q.db.QueryRowContext(ctx, insertIngredient, name)
	var i Ingredient
	err := row.Scan(&i.ID, &i.Name)
	return i, err
}

const insertProductCost = `-- name: InsertProductCost :execrows
;

insert into product_cost_cache (product_id, cost)
values (?, ?)
on conflict (product_id) do update
set cost = excluded.cost
`

type InsertProductCostParams struct {
	ProductID int64   `json:"product_id"`
	Cost      float64 `json:"cost"`
}

func (q *Queries) InsertProductCost(ctx context.Context, arg InsertProductCostParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, insertProductCost, arg.ProductID, arg.Cost)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const insertUnit = `-- name: InsertUnit :one
;

insert into units (name, base_unit_id, factor)
values (?, ?, ?)
returning id, name, base_unit_id, factor
`

type InsertUnitParams struct {
	Name       string  `json:"name"`
	BaseUnitID *int64  `json:"base_unit_id"`
	Factor     float64 `json:"factor"`
}

func (q *Queries) InsertUnit(ctx context.Context, arg InsertUnitParams) (Unit, error) {
	row := q.db.QueryRowContext(ctx, insertUnit, arg.Name, arg.BaseUnitID, arg.Factor)
	var i Unit
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.BaseUnitID,
		&i.Factor,
	)
	return i, err
}

const putCategory = `-- name: PutCategory :one
;

insert into categories (name, vat)
values (?,?)
returning id, name, vat
`

type PutCategoryParams struct {
	Name string `json:"name"`
	Vat  int64  `json:"vat"`
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
	Quantity     float64 `json:"quantity"`
	UnitID       int64   `json:"unit_id"`
	IngredientID int64   `json:"ingredient_id"`
	ProductID    int64   `json:"product_id"`
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

const putIngredientPrice = `-- name: PutIngredientPrice :one
;

insert into ingredient_prices (ingredient_id, price, quantity, unit_id, base_product_id)
values (?, ?, ?, ?, ?)
returning id, time_stamp, price, quantity, unit_id, ingredient_id, base_product_id
`

type PutIngredientPriceParams struct {
	IngredientID  int64    `json:"ingredient_id"`
	Price         *float64 `json:"price"`
	Quantity      float64  `json:"quantity"`
	UnitID        int64    `json:"unit_id"`
	BaseProductID *int64   `json:"base_product_id"`
}

func (q *Queries) PutIngredientPrice(ctx context.Context, arg PutIngredientPriceParams) (IngredientPrice, error) {
	row := q.db.QueryRowContext(ctx, putIngredientPrice,
		arg.IngredientID,
		arg.Price,
		arg.Quantity,
		arg.UnitID,
		arg.BaseProductID,
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
	Name       string `json:"name"`
	CategoryID int64  `json:"category_id"`
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
	Name string `json:"name"`
	Vat  int64  `json:"vat"`
	ID   int64  `json:"id"`
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
	Name string `json:"name"`
	ID   int64  `json:"id"`
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
	Quantity float64 `json:"quantity"`
	UnitID   int64   `json:"unit_id"`
	ID       int64   `json:"id"`
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
	Name          string  `json:"name"`
	CategoryID    int64   `json:"category_id"`
	Price         float64 `json:"price"`
	Multiplicator float64 `json:"multiplicator"`
	ID            int64   `json:"id"`
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

const updateUnit = `-- name: UpdateUnit :one
;

update units
set name=?, base_unit_id=?, factor=?
where id=?
returning id, name, base_unit_id, factor
`

type UpdateUnitParams struct {
	Name       string  `json:"name"`
	BaseUnitID *int64  `json:"base_unit_id"`
	Factor     float64 `json:"factor"`
	ID         int64   `json:"id"`
}

func (q *Queries) UpdateUnit(ctx context.Context, arg UpdateUnitParams) (Unit, error) {
	row := q.db.QueryRowContext(ctx, updateUnit,
		arg.Name,
		arg.BaseUnitID,
		arg.Factor,
		arg.ID,
	)
	var i Unit
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.BaseUnitID,
		&i.Factor,
	)
	return i, err
}
