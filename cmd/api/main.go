package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/ShyamGuna77/rest-sms/internal/web"
)

func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")

	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &web.Application{
		Logger: logger,
	}
	

	logger.Info("server started on :", "addr", *addr)

	router := app.Routes()

	err := http.ListenAndServe(*addr, router)
	log.Fatal(err)

}
