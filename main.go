package main

import (
	"flag"
	log "github.com/inconshreveable/log15"
	"github.com/rs/cors"
	"github.com/tomochain/proxy/cache"
	"github.com/tomochain/proxy/cache/lru"
	"github.com/tomochain/proxy/config"
	"github.com/tomochain/proxy/healthcheck"
	"net/http"
	"net/url"
	"time"
)

var (
	NWorkers        = flag.Int("n", 16, "The number of workers to start")
	HTTPAddr        = flag.String("http", "0.0.0.0:3000", "Address to listen for HTTP requests on")
	HTTPSAddr       = flag.String("https", "", "Address to listen for HTTPS requests on")
	ConfigFile      = flag.String("config", "./config/default.json", "Path to config file")
	CacheLimit      = flag.Int("cacheLimit", 100000, "Cache limit")
	CacheExpiration = flag.String("cacheExpiration", "2s", "Cache expiration")
	Verbosity       = flag.Int("verbosity", 3, "Log Verbosity")
)

type arrEndpointFlags []string

func (i *arrEndpointFlags) String() string {
	return "List of endpoint urls"
}

func (i *arrEndpointFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var endpoints arrEndpointFlags

var storage cache.Storage

func main() {
	// Parse the command-line flags.
	flag.Var(&endpoints, "endpoint", "List of endpoint urls")
	flag.Parse()

	// setup config
	config.Init(*ConfigFile)
	c := config.GetConfig()
	var urls []*url.URL
	if len(endpoints) > 0 {
		// overide config file
		c.Fullnode = endpoints
		c.Masternode = endpoints
	}

	for i := 0; i < len(c.Masternode); i++ {
		url, _ := url.Parse(c.Masternode[i])
		urls = append(urls, url)
	}
	backend.Masternode = urls

	urls = []*url.URL{}
	for i := 0; i < len(c.Fullnode); i++ {
		url, _ := url.Parse(c.Fullnode[i])
		urls = append(urls, url)
	}
	backend.Fullnode = urls

	// setup log
	h := log.LvlFilterHandler(log.Lvl(*Verbosity), log.StdoutHandler)
	log.Root().SetHandler(h)
	log.Debug("Starting the proxy", "workers", *NWorkers, "config", *ConfigFile)

	// Cache
	storage, _ = lrucache.NewStorage(*CacheLimit)

	// Healthcheck
	hlm := make(chan *url.URL)
	nhlm := make(chan *url.URL)
	hlf := make(chan *url.URL)
	nhlf := make(chan *url.URL)
	go func() {
		for {
			<-time.After(10 * time.Second)
			for i := 0; i < len(c.Fullnode); i++ {
				url, _ := url.Parse(c.Fullnode[i])
				go func() {
					u, ok := healthcheck.Run(url)
					if !ok {
						nhlf <- u
					} else {
						hlf <- u
					}

				}()
			}
			for i := 0; i < len(c.Masternode); i++ {
				url, _ := url.Parse(c.Masternode[i])
				go func() {
					u, ok := healthcheck.Run(url)
					if !ok {
						nhlm <- u
					} else {
						hlm <- u
					}
				}()
			}
		}
	}()

	go func() {
		for {
			select {
			case u := <-nhlm:
				for i := 0; i < len(backend.Masternode); i++ {
					if u.String() == backend.Masternode[i].String() {
						backend.Lock()
						backend.Masternode = append(backend.Masternode[:i], backend.Masternode[i+1:]...)
						backend.Unlock()
						break
					}
				}
			case u := <-nhlf:
				for i := 0; i < len(backend.Fullnode); i++ {
					if u.String() == backend.Fullnode[i].String() {
						backend.Lock()
						backend.Fullnode = append(backend.Fullnode[:i], backend.Fullnode[i+1:]...)
						backend.Unlock()
						break
					}
				}
			case u := <-hlm:
				b := true
				for i := 0; i < len(backend.Masternode); i++ {
					if u.String() == backend.Masternode[i].String() {
						b = false
						break
					}
				}
				if b {
					backend.Lock()
					backend.Masternode = append(backend.Masternode, u)
					backend.Unlock()
				}
			case u := <-hlf:
				b := true
				for i := 0; i < len(backend.Fullnode); i++ {
					if u.String() == backend.Fullnode[i].String() {
						b = false
						break
					}
				}
				if b {
					backend.Lock()
					backend.Fullnode = append(backend.Fullnode, u)
					backend.Unlock()
				}
			}
		}
	}()

	StartDispatcher(*NWorkers)

	mux := http.NewServeMux()
	mux.HandleFunc("/proxystatus", proxystatus)
	mux.HandleFunc("/endpointstatus", healthcheck.GetEndpointStatus)

	mux.HandleFunc("/", Collector)
	handler := cors.Default().Handler(mux)

	if *HTTPSAddr != "" {
		go func() {
			log.Info("HTTPS server listening on", "addr", *HTTPSAddr)
			if err := http.ListenAndServeTLS(*HTTPSAddr, "certs/domain.crt", "certs/domain.key", handler); err != nil {
				log.Error("Failed start https server", "error", err.Error())
			}
		}()
	}

	log.Info("HTTP server listening on", "addr", *HTTPAddr)
	if err := http.ListenAndServe(*HTTPAddr, handler); err != nil {
		log.Error("Failed start http server", "error", err.Error())
	}
}
