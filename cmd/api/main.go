package main

import (
	"fmt"
	"log"
	"net/http"
)



func main() {
	fmt.Println("Hello, Go project!")
}




//middleware runs before handlers 

// middleware takes a http.Handler and returns a httphndler

// it also returns a http.HandlerFunc which takes resp (w) and req (r) 

// we return middleware usng next like next.ServeHttp(w,r)



func LoggingMiddleware(next http.Handler) http.Handler {

   return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	log.Println("Before server started")
	next.ServeHTTP(w,r)
	log.Println("After server started")
   })
}