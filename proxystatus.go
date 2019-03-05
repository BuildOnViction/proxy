package main

import (
	"encoding/json"
	"net/http"
)

type ProxyStatus struct {
	Status   bool `json:"status"`
	NWorkers int  `json:"number_workers"`
}

func proxystatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := ProxyStatus{
		true,
		*NWorkers,
	}
	data, _ := json.Marshal(status)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
