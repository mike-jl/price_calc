// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.857
package components

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"github.com/mike-jl/price_calc/db"
	"github.com/mike-jl/price_calc/viewModels"
)

func ProductEdit(viewModel viewmodels.ProductEditViewModel) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templ.JSONScript("viewModel", viewModel).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div x-data=\"productEditData\"><section class=\"section hero is-info custom block\"><div class=\"container\"><div class=\"hero-body p-0\"><div class=\"columns\"><div class=\"column\"><div class=\"field\"><label class=\"label\">Name</label><div class=\"control\"><input class=\"input\" x-model=\"product.product.name\" name=\"name\" form=\"product-edit-form\"></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label\">Category</label><div class=\"control is-expanded\"><div class=\"select is-fullwidth\"><select name=\"category\" x-model=\"selectedCat\" form=\"product-edit-form\"><template x-for=\"(cat, i) in categories\" :key=\"i\"><option :value=\"i\" :selected=\"selectedCat === i\" x-text=\"cat.name + &#39; - &#39; + cat.vat + &#39;%&#39;\"></option></template></select></div></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label\">Gross Price (calculated)</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" disabled :value=\"(productCost * product.product.multiplicator * (1+(categories[selectedCat].vat/100))).toFixed(2)\"></p><p class=\"control\"><a class=\"button is-static\">€</a></p></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label\">Real Price</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" name=\"price\" :value=\"product.product.price\" form=\"product-edit-form\"></p><p class=\"control\"><a class=\"button is-static\">€</a></p></div></div></div></div><div class=\"columns border\"><div class=\"column\"><div class=\"field\"><label class=\"label\">Cost</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" disabled :value=\"productCost\"></p><p class=\"control\"><a class=\"button is-static\">€</a></p></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label\">Multiplicator</label><div class=\"field has-addons\"><div class=\"control is-expanded\"><input type=\"text\" name=\"multiplicator\" x-model=\"product.product.multiplicator\" class=\"input\" :class=\"{\n\t\t\t\t\t\t\t\t\t\t&#39;is-danger&#39;: !/^\\s*\\d*(\\.\\d+)?\\s*$/.test(product.product.multiplicator)\n\t\t\t\t\t\t\t\t\t}\" form=\"product-edit-form\"></div></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label\">Net Price (calculated)</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" disabled :value=\"(productCost * product.product.multiplicator).toFixed(2)\"></p><p class=\"control\"><a class=\"button is-static\">€</a></p></div></div></div><form :hx-post=\"`/product/${product.product.id}`\" hx-swap=\"none\" id=\"product-edit-form\" class=\"column responsive-buttons\"><button id=\"product-modal-button\" class=\"button is-link\" type=\"submit\">Safe</button> <button class=\"button is-danger\" :hx-delete=\"`/product/${product.product.id}`\">Delete</button></form></div><div class=\"columns\"><div class=\"column\"><div class=\"field\"><label class=\"label\">New Ingredient</label><div class=\"control is-expanded\"><div class=\"select is-fullwidth\"><select name=\"ingredient\" x-model.number=\"newIngredientId\" :form=\"`ingredient-form-${product.product.id}`\"><option selected value=\"0\" disabled>Select Ingredient</option><template x-for=\"(ing, i) in ingredients\" :key=\"ing.ingredient.id\"><option :value=\"ing.ingredient.id\" x-text=\"ing.ingredient.name\"></option></template></select></div></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label\">Amount</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input is-fullwidth\" type=\"text\" :form=\"`ingredient-form-${product.product.id}`\" name=\"amount\" x-model=\"newIngredientAmount\"></p><p class=\"control\"><span class=\"select\"><select :form=\"`ingredient-form-${product.product.id}`\" name=\"unit\" x-model.number=\"newIngredientUnitId\"><template x-for=\"unit in getFilteredUnitsForUnitId(getSafeUnitIdFromIngredient(newIngredientId))\" :key=\"unit.id\"><option :value=\"unit.id\" x-text=\"unit.name\"></option></template></select></span></p></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label\">Cost</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" disabled :value=\"newIngredientCost\"></p><p class=\"control\"><a class=\"button is-static\">€</a></p></div></div></div><form :id=\"`ingredient-form-${product.product.id}`\" :hx-put=\"`/ingredient-usage/${product.product.id}`\" hx-swap=\"afterbegin\" hx-target=\"#htmx-script-dump\" class=\"responsive-buttons column\"><button class=\"button is-success\" type=\"submit\">Add</button></form></div></div></div></section><section class=\"section\"><div class=\"product-row container\"><template x-for=\"(usage, i) in ingredient_usages_ext\" :key=\"usage.id\"><div class=\"block\"><template x-if=\"!usage.editing\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = IngredientUsageRow().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "</template><template x-if=\"usage.editing\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = IngredientUsageRowEdit().Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</template></div></template></div></section></div><div id=\"htmx-script-dump\" hidden></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func IngredientUsageRow() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var2 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var2 == nil {
			templ_7745c5c3_Var2 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "<div class=\"columns\"><div class=\"column\"><div class=\"field\"><label class=\"label is-hidden-tablet product-label\">Ingredient</label><div class=\"control\"><input class=\"input\" type=\"text\" :value=\"usage.ingredient.ingredient.name\" disabled></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label is-hidden-tablet product-label\">Amount</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" disabled :value=\"(usage.quantity * usage.unit.factor).toFixed(2)\"></p><p class=\"control\"><a class=\"button is-static\" x-text=\"usage.unit.name\"></a></p></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label is-hidden-tablet product-label\">Cost</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" :value=\"(usage.ingredient.prices[0].price * usage.quantity).toFixed(2)\" disabled></p><p class=\"control\"><a class=\"button is-static\">€</a></p></div></div></div><div class=\"responsive-buttons column\"><button class=\"button is-link\" @click=\"startEditing(usage)\">Edit</button></div></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func IngredientUsageRowEdit() templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var3 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var3 == nil {
			templ_7745c5c3_Var3 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "<div class=\"columns\"><div class=\"column\"><div class=\"field\"><label class=\"label is-hidden-tablet product-label\">Ingredient</label><div class=\"control\"><input class=\"input\" type=\"text\" :value=\"usage.ingredient.ingredient.name\" disabled></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label is-hidden-tablet product-label\">Amount</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input is-fullwidth\" type=\"text\" :form=\"`ingredient-usage-form-${usage.id}`\" name=\"amount\" x-model=\"usage.displayAmount\" @input=\"\n\t\t\t\t\t\t\tconst parsed = parseFloat(usage.displayAmount);\n\t\t\t\t\t\t\tif (!isNaN(parsed)) usage.quantity = parsed / usage.unit.factor; \"></p><p class=\"control\"><span class=\"select\"><select :form=\"`ingredient-usage-form-${usage.id}`\" name=\"unit\"><template x-for=\"unit in getFilteredUnitsForUnitId(usage.unit_id)\" :key=\"unit.id\"><option :value=\"unit.id\" x-text=\"unit.name\" :selected=\"unit.id === usage.unit_id\"></option></template></select></span></p></div></div></div><div class=\"column\"><div class=\"field\"><label class=\"label is-hidden-tablet product-label\">Cost</label><div class=\"field has-addons\"><p class=\"control is-expanded\"><input class=\"input\" type=\"text\" :value=\"(usage.ingredient.prices[0].price * usage.quantity).toFixed(2)\" disabled></p><p class=\"control\"><a class=\"button is-static\">€</a></p></div></div></div><form class=\"responsive-buttons form column\" :id=\"`ingredient-usage-form-${usage.id}`\" hx-swap=\"none\" :hx-post=\"`/ingredient-usage/${usage.id}`\"><button type=\"submit\" class=\"button is-link\" title=\"Save\"><span class=\"is-hidden-tablet\">Save</span> <i class=\"fas fa-check is-hidden-mobile\"></i></button> <button type=\"button\" class=\"button\" @click=\"cancelEditing(usage)\" title=\"Cancel\"><span class=\"is-hidden-tablet\">Cancel</span> <i class=\"fas fa-times is-hidden-mobile\"></i></button> <button type=\"button\" class=\"button is-danger\" :hx-delete=\"`/ingredient-usage/${usage.id}`\" hx-swap=\"none\" x-init=\"htmx.process($el)\" @htmx:after-request=\"removeUsage(usage.id)\" title=\"Delete\"><span class=\"is-hidden-tablet\">Delete</span> <i class=\"fas fa-trash is-hidden-mobile\"></i></button></form></div>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

func NewIngredientUsage(usage db.IngredientUsage) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var4 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var4 == nil {
			templ_7745c5c3_Var4 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templ.JSONScript("new-ingredient-usage", usage).Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 6, "<script>\n\t\twindow.dispatchEvent(new CustomEvent(\"ingredient-added\", {\n\t\t\tdetail: {\n\t\t\t\tingredientUsage: JSON.parse(document.getElementById('new-ingredient-usage').textContent),\n\t\t\t},\n\t\t}))\n\t\tconst htmxScriptDump = document.getElementById('htmx-script-dump');\n\t\thtmxScriptDump.innerHTML = \"\";\n\t</script>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
