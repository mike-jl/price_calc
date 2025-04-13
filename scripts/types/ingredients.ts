import { Unit, EditableWithId, Ingredient, IngredientPrice } from "./common";

export interface IngredientWithPrice extends Ingredient {
    price: IngredientPrice;
}

export interface IngredientsViewModel {
    product_names: { [productId: string]: string };
    ingredients: IngredientWithPrice[];
    units: Unit[];
}

export interface IngredientExtended extends IngredientWithPrice, EditableWithId {
    isBase: boolean;
    displayPrice: string;
    displayQuantity: string;
    unit: Unit;
}

export interface IngredientsData extends IngredientsViewModel {
    init: () => void;

    newIngredientType: string;

    ingredients_ext: IngredientExtended[];
    ingredientBackup: Record<number, IngredientExtended>;

    setIngredientPrice(ingredient: IngredientExtended): void
    setIngredientQuantity(ingredient: IngredientExtended): void
    getFilteredUnitsForUnitId(unitId: number): Unit[]

    startEditing: (usage: IngredientExtended) => void;
    cancelEditing: (usage: IngredientExtended) => void;
    removeItem: (itemId: number) => void;
    modifyIngredient(ingredient: IngredientWithPrice): IngredientExtended;
}
