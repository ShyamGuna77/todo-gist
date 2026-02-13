package web

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *Application) Routes() http.Handler {
	mux := http.NewServeMux()

	fileserver := http.FileServer(http.Dir("./ui/static/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileserver))

	dynamic := alice.New(app.SessionManager.LoadAndSave)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.Home))
	mux.Handle("GET /snippet/view/{id}", dynamic.ThenFunc(app.SnippetView))
	mux.Handle("GET /snippet/create", dynamic.ThenFunc(app.SnippetCreate))
	mux.Handle("POST /snippet/create", dynamic.ThenFunc(app.SnippetCreatePost))


	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)

}
