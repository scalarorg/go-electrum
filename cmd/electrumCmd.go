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
		vaultTxCh := make(chan types.VaultTransaction)

		unixSocketServer, err := socket.Start(unixSocketPath, vaultTxCh)
		if err != nil {
			return err
		}
		defer unixSocketServer.Close()
		go func() {
			for tx := range vaultTxCh {
				log.Info().Msgf("Received vault transaction from socket: %v", tx)
			}
		}()

		client, err := electrum.Connect(&electrum.Options{
			Dial: func() (net.Conn, error) {
				return net.DialTimeout("tcp", rpcServer, time.Second)
			},
			MethodTimeout:   time.Second,
			PingInterval:    -1,
			SoftwareVersion: "testclient",
		})
		receivedVaultTxCh := make(chan *types.VaultTransaction)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if err == nil {
			params := []interface{}{}
			go func() {
				onVaultTransaction := func(vaultTtx *types.VaultTransaction, err error) {
					receivedVaultTxCh <- vaultTtx
				}
				client.VaultTransactionSubscribe(ctx, onVaultTransaction, params)
			}()
		}

		// Setup signal handling
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		// Wait for a signal or a vault transaction
		for {
			select {
			case <-sigCh:
				log.Info().Msg("Received shutdown signal, closing...")
				os.Remove(unixSocketPath)
				os.Exit(1)
				return nil
			case vaultTx := <-receivedVaultTxCh:
				log.Info().Msgf("vaultTx: %v", vaultTx)
			}
		}

	},
}
