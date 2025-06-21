package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/store"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {

	snippets, err := app.store.Snippets.Latest()
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Snippets = snippets
	app.render(w, r, http.StatusOK, "home.html", data)

}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	snippet, err := app.store.Snippets.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	data := app.newTemplateData(r)
	data.Snippet = snippet

	app.render(w, r, http.StatusOK, "view.html", data)
}

type snippetCreateForm struct {
	Title       string            `form:"title" validate:"required,max=100"`
	Content     string            `form:"content" validate:"required"`
	Expires     int               `form:"expires" validate:"required"`
	FieldErrors map[string]string `form:"-"`
}

func (app *application) snippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = snippetCreateForm{
		Expires: 365,
	}
	app.render(w, r, http.StatusOK, "create.html", data)
}

func (app *application) snippetCreatePost(w http.ResponseWriter, r *http.Request) {
	var form snippetCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	if err := validate.Struct(form); err != nil {
		form.FieldErrors = make(map[string]string)
		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				field := strings.ToLower(fe.Field())
				switch fe.Tag() {
				case "required":
					form.FieldErrors[field] = "This field cannot be blank"
				case "max":
					form.FieldErrors[field] = "This field cannot be more than 100 characters long"
				case "expires":
					form.FieldErrors[field] = "This field must equal to 1, 7 or 365"
				default:
					form.FieldErrors[field] = "This field is invalid"
				}
			}
		}
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}

	id, err := app.store.Snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")

	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)

}
