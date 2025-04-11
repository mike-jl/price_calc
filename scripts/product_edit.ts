interface Product {
    id: number;
    name: string;
    price: number;
    multiplicator: number;
    category_id: number;
}

interface ProductWithCost {
    product: Product;
    cost: number;
}

interface Category {
    id: number;
    name: string;
    vat: number;
}

interface IngredientUsage {
    id: number;
    quantity: number;
    unit_id: number;
    ingredient_id: number;
    product_id: number;
}

interface Ingredient {
    id: number;
    name: string;
}

interface IngredientPrice {
    id: number;
    time_stamp: number;
    price: number;
    quantity: number;
    unit_id: number;
    ingredient_id: number;
    base_product_id: number | null;
}

interface IngredientWithPrices {
    ingredient: Ingredient;
    prices: IngredientPrice[];
}

interface Unit {
    id: number;
    name: string;
    base_unit_id: number | null;
    factor: number;
}

interface IngredientUsageExtended extends IngredientUsage {
    unit?: Unit;
    ingredient?: IngredientWithPrices;
    editing: boolean;
}

interface ProductEditViewModel {
    product: ProductWithCost;
    categories: Category[];
    ingredient_usages: IngredientUsage[];
    ingredients: IngredientWithPrices[];
    units: Unit[];
}

export function getProductEditData(): ProductEditViewModel & {
    ingredient_usages: IngredientUsageExtended[];
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
    init: () => void;
} {
    const vmText = document.getElementById('viewModel')!.textContent!;
    const parsedVm: ProductEditViewModel = JSON.parse(vmText);

    return {
        ...parsedVm,
        ingredient_usages: (parsedVm.ingredient_usages ?? []).map(usage =>
            modifyIngredientUsage(usage, parsedVm.units, parsedVm.ingredients)
        ),
        selectedCat: parsedVm.categories.map(c => c.id).indexOf(parsedVm.product.product.category_id),
        newIngredientId: 0,
        newIngredientAmount: 0,
        newIngredientUnitId: 0,
        usageBackup: {},

        getFilteredUnitsForUnitId(unitId: number): Unit[] {
            const unit = this.units.find(u => u.id === unitId);
            if (!unit) return [];
            const baseUnitId = unit.base_unit_id ?? unit.id;
            return this.units.filter(
                u => u.id === baseUnitId || u.base_unit_id === baseUnitId
            );
        },

        getSafeUnitIdFromIngredient(ingredientId: number): number | null {
            const ingredient = this.ingredients.find(i => i.ingredient.id === ingredientId);
            if (!ingredient || ingredient.prices.length === 0) return null;
            return ingredient.prices[0].unit_id;
        },

        get newIngredientCost(): string {
            const ingredient = this.ingredients.find(i => i.ingredient.id === this.newIngredientId);
            const unit = this.units.find(u => u.id === this.newIngredientUnitId);
            if (!ingredient || !unit || isNaN(this.newIngredientAmount) || ingredient.prices.length === 0) {
                return "0.00";
            }
            return (ingredient.prices[0].price * this.newIngredientAmount / unit.factor).toFixed(2);
        },

        get productCost(): string {
            return this.ingredient_usages.reduce((cost, usage) => {
                const ingredient = this.ingredients.find(i => i.ingredient.id === usage.ingredient_id);
                if (!ingredient?.prices || ingredient.prices.length === 0) return cost;
                return cost + ingredient.prices[0].price * usage.quantity;
            }, 0).toFixed(2);
        },

        startEditing(usage: IngredientUsageExtended): void {
            this.usageBackup[usage.id] = JSON.parse(JSON.stringify(usage));
            usage.editing = true;
        },

        cancelEditing(usage: IngredientUsageExtended): void {
            const backup = this.usageBackup[usage.id];
            if (backup) {
                Object.assign(usage, backup);
                delete this.usageBackup[usage.id];
            }
            usage.editing = false;
        },

        init(): void {
            // @ts-ignore - Alpine magic property
            this.$watch(() => this.getSafeUnitIdFromIngredient(this.newIngredientId), (unitId: number) => {
                this.newIngredientUnitId = unitId;
            });

            window.addEventListener("ingredient-added", (e) => {
                const { detail } = e as CustomEvent<{ ingredientUsage: IngredientUsage }>;
                const newUsage = detail.ingredientUsage;
                this.ingredient_usages.push(
                    modifyIngredientUsage(newUsage, this.units, this.ingredients)
                );
            });
        }
    };
}


function modifyIngredientUsage(
    usage: IngredientUsage,
    units: Unit[],
    ingredients: IngredientWithPrices[]
): IngredientUsageExtended {
    const unit = units.find(u => u.id === usage.unit_id);
    const ingredient = ingredients.find(i => i.ingredient.id === usage.ingredient_id);
    return {
        ...usage,
        unit,
        ingredient,
        editing: false,
    };
}

