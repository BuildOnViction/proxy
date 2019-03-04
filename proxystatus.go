package main

import (
	"fmt"
	"io"
	"net/http"
)

func proxystatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	io.WriteString(w, fmt.Sprintf(`{
			"status": %s,
			"number_workers": %d,
			"cache_limit": %d,
			"cache_expiration": %s,
		}`, "true", *NWorkers, *CacheLimit, *CacheExpiration))
	return
}
