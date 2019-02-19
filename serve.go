package main

import (
	"bytes"
	"encoding/json"
	log "github.com/inconshreveable/log15"
	"io/ioutil"
	"net/http"
	"net/url"
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

func route(r *http.Request) (*url.URL, string, error) {
	body, _ := ioutil.ReadAll(r.Body)
	var b JsonRpc
	var url *url.URL
	err := json.Unmarshal(body, &b)
	if err != nil {
		return nil, "", err
	}
	if b.Method == "eth_sendRawTransaction" {
		max := len(backend.Masternode) - 1
		pointer.Masternode = point(pointer.Masternode, max)
		url = backend.Masternode[pointer.Masternode]
		log.Info("RPC masternode request", "method", b.Method, "index", pointer.Masternode, "host", url.Host)
	} else {
		max := len(backend.Fullnode) - 1
		pointer.Fullnode = point(pointer.Fullnode, max)
		url = backend.Fullnode[pointer.Fullnode]
		log.Info("RPC fullnode request", "method", b.Method, "index", pointer.Fullnode, "max", max, "host", url.Host)
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return url, b.Method, err
}

func ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	url, method, _ := route(r)
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme

	req, err = http.NewRequest(r.Method, r.URL.String(), r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, err = client.Do(req)
	defer r.Body.Close()

	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		log.Error("Backend error", "url", r.URL.String(), "method", method, "err", err)
	}

	connChannel <- &HttpConnection{r, resp}
}
