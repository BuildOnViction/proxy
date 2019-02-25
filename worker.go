package main

import (
	"bytes"
	"encoding/json"
	"github.com/hashicorp/golang-lru"
	log "github.com/inconshreveable/log15"
	"io/ioutil"
	"net/http"
	"net/url"
)

type WorkRequest struct {
	W http.ResponseWriter
	R *http.Request
}

type HttpConnection struct {
	Request  *http.Request
	Response *http.Response
}

type JsonRpc struct {
	Method string
}

type HttpConnectionChannel chan *HttpConnection

var WorkerQueue chan chan WorkRequest

var WorkQueue = make(chan WorkRequest, 100)

var connChannel = make(HttpConnectionChannel)

var cache *lru.Cache

func route(r *http.Request) (*url.URL, string, string, error) {
	body, _ := ioutil.ReadAll(r.Body)
	cacheKey := string(body)
	var b JsonRpc
	var url *url.URL
	err := json.Unmarshal(body, &b)
	if err != nil {
		return nil, "", "", err
	}
	if b.Method == "eth_sendRawTransaction" {
		max := len(backend.Masternode) - 1
		pointer.Masternode = point(pointer.Masternode, max)
		url = backend.Masternode[pointer.Masternode]
		log.Info("RPC masternode request", "method", b.Method, "index", pointer.Masternode, "host", url.Host)
		cacheKey = ""
	} else {
		max := len(backend.Fullnode) - 1
		pointer.Fullnode = point(pointer.Fullnode, max)
		url = backend.Fullnode[pointer.Fullnode]
		log.Info("RPC fullnode request", "method", b.Method, "index", pointer.Fullnode, "max", max, "host", url.Host)
	}
	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return url, b.Method, cacheKey, err
}

func Collector(w http.ResponseWriter, r *http.Request) {

	// TODO: https://goenning.net/2017/03/18/server-side-cache-go/
	url, _, cacheKey, _ := route(r)
	r.URL.Host = url.Host
	r.URL.Scheme = url.Scheme

	if c, ok := cache.Get(cacheKey); ok && cacheKey != "" {
		log.Debug("Get from cache", "key", cacheKey)
		w.Header().Set("Content-Type", "application/json")
		w.Write(c.([]byte))
		return
	}

	// serve only JSON RPC request
	if r.Method != "POST" {
		w.WriteHeader(http.StatusOK)
		return
	}

	work := WorkRequest{W: w, R: r}

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
				body, _ := ioutil.ReadAll(conn.Response.Body)
				cache.Add(cacheKey, body)
				w.Write(body)
				defer conn.Response.Body.Close()
			} else {
				w.WriteHeader(http.StatusOK)
			}
			return
		}
	}
}

func StartDispatcher(nworkers int) {
	// First, initialize the channel we are going to but the workers' work channels into.
	WorkerQueue = make(chan chan WorkRequest, nworkers)

	// Now, create all of our workers.
	for i := 0; i < nworkers; i++ {
		// log.Debugf("Starting worker %d", i+1)
		worker := NewWorker(i+1, WorkerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-WorkQueue:
				go func() {
					worker := <-WorkerQueue

					worker <- work
				}()
			}
		}
	}()
}

func NewWorker(id int, workerQueue chan chan WorkRequest) Worker {
	// Create, and return the worker.
	worker := Worker{
		ID:          id,
		Work:        make(chan WorkRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
	}

	return worker
}

type Worker struct {
	ID          int
	Work        chan WorkRequest
	WorkerQueue chan chan WorkRequest
	QuitChan    chan bool
}

func (w Worker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				ServeHTTP(work.W, work.R)

			case <-w.QuitChan:
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}
