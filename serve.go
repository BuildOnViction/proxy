package main

import (
	"crypto/tls"
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

func ServeHTTP(wr http.ResponseWriter, r *http.Request, c HttpConnectionChannel) {
	var resp *http.Response
	var req *http.Request
	var err error
	start := time.Now()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 60,
	}
	defer r.Body.Close()

	req, err = http.NewRequest(r.Method, r.URL.String(), r.Body)
	req.Header.Set("Connection", "close")

	if err != nil {
		c <- &HttpConnection{nil, nil, err, nil}
		return
	}
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, _ = client.Do(req)
	elapsed := time.Since(start)

	c <- &HttpConnection{r, resp, nil, &elapsed}
}
