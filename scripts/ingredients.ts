import { IngredientsData, IngredientsViewModel, IngredientExtended, IngredientWithPrice } from "./types/ingredients";
import { Unit } from "./types/common";
import { createEditingHelpers } from "./utils";

export function getIngredientsData(): IngredientsData {
    const vmText = document.getElementById('viewModel')!.textContent!;
    const parsedVm: IngredientsViewModel = JSON.parse(vmText);
    console.log(parsedVm);

    return {
        ...parsedVm,
        newIngredientType: "price",
        ingredients_ext: [],
        ingredientBackup: {},
        startEditing: () => { },
        cancelEditing: () => { },
        removeItem: () => { },

        init(): void {
            this.transformIngredients();
            this.listenForIngredientEvents();

            const helpers = createEditingHelpers(this.ingredients_ext, this.ingredientBackup);
            Object.assign(this, helpers);
        },

        transformIngredients(): void {
            this.ingredients_ext = (this.ingredients ?? []).map((ingredient) =>
                this.modifyIngredient(ingredient)
            );
        },

        listenForIngredientEvents(): void {
            window.addEventListener("ingredient-added", (e) => {
                const { detail } = e as CustomEvent<{ newIngredient: IngredientWithPrice }>;
                const newIngredient = detail.newIngredient;
                console.log(newIngredient);
                this.ingredients_ext.push(
                    this.modifyIngredient(newIngredient)
                );
            });

        },

        getFilteredUnitsForUnitId(unitId: number): Unit[] {
            const unit = this.units[unitId];
            if (!unit) return [];
            const baseUnitId = unit.base_unit_id ?? unit.id;
            console.log(baseUnitId);
            console.log(unitId);
            const units = Object.values(this.units).filter(
                u => u.id === baseUnitId || u.base_unit_id === baseUnitId
            );
            console.log(units);
            return units;
        },

        setIngredientPrice(ingredient: IngredientExtended): void {
            const ingredientPrice = ingredient.price;
            const unit = this.units[ingredientPrice.unit_id];
            if (!unit) return;
            const parsed = parseFloat(ingredient.displayPrice);
            if (!Number.isNaN(parsed)) {
                ingredientPrice.price = (parsed / ingredientPrice.quantity) * unit.factor;
                console.log(ingredientPrice.price);
            }
        },

        setIngredientQuantity(ingredient: IngredientExtended): void {
            const parsed = parseFloat(ingredient.displayQuantity);
            if (!Number.isNaN(parsed)) ingredient.price.quantity = parsed
        },

        modifyIngredient(ingredient: IngredientWithPrice): IngredientExtended {
            const isBase = ingredient.price.base_product_id === null;
            const ingredientPrice = ingredient.price;
            const unit = this.units[ingredientPrice.unit_id];
            if (!unit) return ingredient as IngredientExtended;
            const displayPrice = ((ingredientPrice.price / unit.factor) * ingredientPrice.quantity).toFixed(2);
            return {
                ...ingredient,
                isBase: isBase,
                editing: false,
                displayPrice: displayPrice,
                displayQuantity: ingredientPrice.quantity.toFixed(2),
                unit: unit,
            };
        },

    };
}
