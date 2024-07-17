-- +goose Up
-- +goose StatementBegin
CREATE TABLE categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    vat INTEGER NOT NULL
);

CREATE TABLE products (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    category_id INTEGER,
    FOREIGN KEY(category_id) REFERENCES categories(id)
    ON DELETE SET NULL
    ON UPDATE CASCADE
);

CREATE TABLE ingredients (
    id  INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE units (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    base_unit_id INTEGER,
    factor NOT NULL DEFAULT 1
);

INSERT INTO units(id, name, base_unit_id, factor)
values
    (1, "l", NULL, 1),
    (2, "ml", 1, 1000),
    (3, "cl", 1, 100),
    (10, "kg", NULL, 1),
    (11, "g", 1, 1000)
;

CREATE TABLE ingredient_prices (
    id INTEGER PRIMARY KEY,
    time_stamp INTEGER NOT NULL DEFAULT ( unixepoch('now') ),
    price REAL NOT NULL,
    quantity REAL NOT NULL,
    unit_id INTEGER NOT NULL,
    ingredient_id INTEGER NOT NULL,
    FOREIGN KEY(ingredient_id) REFERENCES ingredients(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
    FOREIGN KEY(unit_id) REFERENCES units(id)
);

CREATE TABLE ingredient_usage (
    id INTEGER PRIMARY KEY,
    quantity REAL NOT NULL,
    unit_id INTEGER NOT NULL,
    ingredient_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    FOREIGN KEY (ingredient_id) REFERENCES ingredients(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE,
    FOREIGN KEY(product_id) REFERENCES products(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
);
-- +goose StatementEnd


