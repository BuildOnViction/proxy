package main

import (
	log "github.com/inconshreveable/log15"
	"io"
	"net/http"
)

type WorkRequest struct {
	W http.ResponseWriter
	R *http.Request
}

type HttpConnection struct {
	Request  *http.Request
	Response *http.Response
}

type HttpConnectionChannel chan *HttpConnection

var WorkerQueue chan chan WorkRequest

var WorkQueue = make(chan WorkRequest, 100)

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
