package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/html"
)

const (
	jsTemplate = `<script data-no-instant="">document.write('<script src="/livereload.js?port=%d&mindelay=%d"></' + 'script>')</script>`
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
