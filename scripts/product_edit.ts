import Alpine from "alpinejs";
import {
    ProductEditViewModel,
    ProductEditData,
    IngredientUsageExtended,
    IngredientUsage,
} from "./types/product_edit"

import { Unit } from "./types/common";
import { createEditingHelpers } from "./utils";

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
        startEditing: () => { },
        cancelEditing: () => { },
        removeItem: () => { },

        init(): void {
            this.watchNewIngredientId();
            this.transformInitialUsages();
            this.listenForIngredientEvents();

            const helpers = createEditingHelpers(this.ingredient_usages_ext, this.usageBackup);
            Object.assign(this, helpers);
        },

        watchNewIngredientId(): void {
            const $this = this as Alpine.Magics<ProductEditData> & ProductEditData;
            $this.$watch('newIngredientId', (newId: number) => {
                this.newIngredientUnitId = this.getSafeUnitIdFromIngredient(newId) ?? 0;
            });
        },

        transformInitialUsages(): void {
            this.ingredient_usages_ext = (this.ingredient_usages ?? []).map((usage: IngredientUsage) =>
                this.modifyIngredientUsage(usage)
            );
        },

        listenForIngredientEvents(): void {
            window.addEventListener("ingredient-added", (e) => {
                const { detail } = e as CustomEvent<{ ingredientUsage: IngredientUsage }>;
                const newUsage = detail.ingredientUsage;
                this.ingredient_usages_ext.push(
                    this.modifyIngredientUsage(newUsage)
                );
            });
        },

        getFilteredUnitsForUnitId(unitId: number): Unit[] {
            const unit = this.units[unitId];
            if (!unit) return [];
            const baseUnitId = unit.base_unit_id ?? unit.id;
            return Object.values(this.units).filter(
                u => u.id === baseUnitId || u.base_unit_id === baseUnitId
            );
        },

        getSafeUnitIdFromIngredient(ingredientId: number): number | null {
            const ingredient = this.ingredients[ingredientId];
            if (!ingredient || ingredient.prices.length === 0) return null;
            return ingredient.prices[0].unit_id;
        },

        get newIngredientCost(): string {
            const ingredient = this.ingredients[this.newIngredientId];
            const unit = this.units[this.newIngredientUnitId];
            if (!ingredient || !unit || Number.isNaN(this.newIngredientAmount) || ingredient.prices.length === 0) {
                return "0.00";
            }
            return (ingredient.prices[0].price * this.newIngredientAmount / unit.factor).toFixed(2);
        },

        get productCost(): string {
            return this.ingredient_usages_ext.reduce((cost, usage) => {
                const ingredient = this.ingredients[usage.ingredient_id];
                if (!ingredient?.prices || ingredient.prices.length === 0) return cost;
                return cost + ingredient.prices[0].price * usage.quantity;
            }, 0).toFixed(2);
        },

        modifyIngredientUsage(
            usage: IngredientUsage,
        ): IngredientUsageExtended {
            const unit = this.units[usage.unit_id];
            const ingredient = this.ingredients[usage.ingredient_id];
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

