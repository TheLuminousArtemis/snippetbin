package main

import "net/http"

func (app *application) routes() *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(&neuteredFileSystem{http.Dir("./ui/static/")})
	mux.HandleFunc("GET /{$}", app.home)
	mux.HandleFunc("GET /snippet/view/{id}", app.snippetView)
	mux.HandleFunc("GET /snippet/create", app.snippetCreate)
	mux.HandleFunc("POST /snippet/create", app.snippetCreatePost)
	mux.HandleFunc("GET /about", app.about)
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	return mux
}
