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
    Count int `json:"count"`
    Status string `json:"status"`
}

var es map[string]EndpointState = make(map[string]EndpointState)

func Run(u *url.URL) {
	var err error
	var b EthBlockNumber

	byt := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	body := bytes.NewReader(byt)
	resp, err := http.Post(u.String(), "application/json", body)
	bd, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bd, &b)
	bn, err := hexutil.DecodeUint64(b.Result)

    // save state
    if bn == es[u.String()].BlockNumber {
        c := es[u.String()].Count + 1
        status := "OK"
        if c > 10 {
            status = "NOK"
        }
        es[u.String()] = EndpointState{ bn, c, status }
    } else {
        es[u.String()] = EndpointState{ bn, 1, "OK" }
    }

	if err != nil {
		log.Error("Healthcheck", "url", u.String(), "number", bn, "count", es[u.String()].Count, "status", "NOK")
	} else {
		log.Error("Healthcheck", "url", u.String(), "number", bn, "count", es[u.String()].Count, "status", es[u.String()].Status)
	}

	defer resp.Body.Close()
}

func GetEndpointStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data, _ := json.Marshal(es)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}
