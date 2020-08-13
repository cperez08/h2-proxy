package proxy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/cperez08/h2-proxy/config"
)

const (
	// version 1.0 just support http, future versions will include https as well
	defaultScheme       = "http"
	forwardedForHeder   = "X-Forwarded-For"
	proxiedByForHeder   = "X-Proxied-By"
	forwardedHostHeader = "X-Forwarded-Host"
)

func Handler(config *config.ProxyConfig, cli *http.Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		url := r.URL
		url.Host = config.TargetHost + ":" + config.TargetPort
		url.Scheme = defaultScheme

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			HandleError(&w, r, fmt.Sprintf("[%s] error reading request", config.ProxyName))
			return
		}

		proxyReq, err := http.NewRequest(r.Method, url.String(), ioutil.NopCloser(bytes.NewReader(reqBody)))
		if err != nil {
			HandleError(&w, r, fmt.Sprintf("[%s] error parsing request", config.ProxyName))
			return
		}

		proxyReq.Header = r.Header.Clone()
		proxyReq.Header.Set(forwardedHostHeader, r.Host)
		proxyReq.Header.Set(forwardedForHeder, r.RemoteAddr)
		proxyReq.Header.Set(proxiedByForHeder, config.ProxyName)

		proxyReq.Trailer = r.Trailer.Clone()
		rs, er := cli.Do(proxyReq)
		if er != nil {
			HandleError(&w, r, fmt.Sprintf("[%s] error performing request to target: "+err.Error(), config.ProxyName))
			return
		}

		for k, vals := range rs.Header {
			for _, val := range vals {
				w.Header().Add(k, val)
			}
		}

		w.Header().Add(proxiedByForHeder, config.ProxyName)
		rsBody, err := ioutil.ReadAll(rs.Body)
		if err != nil {
			HandleError(&w, r, fmt.Sprintf("[%s] error reading target response", config.ProxyName))
			return
		}

		w.WriteHeader(rs.StatusCode)
		defer rs.Body.Close()
		w.Write(rsBody)

		for t, vals := range rs.Trailer {
			for _, val := range vals {
				w.Header().Add(http.TrailerPrefix+t, val)
			}
		}

		if config.PrintLogs {
			PrintLog(start, len(reqBody), len(rsBody), r, config.CompactLogs)
		}
	})
}

func PrintLog(t time.Time, reqSize int, resSize int, r *http.Request, compact bool) {
	reqID := r.Header.Get("X-Request-Id")
	logStr := ""

	if compact {
		logStr = `id: %s | p: %s | pr: %s | ms: %d | rq_ln: %d | rs_ln: %d`
	} else {
		logStr = `{"rq_id": "%s", "rq_path": "%s", "rq_proto": "%s", "elapsed_time_ms": %d, "rq_lenght": %d, "rs_length": %d}`
	}

	logStr = fmt.Sprintf(logStr,
		reqID,
		r.URL.Path,
		r.Proto,
		time.Since(t).Milliseconds(),
		reqSize,
		resSize,
	)

	log.Println(logStr)
}
