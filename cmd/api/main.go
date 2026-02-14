package main

import (
	"database/sql"
	"flag"

	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/alexedwards/scs/v2"

	"github.com/ShyamGuna77/rest-sms/internal/models"
	"github.com/ShyamGuna77/rest-sms/internal/web"
	"github.com/go-playground/form/v4"
	_ "github.com/go-sql-driver/mysql"
)

func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "web:pass@/snippetbox?parseTime=true", "MySQL data source name")
	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	db, err := openDB(*dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()
	templateCache, err := web.NewTemplateCache()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("database connection pool established")
	formDecoder := form.NewDecoder()

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	app := &web.Application{
		Logger:        logger,
		Snippets:      &models.SnippetModel{DB: db},
		TemplateCache: templateCache,
		FormDecoder:   formDecoder,
		SessionManager: sessionManager,
	}

	srv := &http.Server{
		Addr:     *addr,
		Handler:  app.Routes(),
		ErrorLog: slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	logger.Info("server started on :", "addr", srv.Addr)

	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
