export interface Ingredient {
    id: number;
    name: string;
}

export interface IngredientPrice {
    id: number;
    time_stamp: number;
    price: number;
    quantity: number;
    unit_id: number;
    ingredient_id: number;
    base_product_id: number | null;
}

export interface IngredientWithPrices {
    ingredient: Ingredient;
    prices: IngredientPrice[];
}

export interface Unit {
    id: number;
    name: string;
    base_unit_id: number | null;
    factor: number;
}

export interface EditableWithId {
    id: number;
    editing: boolean;
}
