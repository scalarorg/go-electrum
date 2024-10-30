package electrum

type Method string

const (
	VaultTransactionSubscribe Method = "vault.transactions.subscribe"
	VaultTransactionGet       Method = "vault.transaction.get"
)

func (m Method) String() string {
	return string(m)
}
