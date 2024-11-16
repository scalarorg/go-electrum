package cmd

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/scalarorg/go-electrum/electrum"
	"github.com/scalarorg/go-electrum/electrum/types"
	"github.com/scalarorg/go-electrum/socket"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runElectrumCmd)
}

var runElectrumCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the electrum client",
	RunE: func(cmd *cobra.Command, args []string) error {
		rpcServer, err := cmd.Flags().GetString(rpcServerKey)
		if err != nil {
			return err
		}
		unixSocketPath, err := cmd.Flags().GetString(unixSocketKey)
		if err != nil {
			return err
		}
		lastVaultTx, err := cmd.Flags().GetString(lastVaultTxKey)
		if err != nil {
			return err
		}
		vaultTxCh := make(chan *types.VaultTransaction)
		unixSocketServer, err := socket.Start(unixSocketPath, vaultTxCh)
		if err != nil {
			return err
		}
		defer unixSocketServer.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// delay starting the electrum client until the unix socket is ready and there is some connected client
		go startElectrumClient(ctx, rpcServer, unixSocketServer, vaultTxCh, lastVaultTx)

		// Setup signal handling
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Wait for a signal or a vault transaction
		for {
			<-sigCh
			log.Info().Msg("Received shutdown signal, closing...")
			os.Remove(unixSocketPath)
			os.Exit(1)
			return nil
		}

	},
}

// Waiting for first client to connect before subscribing to vault transactions
func startElectrumClient(ctx context.Context, rpcServer string, socketServer *socket.UnixSocketServer, vaultTxCh chan<- *types.VaultTransaction, lastVaultTx string) {
	params := []interface{}{10}
	if lastVaultTx != "" {
		params = append(params, lastVaultTx)
	}
	client, err := electrum.Connect(&electrum.Options{
		Dial: func() (net.Conn, error) {
			return net.DialTimeout("tcp", rpcServer, time.Second)
		},
		MethodTimeout:   time.Second,
		PingInterval:    -1,
		SoftwareVersion: "testclient",
	})
	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to electrum server at %s", rpcServer)
		return
	}
	go func() {
		onVaultTransaction := func(vaultTxs []types.VaultTransaction, err error) error {
			if err != nil {
				log.Error().Err(err).Msg("Failed to receive vault transaction")
				return err
			}
			for _, vaultTx := range vaultTxs {
				log.Debug().Msgf("Received vaultTx staker address: %v", vaultTx.StakerAddress)
				log.Debug().Msgf("Received vaultTx key: %v", vaultTx.Key)
				log.Debug().Msgf("Send vaultTx staker pubkey: %v", vaultTx.StakerPubkey)
				vaultTxCh <- &vaultTx
			}
			return nil
		}
		log.Debug().Msgf("Subscribing to vault transactions with params: %v", params)
		client.VaultTransactionSubscribe(ctx, onVaultTransaction, params...)
	}()
}
