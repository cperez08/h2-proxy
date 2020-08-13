package pool

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/cperez08/h2-proxy/config"
	"golang.org/x/net/http2"
)

const defaultAddr = "127.0.0.1:8080"

func TestNewConnectionPoolByIP(t *testing.T) {
	ctx := context.Background()
	cfg := getProxyConfig()
	tr := getTransport()
	l := fakeListener("8080")
	defer l.Close()

	cfg.TargetHost = "127.0.0.1"

	cp, err := NewConnectionPool(ctx, cfg, tr)
	if cp == nil || err != nil {
		t.Log("error creating the connection")
		t.Fail()
	}

	conn, err := cp.GetClientConn(&http.Request{}, defaultAddr)
	if conn == nil || err != nil {
		t.Log("error grabbing the connection")
		t.Fail()
	}

	cfg.TargetPort = "8090"
	cp, err = NewConnectionPool(ctx, cfg, tr)
	if cp != nil || err == nil {
		t.Log("should return error initializing the pool")
		t.Fail()
	}
}

func TestNewConnectionPoolByHost(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := getProxyConfig()
	tr := getTransport()
	l := fakeListener("8080")

	defer l.Close()
	defer cancel()

	cp, err := NewConnectionPool(ctx, cfg, tr)
	if cp == nil || err != nil {
		t.Log("error creating the connection", err)
		t.FailNow()
	}

	conn, err := cp.GetClientConn(&http.Request{}, defaultAddr)
	if conn == nil || err != nil {
		t.Log("error grabbing connection")
		t.Fail()
	}

	cfg.TargetPort = "8090"
	cp, err = NewConnectionPool(ctx, cfg, tr)
	if cp != nil || err == nil {
		t.Log("should return error initializing the pool")
		t.Fail()
	}
}

func TestNewConnectionPoolByHost2(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := getProxyConfig()
	tr := getTransport()
	l := fakeListener("8080")

	defer l.Close()
	defer cancel()

	cp, err := NewConnectionPool(ctx, cfg, tr)
	if cp == nil || err != nil {
		t.Log("error creating the connection", err)
		t.FailNow()
	}

	conn, err := cp.GetClientConn(&http.Request{}, defaultAddr)
	if conn == nil || err != nil {
		t.Log("error grabbing connection")
		t.Fail()
	}
	cp.MarkDead(conn)

	// Optional funnel for hosts with IPv4 and IPv6
	conn, err = cp.GetClientConn(&http.Request{}, "[::1]:8080")
	if conn != nil || err == nil {
		cp.MarkDead(conn)
		if _, err2 := cp.GetClientConn(&http.Request{}, "[::1]:8080"); err2 == nil {
			t.Log("should fail since there are no connections")
			t.Fail()
		}
	}

	casted := cp.(*connectionPool)
	casted.refreshConnections([]string{defaultAddr})
	casted.r.C <- true
	time.Sleep(time.Millisecond * 250) // let the refresh to finish before the cancel and listerner close are called in the defer
}

func TestGetNoConnection(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := getProxyConfig()
	tr := getTransport()
	l := fakeListener("8080")

	defer l.Close()
	defer cancel()

	cp, err := NewConnectionPool(ctx, cfg, tr)
	if cp == nil || err != nil {
		t.Log("error creating the connection", err)
		t.FailNow()
	}

	casted := cp.(*connectionPool)
	for _, c := range casted.connections {
		c.IsActive = false
	}

	_, err = casted.GetClientConn(&http.Request{}, defaultAddr)
	if err == nil {
		t.Log("expecting error due to no active connection")
		t.Fail()
	}
}

func TestKillConnectionError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := getProxyConfig()
	tr := getTransport()
	l := fakeListener("8080")

	defer l.Close()
	defer cancel()

	cp, err := NewConnectionPool(ctx, cfg, tr)
	if cp == nil || err != nil {
		t.Log("error creating the connection", err)
		t.FailNow()
	}

	casted := cp.(*connectionPool)
	for _, c := range casted.connections {
		c.Conn.Close()
	}

	// should log error trying to close closed connection
	casted.MarkDead(casted.connections[0].Conn)
}

func getTransport() *http2.Transport {
	return &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}
}

func getProxyConfig() *config.ProxyConfig {
	cfg := &config.ProxyConfig{TargetHost: "localhost", TargetPort: "8080"}
	cfg.SetDefaults()
	return cfg
}

func fakeListener(port string) net.Listener {
	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatalln(err)
	}

	return l
}
