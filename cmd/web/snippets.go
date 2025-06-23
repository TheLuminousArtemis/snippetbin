package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/theluminousartemis/letsgo_snippetbox/internal/store"
)

type SnippetView struct {
	ID      int
	Title   string
	Content string
	Created time.Time
	Expires time.Time
}

func (app *application) snippetView(w http.ResponseWriter, r *http.Request) {
	// id, err := strconv.Atoi(r.PathValue("id"))
	idParam := chi.URLParam(r, "id")
	keyParam := r.URL.Query().Get("key")
	id, err := strconv.Atoi(idParam)
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	dsnippet, err := app.store.Snippets.Get(id)
	if err != nil {
		if errors.Is(err, store.ErrNoRecord) {
			http.NotFound(w, r)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	key, err := base64.RawURLEncoding.DecodeString(keyParam)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	plaintext, err := decryptAESGCM(dsnippet.Ciphertext, key, dsnippet.IV)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	snippet := &SnippetView{
		ID:      dsnippet.ID,
		Title:   dsnippet.Title,
		Content: string(plaintext),
		Created: dsnippet.Created,
		Expires: dsnippet.Expires,
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

	key, err := generateKey()
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	plaintext := form.Content
	ciphertext, nonce, err := encryptAESGCM([]byte(plaintext), key)

	if err != nil {
		app.serverError(w, r, err)
		return
	}

	snippet := &store.Snippet{
		Title:      form.Title,
		Ciphertext: ciphertext,
		IV:         nonce,
		Expires:    time.Now().AddDate(0, 0, form.Expires),
	}

	id, err := app.store.Snippets.Insert(snippet)
	if err != nil {
		app.serverError(w, r, err)
		return
	}
	app.sessionManager.Put(r.Context(), "flash", "Snippet successfully created!")
	// encodedKey := base64.StdEncoding.EncodeToString(key)
	encodedKey := base64.RawURLEncoding.EncodeToString(key)
	// http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d?key=%s", id, encodedKey), http.StatusSeeOther)

}

func generateKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	return key, err
}

func encryptAESGCM(plaintext, key []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}

	nonce = make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}

	ciphertext = gcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
}

func decryptAESGCM(ciphertext, key, nonce []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, errors.New("invalid nonce size")
	}

	return gcm.Open(nil, nonce, ciphertext, nil)
}
