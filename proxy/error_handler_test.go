package proxy

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleErrorGrpc(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", ioutil.NopCloser(bytes.NewReader([]byte(``))))
	req.Header.Set(contentType, "application/grpc")
	var writer = NewCustomeRsWriter()
	HandleError(writer, req, "error", true)

	assert.Equal(t, writer.Header().Get(grpcMessage), "error")
	assert.Equal(t, writer.Header().Get(grpcStatus), "2")
}

func TestHandleErrorHttp(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", ioutil.NopCloser(bytes.NewReader([]byte(``))))
	req.Header.Set(contentType, "application/json")
	var writer = NewCustomeRsWriter()
	HandleError(writer, req, "error", false)

	assert.Equal(t, writer.Header().Get(contentType), "application/json")
}
