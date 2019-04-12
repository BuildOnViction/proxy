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
	log.Debug("Healthcheck test", "url", u.String())

	byt := []byte(`{"jsonrpc":"2.0","method":"eth_blockNumber","params":[],"id":1}`)
	body := bytes.NewReader(byt)
	resp, _ := http.Post(u.String(), "application/json", body)
	var b EthBlockNumber
	bd, _ := ioutil.ReadAll(resp.Body)
	_ = json.Unmarshal(bd, &b)
	bn, _ := hexutil.DecodeUint64(b.Result)
	log.Debug("Body", "number", bn)
	defer resp.Body.Close()
}
