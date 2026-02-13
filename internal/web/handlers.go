package web

import (
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/ShyamGuna77/rest-sms/internal/models"
	"github.com/ShyamGuna77/rest-sms/internal/validator"
	"github.com/go-playground/form/v4" 
)

type Application struct {
	Logger        *slog.Logger
	Snippets      *models.SnippetModel
	TemplateCache map[string]*template.Template
	FormDecoder   *form.Decoder
}

type SnippetCreateForm struct {
	Title   string `form:"title"`
	Content string `form:"content"`
	Expires int `form:"expires"`
	validator.Validator `form:"-"`
}

func (app *Application) Home(w http.ResponseWriter, r *http.Request) {

	snippets, err := app.Snippets.Latest()
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
	data := app.newTemplateData(r)
	data.Snippets = snippets

	app.render(w, r, http.StatusOK, "home.html", data)
}
func (app *Application) SnippetView(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		app.NotFound(w, r)
		return
	}
	snippet, err := app.Snippets.Get(id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.NotFound(w, r)
		} else {
			app.ServerError(w, r, err)
		}
		return
	}
	data := app.newTemplateData(r)
	data.Snippet = snippet
	app.render(w, r, http.StatusOK, "view.html", data)
}

func (app *Application) SnippetCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = SnippetCreateForm{
		Expires: 365,
	}
	app.render(w, r, http.StatusOK, "create.html", data)
}

func (app *Application) SnippetCreatePost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.ClientError(w, http.StatusBadRequest)
		return
	}
	
	 var form SnippetCreateForm
    // Call the Decode() method of the form decoder, passing in the current
    // request and *a pointer* to our snippetCreateForm struct. This will
    // essentially fill our struct with the relevant values from the HTML form.
    // If there is a problem, we return a 400 Bad Request response to the client.
    err = app.decodeError(r, &form)
    if err != nil {
        app.ClientError(w, http.StatusBadRequest)
        return
    }

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be blank")
	form.CheckField(validator.MaxChars(form.Title, 100), "title", "This field cannot be more than 100 characters")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be blank")
	form.CheckField(validator.PermittedValue(form.Expires, 1, 7, 365), "expires", "This field must be 1, 7 or 365")
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		app.render(w, r, http.StatusUnprocessableEntity, "create.html", data)
		return
	}
	id, err := app.Snippets.Insert(form.Title, form.Content, form.Expires)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}
