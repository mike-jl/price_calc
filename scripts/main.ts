declare global {
    interface Window {
        Alpine: typeof Alpine;
        htmx: typeof import('htmx.org');
    }
}

import 'bulma/css/bulma.css';
import '@fortawesome/fontawesome-free/webfonts/fa-solid-900.woff2';
import '@fortawesome/fontawesome-free/webfonts/fa-solid-900.ttf';
import '@fortawesome/fontawesome-free/css/all.css';

import htmx from 'htmx.org'
window.htmx = htmx

import Alpine from 'alpinejs'
window.Alpine = Alpine

import { getProductEditData } from './product_edit'
Alpine.data('productEditData', getProductEditData)

import { getIngredientsData } from './ingredients'
Alpine.data('ingredientsData', getIngredientsData)

Alpine.start()
