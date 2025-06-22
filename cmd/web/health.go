package main

import (
	"net/http"
)

func (app *application) health(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"status": "ok",
	}
	err := app.writeJSON(w, http.StatusOK, data)
	if err != nil {
		app.serverError(w, r, err)
	}
}
