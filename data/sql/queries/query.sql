-- name: GetIngredients :many
select
    i.id, name, ip.id as price_id, time_stamp, price, quantity, unit_id, ingredient_id
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

-- name: GetProductsWithPrice :many
select p.*, cast(ifnull(sum(ip.price * iu.quantity), 0) as real) as price
from products p
left join ingredient_usage iu on iu.product_id = p.id
left join
    (select price, ingredient_id from ingredient_prices order by time_stamp limit 1) ip
    on iu.ingredient_id = ip.ingredient_id
group by iu.id
;

-- name: PutProduct :one
insert into products(name, category_id)
values (?, ?)
returning *
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

