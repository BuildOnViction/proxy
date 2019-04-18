package healthcheck

import (
	"bytes"
	"encoding/json"
	log "github.com/inconshreveable/log15"
	"github.com/tomochain/proxy/utils/hexutil"
	"io/ioutil"
	"net/http"
	"net/url"
)

type EthBlockNumber struct {
	Result string
}

type EndpointState struct {
	BlockNumber uint64 `json:"blockNumber"`
	Count       int    `json:"count"`
	Status      string `json:"status"`
}

var es map[string]EndpointState = make(map[string]EndpointState)

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
	if bn == es[u.String()].BlockNumber {
		c := es[u.String()].Count + 1
		status := "OK"
		if c > 10 {
			status = "NOK"
		}
		es[u.String()] = EndpointState{bn, c, status}
	} else {
		es[u.String()] = EndpointState{bn, 1, "OK"}
	}

	if err != nil {
		es[u.String()] = EndpointState{bn, 0, "NOK"}
	}

	if err != nil || es[u.String()].Status == "NOK" {
		log.Error("Healthcheck", "url", u.String(), "number", bn, "count", es[u.String()].Count, "status", "NOK", "err", err)
	} else {
		log.Info("Healthcheck", "url", u.String(), "number", bn, "count", es[u.String()].Count, "status", "OK")
	}

	if es[u.String()].Status == "NOK" {
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
