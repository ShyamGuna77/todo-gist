package web

import (
	"github.com/ShyamGuna77/rest-sms/internal/models"
	"html/template"
	"path/filepath"
)

type TemplateData struct {
	Snippet  models.Snippet
	Snippets []models.Snippet
}

func NewTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, err := filepath.Glob("./ui/html/pages/*.html")
	if err != nil {
	    return nil, err
	}
	for _, page := range pages {
	    name := filepath.Base(page)
	    // Parse the base template file into a template set.
	    ts, err := template.ParseFiles("./ui/html/base.html")
	    if err != nil {
		   return nil, err
	    }
	    // Call ParseGlob() *on this template set* to add any partials.
	    ts, err = ts.ParseGlob("./ui/html/partials/*.html")
	    if err != nil {
		   return nil, err
	    }
	    // Call ParseFiles() *on this template set* to add the  page template.
	    ts, err = ts.ParseFiles(page)
	    if err != nil {
		   return nil, err
	    }
	    // Add the template set to the map as normal...
	    cache[name] = ts
	}
	return cache, nil
}
