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

func Run(u *url.URL) {
	var err error
	var b EthBlockNumber

	byt := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	body := bytes.NewReader(byt)
	resp, err := http.Post(u.String(), "application/json", body)
	bd, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(bd, &b)
	bn, err := hexutil.DecodeUint64(b.Result)
	log.Debug("Body", "number", bn)
	if err != nil {
		log.Error("Healthcheck", "url", u.String(), "status", "NOK")
	} else {
		log.Debug("Healthcheck", "url", u.String(), "status", "OK")
	}
	defer resp.Body.Close()
}
