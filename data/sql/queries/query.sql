-- name: GetIngredientsWithPriceUnit :many
select i.*, ip.id as price_id, ip.price, ip.unit_id, ip.quantity, ip.time_stamp
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
;

-- name: GetIngredientWithPriceUnit :one
select i.*, ip.id as price_id, ip.price, ip.unit_id, ip.quantity, ip.time_stamp
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

-- name: DeleteIngredient :execrows
delete from ingredients
where (id = ?)
;

-- name: PutIngredientPrice :one
insert into ingredient_prices (ingredient_id, price, quantity, unit_id)
values (?, ?, ?, ?)
returning *
;

-- name: PutIngredeintUsage :one
insert into ingredient_usage (quantity, unit_id, ingredient_id, product_id)
values (?, ?, ?, ?)
returning *
;

-- name: GetUnits :many
select *
from units
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

-- name: GetIngredientUsageForProduct :many
select *
from ingredient_usage iu
where iu.product_id = ?
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

-- name: DeleteIngredientUsage :execrows
delete from ingredient_usage
where (id = ?)
;

-- name: GetProductsWithCost :many
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

