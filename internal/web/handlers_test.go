package web

import (
	"bytes"
	"database/sql"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/ShyamGuna77/rest-sms/internal/models"
	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
)

func newDiscardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestTemplateCache(t *testing.T, pages ...string) map[string]*template.Template {
	t.Helper()
	cache := make(map[string]*template.Template, len(pages))
	for _, name := range pages {
		ts, err := template.New("base.html").Parse(`{{define "base"}}{{template "main" .}}{{end}}`)
		if err != nil {
			t.Fatalf("parse base: %v", err)
		}
		if _, err := ts.Parse(`{{define "main"}}ok{{end}}`); err != nil {
			t.Fatalf("parse main: %v", err)
		}
		cache[name] = ts
	}
	return cache
}

func newTestApp(t *testing.T) (*Application, *sql.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New(): %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	app := &Application{
		Logger:         newDiscardLogger(),
		Snippets:       &models.SnippetModel{DB: db},
		Users:          &models.UserModel{DB: db},
		TemplateCache:  newTestTemplateCache(t, "home.html", "view.html", "create.html", "signup.html", "login.html"),
		FormDecoder:    form.NewDecoder(),
		SessionManager: scs.New(),
		RateLimitEnabled: false,
	}
	return app, db, mock
}

func TestHome_OK(t *testing.T) {
	app, _, mock := newTestApp(t)

	stmt := `SELECT id, title, content, created, expires FROM snippets
	WHERE expires > UTC_TIMESTAMP() ORDER BY id DESC LIMIT 10`

	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "title", "content", "created", "expires"}).
		AddRow(1, "t1", "c1", now, now.Add(24*time.Hour))

	mock.ExpectQuery(regexp.QuoteMeta(stmt)).WillReturnRows(rows)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	app.SessionManager.LoadAndSave(http.HandlerFunc(app.Home)).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestSnippetView_NotFoundForBadID(t *testing.T) {
	app, _, _ := newTestApp(t)

	srv := app.Routes()
	r := httptest.NewRequest(http.MethodGet, "/snippet/view/abc", nil)
	w := httptest.NewRecorder()

	srv.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSnippetCreatePost_RedirectsOnSuccess(t *testing.T) {
	app, _, mock := newTestApp(t)

	stmt := `INSERT INTO snippets (title, content, created, expires)
	VALUES(?, ?, UTC_TIMESTAMP(), DATE_ADD(UTC_TIMESTAMP(), INTERVAL ? DAY))`

	mock.ExpectExec(regexp.QuoteMeta(stmt)).
		WithArgs("My title", "My content", 7).
		WillReturnResult(sqlmock.NewResult(123, 1))

	formVals := url.Values{}
	formVals.Set("title", "My title")
	formVals.Set("content", "My content")
	formVals.Set("expires", "7")

	r := httptest.NewRequest(http.MethodPost, "/snippet/create", bytes.NewBufferString(formVals.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.RemoteAddr = "127.0.0.1:1234"

	// Load a session and mark user authenticated.
	ctx, err := app.SessionManager.Load(r.Context(), "")
	if err != nil {
		t.Fatalf("session load: %v", err)
	}
	app.SessionManager.Put(ctx, "authenticatedUserID", 1)
	r = r.WithContext(ctx)

	w := httptest.NewRecorder()
	app.SessionManager.LoadAndSave(app.Routes()).ServeHTTP(w, r)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusSeeOther)
	}
	if loc := w.Header().Get("Location"); !strings.HasPrefix(loc, "/snippet/view/") {
		t.Fatalf("Location = %q, want redirect to snippet view", loc)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}

func TestRateLimit_TooManyRequests(t *testing.T) {
	app, _, _ := newTestApp(t)
	app.RateLimitEnabled = true
	app.RateLimitRPS = 1
	app.RateLimitBurst = 1

	h := app.rateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r1 := httptest.NewRequest(http.MethodGet, "/", nil)
	r1.RemoteAddr = "127.0.0.1:1111"
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first request status = %d, want %d", w1.Code, http.StatusOK)
	}

	r2 := httptest.NewRequest(http.MethodGet, "/", nil)
	r2.RemoteAddr = "127.0.0.1:1111"
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, r2)
	if w2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want %d", w2.Code, http.StatusTooManyRequests)
	}
}

