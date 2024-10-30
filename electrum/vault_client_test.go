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
	"net"
	"strings"
	"testing"
	"time"

	"github/scalar.org/go-electrum/electrum/types"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestPingElectrum(t *testing.T) {
	// Local electrum server for btc testnet4
	electrsRpcServer := "127.0.0.1:60001"
	client, err := Connect(&Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", electrsRpcServer, time.Second)
		},
		PingInterval:    time.Millisecond,
		SoftwareVersion: "testclient",
	})
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
	electrsRpcServer := "127.0.0.1:60001"
	client, err := Connect(&Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", electrsRpcServer, time.Second)
		},
		MethodTimeout:   time.Second,
		PingInterval:    -1,
		SoftwareVersion: "testclient",
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *vaultClientTestsuite) TearDownTest() {
	s.client.Close()
}

func TestElectrsClient(t *testing.T) {
	suite.Run(t, &vaultClientTestsuite{})
}

func (s *vaultClientTestsuite) TestTransactionGet() {
	expectedResponse := []byte("\xaa\xbb\xcc")
	response, err := s.client.VaultTransactionGet(context.Background(), "testtxhash")
	require.NoError(s.T(), err)
	require.Equal(s.T(), expectedResponse, response)
}

func (s *vaultClientTestsuite) TestVaultTransactionGetCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		_, err := s.client.VaultTransactionGet(ctx, "testtxhash")
		errCh <- err
	}()
	cancel()
	select {
	case err := <-errCh:
		require.ErrorIs(s.T(), err, context.Canceled)
	case <-time.After(2 * time.Second):
		require.Fail(s.T(), "timeoout")
	}
}

func (s *vaultClientTestsuite) TestVaultTransactionSubscribe() {
	latestTxHeigh := 52669 //First vaultTx is in the block 52670
	lastestTxPos := 0
	numNotifiedVaultTxs := 1000
	receivedVaultTxCh := make(chan *types.VaultTransaction)
	onVaultTransaction := func(vaultTtx *types.VaultTransaction, err error) {
		require.NoError(s.T(), err)
		receivedVaultTxCh <- vaultTtx
	}
	s.client.VaultTransactionSubscribe(context.Background(), latestTxHeigh, lastestTxPos, onVaultTransaction)
	// A set structure as the notifications can come out of order.
	receivedVaultTxs := map[int]types.VaultTransaction{}
	for i := 0; i < numNotifiedVaultTxs+1; i++ {
		select {
		case vaultTtx := <-receivedVaultTxCh:
			receivedVaultTxs[vaultTtx.Height] = *vaultTtx
		case <-time.After(time.Second):
			require.Fail(s.T(), "timeout")
		}
	}
	firstVaultTx := receivedVaultTxs[52670]
	require.Equal(s.T(), 10000, firstVaultTx.Amount)
}
