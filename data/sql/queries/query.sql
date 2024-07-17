-- name: GetIngredients :many
select *
from ingredients i
left join ingredient_prices ip on ip.ingredient_id = i.id
order by i.name collate nocase asc, ip.time_stamp desc
;

-- name: PutIngredient :one
insert into ingredients (name)
values (?)
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

