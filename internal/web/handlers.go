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
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

type Application struct {
	Logger         *slog.Logger
	Snippets       *models.SnippetModel
	Users          *models.UserModel
	TemplateCache  map[string]*template.Template
	FormDecoder    *form.Decoder
	SessionManager *scs.SessionManager
}

type SnippetCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	Expires             int    `form:"expires"`
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
	err = app.DecodeForm(r, &form)
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

	app.SessionManager.Put(r.Context(), "flash", "Snippet successfully created!")
	http.Redirect(w, r, fmt.Sprintf("/snippet/view/%d", id), http.StatusSeeOther)
}


// userr authentication 
type userSignupForm struct {
    Name                string `form:"name"`
    Email               string `form:"email"`
    Password            string `form:"password"`
    validator.Validator `form:"-"`
}


// Create a new userLoginForm struct.
type userLoginForm struct {
    Email               string `form:"email"`
    Password            string `form:"password"`
    validator.Validator `form:"-"`
}


func (app *Application) userSignup(w http.ResponseWriter, r *http.Request) {
    data := app.newTemplateData(r)
    data.Form = userSignupForm{}
    app.render(w, r, http.StatusOK, "signup.html", data)	
}


func (app *Application) userSignupPost(w http.ResponseWriter, r *http.Request) {
    // Declare a zero-valued instance of our userSignupForm struct.
    var form userSignupForm
    // Parse the form data into the userSignupForm struct.
    err := app.DecodeForm(r, &form)
    if err != nil {
        app.ClientError(w, http.StatusBadRequest)
        return
    }
    // Validate the form contents using our helper functions.
    form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
    form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
    form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
    form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
    form.CheckField(validator.MinChars(form.Password, 8), "password", "This field must be at least 8 characters long")
    // If there are any errors, redisplay the signup form along with a 422
    // status code.
    if !form.Valid() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
        return
    }
     // Try to create a new user record in the database. If the email already
    // exists then add an error message to the form and redisplay it.
    err = app.Users.Insert(form.Name, form.Email, form.Password)
    if err != nil {
        if errors.Is(err, models.ErrDuplicateEmail) {
            form.AddFieldError("email", "Email address is already in use")
            data := app.newTemplateData(r)
            data.Form = form
            app.render(w, r, http.StatusUnprocessableEntity, "signup.html", data)
        } else {
            app.ServerError(w, r, err)
        }
        return
    }
    // Otherwise add a confirmation flash message to the session confirming that
    // their signup worked.
    app.SessionManager.Put(r.Context(), "flash", "Your signup was successful. Please log in.")
    // And redirect the user to the login page.
    http.Redirect(w, r, "/user/login", http.StatusSeeOther)
}





func (app *Application) userLogin(w http.ResponseWriter, r *http.Request) {
     data := app.newTemplateData(r)
    data.Form = userLoginForm{}
    app.render(w, r, http.StatusOK, "login.html", data)
}

// Login Post Handler
func (app *Application) userLoginPost(w http.ResponseWriter, r *http.Request) {
    var form userLoginForm
    err := app.DecodeForm(r, &form)
    if err != nil {
        app.ClientError(w, http.StatusBadRequest)
        return
    }
    // Do some validation checks on the form. We check that both email and
    // password are provided, and also check the format of the email address as
    // a UX-nicety (in case the user makes a typo).
    form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
    form.CheckField(validator.Matches(form.Email, validator.EmailRX), "email", "This field must be a valid email address")
    form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
    if !form.Valid() {
        data := app.newTemplateData(r)
        data.Form = form
        app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
        return
    }
    // Check whether the credentials are valid. If they're not, add a generic
    // non-field error message and redisplay the login page.
    id, err := app.Users.Authenticate(form.Email, form.Password)
    if err != nil {
        if errors.Is(err, models.ErrInvalidCredentials) {
            form.AddNonFieldError("Email or password is incorrect")
            data := app.newTemplateData(r)
            data.Form = form
            app.render(w, r, http.StatusUnprocessableEntity, "login.html", data)
        } else {
            app.ServerError(w, r, err)
        }
        return
    }
    // Use the RenewToken() method on the current session to change the session	
    // ID. It's good practice to generate a new session ID when the 
    // authentication state or privilege levels change for the user (e.g. login
    // and logout operations).
    err = app.SessionManager.RenewToken(r.Context())
    if err != nil {
        app.ServerError(w, r, err)
        return
    }
    // Add the ID of the current user to the session, so that they are now
    // 'logged in'.
    app.SessionManager.Put(r.Context(), "authenticatedUserID", id)
    // Redirect the user to the create snippet page.
    http.Redirect(w, r, "/snippet/create", http.StatusSeeOther)
}
//user logout post handler

func (app *Application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
    // Use the RenewToken() method on the current session to change the session
    // ID again.
    err := app.SessionManager.RenewToken(r.Context())
    if err != nil {
        app.ServerError(w, r, err)
        return
    }
    // Remove the authenticatedUserID from the session data so that the user is
    // 'logged out'.
    app.SessionManager.Remove(r.Context(), "authenticatedUserID")
    // Add a flash message to the session to confirm to the user that they've been
    // logged out.
    app.SessionManager.Put(r.Context(), "flash", "You've been logged out successfully!")
    // Redirect the user to the application home page.
    http.Redirect(w, r, "/", http.StatusSeeOther)
}