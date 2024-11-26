package main

import (
	"net/http"
)

type writer struct {
	w      http.ResponseWriter
	Status int
}

func (w *writer) Header() http.Header {
	return w.w.Header()
}

func (w *writer) Write(b []byte) (int, error) {
	return w.w.Write(b)
}

func (w *writer) WriteHeader(statusCode int) {
	w.Status = statusCode
	w.w.WriteHeader(statusCode)
}
