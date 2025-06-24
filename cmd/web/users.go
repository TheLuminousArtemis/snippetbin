package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/store"
	"golang.org/x/crypto/bcrypt"
)

// users

type userSignupForm struct {
	Username    string            `form:"username" validate:"required,min=3"`
	Email       string            `form:"email" validate:"required,email"`
	Password    string            `form:"password" validate:"required,min=8"`
	FieldErrors map[string]string `form:"-"`
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	app.render(w, r, http.StatusOK, "signup.html", data)

}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusAccepted)
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
				case "email":
					form.FieldErrors[field] = "This field must be a valid email address"
				case "min":
					form.FieldErrors[field] = "This field must be atleast 8 characters long"
				}
			}
		}
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		return
	}

	user := &store.User{
		Username: form.Username,
		Email:    form.Email,
	}

	if err := user.Password.Set(form.Password); err != nil {
		app.serverError(w, r, err)
		return
	}
	ctx := r.Context()
	err = app.store.Users.Insert(ctx, user)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			form.FieldErrors = make(map[string]string)
			form.FieldErrors["email"] = "Email is already in use"
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		case store.ErrDuplicateUsername:
			form.FieldErrors = make(map[string]string)
			form.FieldErrors["username"] = "Username is already in use"
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
		default:
			app.serverError(w, r, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Sign up successful. Please login")
	http.Redirect(w, r, "/user/login", http.StatusSeeOther)

}

type userLoginForm struct {
	Email          string            `form:"email" validate:"required,email"`
	Password       string            `form:"password" validate:"required"`
	FieldErrors    map[string]string `form:"-"`
	NonFieldErrors []string          `form:"-"`
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	app.render(w, r, http.StatusOK, "login.html", data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {

	var form userLoginForm
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
				case "email":
					form.FieldErrors[field] = "This field must be a valid email address"
				default:
					form.FieldErrors[field] = "This field is invalid"
				}
			}
		}
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		return
	}
	ctx := r.Context()
	user, err := app.store.Users.GetByEmail(ctx, form.Email)
	if err != nil {
		if errors.Is(err, store.ErrInvalidCredentials) {
			form.NonFieldErrors = []string{"Email or password is incorrect"}
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
			return
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	if err := user.Password.Compare(form.Password); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			form.NonFieldErrors = []string{"Email or password is incorrect"}
			data := app.newTemplateData(r)
			data.Form = form
			app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", user.ID)
	http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out successfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
