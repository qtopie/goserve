package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gohugoio/hugo/livereload"
)

var (
	dir        string
	port       int
	cors       string
	basic      string
	watch      bool
	auth       bool
	exclusions string
)

func main() {
	flag.IntVar(&port, "port", 7070, "specify the port to listen on")
	flag.StringVar(&cors, "cors", "", "enable CORS support with specified origin")
	flag.StringVar(&basic, "basic", "", "use basic auth, format username:password, ignored if invalid")
	flag.StringVar(&dir, "dir", "./", "root dir to serve")
	flag.BoolVar(&watch, "watch", false, "watch file changes")
	flag.StringVar(&exclusions, "excludes", "", "file or folders to exclude")
	flag.Parse()

	if basic != "" {
		basicAuthRegex, _ := regexp.Compile("([a-z]+):([a-z]+)")
		if basicAuthRegex.MatchString(basic) {
			auth = true
		} else {
			log.Println("Ignored invalid basic auth value", basic)
		}
	}

	// Register handlers
	fsHandler := http.FileServer(http.Dir(dir))
	http.HandleFunc("/", buildRootHandler(fsHandler.ServeHTTP))

	// livereload
	if watch {
		http.HandleFunc("/livereload.js", livereload.ServeJS)
		http.HandleFunc("/livereload", livereload.Handler)

		livereload.Initialize()
		go watchChanges(dir)
	}

	// Start listening
	addr := ":" + strconv.Itoa(port)
	log.Println("Serving at 0.0.0.0", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// buildRootHandler wraps the core handler with additional features
func buildRootHandler(coreHandler http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if auth {
			username, password, ok := r.BasicAuth()
			basicAuth := fmt.Sprintf("%s:%s", username, password)

			if !ok || basicAuth != basic {
				w.Header().Set("WWW-Authenticate", `Basic realm="realm"`)
				http.Error(w, "Unauthorized request.", http.StatusUnauthorized)
				return
			}
		}

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
