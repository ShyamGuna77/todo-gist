package web

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/ShyamGuna77/rest-sms/internal/models"
)

type Application struct {
	Logger *slog.Logger
	Snippets *models.SnippetModel
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")

	files := []string{
		filepath.Join("ui", "html", "base.html"),
		filepath.Join("ui", "html", "partials", "nav.html"),
		filepath.Join("ui", "html", "pages", "home.html"),
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		app.ServerError(w, r, err)
	}

}

func (app *Application) SnippetView(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func (app *Application) SnippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a snippet creater  box"))
}

func (app *Application) SnippetCreatePost(w http.ResponseWriter, r *http.Request) {
	// w.WriteHeader(http.StatusCreated)
	// w.Write([]byte("THis is a Post request"))
	title := "O snail"
	content := "O snail\nClimb Mount Fuji,\nBut slowly, slowly!\n\n– Kobayashi Issa"
	expires := 7
	// Pass the data to the SnippetModel.Insert() method, receiving the
	// ID of the new record back.
	id, err := app.Snippets.Insert(title, content, expires)
	if err != nil {
	    app.ServerError(w, r, err)
	    return
	}
	// Redirect the user to the relevant page for the snippet.
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
