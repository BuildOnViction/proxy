package main

import (
	"flag"
	"fmt"
	"github.com/tomochain/proxy/config"
	"net/http"
)

var (
	NWorkers   = flag.Int("n", 16, "The number of workers to start")
	HTTPAddr   = flag.String("http", "0.0.0.0:3000", "Address to listen for HTTP requests on")
	ConfigFile = flag.String("config", "./config/default.json", "Path to config file")
)

func main() {
	// Parse the command-line flags.
	flag.Parse()

	config.Init(*ConfigFile)
	c := config.GetConfig()
	backend.Masternode = c.Masternode
	backend.Fullnode = c.Fullnode
	fmt.Println(*ConfigFile)

	fmt.Println("Starting the dispatcher")
	StartDispatcher(*NWorkers)

	fmt.Println("Registering the collector")
	http.HandleFunc("/", Collector)

	fmt.Println("HTTP server listening on", *HTTPAddr)
	if err := http.ListenAndServe(*HTTPAddr, nil); err != nil {
		fmt.Println(err.Error())
	}
}
