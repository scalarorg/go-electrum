package cmd

import "github.com/spf13/cobra"

var (
	// Used for flags.
	rpcServer      string
	rpcServerKey   = "rpc-server"
	unixSocket     string
	unixSocketKey  = "unix-socket"
	lastVaultTx    string
	lastVaultTxKey = "last-vault-tx"
	rootCmd        = &cobra.Command{
		Use:   "electrum",
		Short: "Electrum client for Scalar's version of the Mempool electrs",
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		&rpcServer,
		rpcServerKey,
		"127.0.0.1:60001",
		"path to the configuration file",
	)
	rootCmd.PersistentFlags().StringVar(
		&unixSocket,
		unixSocketKey,
		"/tmp/electrs.sock",
		"unix socket path",
	)
	rootCmd.PersistentFlags().StringVar(
		&lastVaultTx,
		lastVaultTxKey,
		"",
		"last vault tx key (height:position:txid)",
	)
}
