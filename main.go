package main

import (
	"fmt"
	"net/http"
)

func main() {
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	port := ":7070"
	fmt.Println("Serving at", port)
	http.ListenAndServe(port, nil)

}
