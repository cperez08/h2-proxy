package proxy

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/http2"
)

const (
	contentType = "Content-Type"
	grpcMessage = "grpc-message"
	grpcStatus  = "grpc-status"
)

// HandleError ...
func HandleError(w http.ResponseWriter, r *http.Request, errMsg string, printLogs bool) {
	if printLogs {
		log.Println(fmt.Sprintf(`{"rq_id": "%s", "rq_path": "%s", "rq_proto": "%s", "message": %s}`,
			r.Header.Get("X-Request-Id"),
			r.URL.Path,
			r.Proto,
			errMsg,
		))
	}

	ct := r.Header.Get(contentType)
	if strings.Contains(ct, "grpc") {
		HandleGRPCError(w, ct, errMsg)
		return
	}

	HandleHTTPError(w, errMsg)
}

// HandleGRPCError ...
func HandleGRPCError(w http.ResponseWriter, ct, errMsg string) {
	// Add headers empty body and trailers
	w.Header().Set(contentType, ct)
	w.Header().Set(grpcMessage, errMsg)
	w.Header().Set(grpcStatus, fmt.Sprintf("%d", http2.ErrCodeInternal))
	w.Write([]byte(``))
	w.Header().Add(http.TrailerPrefix+grpcStatus, fmt.Sprintf("%d", http2.ErrCodeInternal))
	w.Header().Add(http.TrailerPrefix+grpcMessage, errMsg)
}

// HandleHTTPError ...
func HandleHTTPError(w http.ResponseWriter, errMsg string) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Add(contentType, "application/json")
	w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, errMsg)))
}
