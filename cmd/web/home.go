package main

import "net/http"

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	// data.Snippets = snippets
	app.render(w, r, http.StatusOK, "home.html", data)

}
