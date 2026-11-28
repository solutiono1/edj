package chain

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Electrum talks to a Bitcoin Cash Electrum (Fulcrum) server over TLS. It is
// used only to read UTXOs and broadcast raw transactions that the wallet has
// already signed. Hosts may contain several host:port entries. They are tried
// in order so one public server going down does not break wallet reads.
type Electrum struct {
	Host  string
	Hosts []string
}

func NewElectrum(spec string) *Electrum {
	var hosts []string
	for _, part := range strings.Split(spec, ",") {
		if host := strings.TrimSpace(part); host != "" {
			hosts = append(hosts, host)
		}
	}
	return &Electrum{Hosts: hosts}
}

func (e Electrum) endpoints() []string {
	var out []string
	for _, host := range e.Hosts {
		if host = strings.TrimSpace(host); host != "" {
			out = append(out, host)
		}
	}
	if len(out) == 0 && strings.TrimSpace(e.Host) != "" {
		out = append(out, strings.TrimSpace(e.Host))
	}
	return out
}

type ElectrumUtxo struct {
	TxHash string `json:"tx_hash"`
	TxPos  int    `json:"tx_pos"`
	Height int    `json:"height"`
	Value  int64  `json:"value"`
}

type electrumRequest struct {
	ID     int64  `json:"id"`
	Method string `json:"method"`
	Params []any  `json:"params"`
}

type electrumResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (e Electrum) call(method string, params []any) (json.RawMessage, error) {
	endpoints := e.endpoints()
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("electrum host is not configured")
	}
	var lastErr error
	for _, host := range endpoints {
		raw, retry, err := e.callOne(host, method, params)
		if err == nil {
			return raw, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, lastErr
}

func (e Electrum) callOne(host, method string, params []any) (json.RawMessage, bool, error) {
	conn, err := tls.Dial("tcp", host, &tls.Config{MinVersion: tls.VersionTLS12})
	if err != nil {
		return nil, true, fmt.Errorf("electrum dial %s: %w", host, err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))

	req := electrumRequest{ID: 1, Method: method, Params: params}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, false, err
	}
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return nil, true, fmt.Errorf("electrum write %s: %w", host, err)
	}

	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return nil, true, fmt.Errorf("electrum read %s: %w", host, err)
	}
	var resp electrumResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, false, fmt.Errorf("electrum decode %s: %w", host, err)
	}
	if resp.Error != nil {
		return nil, false, fmt.Errorf("electrum %s on %s: %s", method, host, resp.Error.Message)
	}
	return resp.Result, false, nil
}

func (e Electrum) GetUtxos(address string) ([]ElectrumUtxo, error) {
	raw, err := e.call("blockchain.address.listunspent", []any{address})
	if err != nil {
		return nil, err
	}
	var out []ElectrumUtxo
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("electrum utxos decode: %w", err)
	}
	return out, nil
}

func (e Electrum) Broadcast(rawTx string) (string, error) {
	raw, err := e.call("blockchain.transaction.broadcast", []any{rawTx})
	if err != nil {
		return "", err
	}
	var txid string
	if err := json.Unmarshal(raw, &txid); err != nil {
		return "", fmt.Errorf("electrum broadcast decode: %w", err)
	}
	return txid, nil
}
