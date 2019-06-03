package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gohugoio/hugo/livereload"
)

var (
	dir        string
	port       int
	cors       string
	watch      bool
	exclusions string
)

func main() {
	flag.IntVar(&port, "port", 7070, "specify the port to listen on")
	flag.StringVar(&cors, "cors", "", "enable CORS support")
	flag.StringVar(&dir, "dir", "./", "dir")
	flag.BoolVar(&watch, "watch", true, "watch file changes")
	flag.StringVar(&exclusions, "excludes", "", "file or folders to exclude")
	flag.Parse()

	// Register handlers
	fsHandler := http.FileServer(http.Dir(dir))
	http.HandleFunc("/", buildRootHandler(fsHandler.ServeHTTP))

	// livereload
	if watch {
		http.HandleFunc("/livereload.js", livereload.ServeJS)
		http.HandleFunc("/livereload", livereload.Handler)

		livereload.Initialize()
		done := make(chan bool)
		ticker := time.NewTicker(10 * time.Second)
		go func() {
			// watching
			for {
				select {
				case <-ticker.C:
					livereload.ForceRefresh()
				case <-done:
					return
				}
			}
		}()
	}

	// Start listening
	addr := ":" + strconv.Itoa(port)
	log.Println("Serving at 0.0.0.0", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// buildRootHandler wraps the core handler with additional features
func buildRootHandler(coreHandler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if cors != "" {
			w.Header().Add("Access-Control-Allow-Origin", cors)
			w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Range")
		}

		if watch {
			upath := path.Clean(r.URL.Path)
			filename := filepath.Join(dir, upath[1:])
			filePtr := &filename

			if ok, _ := checkHtmlPage(filePtr); ok {
				file, err := os.Open(*filePtr)
				if err != nil {
					log.Println(err)
				}
				defer file.Close()

				delay := 10
				err = injectHtml(file, w, port, delay)
				if err != nil {
					w.Write([]byte(err.Error()))
				}
				return
			}
		}

		coreHandler.ServeHTTP(w, r)
	}
}
