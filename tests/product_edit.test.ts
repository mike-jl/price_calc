import { describe, expect, it } from 'vitest';
import { createProductEditModel } from '../scripts/product_edit';
import type { ProductEditViewModel } from '../scripts/types/product_edit';

const minimalModel: ProductEditViewModel = {
    product: {
        product: {
            id: 1,
            name: 'Test Product',
            price: 0,
            multiplicator: 1,
            category_id: 1,
        },
        cost: 0,
    },
    categories: [
        { id: 1, name: 'Test Category', vat: 0 },
        { id: 2, name: 'Another Category', vat: 0 },
    ],
    ingredient_usages: [],
    ingredients: {},
    units: {},
};

describe('productCost', () => {
    it('computes cost correctly', () => {
        const vm = createProductEditModel(minimalModel); // if needed to suppress missing fields
        // or define a complete stub above with all fields filled in

        vm.ingredient_usages_ext = [
            {
                id: 1,
                ingredient_id: 1,
                quantity: 2,
                unit_id: 1,
                product_id: 1,
                editing: false,
                displayAmount: '2.00',
            },
            {
                id: 2,
                ingredient_id: 2,
                quantity: 3,
                unit_id: 1,
                product_id: 1,
                editing: false,
                displayAmount: '3.00',
            },
        ];
        vm.ingredients = {
            1: {
                ingredient: { id: 1, name: 'test1' },
                prices: [{
                    id: 1,
                    price: 5,
                    time_stamp: 5,
                    quantity: 3,
                    unit_id: 1,
                    ingredient_id: 1,
                    base_product_id: null
                }],
            },
            2: {
                ingredient: { id: 2, name: 'test2' },
                prices: [{
                    id: 2,
                    price: 3,
                    time_stamp: 5,
                    quantity: 3,
                    unit_id: 1,
                    ingredient_id: 2,
                    base_product_id: null
                }],
            },
        };

        expect(vm.productCost).toBe('19.00'); // 5*2 + 3*3 = 10 + 9 = 19
    });
});

