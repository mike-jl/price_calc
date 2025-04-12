declare global {
    interface Window {
        Alpine: typeof Alpine;
        htmx: typeof import('htmx.org');
    }
}

import 'bulma/css/bulma.css';
import '@fortawesome/fontawesome-free/css/all.css';

import Alpine from 'alpinejs'
import htmx from 'htmx.org'

window.htmx = htmx

import { getProductEditData } from './product_edit'

Alpine.data('productEditData', getProductEditData)

window.Alpine = Alpine

Alpine.start()

