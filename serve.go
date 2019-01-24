package main

import (
	"fmt"
	"net/http"
)

func ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	var resp *http.Response
	var err error
	var req *http.Request
	client := &http.Client{}

	r.URL.Host = "testnet.tomochain.com"
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
