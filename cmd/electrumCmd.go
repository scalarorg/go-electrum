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

		client, err := electrum.Connect(&electrum.Options{
			Dial: func() (net.Conn, error) {
				return net.DialTimeout("tcp", rpcServer, time.Second)
			},
			MethodTimeout:   time.Second,
			PingInterval:    -1,
			SoftwareVersion: "testclient",
		})
		if err == nil {
			// delay starting the electrum client until the unix socket is ready and there is some connected client
			go startElectrumClient(ctx, client, unixSocketServer, vaultTxCh, lastVaultTx)

		}
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
func startElectrumClient(ctx context.Context, client *electrum.Client, socketServer *socket.UnixSocketServer, vaultTxCh chan<- *types.VaultTransaction, lastVaultTx string) {
	for {
		if socketServer.ConnectionCount() > 0 {
			params := []interface{}{}
			if lastVaultTx != "" {
				params = append(params, lastVaultTx)
			}
			go func() {
				onVaultTransaction := func(vaultTxInfo *types.VaultTxInfo, err error) {
					if err != nil {
						log.Error().Err(err).Msg("Failed to receive vault transaction")
						return
					}
					log.Debug().Msgf("Received vaultTx: %v", vaultTxInfo.Key)
					vaultTx, err := types.NewVaultTransactionFromInfo(vaultTxInfo)
					if err != nil {
						log.Error().Err(err).Msgf("Failed to create vault transaction from info: %v", vaultTxInfo)
						return
					}
					vaultTxCh <- vaultTx
				}
				log.Debug().Msgf("Subscribing to vault transactions with params: %v", params)
				client.VaultTransactionSubscribe(ctx, onVaultTransaction, params...)
			}()
			break
		} else {
			log.Debug().Msg("No connected clients, skipping vault transaction subscription")
			time.Sleep(time.Second)
		}
	}
}
