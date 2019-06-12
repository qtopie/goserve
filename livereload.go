package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gohugoio/hugo/livereload"

	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/html"
)

const (
	MAX_WATCH_LIMIT = 12
	jsTemplate      = `<script data-no-instant="">document.write('<script src="/livereload.js?port=%d&mindelay=%d"></' + 'script>')</script>`
)

// injectHtml injects livereload js into html file
func injectHtml(r io.Reader, w io.Writer, port, delay int) (err error) {
	doc, err := html.Parse(r)
	if err != nil {
		return
	}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			livereloadScript := fmt.Sprintf(jsTemplate, port, delay)
			nodes, _ := html.ParseFragment(strings.NewReader(livereloadScript), n)
			n.AppendChild(nodes[0])
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	return html.Render(w, doc)
}

// checkHtmlPage find whether need to inject html page
func checkHtmlPage(filename *string) (found bool, err error) {
	d, err := os.Stat(*filename)
	if err != nil {
		return false, err
	}

	if d.IsDir() {
		index := filepath.Join(*filename, "index.html")
		if _, err := os.Stat(index); err == nil {
			*filename = index
			return true, nil
		}
	} else if strings.HasSuffix(*filename, ".html") {
		return true, nil
	}

	return
}

func watchChanges(dir string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalln(err)
	}
	defer watcher.Close()

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// ignore hidden dir such as .git .vscode
		return watcher.Add(path)
	})

	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan bool)
	q := make(chan string, 10)
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ticker.C:
				if len(q) > MAX_WATCH_LIMIT {
					// refresh all
					livereload.ForceRefresh()
				}

				files := readUniqueVals(q)
				for f := range files {
					livereload.RefreshPath(f)
					log.Println("Refreshed", f)
				}
			}
		}
	}()

	go func() {

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					log.Println("Err")
					return
				}

				if event.Op&(fsnotify.Remove|fsnotify.Create|fsnotify.Write|fsnotify.Rename) > 0 {
					if len(q) < cap(q) {
						q <- event.Name
					}
				}
			case err, ok := <-watcher.Errors:
				log.Println(err)
				if !ok {
					return
				}
			}
		}
	}()

	<-done
}

func readUniqueVals(ch chan string) map[string]bool {
	// set limit
	const max = 100
	count := 0
	vals := make(map[string]bool)

	for {
		select {
		case v := <-ch:
			vals[v] = true
			count++
		default:
			return vals
		}

		if count > max {
			break
		}
	}

	return vals
}
