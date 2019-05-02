package healthcheck

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	log "github.com/inconshreveable/log15"
	"github.com/tomochain/proxy/utils/hexutil"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type EthBlockNumber struct {
	Result string
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

func Run(u *url.URL) (*url.URL, bool) {
	var err error
	var b EthBlockNumber
	var bn uint64
	var bd []byte
	var req *http.Request
	var resp *http.Response

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 60,
	}
	byt := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	body := bytes.NewReader(byt)
	req, _ = http.NewRequest("POST", u.String(), body)
	req.Header.Set("Connection", "close")
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err == nil {
		defer resp.Body.Close()
		bd, err = ioutil.ReadAll(resp.Body)
		if err == nil {
			err = json.Unmarshal(bd, &b)
		}
		if err == nil {
			bn, err = hexutil.DecodeUint64(b.Result)
		}
	}

	// save state
	es.Lock()
	defer es.Unlock()
	if bn == es.state[u.String()].BlockNumber {
		c := es.state[u.String()].Count + 1
		status := "OK"
		if c > 10 {
			status = "NOK"
		}
		es.state[u.String()] = EndpointState{bn, c, status}
	} else {
		es.state[u.String()] = EndpointState{bn, 1, "OK"}
	}

	if err != nil {
		es.state[u.String()] = EndpointState{bn, 0, "NOK"}
	}

	if err != nil || es.state[u.String()].Status == "NOK" {
		log.Error("Healthcheck", "url", u.String(), "number", bn, "count", es.state[u.String()].Count, "status", "NOK", "err", err)
	} else {
		log.Info("Healthcheck", "url", u.String(), "number", bn, "count", es.state[u.String()].Count, "status", "OK")
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
