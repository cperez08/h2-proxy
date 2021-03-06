package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/http2"

	"github.com/cperez08/h2-proxy/config"
	"github.com/cperez08/h2-proxy/pool"
	"github.com/cperez08/h2-proxy/proxy"
)

const configDefaultLocation = "/etc/h2-proxy/config.yaml"

var sigs = make(chan os.Signal, 1)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	cfg, err := proxy.NewProxyFromFile(getFileLocation())
	if err != nil {
		log.Fatal("error loading yaml config", err)
	}

	server, cli := initClientAndServer(ctx, cfg)

	l, err := net.Listen("tcp", cfg.ProxyAddres)
	if err != nil {
		log.Fatalln(err)
	}

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		log.Println(sig)
		log.Println("shutting down proxy")
		cancel()
		if err := l.Close(); err != nil {
			log.Println("error closing listener")
			os.Exit(1)
		}
		os.Exit(0)
	}()

	log.Println("starting proxy on ", cfg.ProxyAddres)

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("error accepting new connection ", err)
		}

		log.Println("accepted new connection from", conn.RemoteAddr().String())
		server.ServeConn(conn, &http2.ServeConnOpts{
			Handler:    proxy.Handler(cfg, cli),
			BaseConfig: &http.Server{},
		})
	}
}

func getFileLocation() string {
	location := os.Getenv("H2_PROXY_CFG_LOCATION")
	if location == "" {
		return configDefaultLocation
	}

	return location
}

func initClientAndServer(ctx context.Context, cfg *config.ProxyConfig) (*http2.Server, *http.Client) {
	t := &http2.Transport{
		DisableCompression: true,
		AllowHTTP:          true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}

	pool, err := pool.NewConnectionPool(ctx, cfg, t)
	if err != nil {
		log.Fatalln(err)
	}

	t.ConnPool = pool
	cli := &http.Client{Transport: t}

	server := &http2.Server{
		IdleTimeout: time.Second * time.Duration(cfg.IdleTimeout),
	}

	return server, cli
}
