package healthcheck

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	log "github.com/inconshreveable/log15"
	"github.com/tomochain/proxy/config"
	"github.com/tomochain/proxy/utils/hexutil"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type EthBlockNumber struct {
	Result string
}

type RpcBlock struct {
	Result map[string]interface{} `json:"result"`
}

type EndpointState struct {
	BlockNumber uint64 `json:"blockNumber"`
	Count       int    `json:"count"`
	Status      string `json:"status"`
}

type StateStore struct {
	sync.Mutex
	state map[string]EndpointState
}

var es *StateStore = &StateStore{state: make(map[string]EndpointState)}

type ProxyStatus struct {
	Status bool `json:"status"`
}

func Run(u *url.URL) (*url.URL, bool) {
	var err error
	var b EthBlockNumber
	var block RpcBlock
	var bn uint64
	var timestamp uint64
	var bd []byte
	var req *http.Request
	var resp *http.Response

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 10 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 4 * time.Second,
		ResponseHeaderTimeout: 3 * time.Second,
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 60,
	}
	byt := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	body := bytes.NewReader(byt)
	req, _ = http.NewRequest("POST", u.String(), body)
	req.Header.Set("Connection", "close")
	req.Close = true
	req.Header.Set("Content-Type", "application/json")

	c := config.GetConfig()
	if c.Headers != nil {
		for key, value := range *c.Headers {
			req.Header.Set(key, value)
			if host := req.Header.Get("Host"); host != "" {
				req.Host = host
			}
		}
	}

	resp, err = client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		defer req.Body.Close()
		bd, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(bd, &b)
		}
		if err == nil {
			bn, err = hexutil.DecodeUint64(b.Result)
		}
	}

	byt = []byte(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}`)
	body = bytes.NewReader(byt)
	req, _ = http.NewRequest("POST", u.String(), body)
	req.Header.Set("Connection", "close")
	req.Close = true
	req.Header.Set("Content-Type", "application/json")

	c = config.GetConfig()
	if c.Headers != nil {
		for key, value := range *c.Headers {
			req.Header.Set(key, value)
			if host := req.Header.Get("Host"); host != "" {
				req.Host = host
			}
		}
	}

	resp, err = client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		defer req.Body.Close()
		bd, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(bd, &block)
		}
		if err == nil {
			timestamp, err = hexutil.DecodeUint64(block.Result["timestamp"].(string))
		}
	}

	delta := uint64(time.Now().Unix()) - timestamp

	// save state
	es.Lock()
	defer es.Unlock()
	status := "OK"
	count := 1
	if bn == es.state[u.String()].BlockNumber {
		count = es.state[u.String()].Count + 1
		if count > 10 {
			status = "NOK"
		}
	}

	if delta > 200 {
		status = "NOK"
	}

	es.state[u.String()] = EndpointState{bn, count, status}

	if err != nil {
		es.state[u.String()] = EndpointState{bn, 0, "NOK"}
	}

	if err != nil || es.state[u.String()].Status == "NOK" {
		log.Error("Healthcheck", "url", u.String(), "number", bn, "count", es.state[u.String()].Count, "status", "NOK", "delta", delta, "err", err)
	} else {
		log.Info("Healthcheck", "url", u.String(), "number", bn, "count", es.state[u.String()].Count, "status", "OK", "delta", delta)
	}

	if es.state[u.String()].Status == "NOK" {
		return u, false
	} else {
		return u, true
	}

}

func GetEndpointStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()

	endpoint := r.URL.Query().Get("u")
	data, _ := json.Marshal(es.state[endpoint])
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func GetProxyStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	defer r.Body.Close()

	status := ProxyStatus{
		true,
	}

	for _, value := range es.state {
		if value.Status == "NOK" {
			status.Status = false
			break
		}
	}

	if status.Status {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusBadGateway)
	}

	data, _ := json.Marshal(status)
	w.Write(data)
	return
}
