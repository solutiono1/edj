package chain

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"
)

// Electrum talks to a Bitcoin Cash Electrum (Fulcrum) server over TLS. It is
// used only to read UTXOs and broadcast raw transactions that the wallet has
// already signed.
type Electrum struct {
	Host string
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
	conn, err := tls.Dial("tcp", e.Host, &tls.Config{MinVersion: tls.VersionTLS12})
	if err != nil {
		return nil, fmt.Errorf("electrum dial: %w", err)
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(15 * time.Second))

	req := electrumRequest{ID: 1, Method: method, Params: params}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("electrum write: %w", err)
	}

	line, err := bufio.NewReader(conn).ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("electrum read: %w", err)
	}
	var resp electrumResponse
	if err := json.Unmarshal(line, &resp); err != nil {
		return nil, fmt.Errorf("electrum decode: %w", err)
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("electrum %s: %s", method, resp.Error.Message)
	}
	return resp.Result, nil
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
