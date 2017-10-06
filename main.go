package main

import (
	"strconv"
	"net/http"
	"flag"
	"log"
)

var (
	port int
	enableCORS bool
)

// configure CORS support (wild)
func configureCORS() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Range")
		http.DefaultServeMux.ServeHTTP(w, r)
	})

}

func buildHandler() http.Handler {
	var handler http.Handler
	if enableCORS {
		handler = configureCORS()
	}
	
	return handler
}

func main() {
	flag.IntVar(&port, "p", 7070, "specify the port to listen on")
	flag.BoolVar(&enableCORS,"cors", false, "enable CORS support")
	flag.Parse()

	fsHandler := http.FileServer(http.Dir("."))
	http.Handle("/", fsHandler)

	addr := ":" + strconv.Itoa(port)
	log.Println("Serving port", addr)

	handler := buildHandler()
	log.Fatal(http.ListenAndServe(addr, handler))
}
