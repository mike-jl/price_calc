export function getProductEditData() {
    const vm = document.getElementById('viewModel').textContent;
    // console.log('View model:', vm);

    /** @type {ProductEditViewModel} */
    const parsedVm = JSON.parse(vm);
    // console.log('Parsed view model:', parsedVm);
    return {
        ...parsedVm,
        ingredient_usages: (parsedVm.ingredient_usages ?? []).map(usage => modifyIngredientUsage(usage, parsedVm.units, parsedVm.ingredients)),
        selectedCat: parsedVm.categories.map(c => c.id).indexOf(parsedVm.product.product.category_id),
        newIngredientId: 0,
        newIngredientAmount: 0,
        newIngredientUnitId: 0,
        /**
         * @param {number} unitId
         * @returns {Unit[]}
         */
        getFilteredUnitsForUnitId(unitId) {
            const unit = this.units.find(u => u.id === unitId)
            if (!unit) return []
            let baseUnitId = unit.base_unit_id
            if (baseUnitId === null || baseUnitId === undefined) {
                baseUnitId = unit.id
            }

            const out = this.units.filter(u =>
                u.id === baseUnitId || u.base_unit_id === baseUnitId
            );

            return out;
        },
        /**
         * @param {number} ingredientId
         * @returns {number|null}
         */
        getSafeUnitIdFromIngredient(ingredientId) {
            const ingredient = this.ingredients.find(i => i.ingredient.id == ingredientId);
            // console.log('Ingredient:', ingredient);
            if (!ingredient || !ingredient.prices || ingredient.prices.length === 0) {
                return null;
            }
            return ingredient.prices[0].unit_id;
        },
        get newIngredientCost() {
            const ingredient = this.ingredients.find(i => i.ingredient.id === this.newIngredientId);
            const unit = this.units.find(u => u.id === this.newIngredientUnitId);
            if (!ingredient || !ingredient.prices || ingredient.prices.length === 0 || !unit || isNaN(this.newIngredientAmount)) {
                return '0.00';
            }

            return (ingredient.prices[0].price * this.newIngredientAmount / unit.factor).toFixed(2);
        },
        get productCost() {
            let cost = 0;
            for (const usage of this.ingredient_usages) {
                if (!usage.ingredient.prices || usage.ingredient.prices.length === 0) {
                    continue;
                }
                cost += usage.ingredient.prices[0].price * usage.quantity;
            }
            return cost.toFixed(2);
        },
        /**
         * @param {IngredientUsageExtended} usage
         * @returns {void}
        **/
        startEditing(usage) {
            this.usageBackup[usage.id] = JSON.parse(JSON.stringify(usage));
            usage.editing = true;
        },
        /**
         * @param {IngredientUsageExtended} usage
         * @returns {void}
        **/
        cancelEditing(usage) {
            if (this.usageBackup[usage.id]) {
                Object.assign(usage, this.usageBackup[usage.id]);
                delete this.usageBackup[usage.id];
            }
            usage.editing = false;
        },
        usageBackup: {},
        init() {
            // @ts-ignore
            this.$watch(() => this.getSafeUnitIdFromIngredient(this.newIngredientId), (/** @type {Number} unitId */  unitId) => {
                // console.log('New unit ID:', unitId);
                this.newIngredientUnitId = unitId;
            });
            window.addEventListener('ingredient-added', (/** @type {CustomEvent} */ e) => {
                const newUsage = e.detail.ingredientUsage;
                this.ingredient_usages.push(modifyIngredientUsage(newUsage, this.units, this.ingredients));
            });
        }

    };
}

/**
 * @param {IngredientUsage} usage
 * @param {Unit[]} units
 * @param {IngredientWithPrices[]} ingredients
 * @returns {IngredientUsageExtended}
 **/
function modifyIngredientUsage(usage, units, ingredients) {
    const unit = units.find(u => u.id === usage.unit_id);
    const ingredient = ingredients.find(i => i.ingredient.id === usage.ingredient_id);
    return {
        ...usage,
        unit: unit,
        ingredient: ingredient,
        editing: false,
    };
}


/**
 * @typedef {object} Product
 * @property {number} id
 * @property {string} name
 * @property {number} price
 * @property {number} multiplicator
 * @property {number} category_id
 */

/**
 * @typedef {object} ProductWithCost
 * @property {Product} product
 * @property {number} cost
 */

/**
 * @typedef {object} Category
 * @property {number} id
 * @property {string} name
 * @property {number} vat
 */

/**
 * @typedef {object} IngredientUsage
 * @property {number} id
 * @property {number} quantity
 * @property {number} unit_id
 * @property {number} ingredient_id
 * @property {number} product_id
 */

/**
 * @typedef {object} IngredientUsageExtended
 * @augments {IngredientUsage}
 * @property {Unit} unit
 * @property {IngredientWithPrices} ingredient
 * @property {boolean} editing
 */

/**
 * @typedef {object} Ingredient
 * @property {number} id
 * @property {string} name
 */

/**
 * @typedef {object} IngredientPrice
 * @property {number} id
 * @property {number} time_stamp
 * @property {number} price
 * @property {number} quantity
 * @property {number} unit_id
 * @property {number} ingredient_id
 * @property {number|null} base_product_id
 */

/**
 * @typedef {object} IngredientWithPrices
 * @property {Ingredient} ingredient
 * @property {IngredientPrice[]} prices
 */

/**
 * @typedef {object} Unit
 * @property {number} id
 * @property {string} name
 * @property {number|null} base_unit_id
 * @property {number} factor
 */

/**
 * @typedef {object} ProductEditViewModel
 * @property {ProductWithCost} product
 * @property {Category[]} categories
 * @property {IngredientUsage[]} ingredient_usages
 * @property {IngredientWithPrices[]} ingredients
 * @property {Unit[]} units
 */

