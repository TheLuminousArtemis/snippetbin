package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"time"

	"github.com/theluminousartemis/letsgo_snippetbox/ui"
)

type templateData struct {
	Snippet *SnippetView
	// Snippets        []store.Snippet
	CurrentYear     int
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := fs.Glob(ui.Files, "html/pages/*.html")
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)
		// ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.html")
		// if err != nil {
		// 	return nil, err
		// }
		// ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		// if err != nil {
		// 	return nil, err
		// }
		// ts, err = ts.ParseFiles(page)
		// if err != nil {
		// 	return nil, err
		// }
		// cache[name] = ts
		patterns := []string{
			"html/base.html",
			// "html/pages/*.html",
			"html/partials/*.html",
			page,
		}
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}
		cache[name] = ts

	}
	return cache, nil
}

func humanDate(t time.Time) string {
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}
