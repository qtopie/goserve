package main

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func Test_injectHtml(t *testing.T) {
	var (
		integralHtml = `<!DOCTYPE html>
			<html>
				<body>
					Hello there!
				</body>
			</html>`
		integralHtmlWant = `<script data-no-instant="">document.write('<script src="/livereload.js?port=7070&mindelay=10"></' + 'script>')</script>`
		fragmentHtml     = `<div>Hello World!</div>`
		fragmentHtmlWant = `<script data-no-instant="">document.write('<script src="/livereload.js?port=8080&mindelay=7"></' + 'script>')</script>`
	)
	type args struct {
		r     io.Reader
		port  int
		delay int
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{"integral_html", args{strings.NewReader(integralHtml), 7070, 10}, integralHtmlWant, false},
		{"no_body", args{strings.NewReader(fragmentHtml), 8080, 7}, fragmentHtmlWant, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if err := injectHtml(tt.args.r, w, tt.args.port, tt.args.delay); (err != nil) != tt.wantErr {
				t.Errorf("injectHtml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// should find a way to compare compact html
			if gotW := w.String(); !strings.Contains(gotW, tt.wantW) {
				t.Errorf("injectHtml() = %v, want contains %v", gotW, tt.wantW)
			}
		})
	}
}
