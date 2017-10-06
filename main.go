package main

import (
	"strconv"
	"net/http"
	"flag"
	"log"
)

var (
	port int
)

// func configureCORS(handler http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Add("Access-Control-Allow-Origin", "*")
// 		handler.ServeHTTP(w, r)
// 	})

// }

// func buildHandler() http.Handler {
// 	var handler http.Handler
// 	handler = http.DefaultServeMux
// 	handler = configureCORS(handler)
// 	return handler
// }



func main() {
	flag.IntVar(&port, "p", 7070, "specify the port to listen on")
	flag.Parse()

	fsHandler := http.FileServer(http.Dir("."))
	http.Handle("/", fsHandler)

	addr := ":" + strconv.Itoa(port)
	log.Println("Serving port", addr)

	err := http.ListenAndServe(addr, nil)
	if err != nil{
		log.Println("Error", err)
	}
}
