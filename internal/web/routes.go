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


	//user authentication routes
	mux.Handle("GET /user/signup", dynamic.ThenFunc(app.userSignup))
	mux.Handle("POST /user/signup", dynamic.ThenFunc(app.userSignupPost))
	mux.Handle("GET /user/login", dynamic.ThenFunc(app.userLogin))
	mux.Handle("POST /user/login", dynamic.ThenFunc(app.userLoginPost))
	mux.Handle("POST /user/logout", dynamic.ThenFunc(app.userLogoutPost))

	standard := alice.New(app.recoverPanic, app.logRequest, commonHeaders)
	return standard.Then(mux)

}
