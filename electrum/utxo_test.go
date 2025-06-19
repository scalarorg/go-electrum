package electrum_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/scalarorg/go-electrum/electrum"
	"github.com/scalarorg/go-electrum/electrum/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type utxoTestsuite struct {
	suite.Suite

	client *electrum.Client
}

func (s *utxoTestsuite) SetupTest() {
	client, err := electrum.Connect(&electrum.Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", "127.0.0.1:60001", time.Second)
		},
		MethodTimeout:   time.Second,
		PingInterval:    -1,
		SoftwareVersion: "testclient",
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *utxoTestsuite) TearDownTest() {
	s.client.Close()
}

func TestUtxoClient(t *testing.T) {
	suite.Run(t, &utxoTestsuite{})
}
func (s *utxoTestsuite) TestUtxoList() {
	var response interface{}
	err := s.client.CallRpcBlockingMethod(context.Background(), response,
		types.BlockchainScripthashListunspent,
		"tb1p4r79pkrlzmvf9dx56zraykwq4dqhuyrtq39jjxnh9rf2uy6rmelspm6j23",
	)
	require.NoError(s.T(), err)
	require.NotNil(s.T(), response)
	println(response)
}
