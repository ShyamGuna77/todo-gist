package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ShyamGuna77/rest-sms/cmd/web"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", web.Home)
	mux.HandleFunc("GET /snippet/view/{id}", web.SnippetView)
	mux.HandleFunc("GET /snippet/create", web.SnippetCreate)
	mux.HandleFunc("POST /snippet/create", web.SnippetCreatePost)
	port := ":3000"

	fmt.Println("server started on :", port)

	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("Error occured on :", err)

	}

}
