// Copyright 2024 Scalar Org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// TxInfo is returned by ScriptHashGetHistory.
type TxInfo struct {
	// >0 for a confirmed transaction. 0 for an unconfirmed transaction. -1 for an unconfirmed
	// transaction with an unconfirmed parent transaction.
	Height int    `json:"height"`
	TxHash string `json:"tx_hash"`
}

// TxHistory is returned by ScriptHashGetHistory.
type TxHistory []*TxInfo

// Status encodes the status of the address history as a hex-encoded hash, according to the Electrum
// specification.
// https://github.com/spesmilo/electrumx/blob/75d0e25bf022355941ae4ecb4d5eba41babcf041/docs/protocol-basics.rst#status
func (history TxHistory) Status() string {
	if len(history) == 0 {
		return ""
	}
	status := bytes.Buffer{}
	for _, tx := range history {
		status.WriteString(fmt.Sprintf("%s:%d:", tx.TxHash, tx.Height))
	}
	hash := sha256.Sum256(status.Bytes())
	return hex.EncodeToString(hash[:])
}

// Header is used by HeadersSubscribe().
type Header struct {
	Height int `json:"height"`
}

type VaultTransaction struct {
	Height        int    `json:"confirmed_height"`
	TxHash        string `json:"txid"`
	TxPosition    int32  `json:"tx_position"`
	Amount        int64  `json:"amount"`
	SenderAddress string `json:"sender_address"`
	SenderPubkey  string `json:"sender_pubkey"`
	DestChainId   []byte `json:"destination_chain_id"`
	DestChainHash []byte `json:"destination_contract_address"`
	DestAddress   []byte `json:"destination_recipient_address"`
	Timestamp     int64  `json:"timestamp"`
	TxContent     string `json:"tx_content"`
}

func (tx *VaultTransaction) Marshal() ([]byte, error) {
	return json.Marshal(tx)
}
