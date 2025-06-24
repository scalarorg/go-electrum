// Copyright 2024 Scalar org
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

package electrum

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/scalarorg/go-electrum/electrum/types"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Before running the test, you need to start the electrum server:
// Todo: build a docker image for btc regtest and electrum server
// const electrsRpcServer = "electrs4.btc.scalar.org:80"
// const electrsRpcServer = "192.168.1.254:60001"
const electrsRpcServer = "18.140.72.123:60001"

// const electrsRpcServer = "127.0.0.1:60001"

func TestPingElectrum(t *testing.T) {
	// Local electrum server for btc testnet4
	// electrsRpcServer := "127.0.0.1:60001"
	client, err := Connect(&Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", electrsRpcServer, time.Second)
		},
		PingInterval:    time.Millisecond,
		SoftwareVersion: "testclient",
	})
	log.Info().Msgf("err: %v", err)
	require.NoError(t, err)
	err = client.ping()
	require.NoError(t, err)
	serverVersion := client.ServerVersion()
	parts := strings.Split(serverVersion.String(), ";")
	require.Equal(t, 2, len(parts))
	require.Equal(t, "1.4", parts[1])
	client.Close()
}

type vaultClientTestsuite struct {
	suite.Suite

	client *Client
}

func (s *vaultClientTestsuite) SetupTest() {
	client, err := Connect(&Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", electrsRpcServer, time.Second)
		},
		MethodTimeout:   30 * time.Second,
		PingInterval:    -1,
		SoftwareVersion: "testclient",
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *vaultClientTestsuite) TearDownTest() {
	if s.client == nil {
		return
	}
	s.client.Close()
}

func TestElectrsClient(t *testing.T) {
	suite.Run(t, &vaultClientTestsuite{})
}

// func (s *vaultClientTestsuite) TestTransactionGetFrom() {
// 	expectedResponse := []byte("\xaa\xbb\xcc")
// 	// hash := "b43da04e4968227daed5f667f68af19988af4201b36ca552ca15e07e8c70a4fd"
// 	hash := "108e40ac667908a4c1e0b2503371ed70651d2467fc02e09f54d80208afb75a31"
// 	length := 10
// 	response, err := s.client.VaultTransactionsGetFrom(context.Background(), hash, length)
// 	require.NoError(s.T(), err)
// 	// rawTx, err := hex.DecodeString(rawTxHex)
// 	// if err != nil {
// 	// 	return nil, fmt.Errorf("failed to decode transaction hex: %w", err)
// 	// }
// 	// println(hex.EncodeToString(response))
// 	// for _, tx := range response {
// 	// 	println(tx["change_amount"].(*float64))
// 	// }
// 	require.Equal(s.T(), expectedResponse, response)
// }

// func (s *vaultClientTestsuite) TestVaultTransactionGetCancel() {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	errCh := make(chan error)
// 	go func() {
// 		hash := "b43da04e4968227daed5f667f68af19988af4201b36ca552ca15e07e8c70a4fd"
// 		length := 10
// 		_, err := s.client.VaultTransactionsGetFrom(ctx, hash, length)
// 		errCh <- err
// 	}()
// 	cancel()
// 	select {
// 	case err := <-errCh:
// 		require.ErrorIs(s.T(), err, context.Canceled)
// 	case <-time.After(2 * time.Second):
// 		require.Fail(s.T(), "timeoout")
// 	}
// }

//	func (s *vaultClientTestsuite) TestVaultTransactionSubscribe() {
//		ctx, cancel := context.WithCancel(context.Background())
//		errCh := make(chan error)
//		receivedVaultTxCh := make(chan *types.VaultTransaction)
//		params := []interface{}{}
//		go func() {
//			onVaultTransaction := func(vaultTxs []types.VaultTransaction, err error) error {
//				require.NoError(s.T(), err)
//				for _, vaultTx := range vaultTxs {
//					log.Info().Msgf("vaultTx: %v", vaultTx)
//					receivedVaultTxCh <- &vaultTx
//				}
//				return nil
//			}
//			s.client.VaultTransactionSubscribe(ctx, onVaultTransaction, params)
//		}()
//		cancel()
//		select {
//		case vaultTx := <-receivedVaultTxCh:
//			log.Info().Msgf("vaultTx: %v", vaultTx)
//		case err := <-errCh:
//			require.ErrorIs(s.T(), err, context.Canceled)
//		case <-time.After(24 * time.Hour):
//			require.Fail(s.T(), "timeoout")
//		}
//	}
type TxMerkle struct {
	TxHash string   `json:"tx_hash"`
	Merkle []string `json:"merkle"`
}

func (s *vaultClientTestsuite) TestMerkleRoot() {
	method := "blockchain.transaction.id_from_pos"
	height := 85732
	//for i := 0; i < 100; i++ {
	params := []interface{}{height, 10000, true}
	var response TxMerkle
	err := s.client.CallRpcBlockingMethod(context.Background(), &response, method, params...)
	require.NoError(s.T(), err)
	log.Info().Msgf("response {%d}: %+v", 10000, response)

	//}
}

func (s *vaultClientTestsuite) TestBlock86055MerkleVerification() {
	log.Info().Msgf("TestBlock86055MerkleVerification")
	txHashMethod := "blockchain.transaction.id_from_pos"
	merkleMethod := "blockchain.transaction.get_merkle"
	blockHeight := 86055
	numTransactions := 16

	// Step 1: Get all transaction hashes and their merkle proofs
	type TransactionInfo struct {
		TxHash     string
		MerklePath []string
		Position   int
	}

	transactions := make([]TransactionInfo, numTransactions)

	// Get each transaction hash and merkle proof
	for i := 0; i < numTransactions; i++ {
		// Get transaction hash by position

		txHashParams := []interface{}{blockHeight, i, true}
		var txHashResponse string

		err := s.client.CallRpcBlockingMethod(context.Background(), &txHashResponse, txHashMethod, txHashParams...)
		require.NoError(s.T(), err)

		// Get merkle proof for this transaction

		merkleParams := []interface{}{txHashResponse, blockHeight}
		var merkleResponse GetMerkleResult

		err = s.client.CallRpcBlockingMethod(context.Background(), &merkleResponse, merkleMethod, merkleParams...)
		require.NoError(s.T(), err)

		transactions[i] = TransactionInfo{
			TxHash:     txHashResponse,
			MerklePath: merkleResponse.Merkle,
			Position:   merkleResponse.Pos,
		}

		log.Info().Msgf("Transaction %d: Hash=%s, Position=%d, MerklePath=%v",
			i, txHashResponse, merkleResponse.Pos, merkleResponse.Merkle)
	}

	// Step 2: Construct merkle tree from all transaction hashes
	log.Info().Msgf("Constructing merkle tree from %d transactions", numTransactions)

	// Convert transaction hashes to byte arrays
	txHashes := make([][]byte, numTransactions)
	for i, tx := range transactions {
		hashBytes, err := hex.DecodeString(tx.TxHash)
		require.NoError(s.T(), err)
		txHashes[i] = hashBytes
	}

	// Build merkle tree
	merkleRoot := s.buildMerkleTree(txHashes)
	log.Info().Msgf("Calculated merkle root: %x", merkleRoot)

	// Step 3: Verify each transaction's merkle proof
	log.Info().Msgf("Verifying merkle proofs for all transactions")

	for i, tx := range transactions {
		// Convert merkle path to byte arrays
		merklePath := make([][32]byte, len(tx.MerklePath))
		for j, hash := range tx.MerklePath {
			hashBytes, err := hex.DecodeString(hash)
			require.NoError(s.T(), err)
			// Reverse bytes for Bitcoin format
			reversedHash := s.reverseBytes(hashBytes)
			copy(merklePath[j][:], reversedHash)
		}

		// Get transaction hash as bytes
		txHashBytes, err := hex.DecodeString(tx.TxHash)
		require.NoError(s.T(), err)

		// Verify merkle proof
		calculatedRoot := s.GetMerkleRootFromPath(txHashBytes, uint64(tx.Position), merklePath, true)

		// Compare with our calculated root
		require.Equal(s.T(), merkleRoot, calculatedRoot,
			"Transaction %d merkle proof verification failed", i)

		log.Info().Msgf("Transaction %d merkle proof verified successfully", i)
	}

	log.Info().Msgf("All %d transaction merkle proofs verified successfully!", numTransactions)
}

// buildMerkleTree constructs a merkle tree from transaction hashes
func (s *vaultClientTestsuite) buildMerkleTree(txHashes [][]byte) []byte {
	if len(txHashes) == 0 {
		return nil
	}
	if len(txHashes) == 1 {
		return txHashes[0]
	}

	// Double SHA256 each transaction hash
	hashes := make([][]byte, len(txHashes))
	for i, txHash := range txHashes {
		hashes[i] = s.doubleSha256(txHash)
	}

	// Build tree level by level
	for len(hashes) > 1 {
		nextLevel := make([][]byte, 0, (len(hashes)+1)/2)

		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				// Concatenate two hashes and hash them
				combined := append(hashes[i], hashes[i+1]...)
				nextLevel = append(nextLevel, s.doubleSha256(combined))
			} else {
				// Odd number of hashes, duplicate the last one
				combined := append(hashes[i], hashes[i]...)
				nextLevel = append(nextLevel, s.doubleSha256(combined))
			}
		}

		hashes = nextLevel
	}

	return hashes[0]
}

// doubleSha256 performs double SHA256 hashing
func (s *vaultClientTestsuite) doubleSha256(data []byte) []byte {
	firstHash := s.sha256(data)
	return s.sha256(firstHash)
}

// sha256 performs SHA256 hashing
func (s *vaultClientTestsuite) sha256(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// reverseBytes reverses the byte order
func (s *vaultClientTestsuite) reverseBytes(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[len(data)-1-i] = b
	}
	return result
}

// GetMerkleRootFromPath calculates merkle root from transaction hash, position, and merkle path
func (s *vaultClientTestsuite) GetMerkleRootFromPath(txid []byte, txIndex uint64, merklePath [][32]byte, reverse bool) []byte {
	currentHash := make([]byte, 32)
	copy(currentHash, txid)

	if reverse {
		currentHash = s.reverseBytes(currentHash)
	}

	currentHash = s.doubleSha256(currentHash)

	for _, siblingHash := range merklePath {
		var combined []byte

		if txIndex&1 == 0 {
			// Current hash is on the left
			combined = append(currentHash, siblingHash[:]...)
		} else {
			// Current hash is on the right
			combined = append(siblingHash[:], currentHash...)
		}

		currentHash = s.doubleSha256(combined)
		txIndex >>= 1
	}

	return currentHash
}

func (s *vaultClientTestsuite) TestMerkTreeProof() {
	log.Info().Msgf("TestMerkTreeProof")
	method := "blockchain.transaction.get_merkle"
	params := []interface{}{"04fde533acb663bc0b72f049e8c9d5f53ecf53da7b7d8cf15214558b0a0d2f29", 85732}

	// Use the blocking method instead of async
	var response GetMerkleResult
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := s.client.CallRpcBlockingMethod(ctx, &response, method, params...)
	require.NoError(s.T(), err)

	log.Info().Msgf("merkle: %+v", response)
}
func (s *vaultClientTestsuite) TestVaultBlocksSubscribe() {
	log.Info().Msgf("TestVaultBlocksSubscribe")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Make sure to cancel the context when the function exits

	errCh := make(chan error)
	receivedVaultTxCh := make(chan *types.VaultBlock)
	params := []interface{}{1, 86983}
	onVaultBlocks := func(params json.RawMessage, err error) {
		require.NoError(s.T(), err)
		var vaultBlocks []types.VaultBlock
		err = json.Unmarshal(params, &vaultBlocks)
		if err != nil {
			log.Info().Msgf("error unmarshalling params: %v, response: %v", err, string(params))
			return
		} else {
			log.Info().Msgf("Unmarshalled VaultBlocks: %d", len(vaultBlocks))
		}
		require.NoError(s.T(), err)
		for _, vaultBlock := range vaultBlocks {
			receivedVaultTxCh <- &vaultBlock
		}
	}
	err := s.client.SubscribeEvent(ctx, "vault.blocks.subscribe", onVaultBlocks, params...)
	require.NoError(s.T(), err)

	// Wait for a reasonable amount of time for data, then exit
	timeout := time.After(10 * time.Second)
	count := 0
	for {
		select {
		case vaultBlock := <-receivedVaultTxCh:
			log.Info().Msgf("Received vaultBlock: %d, %v", vaultBlock.Height, vaultBlock.Hash)
			count++
			// Exit after receiving some data
			if count >= 5 {
				return
			}
		case err := <-errCh:
			require.ErrorIs(s.T(), err, context.Canceled)
			return
		case <-timeout:
			log.Info().Msgf("Test completed after timeout, received %d blocks", count)
			return
		}
	}
}
