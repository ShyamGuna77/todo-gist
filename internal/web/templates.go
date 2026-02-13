package web

import (
	"github.com/ShyamGuna77/rest-sms/internal/models"
	"html/template"
	"path/filepath"
	"time"
	"net/http"
)

type TemplateData struct {
	Snippet  models.Snippet
	Snippets []models.Snippet
	CurrentYear int
	Form any
}

// humanDate formats a time in the form "11 Feb 2026".
func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02 Jan 2006")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}

func (app *Application) newTemplateData(r *http.Request) TemplateData {
return TemplateData{
CurrentYear: time.Now().Year(),
}
}

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
		return nil, err
	}
	for _, page := range pages {
		name := filepath.Base(page)

		// Create a new template set, registering our template functions.
		ts, err := template.New("base.html").Funcs(functions).ParseFiles("./ui/html/base.html")
		if err != nil {
			return nil, err
		}

		// Add any partials.
		ts, err = ts.ParseGlob("./ui/html/partials/*.html")
		if err != nil {
			return nil, err
		}

		// Add the specific page template.
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		// Store in the cache under the page name.
		cache[name] = ts
	}
	return cache, nil
}
