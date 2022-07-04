package main

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"time"

	"github.com/hherman1/gorgina/db/persist"
)

func putForm(existing persist.Catalog) string {
	existing.Price.Float64 = math.Round(existing.Price.Float64*100) / 100
	return fmt.Sprintf(`
<div class="grid place-items-center">
<form hx-post="api/put" hx-target="#viewport" class="w-96 grid grid-cols-1 place-content-center">
	<input type="hidden" name="id" value="%v" />
	<label for="title"> Title </label> <input type="text" id="title" name="title" class="border-2 p-2" value="%v"> </input> <br/>
	<label for="description"> Description </label> <br /> <textarea name="description" id="description" class="border-2 p-2" value="%v">%v</textarea><br/>
	<label for="category"> Category </label> <select name="category" id="category" class="border-2 p-2" value="%v">
		<option value="tops">Tops</option>
		<option value="bottoms">Bottoms</option>
		<option value="dresses">Dresses</option>
		<option value="accessories">Accessories</option>
		<option value="shoes">Shoes</option>
	</select> <br />
	<label for="brand"> Brand </label> <input type="text" name="brand" id="brand" class="border-2 p-2" value="%v"/> <br/>
	<label for="color"> Color </label> <input type="text" name="color" id="color" class="border-2 p-2" value="%v"/> <br/>
	<label for="pattern"> Pattern </label> <input type="text" name="pattern" id="pattern" class="border-2 p-2" value="%v"/> <br/>
	<label for="price"> Price </label> <input type="text" name="price" id="price" class="border-2 p-2" value="%v" placeholder="30.99" /> <br/>
	<div class="border-2 p-2"> <input type="checkbox" name="used" id="used" value="true"/> <label for="used"> Use now </label> </div> <br/>
	<input type="submit"  class="border-2 p-2 rounded-full text-blue-100 bg-blue-600 hover:bg-blue-500 mb-4 cursor-pointer"/>
	<button class="border-2 p-2 rounded-full bg-slate-50 hover:bg-slate-100" hx-get="component/list" hx-target="#viewport"> Cancel </button>
</form>
</div>`,
		existing.ID,
		existing.Title.String,
		existing.Description.String,
		existing.Description.String,
		existing.Category.String,
		existing.Brand.String,
		existing.Color.String,
		existing.Pattern.String,
		existing.Price.Float64)
}

func listCatalog(items []persist.Catalog) (string, error) {
	const tmpl = `
<div class="flex flex-wrap">
	{{- range .}}
	{{template "item" .}}
	{{- end}}
</div>`

	t := template.Must(template.New("catalog").Parse(tmpl))
	template.Must(t.New("item").Funcs(template.FuncMap{
		"used": itemUsedRecently,
	}).Parse(itemTmpl))
	var bs bytes.Buffer
	err := t.Execute(&bs, items)
	if err != nil {
		return "", fmt.Errorf("execute tmpl: %w", err)
	}
	return bs.String(), nil
}

func itemUsedRecently(t time.Time) bool {
	return time.Now().Sub(t) < 30*time.Minute
}

const itemTmpl = `
	<div class="p-3 m-3 max-w-xs" id="list-{{.ID}}">
		<div class="text-lg {{ if used .LastActivity.Time }}decoration-green-500 underline decoration-2{{end}}"> <b>{{.Title.String}}</b> </div>
		{{if .LastActivity.Valid}}
		<div class="text-sm p-1">
			<span class="italic text-slate-400" timestamp="{{.LastActivity.Time.UnixMilli}}"> </span>
			<span hx-target="#viewport" hx-get="component/useHistory?id={{.ID}}" class="cursor-pointer bg-slate-100 p-1 rounded-lg hover:bg-slate-50">⏱</span>
		</div>
		{{end}}
		<div class="p-1"> {{.Description.String}} </div>
		<div class="p-1 italic">
			<span class="italic"> {{.Category.String}} </span>   ⸱ {{.Brand.String}} ⸱ {{.Color.String}} ⸱ {{.Pattern.String}}
		</div>
		<div class="p-1 text-green-800"> ${{printf "%.2f" .Price.Float64}} </div>

		<button hx-target="#viewport" hx-get="component/putCatalog?id={{.ID}}" class="p-2 text-slate-500 rounded-lg bg-slate-50 hover:bg-slate-100"> Edit </button>
		{{- if eq (used .LastActivity.Time) false }}
		<button hx-target="#list-{{.ID}}" hx-get="api/use?id={{.ID}}" hx-swap="outerHTML" class="p-2 rounded-lg text-green-600 bg-green-100 hover:bg-green-200"> Use </button>
		{{- end}}
		{{- if not .Hidden }}
		<button hx-target="#viewport" hx-get="api/hide?hidden=true&id={{.ID}}" class="p-2 rounded-lg text-slate-600 bg-slate-50 hover:bg-red-200"> Hide </button>
		{{- else }}
		<button hx-target="#viewport" hx-get="api/hide?hidden=false&id={{.ID}}" class="p-2 rounded-lg text-red-100 bg-red-500 hover:bg-red-400"> Unhide </button>
		{{- end }}

		<!-- Allows setting a description on the last use if there was something noteworthy -->
		{{- if used .LastActivity.Time }}
		<input type="text" name="note" placeholder="Use notes" class="p-2 mt-2 border-2 rounded-lg" hx-swap="none" hx-post="api/use/note?id={{.ID}}" hx-trigger="input" value="{{.LastNote.String}}" />
		{{- end}}

	</div>`

func renderCatalogItem(item persist.Catalog) (string, error) {
	t := template.Must(template.New("item").Funcs(template.FuncMap{
		"used": itemUsedRecently,
	}).Parse(itemTmpl))
	var bs bytes.Buffer
	err := t.Execute(&bs, item)
	if err != nil {
		return "", fmt.Errorf("execute tmpl: %w", err)
	}
	return bs.String(), nil
}

func renderEditableHistory(item persist.Catalog, history []persist.Activity) (string, error) {
	dot := struct {
		Item    persist.Catalog
		History []persist.Activity
	}{item, history}
	const tmpl = `
<div class="p-4">
	<h2 class="font-bold text-lg"> {{.Item.Title.String}} </h2>
	<table>
		<tr> <th class="p-2"> Time </th> <th class="p-2"> Note </th> </tr>
		{{- range .History }}
		<tr>
			<form id="form-{{.ID}}" hx-swap="none" hx-post="api/putUse" hx-trigger="input"/> </form>
			<input type="hidden" name="id" value="{{.ID}}" form="form-{{.ID}}"/>
			<input type="hidden" name="timezoneMs" form="form-{{.ID}}"/>
			<td class="p-2"> <input type="datetime-local" timestamp="{{.Ts.UnixMilli}}" hx-swap="none" hx-post="api/use/put" hx-trigger="input" name="time" form="form-{{.ID}}" hx-include="[form=form-{{.ID}}]"/> </td>
			<td class="p-2"> <input type="text" hx-swap="none" hx-post="api/use/put" hx-trigger="input" name="note" form="form-{{.ID}}" hx-include="[form=form-{{.ID}}]" value="{{.Note.String}}" class="p-2"/> </td>
		</tr>
		{{- end}}
	</div>
</div>
<script type="text/javascript">
document.querySelectorAll("[name=timezoneMs]").forEach(el => {
	el.value = new Date().getTimezoneOffset() * 60 * 1000
})
</script>
`
	t := template.Must(template.New("history").Parse(tmpl))
	var bs bytes.Buffer
	err := t.Execute(&bs, dot)
	if err != nil {
		return "", fmt.Errorf("execute tmpl: %w", err)
	}
	return bs.String(), nil
}
