package auth

import (
	"fmt"
	"net/http"
)

type writer struct {
	response http.ResponseWriter
}

func (w *writer) code(code int) {
	w.response.WriteHeader(code)
}

func (w *writer) line(mesg string, args ...any) {
	w.text(mesg, args...)
	w.text("\n")
}

func (w *writer) end() {
	w.line("</html>")
	if flusher, ok := w.response.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *writer) start() {
	w.line("<html>")
}

func (w *writer) text(mesg string, args ...any) {
	if len(args) > 0 {
		mesg = fmt.Sprintf(mesg, args...)
	}
	w.response.Write([]byte(mesg))
}
