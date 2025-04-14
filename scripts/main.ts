import 'bulma/css/bulma.css';
import '@fortawesome/fontawesome-free/webfonts/fa-solid-900.woff2';
import '@fortawesome/fontawesome-free/webfonts/fa-solid-900.ttf';
import '@fortawesome/fontawesome-free/css/all.css';

import Alpine from 'alpinejs';
import { getProductEditData } from './product_edit';
Alpine.data('productEditData', getProductEditData);

import { getIngredientsData } from './ingredients';
Alpine.data('ingredientsData', getIngredientsData);

Alpine.start();

document.addEventListener('DOMContentLoaded', () => {
    // Global htmx error handler
    document.body.addEventListener('htmx:beforeSwap', (evt: Event) => {
        const detail = (evt as CustomEvent<{ xhr: XMLHttpRequest; isError: boolean }>).detail;
        if (detail?.isError) {
            alert(detail.xhr.responseText);
        }
    });

    // Bulma navbar burger toggling
    const burgers = Array.from(document.querySelectorAll('.navbar-burger'));

    for (const burger of burgers) {
        burger.addEventListener('click', () => {
            const targetId = (burger as HTMLElement).dataset.target;
            if (!targetId) return;

            const target = document.getElementById(targetId);
            if (!target) return;

            burger.classList.toggle('is-active');
            target.classList.toggle('is-active');
        });
    }
});

