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
<form hx-post="/api/put" hx-target="#viewport" class="w-96 grid grid-cols-1 place-content-center">
	<input type="hidden" name="id" value="%v" />
	<label for="title"> Title </label> <input type="text" name="title" class="border-2 p-2" value="%v"> </input> <br/>
	<label for="description"> Description </label> <br /> <textarea name="description" class="border-2 p-2" value="%v">%v</textarea><br/>
	<label for="category"> Category </label> <select name="category" class="border-2 p-2" value="%v">
		<option value="tops">Tops</option>
		<option value="bottoms">Bottoms</option>
		<option value="dresses">Dresses</option>
		<option value="accessories">Accessories</option>
		<option value="shoes">Shoes</option>
	</select> <br />
	<label for="brand"> Brand </label> <input type="text" name="brand" class="border-2 p-2" value="%v"/> <br/>
	<label for="color"> Color </label> <input type="text" name="color" class="border-2 p-2" value="%v"/> <br/>
	<label for="pattern"> Pattern </label> <input type="text" name="pattern" class="border-2 p-2" value="%v"/> <br/>
	<label for="price"> Price </label> <input type="text" name="price" class="border-2 p-2" value="%v" placeholder="30.99" /> <br/>
	<input type="submit"  class="border-2 p-2 rounded-full text-blue-100 bg-blue-600 hover:bg-blue-500 mb-4 cursor-pointer"/>
	<button class="border-2 p-2 rounded-full bg-slate-50 hover:bg-slate-100" hx-get="/component/list" hx-target="#viewport"> Cancel </button>
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
		<div>
			<span {{ if used .LastActivity.Time }}class="decoration-green-500 underline decoration-2"{{end}}> <b>{{.Title.String}}</b> </span>
			{{if .LastActivity.Valid}}<span class="ml-4 text-slate-400"> {{.LastActivity.Time.Local.Format "3:04PM 01/02/06"}}</span>{{end}}
		</div>
		<div class="p-1"> {{.Description.String}} </div>
		<div class="p-1"> {{.Category.String}} </div>
		<div class="p-1"> {{.Brand.String}} </div>
		<div class="p-1"> {{.Color.String}} </div>
		<div class="p-1"> {{.Pattern.String}} </div>
		<div class="p-1 text-green-800"> ${{printf "%.2f" .Price.Float64}} </div>

		<button hx-target="#viewport" hx-get="/component/putCatalog?id={{.ID}}" class="p-2 text-slate-500 rounded-lg bg-slate-50 hover:bg-slate-100"> Edit </button>
		{{- if eq (used .LastActivity.Time) false }}
		<button hx-target="#list-{{.ID}}" hx-get="/api/use?id={{.ID}}" class="p-2 rounded-lg text-green-600 bg-green-100 hover:bg-green-200"> Use </button>
		{{- end}}
	</div>`

func loggedCatalogItem(item persist.Catalog) (string, error) {
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
