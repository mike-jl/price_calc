import { IngredientWithPrices, Unit } from './common';

export interface Product {
    id: number;
    name: string;
    price: number;
    multiplicator: number;
    category_id: number;
}

export interface ProductWithCost {
    product: Product;
    cost: number;
}

export interface Category {
    id: number;
    name: string;
    vat: number;
}

export interface IngredientUsage {
    id: number;
    quantity: number;
    unit_id: number;
    ingredient_id: number;
    product_id: number;
}

export interface IngredientUsageExtended extends IngredientUsage {
    unit?: Unit;
    ingredient?: IngredientWithPrices;
    editing: boolean;
    displayAmount: string
}

export interface ProductEditViewModel {
    product: ProductWithCost;
    categories: Category[];
    ingredient_usages: IngredientUsage[];
    ingredients: IngredientWithPrices[];
    units: Unit[];
}

export type ProductEditData = ProductEditViewModel & {
    ingredient_usages_ext: IngredientUsageExtended[];
    selectedCat: number;
    newIngredientId: number;
    newIngredientAmount: number;
    newIngredientUnitId: number;
    usageBackup: Record<number, IngredientUsageExtended>;
    getFilteredUnitsForUnitId: (unitId: number) => Unit[];
    getSafeUnitIdFromIngredient: (ingredientId: number) => number | null;
    readonly newIngredientCost: string;
    readonly productCost: string;
    startEditing: (usage: IngredientUsageExtended) => void;
    cancelEditing: (usage: IngredientUsageExtended) => void;
    removeUsage: (usageId: number) => void;
    init: () => void;
    modifyIngredientUsage: (usage: IngredientUsage, units: Unit[], ingredients: IngredientWithPrices[]) => IngredientUsageExtended;
};

