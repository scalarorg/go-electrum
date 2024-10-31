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

	"github.com/rs/zerolog/log"
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

func (s *vaultClientTestsuite) TestTransactionGetFrom() {
	expectedResponse := []byte("\xaa\xbb\xcc")
	// hash := "b43da04e4968227daed5f667f68af19988af4201b36ca552ca15e07e8c70a4fd"
	hash := "7490fc0a42ea90d9d3f712f556489a9b51081f1d3e9fad4cd38ad440dc4cd665"
	length := 10
	response, err := s.client.VaultTransactionsGetFrom(context.Background(), hash, length)
	require.NoError(s.T(), err)
	// rawTx, err := hex.DecodeString(rawTxHex)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to decode transaction hex: %w", err)
	// }
	// println(hex.EncodeToString(response))
	// for _, tx := range response {
	// 	println(tx["change_amount"].(*float64))
	// }
	require.Equal(s.T(), expectedResponse, response)
}

func (s *vaultClientTestsuite) TestVaultTransactionGetCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	go func() {
		hash := "b43da04e4968227daed5f667f68af19988af4201b36ca552ca15e07e8c70a4fd"
		length := 10
		_, err := s.client.VaultTransactionsGetFrom(ctx, hash, length)
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
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error)
	receivedVaultTxCh := make(chan *types.VaultTransaction)
	params := []interface{}{}
	go func() {
		onVaultTransaction := func(vaultTtx *types.VaultTransaction, err error) {
			require.NoError(s.T(), err)
			receivedVaultTxCh <- vaultTtx
		}
		s.client.VaultTransactionSubscribe(ctx, onVaultTransaction, params)
	}()
	cancel()
	select {
	case vaultTx := <-receivedVaultTxCh:
		log.Info().Msgf("vaultTx: %v", vaultTx)
	case err := <-errCh:
		require.ErrorIs(s.T(), err, context.Canceled)
	case <-time.After(24 * time.Hour):
		require.Fail(s.T(), "timeoout")
	}
}
