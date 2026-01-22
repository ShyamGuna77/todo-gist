package main

import (
	"fmt"
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from Server "))
}

func snippetView(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a snippet viewer box"))
}

func snippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a snippet creater  box"))
}
func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", home)
	mux.HandleFunc("/view", snippetView)
	mux.HandleFunc("/create", snippetCreate)
	port := ":3000"

	fmt.Println("server started on :", port)

	err := http.ListenAndServe(port, mux)
	if err != nil {
		log.Fatal("Error occured on :", err)

	}

}
