package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Backend struct {
	Masternode []string
	Fullnode   []string
}

type Pointer struct {
	Masternode int
	Fullnode   int
}

type JsonRpc struct {
	Method string
}

var (
	backend = Backend{[]string{"testnet.tomochain.com"}, []string{"testnet.tomochain.com"}}
	pointer = Pointer{0, 0}
)

func point(p int, max int) int {
	if p == max {
		return 0
	}
	return p + 1
}

func route(r *http.Request) (string, error) {
	decoder := json.NewDecoder(r.Body)
	var b JsonRpc
	err := decoder.Decode(&b)
	if err != nil {
		return "", err
	}
	if b.Method == "eth_sendRawTransaction" {
		max := len(backend.Masternode) - 1
		i := point(pointer.Masternode, max)
		return backend.Masternode[i], err
	}
	max := len(backend.Fullnode) - 1
	i := point(pointer.Fullnode, max)
	return backend.Fullnode[i], err
}

func ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	r.URL.Host, _ = route(r)
	r.URL.Scheme = "https"

	req, err = http.NewRequest(r.Method, r.URL.String(), r.Body)
	for name, value := range r.Header {
		req.Header.Set(name, value[0])
	}
	resp, err = client.Do(req)
	r.Body.Close()

	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		fmt.Println(err)
	}

	connChannel <- &HttpConnection{r, resp}
}
