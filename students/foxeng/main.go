package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime/debug"
	"strings"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/panic/", panicDemo)
	mux.HandleFunc("/panic-after/", panicAfterDemo)
	mux.HandleFunc("/", hello)
	log.Fatal(http.ListenAndServe(":3000", recoverHandler(mux)))
}

func recoverHandler(m *http.ServeMux) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				stack := string(debug.Stack())
				log.Println(r)
				log.Print(stack)
				fmt.Println(os.Getenv("ENV"))
				w.WriteHeader(http.StatusInternalServerError)
				if strings.Contains(strings.ToLower(os.Getenv("ENV")), "dev") {
					fmt.Fprintln(w, r)
					fmt.Fprint(w, stack)
				} else {
					fmt.Fprintln(w, "Something went wrong")
				}
			}
		}()

		// NOTE: To ensure no partial writes to w in case the handler panics, we wrap
		// w in a buffered version of ResponseWriter and only write if no panic occurs
		wb := &wbuf{ResponseWriter: w}
		m.ServeHTTP(wb, r)
		if wb.statusCode != 0 {
			w.WriteHeader(wb.statusCode)
		}
		wb.buf.WriteTo(w)
	}
}

type wbuf struct {
	http.ResponseWriter
	buf        bytes.Buffer
	statusCode int
}

func (wb *wbuf) Write(b []byte) (int, error) {
	return wb.buf.Write(b)
}

func (wb *wbuf) WriteHeader(statusCode int) {
	wb.statusCode = statusCode
}

func panicDemo(w http.ResponseWriter, r *http.Request) {
	funcThatPanics()
}

func panicAfterDemo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Hello!</h1>")
	funcThatPanics()
}

func funcThatPanics() {
	panic("Oh no!")
}

func hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Hello!</h1>")
}
