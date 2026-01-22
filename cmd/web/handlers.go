package web

import (
	"fmt"
	"net/http"
	"strconv"
)

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Server", "Go")
	w.Write([]byte("This is Home  snippet box"))
}

func SnippetView(w http.ResponseWriter, r *http.Request) {

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id < 1 {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "Display a specific snippet with ID %d...", id)
}

func SnippetCreate(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a snippet creater  box"))
}

func SnippetCreatePost(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("THis is a Post request"))
}
