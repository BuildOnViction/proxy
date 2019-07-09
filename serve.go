package main

import (
	"crypto/tls"
	"github.com/tomochain/proxy/config"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	sync.Mutex
	Masternode []*url.URL
	Fullnode   []*url.URL
	Websocket  []*url.URL
}

type Pointer struct {
	Masternode int
	Fullnode   int
	Websocket  int
}

var (
	backend Backend
	pointer = Pointer{0, 0, 0}
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
	req.Close = true

	cfg := config.GetConfig()
	if cfg.Headers != nil {
		for key, value := range *cfg.Headers {
			req.Header.Set(key, value)
			if host := req.Header.Get("Host"); host != "" {
				req.Host = host
			}
		}
	}

	if err != nil {
		c <- &HttpConnection{nil, nil, err, nil}
		return
	}
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, _ = client.Do(req)
	elapsed := time.Since(start)

	c <- &HttpConnection{req, resp, nil, &elapsed}
}
