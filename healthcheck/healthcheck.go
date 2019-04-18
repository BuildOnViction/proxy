package healthcheck

import (
	"bytes"
	"encoding/json"
	log "github.com/inconshreveable/log15"
	"github.com/tomochain/proxy/utils/hexutil"
	"io/ioutil"
	"net/http"
	"net/url"
    "sync"
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

var es *StateStore = &StateStore{ state: make(map[string]EndpointState) }

func Run(u *url.URL) (*url.URL, bool) {
	var err error
	var b EthBlockNumber
	var bn uint64
	var bd []byte

	byt := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	body := bytes.NewReader(byt)
	resp, err := http.Post(u.String(), "application/json", body)
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
    es.Unlock()

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

	data, _ := json.Marshal(es)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
