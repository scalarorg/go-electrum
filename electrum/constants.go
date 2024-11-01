package electrum

type Method string

const (
	VaultTransactionsSubscribe Method = "vault.transactions.subscribe"
	VaultTransactionsGet       Method = "vault.transactions.get"
)

func (m Method) String() string {
	return string(m)
}
