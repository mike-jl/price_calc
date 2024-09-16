function refreshModalHandler() {
    // Add a click event on buttons to open a specific modal
    (document.querySelectorAll('.js-modal-trigger') || []).forEach(($trigger) => {
        const modal = $trigger.dataset.target;
        const $target = document.getElementById(modal);

        $trigger.removeEventListener('click', () => {
            openModal($target);
        });

        $trigger.addEventListener('click', () => {
            openModal($target);
        });
    });

    // Add a click event on various child elements to close the parent modal
    (document.querySelectorAll('.modal-background, .modal-close, .modal-card-head .delete, .modal-card-foot .button') || []).forEach(($close) => {
        const $target = $close.closest('.modal');

        $close.removeEventListener('click', () => {
            closeModal($target);
        });

        $close.addEventListener('click', () => {
            closeModal($target);
        });
    });
}

// Functions to open and close a modal
function openModal($el) {
    $el.classList.add('is-active');
}

function closeModal($el) {
    $el.classList.remove('is-active');
}

function closeAllModals() {
    (document.querySelectorAll('.modal') || []).forEach(($modal) => {
        closeModal($modal);
    });
}

function refreshIngredientCost(e) {
    let productId = e.getAttribute('data-product-id')
    let cost = document.getElementById("cost-" + productId)
    let unit = document.getElementById("unit-" + productId)
    let amount = document.getElementById("amount-" + productId)
    let ingredient = document.getElementById("ingredient-" + productId)

    let newCost = 0.0
    try {
        let factor = unit.options[unit.selectedIndex].getAttribute('data-factor')
        let baseCost = ingredient.options[ingredient.selectedIndex].getAttribute('data-price')
        newCost = (amount.value / factor) * baseCost
    } catch (error) {

    }
    cost.value = ((Math.round(newCost * 100) / 100).toFixed(2))
}

function refreshProductCost(e) {
    let productId = e.getAttribute('data-product-id');
    let cost = document.getElementById('product-cost-input-' + productId);
    newCost = 0.0;
    (document.querySelectorAll(".ingredient-usage-column-" + productId) || []).forEach((ingredient) => {
        newCost += Number(ingredient.getAttribute('data-cost'));
    });
    // console.log(newCost)
    cost.value = ((Math.round(newCost * 100) / 100).toFixed(2));
    cost.setAttribute("data-cost", newCost);

    let gross = document.getElementById('product-gross-input-' + productId)
    let net = document.getElementById('product-net-input-' + productId)
    let mult = document.getElementById('product-multiplicator-input-' + productId)
    let cat = document.getElementById('product-cat-select-' + productId)
    let vat = cat.options[cat.selectedIndex].getAttribute('data-vat')

    let newNet = newCost * mult.value
    let newGross = newNet * (1 + (vat / 100))

    net.value = ((Math.round(newNet * 100) / 100).toFixed(2))
    gross.value = ((Math.round(newGross * 100) / 100).toFixed(2))
}

function getUnitIdFromSelect(productId) {
    let e = document.getElementById("ingredient-" + productId)
    let selected = e.options[e.selectedIndex]
    if (selected) {
        return unitId = selected.getAttribute('data-unit-id')
    }
    return -1
}

document.addEventListener('DOMContentLoaded', () => {
    //htmx.logger = function(elt, event, data) {
    //    if (console) {
    //        console.log(event, elt, data);
    //    }
    //}

    document.body.addEventListener('htmx:beforeSwap', function(evt) {
        if (evt.detail.isError) {
            alert(evt.detail.xhr.response);
        }
    });


    // Get all "navbar-burger" elements
    const $navbarBurgers = Array.prototype.slice.call(document.querySelectorAll('.navbar-burger'), 0);

    // Add a click event on each of them
    $navbarBurgers.forEach(el => {
        el.addEventListener('click', () => {

            // Get the target from the "data-target" attribute
            const target = el.dataset.target;
            const $target = document.getElementById(target);

            // Toggle the "is-active" class on both the "navbar-burger" and the "navbar-menu"
            el.classList.toggle('is-active');
            $target.classList.toggle('is-active');

        });
    });

    refreshModalHandler()

    // Add a keyboard event to close all modals
    document.addEventListener('keydown', (event) => {
        if (event.key === "Escape") {
            closeAllModals();
        }
    });
});
