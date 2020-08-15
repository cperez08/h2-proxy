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

// Handler handles the proxy requests
func Handler(config *config.ProxyConfig, cli *http.Client) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		proxyReq, reqSize, err := createRequest(r, config)
		if err != nil {
			HandleError(w, r, err.Error(), config.PrintLogs)
			return
		}

		rs, err := cli.Do(proxyReq)
		if err != nil {
			HandleError(w, r, fmt.Sprintf("[%s] error performing request to target: "+err.Error(), config.ProxyName), config.PrintLogs)
			return
		}

		rsSize, err := writeResponse(w, rs, config)
		if err != nil {
			HandleError(w, r, err.Error(), config.PrintLogs)
			return
		}

		if config.PrintLogs {
			PrintLog(start, reqSize, rsSize, r, config.CompactLogs)
		}
	})
}

func createRequest(r *http.Request, config *config.ProxyConfig) (_ *http.Request, requestSize int, _ error) {
	url := r.URL
	url.Host = config.TargetHost + ":" + config.TargetPort
	url.Scheme = defaultScheme

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("[%s] error reading request", config.ProxyName)
	}

	proxyReq, err := http.NewRequest(r.Method, url.String(), ioutil.NopCloser(bytes.NewReader(reqBody)))
	if err != nil {
		return nil, 0, fmt.Errorf("[%s] error parsing request", config.ProxyName)
	}

	proxyReq.Header = r.Header.Clone()
	proxyReq.Header.Set(forwardedHostHeader, r.Host)
	proxyReq.Header.Set(forwardedForHeder, r.RemoteAddr)
	proxyReq.Header.Set(proxiedByForHeder, config.ProxyName)

	proxyReq.Trailer = r.Trailer.Clone()

	return proxyReq, len(reqBody), nil
}

func writeResponse(w http.ResponseWriter, rs *http.Response, config *config.ProxyConfig) (responseSize int, _ error) {
	for k, vals := range rs.Header {
		for _, val := range vals {
			w.Header().Add(k, val)
		}
	}

	w.Header().Add(proxiedByForHeder, config.ProxyName)
	rsBody, err := ioutil.ReadAll(rs.Body)
	if err != nil {
		return 0, fmt.Errorf("[%s] error reading target response", config.ProxyName)
	}

	w.WriteHeader(rs.StatusCode)
	defer rs.Body.Close()
	w.Write(rsBody)

	for t, vals := range rs.Trailer {
		for _, val := range vals {
			w.Header().Add(http.TrailerPrefix+t, val)
		}
	}

	return len(rsBody), nil
}

// PrintLog prints in stout basic information about the request and response
func PrintLog(t time.Time, reqSize int, resSize int, r *http.Request, compact bool) {
	logStr := ""

	if compact {
		logStr = `id: %s | p: %s | pr: %s | ms: %d | rq_ln: %d | rs_ln: %d`
	} else {
		logStr = `{"rq_id": "%s", "rq_path": "%s", "rq_proto": "%s", "elapsed_time_ms": %d, "rq_length": %d, "rs_length": %d}`
	}

	logStr = fmt.Sprintf(logStr,
		r.Header.Get("X-Request-Id"),
		r.URL.Path,
		r.Proto,
		time.Since(t).Milliseconds(),
		reqSize,
		resSize,
	)

	log.Println(logStr)
}
