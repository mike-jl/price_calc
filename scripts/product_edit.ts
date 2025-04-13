import Alpine from "alpinejs";
import {
    ProductEditViewModel,
    ProductEditData,
    IngredientUsageExtended,
    IngredientUsage,
} from "./types/product_edit"

import { Unit, IngredientWithPrices } from "./types/common";

export function getProductEditData(): ProductEditData {
    const vmText = document.getElementById('viewModel')!.textContent!;
    const parsedVm: ProductEditViewModel = JSON.parse(vmText);

    return {
        ...parsedVm,
        ingredient_usages_ext: [],
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
            return this.ingredient_usages_ext.reduce((cost, usage) => {
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
        removeUsage(usageId: number): void {
            this.ingredient_usages_ext = this.ingredient_usages_ext.filter((u) => u.id !== usageId);
        },
        init(): void {
            const $this = this as Alpine.Magics<any> & ProductEditData;
            $this.$watch('newIngredientId', (newId: number) => {
                this.newIngredientUnitId = this.getSafeUnitIdFromIngredient(newId) ?? 0;
            });

            this.ingredient_usages_ext = (this.ingredient_usages ?? []).map((usage: IngredientUsage) =>
                this.modifyIngredientUsage(usage, this.units, this.ingredients)
            );

            window.addEventListener("ingredient-added", (e) => {
                const { detail } = e as CustomEvent<{ ingredientUsage: IngredientUsage }>;
                const newUsage = detail.ingredientUsage;
                this.ingredient_usages_ext.push(
                    this.modifyIngredientUsage(newUsage, this.units, this.ingredients)
                );
            });
        },
        modifyIngredientUsage(
            usage: IngredientUsage,
            units: Unit[],
            ingredients: IngredientWithPrices[]
        ): IngredientUsageExtended {
            const unit = units.find(u => u.id === usage.unit_id);
            const ingredient = ingredients.find(i => i.ingredient.id === usage.ingredient_id);
            if (!unit || !ingredient) {
                throw new Error(`Unit or ingredient not found for usage ID: ${usage.id}`);
            }
            const displayAmount = (usage.quantity * unit.factor).toFixed(2);

            return {
                ...usage,
                unit,
                ingredient,
                editing: false,
                displayAmount,
            };
        }
    };
}



