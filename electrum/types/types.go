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
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
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

type HeaderEntry struct {
	Hex    string `json:"hex"` //Header hex
	Hash   string `json:"hash"`
	Height int    `json:"height"` //Height of the block
}
type BlockchainHeader struct {
	Version       int32  `json:"version"`
	PrevBlockhash []byte `json:"prev_blockhash"`
	MerkleRoot    []byte `json:"MerkleRoot"`
	Time          uint32 `json:"time"`
	CompactTarget uint32 `json:"compact_target"`
	Nonce         uint32 `json:"nonce"`
	Hash          string `json:"hash"`
	Height        int    `json:"height"` //Height of the block
}

func ParseHeaderEntry(headerEntry *HeaderEntry) BlockchainHeader {
	bytes, err := hex.DecodeString(headerEntry.Hex)
	if err != nil {
		log.Error().Err(err).Str("Hex", headerEntry.Hex).Msg("Cannot decode header hex")
	}
	version := binary.LittleEndian.Uint32(bytes[0:4])
	time := binary.LittleEndian.Uint32(bytes[68:72])
	compactTarget := binary.LittleEndian.Uint32(bytes[72:76])
	nonce := binary.LittleEndian.Uint32(bytes[76:80])
	return BlockchainHeader{
		Version:       int32(version),
		PrevBlockhash: bytes[4:36],
		MerkleRoot:    bytes[36:68],
		Time:          time,
		CompactTarget: compactTarget,
		Nonce:         nonce,
		Hash:          headerEntry.Hash,
		Height:        headerEntry.Height,
	}
}

type VaultTransaction struct {
	Height               int    `json:"confirmed_height"`
	BlockHash            string `json:"block_hash"`
	TxHash               string `json:"txid"`
	TxPosition           int    `json:"tx_position"`
	Amount               uint64 `json:"amount"`
	StakerAddress        string `json:"staker_address"`
	StakerPubkey         string `json:"staker_pubkey"`
	CovenantQuorum       uint8  `json:"covenant_quorum"`
	ServiceTag           []byte `json:"service_tag"`
	VaultTxType          uint8  `json:"vault_tx_type"`
	DestChain            uint64 `json:"destination_chain"`
	DestTokenAddress     string `json:"destination_token_address"`
	DestRecipientAddress string `json:"destination_recipient_address"`
	SessionSequence      uint64 `json:"session_sequence"` //For redeem session
	CustodianGroupUid    []byte `json:"custodian_group_uid"`
	ScriptPubkey         []byte `json:"script_pubkey"`
	Timestamp            int64  `json:"timestamp"`
	Key                  string `json:"key"`
	TxContent            string `json:"tx_content"`
}

func (tx *VaultTransaction) Marshal() ([]byte, error) {
	return json.Marshal(tx)
}

type VaultTx struct {
	RawTx  []byte           `json:"raw_tx"`
	TxInfo VaultTransaction `json:"tx_info"`
	Proof  []string         `json:"proof"`
}
type VaultBlock struct {
	Hash     string    `json:"hash"`
	Height   int       `json:"height"`
	Time     uint64    `json:"time"`
	VaultTxs []VaultTx `json:"txes"`
}
