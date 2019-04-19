package main

import (
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	sync.Mutex
	Masternode []*url.URL
	Fullnode   []*url.URL
}

type Pointer struct {
	Masternode int
	Fullnode   int
}

var (
	backend Backend
	pointer = Pointer{0, 0}
)

func point(p int, max int) int {
	if p >= max {
		return 0
	}
	return p + 1
}

func ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var req *http.Request
	var err error
	client := &http.Client{
		Timeout: time.Second * 60,
	}
	defer r.Body.Close()

	req, err = http.NewRequest(r.Method, r.URL.String(), r.Body)
	if err != nil {
		connChannel <- &HttpConnection{nil, nil, err}
		return
	}
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, _ = client.Do(req)

	connChannel <- &HttpConnection{r, resp, nil}
}
