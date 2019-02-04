package main

import (
	"io"
	"net/http"
	log "github.com/inconshreveable/log15"
)

// A buffered channel that we can send work requests on.
var WorkQueue = make(chan WorkRequest, 100)

type HttpConnection struct {
	Request  *http.Request
	Response *http.Response
}

type HttpConnectionChannel chan *HttpConnection

var connChannel = make(HttpConnectionChannel)

func Collector(w http.ResponseWriter, r *http.Request) {
	work := WorkRequest{W: w, R: r}

	// serve only JSON RPC request
	if r.Method != "POST" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Push the work onto the queue.
	WorkQueue <- work
	log.Debug("Work request queued")

	for {
		select {
		case conn := <-connChannel:
			if conn.Response != nil {
				for k, v := range conn.Response.Header {
					w.Header().Set(k, v[0])
				}
				w.WriteHeader(conn.Response.StatusCode)
				io.Copy(w, conn.Response.Body)
			} else {
				w.WriteHeader(http.StatusOK)
			}
			return
		}
	}
}
