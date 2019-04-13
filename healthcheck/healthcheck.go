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
    blockNumber uint64
    count int
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
	if err != nil {
		log.Error("Healthcheck", "url", u.String(), "status", "NOK")
	} else {
		log.Debug("Healthcheck", "url", u.String(), "status", "OK")
	}

    // save state
    if bn == es[u.String()].blockNumber {
        c := es[u.String()].count + 1
        es[u.String()] = EndpointState{ bn, c }
    } else {
        es[u.String()] = EndpointState{ bn, 1 }
    }

	defer resp.Body.Close()
}
