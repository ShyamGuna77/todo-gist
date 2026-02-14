package web

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/form/v4"
)

func (app *Application) ServerError(w http.ResponseWriter, r *http.Request, err error) {

	var (
		method = r.Method
		uri    = r.URL.RequestURI()
		trace  = string(debug.Stack())
	)
	app.Logger.Error(err.Error(), "method", method, "uri", uri, "trace", trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

func (app *Application) ClientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

func (app *Application) NotFound(w http.ResponseWriter, r *http.Request) {
	app.ClientError(w, http.StatusNotFound)
}

func (app *Application) render(w http.ResponseWriter, r *http.Request, status int, page string, data TemplateData) {
	// Retrieve the appropriate template set from the cache based on the page
	// name (like 'home.tmpl'). If no entry exists in the cache with the
	// provided name, then create a new error and call the serverError() helper
	// method that we made earlier and return.
	ts, ok := app.TemplateCache[page]
	if !ok {
		err := fmt.Errorf("the template %s does not exist", page)
		app.ServerError(w, r, err)
		return
	}
	buf := new(bytes.Buffer)
	err := ts.ExecuteTemplate(buf, "base", data)
	if err != nil {
		app.ServerError(w, r, err)
		return
	}

	// Write out the provided HTTP status code ('200 OK', '400 Bad Request' etc).
	w.WriteHeader(status)
	buf.WriteTo(w)
	// Execute the template set and write the response body. Again, if there
	// is any error we call the serverError() helper.
	// err := ts.ExecuteTemplate(w, "base", data)
	// if err != nil {
	// 	app.ServerError(w, r, err)
	// }
}


// DecodeForm parses the request form and decodes it into dst. Returns an error on failure.
func (app *Application) DecodeForm(r *http.Request, dst any) error {
	if err := r.ParseForm(); err != nil {
		return err
	}
	if err := app.FormDecoder.Decode(dst, r.PostForm); err != nil {
		var invalidDecodeError *form.InvalidDecoderError
		if errors.As(err, &invalidDecodeError) {
			panic(err)
		}
		return err
	}
	return nil
}
