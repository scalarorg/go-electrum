package electrum_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"testing"
	"time"

	btc "github.com/scalarorg/go-electrum/electrum"
	"github.com/stretchr/testify/require"
)

func reverseBytes(data []byte) []byte {
	result := make([]byte, len(data))
	for i, b := range data {
		result[len(data)-1-i] = b
	}
	return result
}

func doubleSha256(data []byte) []byte {
	firstHash := sha256.Sum256(data)
	secondHash := sha256.Sum256(firstHash[:])
	return secondHash[:]
}

func buildMerkleTree(txHashes [][]byte) []byte {
	if len(txHashes) == 0 {
		return nil
	}
	if len(txHashes) == 1 {
		return txHashes[0]
	}
	// Double SHA256 each transaction hash
	hashes := make([][]byte, len(txHashes))
	for i, txHash := range txHashes {
		hashes[i] = doubleSha256(txHash)
	}
	for len(hashes) > 1 {
		nextLevel := make([][]byte, 0, (len(hashes)+1)/2)
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				combined := append(hashes[i], hashes[i+1]...)
				nextLevel = append(nextLevel, doubleSha256(combined))
			} else {
				combined := append(hashes[i], hashes[i]...)
				nextLevel = append(nextLevel, doubleSha256(combined))
			}
		}
		hashes = nextLevel
	}
	return hashes[0]
}

func GetMerkleRootFromPath(txid []byte, txIndex uint64, merklePath [][32]byte, reverse bool) []byte {
	currentHash := make([]byte, 32)
	copy(currentHash, txid)
	if reverse {
		currentHash = reverseBytes(currentHash)
	}
	currentHash = doubleSha256(currentHash)
	for _, siblingHash := range merklePath {
		var combined []byte
		if txIndex&1 == 0 {
			combined = append(currentHash, siblingHash[:]...)
		} else {
			combined = append(siblingHash[:], currentHash...)
		}
		currentHash = doubleSha256(combined)
		txIndex >>= 1
	}
	return currentHash
}

func TestBlock86055MerkleVerification(t *testing.T) {
	const electrsRpcServer = "127.0.0.1:60001"
	client, err := btc.Connect(&btc.Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", electrsRpcServer, 5*time.Second)
		},
		MethodTimeout:   30 * time.Second,
		PingInterval:    -1,
		SoftwareVersion: "testclient",
	})
	require.NoError(t, err)
	defer client.Close()

	blockHeight := 86055
	numTransactions := 16
	txHashMethod := "blockchain.transaction.id_from_pos"
	mempoolMerkleRoot := "264933d6e59f4e71342000593c1567bea56fe75cf9c70ebde33cdea18a56b51e"
	//merkleMethod := "blockchain.transaction.get_merkle"
	type TransactionInfo struct {
		TxHash   string   `json:"tx_hash"`
		Merkle   []string `json:"merkle"`
		Position int
	}
	transactions := make([]TransactionInfo, numTransactions)
	for i := 0; i < numTransactions; i++ {

		txHashParams := []interface{}{blockHeight, i, true}
		var txHashResponse TransactionInfo
		txHashResponse.Position = i

		err := client.CallRpcBlockingMethod(context.Background(), &txHashResponse, txHashMethod, txHashParams...)
		require.NoError(t, err)
		fmt.Printf("TxHash[%d]=%s\n", txHashResponse.Position, txHashResponse.TxHash)
		fmt.Printf("TxProof[%d]=%v\n", txHashResponse.Position, txHashResponse.Merkle)
		// merkleParams := []interface{}{txHashResponse.TxHash, blockHeight}
		// var merkleResponse btc.GetMerkleResult
		// err = client.CallRpcBlockingMethod(context.Background(), &merkleResponse, merkleMethod, merkleParams...)
		// require.NoError(t, err)
		// fmt.Printf("TxHash[%d]=%+v\n", merkleResponse.Pos, merkleResponse.Merkle)
		// transactions[i] = TransactionInfo{
		// 	TxHash:     txHashResponse.TxHash,
		// 	Merkle: merkleResponse.Merkle,
		// 	Position:   merkleResponse.Pos,
		// }
		//fmt.Printf("Transaction %d: Hash=%s, MerklePath=%v\n", txHashResponse.Position, txHashResponse.TxHash, txHashResponse.Merkle)
		//fmt.Printf("TxHash[%d]=%+v\n", merkleResponse.Pos, merkleResponse.Merkle)
	}
	txHashes := make([][]byte, numTransactions)
	for i, tx := range transactions {
		hashBytes, err := hex.DecodeString(tx.TxHash)
		require.NoError(t, err)
		txHashes[i] = hashBytes
	}
	merkleRoot := buildMerkleTree(txHashes)
	require.Equal(t, mempoolMerkleRoot, hex.EncodeToString(merkleRoot))
	fmt.Printf("Calculated merkle root: %x\n", merkleRoot)
	for i, tx := range transactions {
		merklePath := make([][32]byte, len(tx.Merkle))
		for j, hash := range tx.Merkle {
			hashBytes, err := hex.DecodeString(hash)
			require.NoError(t, err)
			//reversedHash := reverseBytes(hashBytes)
			copy(merklePath[j][:], hashBytes)
		}
		txHashBytes, err := hex.DecodeString(tx.TxHash)
		require.NoError(t, err)
		calculatedRoot := GetMerkleRootFromPath(txHashBytes, uint64(tx.Position), merklePath, true)
		require.Equal(t, merkleRoot, calculatedRoot, "Transaction %d merkle proof verification failed", i)
		fmt.Printf("Transaction %d merkle proof verified successfully\n", i)
	}
	fmt.Printf("All %d transaction merkle proofs verified successfully!\n", numTransactions)
}

// block hash 000000000e3a834f586d716682aad42e58028ad6358f957c879327a03ef18d35
func TestBlock85733MerkleVerification(t *testing.T) {
	const electrsRpcServer = "127.0.0.1:60001"
	client, err := btc.Connect(&btc.Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", electrsRpcServer, 5*time.Second)
		},
		MethodTimeout:   30 * time.Second,
		PingInterval:    -1,
		SoftwareVersion: "testclient",
	})
	require.NoError(t, err)
	defer client.Close()

	blockHeight := 85733
	//numTransactions := 16
	//txHashMethod := "blockchain.transaction.id_from_pos"
	//mempoolMerkleRoot := "264933d6e59f4e71342000593c1567bea56fe75cf9c70ebde33cdea18a56b51e"
	merkleMethod := "blockchain.transaction.get_merkle"
	txHash := "5fdf97225eee035355a28d8653994a7c5132152ab5b79fdf83cff6626658731a"
	type TransactionInfo struct {
		TxHash   string   `json:"tx_hash"`
		Merkle   []string `json:"merkle"`
		Position int
	}

	// txHashParams := []interface{}{blockHeight, i, true}
	// var txHashResponse TransactionInfo
	// txHashResponse.Position = i

	// err := client.CallRpcBlockingMethod(context.Background(), &txHashResponse, txHashMethod, txHashParams...)
	// require.NoError(t, err)
	// fmt.Printf("TxHash[%d]=%s\n", txHashResponse.Position, txHashResponse.TxHash)
	// fmt.Printf("TxProof[%d]=%v\n", txHashResponse.Position, txHashResponse.Merkle)
	merkleParams := []interface{}{txHash, blockHeight}
	var merkleResponse btc.GetMerkleResult
	err = client.CallRpcBlockingMethod(context.Background(), &merkleResponse, merkleMethod, merkleParams...)
	require.NoError(t, err)
	fmt.Printf("TxHash[%d]=%+v\n", merkleResponse.Pos, merkleResponse.Merkle)

	//fmt.Printf("Transaction %d: Hash=%s, MerklePath=%v\n", txHashResponse.Position, txHashResponse.TxHash, txHashResponse.Merkle)
	//fmt.Printf("TxHash[%d]=%+v\n", merkleResponse.Pos, merkleResponse.Merkle)

}
