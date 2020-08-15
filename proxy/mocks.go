package proxy

import (
	"errors"
	"net/http"
)

// CustomResponseWriter implements http.ResponseWriter
type CustomResponseWriter struct {
	Headers http.Header
	Status  int
	Body    []byte
}

// Write ...
func (w *CustomResponseWriter) Write(b []byte) (int, error) {
	w.Body = b
	return 1, nil
}

// Header ...
func (w *CustomResponseWriter) Header() http.Header {
	return w.Headers
}

// WriteHeader ...
func (w *CustomResponseWriter) WriteHeader(s int) {
	w.Status = s
}

// NewCustomeRsWriter ...
func NewCustomeRsWriter() http.ResponseWriter {
	return &CustomResponseWriter{Headers: make(http.Header)}
}

// FakeReader feake reader implements the io.Reader and io.ReadCloser
type FakeReader struct{}

// Read ...
func (r *FakeReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("error from fake reader")
}

// Close ...
func (r *FakeReader) Close() error {
	return nil
}
