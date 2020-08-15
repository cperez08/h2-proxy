package proxy

import (
	"errors"
	"net/http"
)

type CustomResponseWriter struct {
	Headers http.Header
	Status  int
	Body    []byte
}

func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.Body = b
	return 1, nil
}

func (w *CustomResponseWriter) Header() http.Header {
	return w.Headers
}

func (w *CustomResponseWriter) WriteHeader(s int) {
	w.Status = s
}

func NewCustomeRsWriter() http.ResponseWriter {
	return &CustomResponseWriter{Headers: make(http.Header)}
}

type FakeReader struct{}

func (r *FakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error from fake reader")
}

func (r *FakeReader) Close() error {
	return nil
}
