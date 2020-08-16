package proxy

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

var (
	cfg, _ = NewProxyFromFile("../config/config.yaml")
)

func TestProxyHandler(t *testing.T) {
	FakeProxyListerner, err := net.Listen("tcp", "0.0.0.0:7070")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	FakeServerListerner, err := net.Listen("tcp", "0.0.0.0:7090")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	defer FakeProxyListerner.Close()
	defer FakeServerListerner.Close()

	go SetUpFakeServerProxy(FakeProxyListerner)
	go SetUpFakeServer(FakeServerListerner)

	tr := &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}

	cli := &http.Client{Transport: tr}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:7070/ok", ioutil.NopCloser(bytes.NewReader([]byte(`{"req": "1"}`))))

	// make request to proxy which redirets the call to our fake handler
	response, err := cli.Do(req)
	if err != nil {
		t.Log("error performing requests", err)
		t.Fail()
	}

	if response.StatusCode != http.StatusOK {
		t.Log("Unexpected status code", response.StatusCode)
		t.Fail()
	}

	if response.Header.Get("server") != "fake-server" {
		t.Log("expected header not received")
		t.Fail()
	}

	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Log("error reading response")
		t.Fail()
	}

	defer response.Body.Close()
	if response.Trailer.Get("trailer1") != "fake-trailer" {
		t.Log("expected trailer not received")
		t.Fail()
	}
}

func TestProxyHandlerWithResponseError(t *testing.T) {
	FakeProxyListerner, err := net.Listen("tcp", "0.0.0.0:7070")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	FakeServerListerner, err := net.Listen("tcp", "0.0.0.0:7090")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	defer FakeProxyListerner.Close()
	defer FakeServerListerner.Close()

	go SetUpFakeServerProxy(FakeProxyListerner)
	go SetUpFakeServer(FakeServerListerner)

	// increase coverage by seeting compact logs true
	cfg.CompactLogs = true

	tr := &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}

	cli := &http.Client{Transport: tr}
	req, _ := http.NewRequest(http.MethodGet, "http://localhost:7070/fail", ioutil.NopCloser(bytes.NewReader([]byte(`{"req": "1"}`))))

	response, err := cli.Do(req)
	if err != nil {
		t.Log("error performing requests", err)
		t.Fail()
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Log("Unexpected status code", response.StatusCode, response.Status)
		t.Fail()
	}

	if response.Header.Get("server") != "fake-server" {
		t.Log("expected header not received")
		t.Fail()
	}

	_, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Log("error reading response")
		t.Fail()
	}

	defer response.Body.Close()
	if response.Trailer.Get("trailer1") != "fake-trailer" {
		t.Log("expected trailer not received")
		t.Fail()
	}
}

func TestForceRequestError(t *testing.T) {
	FakeProxyListerner, err := net.Listen("tcp", "0.0.0.0:7070")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	FakeServerListerner, err := net.Listen("tcp", "0.0.0.0:7090")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	defer FakeProxyListerner.Close()
	defer FakeServerListerner.Close()

	go SetUpFakeServerProxy(FakeProxyListerner)
	go SetUpTimeoutServer(FakeServerListerner)

	tr := &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}

	cli := &http.Client{Transport: tr}

	req, _ := http.NewRequest(http.MethodGet, "http://localhost:7070/ok", ioutil.NopCloser(bytes.NewReader([]byte(`{"req": "1"}`))))
	response, err := cli.Do(req)
	if err != nil {
		t.Log("error performing requests", err)
		t.Fail()
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Log("expecting connection errors")
		t.Fail()
	}
}

func TestCreateRequest(t *testing.T) {
	var fr io.Reader = &FakeReader{}
	req, err := http.NewRequest("GET", "localhost:9090", fr)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}

	if _, _, err := createRequest(req, cfg); err == nil {
		t.Log("expected error creating request")
		t.Fail()
	}

	req = &http.Request{Method: "¡™¢", Body: ioutil.NopCloser(bytes.NewBuffer([]byte(``))), Header: make(http.Header), URL: &url.URL{Host: "localhost"}}
	if _, _, err := createRequest(req, cfg); err == nil {
		t.Log("expected error creating new proxy request")
		t.Fail()
	}
}

func TestWriteResponse(t *testing.T) {
	wr := NewCustomeRsWriter()
	var fr io.ReadCloser = &FakeReader{}
	defer fr.Close()
	rs := &http.Response{StatusCode: 200, Header: make(http.Header), Body: fr}
	if _, err := writeResponse(wr, rs, cfg); err == nil {
		t.Log("expected error reading response")
		t.Fail()
	}
}

func SetUpFakeServerProxy(lis net.Listener) {
	t := &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}

	cli := &http.Client{Transport: t}

	server := http2.Server{}

	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalln("error accepting new connection ", err)
		}

		log.Println("accepted new connection from", conn.RemoteAddr().String())
		server.ServeConn(conn, &http2.ServeConnOpts{
			Handler:    Handler(cfg, cli),
			BaseConfig: &http.Server{},
		})
	}
}

func SetUpFakeServer(lis net.Listener) {
	server := http2.Server{}
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Fatalln("error accepting new connection ", err)
		}

		server.ServeConn(conn, &http2.ServeConnOpts{
			Handler:    FakeHandler(),
			BaseConfig: &http.Server{},
		})
	}
}

func SetUpTimeoutServer(lis net.Listener) {
	server := http2.Server{}
	count := 1
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("error accepting new connection ", err)
		}

		// force timeout error
		if count == 1 {
			t := time.Now()
			conn.SetDeadline(t.Add(time.Microsecond))
			count++
		}

		// force return in the retry
		if conn == nil {
			return
		}

		server.ServeConn(conn, &http2.ServeConnOpts{
			Handler:    FakeHandler(),
			BaseConfig: &http.Server{},
		})
	}
}

func FakeHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := map[string]string{"server": "fake-server", "version": "v1"}
		trailers := map[string]string{"trailer1": "fake-trailer", "trailer-version": "v1"}

		for k, v := range headers {
			w.Header().Add(k, v)
		}

		if strings.Contains(r.URL.String(), "/ok") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		} else {
			w.Header().Add("Status-Code", fmt.Sprintf("%d", http2.ErrCodeInternal))
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"status":"fail"}`))
		}

		for t, vals := range trailers {
			w.Header().Add(http.TrailerPrefix+t, vals)
		}
	})
}
