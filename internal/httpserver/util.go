// Package httpserver provides functionality common to all storetheindex HTTP
// servers
package httpserver

import (
	"Pando/internal/syserr"
	"errors"
	"fmt"
	"net/http"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("indexer/http")

type ErrorJsonResponse struct {
	Error string
}

func WriteJsonResponse(w http.ResponseWriter, status int, body []byte) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if _, err := w.Write(body); err != nil {
		log.Errorw("cannot write response", "err", err)
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func HandleError(w http.ResponseWriter, err error, reqType string) {
	status := http.StatusBadRequest
	var se *syserr.SysError
	if errors.As(err, &se) {
		if se.Status() >= 500 {
			log.Errorw(fmt.Sprint("cannot handle", reqType, "request"), "err", se.Error(), "status", se.Status())
			http.Error(w, "", se.Status())
			return
		}
		status = se.Status()
	}
	log.Infow(fmt.Sprint("bad", reqType, "request"), "err", err.Error(), "status", status)
	http.Error(w, err.Error(), status)
}
