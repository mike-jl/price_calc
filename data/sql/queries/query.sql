-- name: GetIngredientsWithPriceUnit :many
select
    i.*,
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
        limit:price_limit
    )
where (:ingredient_id is null or i.id =:ingredient_id)
;

-- name: PutIngredient :one
insert into ingredients(name)
values (?)
returning *
;

-- name: UpdateIngredient :one
update ingredients
set name=?
where ( id = ? )
returning *
;

-- name: GetIngredient :one
select *
from ingredients
where id = ?
;

-- name: DeleteIngredient :execrows
delete from ingredients
where (id = ?)
;

-- name: PutIngredientPrice :one
insert into ingredient_prices (ingredient_id, price, quantity, unit_id, base_product_id)
values (?, ?, ?, ?, ?)
returning *
;

-- name: PutIngredeintUsage :one
insert into ingredient_usage (quantity, unit_id, ingredient_id, product_id)
values (?, ?, ?, ?)
returning *
;

-- name: InsertUnit :one
insert into units (name, base_unit_id, factor)
values (?, ?, ?)
returning *
;

-- name: GetUnits :many
select *
from units
;

-- name: GetUnit :one
select *
from units
where id = ?
;

-- name: UpdateUnit :one
update units
set name=?, base_unit_id=?, factor=?
where id=?
returning *
;

-- name: DeleteUnit :execrows
delete from units
where id = ?
;

-- name: GetIngredientsFromUnit :many
select distinct i.id, i.name
from ingredient_prices ip
join ingredients i on i.id = ip.ingredient_id
where ip.unit_id = ?
;

-- name: GetProductsFromUnit :many
select distinct p.id, p.name
from ingredient_usage iu
join products p on p.id = iu.product_id
where iu.unit_id = ?
;

-- name: GetProductsWithIngredients :many
select p.*, iu.*, i.*, ip.*
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
;

-- name: GetIngredientUsageForProductWithPrice :many
select iu.*, i.*, ip.*
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
;

-- name: GetIngredientUsageForProduct :many
select *
from ingredient_usage iu
where iu.product_id = ?
;

-- name: InsertProductCost :execrows
insert into product_cost_cache (product_id, cost)
values (?, ?)
on conflict (product_id) do update
set cost = excluded.cost
;

-- name: GetIngredientUsage :one
select *
from ingredient_usage iu
where iu.id = ?
;

-- name: UpdateIngredientUsage :one
update ingredient_usage
set quantity=?, unit_id=?
where ( id = ? )
returning *
;

-- name: DeleteIngredientUsage :one
delete from ingredient_usage
where (id = ?)
returning product_id
;

-- name: GetProductsFromIngredient :many
select distinct p.id, p.name
from ingredient_usage iu
join products p on p.id = iu.product_id
where iu.ingredient_id = ?
;

-- name: GetProductsWithCost :many
select p.id, p.name, p.price, p.multiplicator, p.category_id, pc.cost
from products p
left join product_cost_cache pc on pc.product_id = p.id
;

-- name: GetProductCost :one
select *
from product_cost_cache
where product_id = ?
;

-- name: GetProductNames :many
select p.id, p.name
from products p
;

-- name: GetProductWithCost :one
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
;

-- name: PutProduct :one
insert into products(name, category_id)
values (?, ?)
returning *
;

-- name: UpdateProduct :one
update products
set name=?, category_id=?, price=?, multiplicator=?
where id=?
returning *
;

-- name: DeleteProduct :execrows
delete from products
where id = ?
;

-- name: GetCategories :many
select *
from categories
;

-- name: PutCategory :one
insert into categories (name, vat)
values (?,?)
returning *
;


-- name: UpdateCategory :one
update categories
set name=?, vat=?
where id=?
returning *
;

-- name: GetCategory :one
select *
from categories
where id = ?
;

