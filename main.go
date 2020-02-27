package main

import (
	"crypto/tls"
	"flag"
	"fmt"
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
	WsAddr          = flag.String("ws", "", "Address to listen for WS requests on")
	WssAddr         = flag.String("wss", "", "Address to listen for WSS requests on")
	ConfigFile      = flag.String("config", "./config/default.json", "Path to config file")
	CacheLimit      = flag.Int("cacheLimit", 100000, "Cache limit")
	CacheExpiration = flag.String("cacheExpiration", "2s", "Cache expiration")
	Verbosity       = flag.Int("verbosity", 3, "Log Verbosity")
	Version         = flag.Bool("version", false, "Version")
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

	if *Version != false {
		fmt.Println("v0.1.8")
		return
	}

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

	urls = []*url.URL{}
	for i := 0; i < len(c.Websocket); i++ {
		url, _ := url.Parse(c.Websocket[i])
		urls = append(urls, url)
	}
	backend.Websocket = urls

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

	wsProxyHandler := WsProxyHandler(backend.Websocket)

	mux := http.NewServeMux()
	mux.HandleFunc("/proxystatus", healthcheck.GetProxyStatus)
	mux.HandleFunc("/endpointstatus", healthcheck.GetEndpointStatus)

	if c.WsServerName != "" && len(backend.Websocket) > 0 {
		mux.HandleFunc(c.WsServerName+"/", wsProxyHandler.ServeHTTP)
	}

	mux.HandleFunc("/", Collector)
	handler := cors.AllowAll().Handler(mux)

	tlsConfig := &tls.Config{}
	for i := 0; i < len(c.Certs); i++ {
		sslCrt := c.Certs[i].Crt
		sslKey := c.Certs[i].Key
		cert, err := tls.LoadX509KeyPair(sslCrt, sslKey)
		if err != nil {
			log.Error("SSL certs", "error", err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)

		tlsConfig.BuildNameToCertificate()
	}

	if *HTTPSAddr != "" {
		server := http.Server{
			Addr:         *HTTPSAddr,
			Handler:      handler,
			TLSConfig:    tlsConfig,
			WriteTimeout: 120 * time.Second,
			ReadTimeout:  120 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		go func() {
			log.Info("HTTPS server listening on", "addr", *HTTPSAddr)
			if err := server.ListenAndServeTLS("", ""); err != nil {
				log.Error("Failed start https server", "error", err.Error())
			}
		}()
	}

	if *WsAddr != "" && len(backend.Websocket) > 0 {
		go func() {
			log.Info("WS server listening on", "addr", *WsAddr)
			if err := http.ListenAndServe(*WsAddr, wsProxyHandler); err != nil {
				log.Error("Failed start ws server", "error", err.Error())
			}
		}()
	}

	if *WssAddr != "" && len(backend.Websocket) > 0 {
		server := http.Server{
			Addr:         *HTTPSAddr,
			Handler:      handler,
			TLSConfig:    tlsConfig,
			WriteTimeout: 120 * time.Second,
			ReadTimeout:  120 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		go func() {
			log.Info("WSS server listening on", "addr", *WssAddr)
			if err := server.ListenAndServeTLS("", ""); err != nil {
				log.Error("Failed start ws server", "error", err.Error())
			}
		}()
	}

	log.Info("HTTP server listening on", "addr", *HTTPAddr)
	server := http.Server{
		Addr:         *HTTPAddr,
		Handler:      handler,
		WriteTimeout: 120 * time.Second,
		ReadTimeout:  120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Error("Failed start http server", "error", err.Error())
	}
}
