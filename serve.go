package main

import (
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
	client := &http.Client{}

	req, _ = http.NewRequest(r.Method, r.URL.String(), r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, _ = client.Do(req)
	defer r.Body.Close()

	connChannel <- &HttpConnection{r, resp}
}
