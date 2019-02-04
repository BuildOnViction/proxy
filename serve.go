package main

import (
	"encoding/json"
	"net/http"
	"net/url"
    log "github.com/inconshreveable/log15"
)

type Backend struct {
	Masternode []*url.URL
	Fullnode   []*url.URL
}

type Pointer struct {
	Masternode int
	Fullnode   int
}

type JsonRpc struct {
	Method string
}

var (
	backend Backend
	pointer = Pointer{0, 0}
)

func point(p int, max int) int {
	if p == max {
		return 0
	}
	return p + 1
}

func route(r *http.Request) (*url.URL, error) {
	decoder := json.NewDecoder(r.Body)
	var b JsonRpc
    var url *url.URL
    var index int
	err := decoder.Decode(&b)
	if err != nil {
		return nil, err
	}
	if b.Method == "eth_sendRawTransaction" {
		max := len(backend.Masternode) - 1
		index = point(pointer.Masternode, max)
		url = backend.Masternode[index]
        log.Info("RPC masternode request", "method", b.Method, "index", index, "host", url.Host)
	} else {
        max := len(backend.Fullnode) - 1
        index = point(pointer.Fullnode, max)
	    url = backend.Fullnode[index]
        log.Info("RPC fullnode request", "method", b.Method, "index", index, "host", url.Host)
    }
    return url, err
}

func ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	url, _ := route(r)
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme

	req, err = http.NewRequest(r.Method, r.URL.String(), r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, err = client.Do(req)
	r.Body.Close()

	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		log.Error("Backend error", "err", err)
	}

	connChannel <- &HttpConnection{r, resp}
}
