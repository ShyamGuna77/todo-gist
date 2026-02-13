package web

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	fileserver := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))

	mux.HandleFunc("GET /{$}", app.Home)
	mux.HandleFunc("GET /snippet/view/{id}", app.SnippetView)
	mux.HandleFunc("GET /snippet/create", app.SnippetCreate)
	mux.HandleFunc("POST /snippet/create", app.SnippetCreatePost)


	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)

}
