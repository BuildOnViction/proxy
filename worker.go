package main

import (
	"net/http"
)

type WorkRequest struct {
	W http.ResponseWriter
	R *http.Request
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
